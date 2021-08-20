.PHONY: all
all: init apply

.PHONY: init init-common init-zone-a init-zone-c init-zone-d init-ecs
init: init-common init-zone-a init-zone-c init-zone-d init-ecs

init-common: common/aws.tf common/backend.tf common/versions.tf
	terraform -chdir=common init
init-zone-a: ec2/zone-a/aws.tf ec2/zone-a/backend.tf ec2/zone-a/versions.tf ec2/zone-a/teams.json
	terraform -chdir=ec2/zone-a init
init-zone-c: ec2/zone-c/aws.tf ec2/zone-c/backend.tf ec2/zone-c/versions.tf ec2/zone-c/teams.json
	terraform -chdir=ec2/zone-c init
init-zone-d: ec2/zone-d/aws.tf ec2/zone-d/backend.tf ec2/zone-d/versions.tf ec2/zone-d/teams.json
	 terraform -chdir=ec2/zone-d init
init-ecs: ecs/aws.tf ecs/backend.tf ecs/versions.tf
	@$(call _validate_envvars)
	terraform -chdir=ecs init \

.PHONY: apply apply-common apply-zone-a apply-zone-c apply-zone-d apply-ecs
apply: apply-common apply-zone-a apply-zone-c apply-zone-d apply-ecs

apply-common: common/aws.tf common/backend.tf common/versions.tf
	terraform -chdir=common apply
apply-zone-a: ec2/zone-a/teams.json ec2/zone-a/aws.tf ec2/zone-a/backend.tf ec2/zone-a/versions.tf
	@$(call _validate_envvars)
	terraform -chdir=ec2/zone-a apply \
		-var="git_tag=$(GIT_TAG)" \
		-var="isuxportal_supervisor_endpoint_url=$(ISUXPORTAL_SUPERVISOR_ENDPOINT_URL)" \
		-var="isuxportal_supervisor_token=$(ISUXPORTAL_SUPERVISOR_TOKEN)"
apply-zone-c: ec2/zone-c/teams.json ec2/zone-c/aws.tf ec2/zone-c/backend.tf ec2/zone-c/versions.tf
	@$(call _validate_envvars)
	terraform -chdir=ec2/zone-c apply \
		-var="git_tag=$(GIT_TAG)" \
		-var="isuxportal_supervisor_endpoint_url=$(ISUXPORTAL_SUPERVISOR_ENDPOINT_URL)" \
		-var="isuxportal_supervisor_token=$(ISUXPORTAL_SUPERVISOR_TOKEN)"
apply-zone-d: ec2/zone-d/teams.json ec2/zone-d/aws.tf ec2/zone-d/backend.tf ec2/zone-d/versions.tf
	@$(call _validate_envvars)
	terraform -chdir=ec2/zone-d apply \
		-var="git_tag=$(GIT_TAG)" \
		-var="isuxportal_supervisor_endpoint_url=$(ISUXPORTAL_SUPERVISOR_ENDPOINT_URL)" \
		-var="isuxportal_supervisor_token=$(ISUXPORTAL_SUPERVISOR_TOKEN)"
apply-ecs: ecs/aws.tf ecs/backend.tf ecs/versions.tf
	@$(call _validate_envvars)
	terraform -chdir=ecs apply

ec2/zone-a/teams.json:
	cat teams.json | jq -r ".zone_a" > ec2/zone-a/teams.json
ec2/zone-c/teams.json:
	cat teams.json | jq -r ".zone_c" > ec2/zone-c/teams.json
ec2/zone-d/teams.json:
	cat teams.json | jq -r ".zone_d" > ec2/zone-d/teams.json

common/aws.tf:
	cp .base/aws.tf common/aws.tf
ec2/zone-a/aws.tf:
	cp .base/aws.tf ec2/zone-a/aws.tf
ec2/zone-c/aws.tf:
	cp .base/aws.tf ec2/zone-c/aws.tf
ec2/zone-d/aws.tf:
	cp .base/aws.tf ec2/zone-d/aws.tf
ecs/aws.tf:
	cp .base/aws.tf ecs/aws.tf

common/backend.tf:
	cat .base/backend.tf | sed -e 's/__KIND__/common/' > common/backend.tf
ec2/zone-a/backend.tf:
	cat .base/backend.tf | sed -e 's/__KIND__/zone-a/' > ec2/zone-a/backend.tf
ec2/zone-c/backend.tf:
	cat .base/backend.tf | sed -e 's/__KIND__/zone-c/' > ec2/zone-c/backend.tf
ec2/zone-d/backend.tf:
	cat .base/backend.tf | sed -e 's/__KIND__/zone-d/' > ec2/zone-d/backend.tf
ecs/backend.tf:
	cat .base/backend.tf | sed -e 's/__KIND__/ecs/' > ecs/backend.tf

common/versions.tf:
	cp .base/versions.tf common/versions.tf
ec2/zone-a/versions.tf:
	cp .base/versions.tf ec2/zone-a/versions.tf
ec2/zone-c/versions.tf:
	cp .base/versions.tf ec2/zone-c/versions.tf
ec2/zone-d/versions.tf:
	cp .base/versions.tf ec2/zone-d/versions.tf
ecs/versions.tf:
	cp .base/versions.tf ecs/versions.tf

.PHONY: destroy destroy-common destroy-zone-a destroy-zone-d destroy-zone-c destroy-ecs
destroy: destroy-ecs destroy-zone-a destroy-zone-c destroy-zone-d destroy-common

destroy-common:
	terraform -chdir=common destroy
destroy-zone-a:
	terraform -chdir=ec2/zone-a destroy \
		-var="git_tag=$(GIT_TAG)" \
		-var="isuxportal_supervisor_endpoint_url=" \
		-var="isuxportal_supervisor_token="
destroy-zone-c:
	terraform -chdir=ec2/zone-c destroy \
		-var="git_tag=$(GIT_TAG)" \
		-var="isuxportal_supervisor_endpoint_url=" \
		-var="isuxportal_supervisor_token="
destroy-zone-d:
	terraform -chdir=ec2/zone-d destroy \
		-var="git_tag=$(GIT_TAG)" \
		-var="isuxportal_supervisor_endpoint_url=" \
		-var="isuxportal_supervisor_token="
destroy-ecs:
	terraform -chdir=ecs destroy

.PHONY: cleanup
cleanup:
	rm -rf ecs/aws.tf
	rm -rf ecs/backend.tf
	rm -rf ecs/versions.tf
	rm -rf ec2/zone-{a,c,d}/teams.json
	rm -rf ec2/zone-{a,c,d}/aws.tf
	rm -rf ec2/zone-{a,c,d}/backend.tf
	rm -rf ec2/zone-{a,c,d}/versions.tf
	rm -rf common/aws.tf
	rm -rf common/backend.tf
	rm -rf common/versions.tf
