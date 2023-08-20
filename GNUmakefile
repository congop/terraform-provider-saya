# Copyright (C) 2023 Patrice Congo <@congop>
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#      http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

PJT_MKFILE_ABSPATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PJT_MKFILE_ABSDIR := $(strip $(patsubst %/,%,$(dir $(PJT_MKFILE_ABSPATH))) )

default: testacc

# Run acceptance tests
.PHONY: testacc, check-license-header
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m


check-license-header:
	@.build-tools/check-license.sh


act-runner-docker-build:
	cd $(PJT_MKFILE_ABSDIR)/.act \
	&& IMG_RUNNER='$(shell docker image ls act_local/runner-ubuntu-22.4 --format "{{.ID}}: {{.Repository}}")' \
	&& 	if [[ -n $$IMG_RUNNER ]]; then \
				echo "Image [ $$IMG_RUNNER ] found --> skipping docker build"; \
			else \
				docker build -f ./Dockerfile.act-runner -t act_local/runner-ubuntu-22.4 .; \
			fi

act-runner-docker-rm-vms:
	docker container stop $$(docker container ls -a --filter "name=act-Tests-Terraform-Provider-Acceptance-Tests*" --filter status=running --format "{{.ID}}") || true
	docker container rm $$(docker container ls -a --filter "name=act-Tests-Terraform-Provider-Acceptance-Tests*" --filter status=exited --format "{{.ID}}") || true

act-runner-docker-clean-img-container: act-runner-docker-rm-vms
	docker image rm -f act_local/runner-ubuntu-22.4:latest

act-install-binary:
	.act/install-act-bin.sh 0.2.48

# .tmp/bin/act --list  ==> t o list jobs
# access denied on /dev/kvm
# - add user to kvm group
# - or use docker flag --group-add
# To use an alternative (local) saya release:
#     SAYA_RELEASE_URL=http://172.17.0.1:9099/saya-teaser-20230808T183627.zip make act-run-github-actions-job-build
act-run-github-actions-job-build:
	cd $(PJT_MKFILE_ABSDIR)
	clear ; mkdir -p .tmp/artifacts ;
	.tmp/bin/act -v --env LC_ALL=C.UTF-8 \
		--env LANG=C.UTF-8 \
		--env LC_TIME=C.UTF-8 \
		--env SAYA_RELEASE_URL="$(SAYA_RELEASE_URL)" \
		--env SAYA_APT_PROXY=http://172.17.0.1:3142 \
		--platform ubuntu-22.04=act_local/runner-ubuntu-22.4 \
		--container-options "--privileged --device /dev/kvm --group-add $(shell grep kvm /etc/group | awk -F ':' '{print $$3}')  --group-add $(shell grep docker /etc/group | awk -F ':' '{print $$3}')" \
	    --pull=false \
		--verbose \
		--artifact-server-path $(PJT_MKFILE_ABSDIR)/.tmp/artifacts \
		--rebuild --rm \
	;