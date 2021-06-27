require 'json'
require 'aws-sdk-s3'

BUCKET = 'isucon10-machine-images'
PREFIX = 'final/'

do_presign = ARGV.delete('--presign')

s3 = Aws::S3::Client.new(use_dualstack_endpoint: true)

manifest_obj = s3.list_objects_v2(bucket: BUCKET, prefix: "#{PREFIX}#{ARGV[0]}/")
  .flat_map(&:contents)
  .select { |_| _.key.match?(/\.json$/) }
  .sort_by { |_| _.last_modified }
  .last
manifest = JSON.parse(s3.get_object(bucket: BUCKET, key: manifest_obj.key).body.read)

qcow2_key = manifest.fetch('qcow2_key')
qcow2_sha256 = manifest.fetch('qcow2_sha256')

url = Aws::S3::Presigner.new(client: s3).presigned_url(:get_object, bucket: BUCKET, key: qcow2_key)
if do_presign
  puts({url: url, sum: qcow2_sha256}.to_json)
else
  File.write "tmp/latest-#{ARGV[0]}.sum", "#{qcow2_sha256}\n"
  system("curl", "-fo", "tmp/latest-#{ARGV[0]}.qcow2", url, exception: true)
end
