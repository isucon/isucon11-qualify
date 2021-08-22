MAKEFILE_DIR:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: default generate build build-contestant build-bench build-standalone clean clean-output

default: generate

files-generated:  files-generated/REVISION files-generated/isucon11-qualify.tar files-generated/isucon11-portal.tar files-generated/initial-data.sql files-generated/initialize.json

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
	git clone --depth=1 https://github.com/isucon/isucon11-qualify.git -b main $(tmpdir)/isucon11-portal
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
