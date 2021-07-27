.PHONY: all
all: init apply

.PHONY: init init-common init-zone-a init-zone-b init-zone-c
init: init-common init-zone-a init-zone-b init-zone-c

init-common: common/aws.tf common/backend.tf common/versions.tf
	terraform -chdir=common init
init-zone-a: zones/a/aws.tf zones/a/backend.tf zones/a/versions.tf
	terraform -chdir=zones/a init
init-zone-b: zones/b/aws.tf zones/b/backend.tf zones/b/versions.tf
	terraform -chdir=zones/b init
init-zone-c: zones/c/aws.tf zones/c/backend.tf zones/c/versions.tf
	terraform -chdir=zones/c init

.PHONY: apply apply-common apply-zone-a apply-zone-b apply-zone-c
apply: apply-common apply-zone-a apply-zone-b apply-zone-c

apply-common: common/aws.tf common/backend.tf common/versions.tf
	terraform -chdir=common apply
apply-zone-a: zones/a/teams.json zones/a/aws.tf zones/a/backend.tf zones/a/versions.tf
	@$(call _validate_envvars)
	terraform -chdir=zones/a apply \
		-var="ami_id=$(AMI_ID)" \
		-var="isuxportal_supervisor_endpoint_url=$(ISUXPORTAL_SUPERVISOR_ENDPOINT_URL)" \
		-var="isuxportal_supervisor_token=$(ISUXPORTAL_SUPERVISOR_TOKEN)"
apply-zone-b: zones/b/teams.json zones/b/aws.tf zones/b/backend.tf zones/b/versions.tf
	@$(call _validate_envvars)
	terraform -chdir=zones/b apply \
		-var="ami_id=$(AMI_ID)" \
		-var="isuxportal_supervisor_endpoint_url=$(ISUXPORTAL_SUPERVISOR_ENDPOINT_URL)" \
		-var="isuxportal_supervisor_token=$(ISUXPORTAL_SUPERVISOR_TOKEN)"
apply-zone-c: zones/c/teams.json zones/c/aws.tf zones/c/backend.tf zones/c/versions.tf
	@$(call _validate_envvars)
	terraform -chdir=zones/c apply \
		-var="ami_id=$(AMI_ID)" \
		-var="isuxportal_supervisor_endpoint_url=$(ISUXPORTAL_SUPERVISOR_ENDPOINT_URL)" \
		-var="isuxportal_supervisor_token=$(ISUXPORTAL_SUPERVISOR_TOKEN)"

zones/a/teams.json:
	cat teams.json | jq -r ".zone_a" > zones/a/teams.json
zones/b/teams.json:
	cat teams.json | jq -r ".zone_b" > zones/b/teams.json
zones/c/teams.json:
	cat teams.json | jq -r ".zone_c" > zones/c/teams.json

common/aws.tf:
	cp base/aws.tf common/aws.tf
zones/a/aws.tf:
	cp base/aws.tf zones/a/aws.tf
zones/b/aws.tf:
	cp base/aws.tf zones/b/aws.tf
zones/c/aws.tf:
	cp base/aws.tf zones/c/aws.tf

common/backend.tf:
	cat base/backend.tf | sed -e 's/__KIND__/common/' > common/backend.tf
zones/a/backend.tf:
	cat base/backend.tf | sed -e 's/__KIND__/zone-a/' > zones/a/backend.tf
zones/b/backend.tf:
	cat base/backend.tf | sed -e 's/__KIND__/zone-b/' > zones/b/backend.tf
zones/c/backend.tf:
	cat base/backend.tf | sed -e 's/__KIND__/zone-c/' > zones/c/backend.tf

common/versions.tf:
	cp base/versions.tf common/versions.tf
zones/a/versions.tf:
	cp base/versions.tf zones/a/versions.tf
zones/b/versions.tf:
	cp base/versions.tf zones/b/versions.tf
zones/c/versions.tf:
	cp base/versions.tf zones/c/versions.tf

.PHONY: destroy destroy-common destroy-zone-a destroy-zone-b destroy-zone-c
destroy: destroy-common destroy-zone-a destroy-zone-b destroy-zone-c

destroy-common:
	terraform -chdir=common destroy
destroy-zone-a:
	terraform -chdir=zones/a destroy -var="ami_id=$(AMI_ID)"
destroy-zone-b:
	terraform -chdir=zones/b destroy -var="ami_id=$(AMI_ID)"
destroy-zone-c:
	terraform -chdir=zones/c destroy -var="ami_id=$(AMI_ID)"

.PHONY: cleanup
cleanup:
	rm -rf zones/{a,b,c}/teams.json
	rm -rf zones/{a,b,c}/aws.tf
	rm -rf zones/{a,b,c}/backend.tf
	rm -rf zones/{a,b,c}/versions.tf
	rm -rf common/aws.tf
	rm -rf common/backend.tf
	rm -rf common/versions.tf
