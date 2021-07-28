.PHONY: all
all: init apply

.PHONY: init init-common init-zone-a init-zone-c init-zone-d
init: init-common init-zone-a init-zone-d init-zone-c

init-common: common/aws.tf common/backend.tf common/versions.tf
	terraform -chdir=common init
init-zone-a: zones/a/aws.tf zones/a/backend.tf zones/a/versions.tf
	terraform -chdir=zones/a init
init-zone-c: zones/c/aws.tf zones/c/backend.tf zones/c/versions.tf
	terraform -chdir=zones/c init
init-zone-d: zones/d/aws.tf zones/d/backend.tf zones/d/versions.tf
	terraform -chdir=zones/d init

.PHONY: apply apply-common apply-zone-a apply-zone-c apply-zone-d
apply: apply-common apply-zone-a apply-zone-c apply-zone-d

apply-common: common/aws.tf common/backend.tf common/versions.tf
	terraform -chdir=common apply
apply-zone-a: zones/a/teams.json zones/a/aws.tf zones/a/backend.tf zones/a/versions.tf
	@$(call _validate_envvars)
	terraform -chdir=zones/a apply \
		-var="ami_id=$(AMI_ID)" \
		-var="isuxportal_supervisor_endpoint_url=$(ISUXPORTAL_SUPERVISOR_ENDPOINT_URL)" \
		-var="isuxportal_supervisor_token=$(ISUXPORTAL_SUPERVISOR_TOKEN)"
apply-zone-c: zones/c/teams.json zones/c/aws.tf zones/c/backend.tf zones/c/versions.tf
	@$(call _validate_envvars)
	terraform -chdir=zones/c apply \
		-var="ami_id=$(AMI_ID)" \
		-var="isuxportal_supervisor_endpoint_url=$(ISUXPORTAL_SUPERVISOR_ENDPOINT_URL)" \
		-var="isuxportal_supervisor_token=$(ISUXPORTAL_SUPERVISOR_TOKEN)"
apply-zone-d: zones/d/teams.json zones/d/aws.tf zones/d/backend.tf zones/d/versions.tf
	@$(call _validate_envvars)
	terraform -chdir=zones/d apply \
		-var="ami_id=$(AMI_ID)" \
		-var="isuxportal_supervisor_endpoint_url=$(ISUXPORTAL_SUPERVISOR_ENDPOINT_URL)" \
		-var="isuxportal_supervisor_token=$(ISUXPORTAL_SUPERVISOR_TOKEN)"

zones/a/teams.json:
	cat teams.json | jq -r ".zone_a" > zones/a/teams.json
zones/c/teams.json:
	cat teams.json | jq -r ".zone_c" > zones/c/teams.json
zones/d/teams.json:
	cat teams.json | jq -r ".zone_b" > zones/d/teams.json

common/aws.tf:
	cp base/aws.tf common/aws.tf
zones/a/aws.tf:
	cp base/aws.tf zones/a/aws.tf
zones/c/aws.tf:
	cp base/aws.tf zones/c/aws.tf
zones/d/aws.tf:
	cp base/aws.tf zones/d/aws.tf

common/backend.tf:
	cat base/backend.tf | sed -e 's/__KIND__/common/' > common/backend.tf
zones/a/backend.tf:
	cat base/backend.tf | sed -e 's/__KIND__/zone-a/' > zones/a/backend.tf
zones/c/backend.tf:
	cat base/backend.tf | sed -e 's/__KIND__/zone-c/' > zones/c/backend.tf
zones/d/backend.tf:
	cat base/backend.tf | sed -e 's/__KIND__/zone-d/' > zones/d/backend.tf

common/versions.tf:
	cp base/versions.tf common/versions.tf
zones/a/versions.tf:
	cp base/versions.tf zones/a/versions.tf
zones/c/versions.tf:
	cp base/versions.tf zones/c/versions.tf
zones/d/versions.tf:
	cp base/versions.tf zones/d/versions.tf

.PHONY: destroy destroy-common destroy-zone-a destroy-zone-d destroy-zone-c
destroy: destroy-zone-a destroy-zone-c destroy-zone-d destroy-common

destroy-common:
	terraform -chdir=common destroy
destroy-zone-a:
	terraform -chdir=zones/a destroy \
		-var="ami_id=" \
		-var="isuxportal_supervisor_endpoint_url=" \
		-var="isuxportal_supervisor_token="
destroy-zone-c:
	terraform -chdir=zones/c destroy \
		-var="ami_id=" \
		-var="isuxportal_supervisor_endpoint_url=" \
		-var="isuxportal_supervisor_token="
destroy-zone-d:
	terraform -chdir=zones/d destroy \
		-var="ami_id=" \
		-var="isuxportal_supervisor_endpoint_url=" \
		-var="isuxportal_supervisor_token="

.PHONY: cleanup
cleanup:
	rm -rf zones/{a,c,d}/teams.json
	rm -rf zones/{a,c,d}/aws.tf
	rm -rf zones/{a,c,d}/backend.tf
	rm -rf zones/{a,c,d}/versions.tf
	rm -rf common/aws.tf
	rm -rf common/backend.tf
	rm -rf common/versions.tf
