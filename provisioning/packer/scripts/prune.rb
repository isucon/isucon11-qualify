require 'json'
require 'aws-sdk-s3'
require 'aws-sdk-ec2'

BUCKET = 'isucon10-machine-images'
PREFIX = 'final/'
KEEP_NUM = 15

s3 = Aws::S3::Client.new(use_dualstack_endpoint: true)
ec2 = Aws::EC2::Client.new()

manifest = JSON.parse(ARGF.read)
family = manifest.fetch('family')

prefix = "#{PREFIX}#{family}/"

manifest_objs = s3.list_objects_v2(bucket: BUCKET, prefix: prefix)
  .flat_map(&:contents)
  .select { |_| _.key.match?(/\.json$/) }
  .sort_by { |_| _.last_modified }

manifest_objs[0...-KEEP_NUM].each do |manifest_obj|
  puts "===> Pruning s3://#{BUCKET}/#{manifest_obj.key}"
  target_manifest = JSON.parse(s3.get_object(bucket: BUCKET, key: manifest_obj.key).body.read)

  # Delete qcow2
  qcow2_key = target_manifest['qcow2_key']
  if qcow2_key
    puts "   * Delete qcow2: #{qcow2_key}"
    begin
      s3.delete_object(bucket: BUCKET, key: qcow2_key)
    rescue Aws::S3::Errors::NotFound => e
      puts "   > it was already deleted. (#{e.inspect})"
    end
  end

  # Deregister AMI
  ami_id = target_manifest['ami_id']
  if ami_id
    puts "   * Deregister AMI: #{ami_id}"
    image = ec2.describe_images(image_ids: [ami_id]).images[0]
    if image
      snapshot_ids = image.block_device_mappings.map do |mapping|
        mapping.ebs && mapping.ebs.snapshot_id
      end.compact

      ec2.deregister_image(image_id: ami_id)

      snapshot_ids.each do |snapshot_id|
        puts "   * Delete snapshot #{snapshot_id} "
        ec2.delete_snapshot(snapshot_id: snapshot_id)
      end
    else
      puts "   > it was already gone."
    end
  end

  puts "   * Delete manifest"
  s3.delete_object(bucket: BUCKET, key: manifest_obj.key)
end
