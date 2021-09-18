MAKEFILE_DIR:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: default generate build build-contestant build-bench build-standalone clean clean-output

default: generate

files-generated:  files-generated/REVISION files-generated/isucon11-qualify.tar files-generated/isucon11-portal.tar files-generated/initial-data.sql files-generated/initialize.json files-generated/images.tgz

files-generated/REVISION:
	mkdir -p files-generated/
	if [ -n "$(GITHUB_SHA)" ]; then echo "$(GITHUB_SHA)" > files-generated/REVISION; fi
	if [ ! -e files-generated/REVISION ]; then git rev-parse HEAD > files-generated/REVISION; fi
	test -e files-generated/REVISION

files-generated/isucon11-qualify.tar: files-generated/REVISION
	mkdir -p files-generated/
	git -C "$(shell git rev-parse --show-toplevel)" archive "$(shell cat files-generated/REVISION)" > files-generated/isucon11-qualify.tar

files-generated/isucon11-portal.tar:
	mkdir -p files-generated/
	$(eval tmpdir=$(shell mktemp -d))
	git -c core.sshCommand="ssh -i $(MAKEFILE_DIR)isucon11-portal.key -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -F /dev/null" clone --depth=1 git@github.com:isucon/isucon11-portal.git -b main $(tmpdir)/isucon11-portal
	cd $(tmpdir)/isucon11-portal && \
		tar cvf $(MAKEFILE_DIR)files-generated/isucon11-portal.tar . && \
		cd -
	rm -rf $(tmpdir)

files-generated/initial-data.sql:
	mkdir -p files-generated/
	aws s3 cp s3://isucon11-qualify-dev/initial-data.sql files-generated/initial-data.sql


files-generated/initialize.json:
	mkdir -p files-generated/
	aws s3 cp s3://isucon11-qualify-dev/initialize.json files-generated/initialize.json

files-generated/images.tgz:
	mkdir -p files-generated/
	aws s3 cp s3://isucon11-qualify-dev/images.tgz files-generated/images.tgz


generate: output target/contestant.json target/bench.json target/standalone.json

output:
	mkdir -p output/

target/contestant.json: files-generated output base.libsonnet contestant.jsonnet
	mkdir -p target/
	jsonnet --jpath . contestant.jsonnet > target/contestant.json
	packer validate target/contestant.json

target/bench.json: files-generated output base.libsonnet bench.jsonnet
	mkdir -p target/
	jsonnet --jpath . bench.jsonnet > target/bench.json
	packer validate target/bench.json


target/standalone.json: files-generated base.libsonnet standalone.jsonnet
	mkdir -p target/
	jsonnet --jpath . standalone.jsonnet > target/standalone.json
	packer validate target/standalone.json


build: build-contestant build-bench

build-contestant: files-generated target/contestant.json
	packer build -var "revision=$(shell cat files-generated/REVISION)" target/contestant.json
build-bench: files-generated target/bench.json
	packer build -var "revision=$(shell cat files-generated/REVISION)" target/bench.json
build-standalone: files-generated target/standalone.json
	packer build -var "revision=$(shell cat files-generated/REVISION)" target/standalone.json


clean:
	rm -rf files-generated/
	rm -rf target/
	rm -rf tmp/

clean-output:
	rm -rf output/


# for kakomon

.PHONY: files-generated_kakomon files-generated_kakomon/initial-data.sql files-generated_kakomon/initialize.json files-generated_kakomon/images.tgz build-contestant_kakomon build-bench_kakomon

files-generated_kakomon:  files-generated/REVISION files-generated/isucon11-qualify.tar files-generated_kakomon/initial-data.sql files-generated_kakomon/initialize.json files-generated_kakomon/images.tgz

files-generated_kakomon/initial-data.sql:
	$(error "files-generated/initial-data.sql does not exist")
files-generated_kakomon/initialize.json:
	$(error "files-generated/initialize.json does not exist")
files-generated_kakomon/images.tgz:
	$(error "files-generated/images.tgz does not exist")

target/contestant_kakomon.json: files-generated output base.libsonnet contestant_kakomon.jsonnet
	mkdir -p target/
	jsonnet --jpath . contestant_kakomon.jsonnet > target/contestant_kakomon.json
	packer validate target/contestant_kakomon.json
target/bench_kakomon.json: files-generated output base.libsonnet bench_kakomon.jsonnet
	mkdir -p target/
	jsonnet --jpath . bench_kakomon.jsonnet > target/bench_kakomon.json
	packer validate target/bench_kakomon.json

build-contestant_kakomon: files-generated_kakomon target/contestant_kakomon.json
	packer build -var "revision=$(shell cat files-generated/REVISION)" target/contestant_kakomon.json
build-bench_kakomon: files-generated_kakomon target/bench_kakomon.json
	packer build -var "revision=$(shell cat files-generated/REVISION)" target/bench_kakomon.json

