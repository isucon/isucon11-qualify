PACKER_QEMU_MEM ?= 4096M
PACKER_QEMU_SMP ?= 4
export PACKER_QEMU_MEM
export PACKER_QEMU_SMP

default: generate

#files-generated:  files-generated/REVISION files-generated/archive.tar files-generated/isuxportal-supervisor files-generated/bench-vendor.tar files/tls-cert.pem files/tls-key.pem
files-generated:  files-generated/REVISION files-generated/archive.tar
	mkdir -p files-cached/

files/tls-cert.pem:
	cp ../secrets/cert.pem files/tls-cert.pem

files/tls-key.pem:
	cp ../secrets/key.pem files/tls-key.pem

files-generated/REVISION:
	mkdir -p files-generated/
	if [ -n "$(GITHUB_SHA)" ]; then echo "$(GITHUB_SHA)" > files-generated/REVISION; fi
	if [ ! -e files-generated/REVISION ]; then git rev-parse HEAD > files-generated/REVISION; fi
	test -e files-generated/REVISION

files-generated/archive.tar: files-generated/REVISION
	mkdir -p files-generated/
	git -C "$(shell git rev-parse --show-toplevel)" archive "$(shell cat files-generated/REVISION)" > files-generated/archive.tar
#
#files-generated/bench-vendor.tar: files-generated/REVISION
#	( cd ../bench && make vendor )
#	tar cf files-generated/bench-vendor.tar -C ../bench vendor
#
#files-generated/isuxportal-supervisor: 
#	mkdir -p files-generated/
#	touch files-generated/isuxportal-supervisor
#	#aws s3 cp s3://isucon10-machine-images/supervisor/isuxportal-supervisor-ubuntu2004-amd64 files-generated/isuxportal-supervisor

generate: output target/contestant.json target/bench.json target/full.json

output:
	mkdir -p output/

target/contestant.json: files-generated *.libsonnet contestant.jsonnet
	mkdir -p target/
	jsonnet --jpath . contestant.jsonnet > target/contestant.json
	packer validate target/contestant.json

target/bench.json: files-generated *.libsonnet bench.jsonnet
	mkdir -p target/
	jsonnet --jpath . bench.jsonnet > target/bench.json
	packer validate target/bench.json

target/full.json: files-generated *.libsonnet full.jsonnet
	mkdir -p target/
	jsonnet --jpath . full.jsonnet > target/full.json
	packer validate target/full.json


build: build-contestant build-bench

build-contestant: files-generated target/contestant.json
	packer build -var "revision=$(shell cat files-generated/REVISION)" target/contestant.json
build-bench: files-generated target/bench.json
	packer build -var "revision=$(shell cat files-generated/REVISION)" target/bench.json
build-full: files-generated target/full.json
	packer build -var "revision=$(shell cat files-generated/REVISION)" target/full.json

build-full-ec2: files-generated target/full.json
	packer build -var "revision=$(shell cat files-generated/REVISION)" -only amazon-ebs target/full.json

build-full-qemu: files-generated target/full.json
	packer build -var "revision=$(shell cat files-generated/REVISION)" -only qemu target/full.json


clean:
	rm -rf files-generated/
	rm -rf target/
	rm -rf tmp/

clean-cache:
	rm -rf files-cached/

clean-output:
	rm -rf output/

upload:
	if [ -e output/manifest-amd64-contestant.json ]; then ruby scripts/upload.rb output/manifest-amd64-contestant.json; fi
	if [ -e output/manifest-amd64-bench.json ]; then ruby scripts/upload.rb output/manifest-amd64-bench.json; fi

upload-full:
	if [ -e output/manifest-amd64-full.json ]; then ruby scripts/upload.rb output/manifest-amd64-full.json; fi


prune:
	if [ -e output/manifest-amd64-contestant.json ]; then ruby scripts/prune.rb output/manifest-amd64-contestant.json; fi
	if [ -e output/manifest-amd64-bench.json ]; then ruby scripts/prune.rb output/manifest-amd64-bench.json; fi
	if [ -e output/manifest-amd64-full.json ]; then ruby scripts/prune.rb output/manifest-amd64-full.json; fi
	if [ -e output/manifest-amd64-ci.json ]; then ruby scripts/prune.rb output/manifest-amd64-ci.json; fi

cache:
	mkdir -p files-cached/
	if [ -e output/cache-amd64-contestant-qemu/files-cached/local.tar.gz ]; then mv output/cache-amd64-contestant-qemu/files-cached/local.tar.gz files-cached/local.tar.gz; fi
