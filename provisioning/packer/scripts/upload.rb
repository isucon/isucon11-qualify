require 'json'
require 'aws-sdk-s3'

BUCKET = 'isucon10-machine-images'
PREFIX = 'final/'

manifest = JSON.parse(File.read(ARGV[0]))
builds = manifest.fetch('builds')

ami = builds.find { |_| _.fetch('name') == 'amazon-ebs' }
qemu =  builds.find { |_| _.fetch('name') == 'qemu' }

name = manifest['name'] = (qemu || ami).fetch('custom_data').fetch('name')
family = manifest['family'] = (qemu || ami).fetch('custom_data').fetch('family')

manifest['ami_id'] = ami.fetch('artifact_id').split(?:,2)[1] if ami

manifest_key = manifest['manifest_key'] = "#{PREFIX}#{family}/#{name}.json"
qcow2_path = "output/#{name}/#{name}"
qcow2_key = manifest['qcow2_key'] = "#{PREFIX}#{family}/#{name}.qcow2" if qemu

s3 = Aws::S3::Client.new(use_dualstack_endpoint: true)

puts "==> loading manifest "
puts "    #{manifest.to_json}"

if qemu
  puts "==> uploading qcow2"
  puts "  * Source:      #{qcow2_path}"
  puts "  * Destination: s3://#{BUCKET}/#{qcow2_key}"
  checksum = manifest['qcow2_sha256'] = Digest::SHA256.file(qcow2_path)
  puts "  * Checksum: #{checksum}"
  begin
    File.open(qcow2_path, 'rb') do |io|
      s3.put_object(bucket: BUCKET, key: qcow2_key, body: io)
    end
  rescue Aws::S3::Errors::EntityTooLarge => e
    puts "  > using aws-cli (#{e.inspect})"
    system("aws", "s3", "cp", "--quiet", qcow2_path, "s3://#{BUCKET}/#{qcow2_key}", exception: true)
  end
end

puts "==> uploading manifest"
puts "  * Destination: s3://#{BUCKET}/#{manifest_key}"
s3.put_object(bucket: BUCKET, key: manifest_key, body: JSON.pretty_generate(manifest))
File.write ARGV[0], JSON.pretty_generate(manifest)
