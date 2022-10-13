SHELL:=/bin/bash
RELEASE_VERSION:=$(shell ../getversion)
RELEASE_DIR=/tmp/greenplum-instance_release
#TAG_PREFIX empty by default
GREENPLUM_OPERATOR_IMAGE_REPO:=greenplum-operator
GREENPLUM_OPERATOR_IMAGE_TAG:=$(GREENPLUM_OPERATOR_IMAGE_REPO):$(TAG_PREFIX)$(shell ../getversion --image)
GOBIN:=$(PWD)/bin
DATE:=$(shell date +%FT%H:%M:%S)
DOCKERTAG_PREFIX:=greenplum-operator
OLD_UPGRADE_VERSION:=v2.2.0
CONCOURSE_BUILD?=false

.PHONY: build
build: build-in-docker

.PHONY: docker
docker: package-in-docker

sources := $(shell find . -name "*.go" -print)

package-in-docker: export DOCKER_BUILDKIT=1
container-structure-test-in-docker: export DOCKER_BUILDKIT=1
package-in-docker container-structure-test-in-docker: $(sources) Dockerfile Dockerfile_container_structure_test Dockerfile_test.yml
	@$(MAKE) generate manifests
	@if [ -z "${GREENPLUM_VERSION}" ]; then \
		GREENPLUM_VERSION=$$(gcloud container images list-tags gcr.io/gp-kubernetes/ubuntu-gpdb-ent --filter='tags=latest' --format='value(tags.filter(NOT "latest"))') ; \
    fi; \
	echo "Using GREENPLUM_VERSION: $${GREENPLUM_VERSION}" ; \
	docker build \
		--iidfile package-in-docker \
		-t "$(GREENPLUM_OPERATOR_IMAGE_TAG)" \
		-t "$(GREENPLUM_OPERATOR_IMAGE_REPO):latest" \
		-f Dockerfile \
		--build-arg CONCOURSE_BUILD=$(CONCOURSE_BUILD) \
		--build-arg VERSION=$(RELEASE_VERSION) \
		--build-arg DATE="$(DATE)" \
		--build-arg GREENPLUM_VERSION="$${GREENPLUM_VERSION}" \
		..  && \
	docker build \
		--iidfile container-structure-test-in-docker \
		--build-arg VERSION="$(RELEASE_VERSION)" \
		--build-arg DATE="$(DATE)" \
		--build-arg GREENPLUM_VERSION="$${GREENPLUM_VERSION}" \
		--build-arg IMAGE_TAG="$(GREENPLUM_OPERATOR_IMAGE_TAG)" \
		-t $(DOCKERTAG_PREFIX):container-structure-test-in-docker \
		-f Dockerfile_container_structure_test \
		..

.PHONY: docker-check
docker-check: $(sources) package-in-docker container-structure-test-in-docker
	docker run --tty --rm -v /var/run/docker.sock:/var/run/docker.sock $(DOCKERTAG_PREFIX):container-structure-test-in-docker
	@echo "Check for up-to-date packages" # can't do this in container-structure-test b/c we need root, and we don't want to install sudo.
	docker run --tty --rm --user root $(GREENPLUM_OPERATOR_IMAGE_TAG) bash -c "apt-get update && apt-get upgrade --just-print | tee /dev/tty | grep -q '^0 upgraded, 0 newly installed, 0 to remove and 0 not upgraded'"

build-in-docker: $(sources) Dockerfile
	docker build --target build-in-docker --iidfile build-in-docker --tag $(DOCKERTAG_PREFIX):build-in-docker --file Dockerfile ..

.PHONY: clean
clean:
	$(info Removing old greenplum-operator artifacts ...)
	$(MAKE) -C cmd/greenplumOperator $@
	for tag in \
			container-structure-test-in-docker \
			package-in-docker \
			build-in-docker \
			; do \
		if [ -f $$tag ] ; then \
			docker rmi -f $$(cat $$tag) ; \
			rm -f $$tag ; \
		fi ; \
	done

.PHONY: lint
lint:
	PATH=$(GOBIN):$$PATH golangci-lint run -c ../.golangci.yml

.PHONY: check
check: unit lint

.PHONY: unit
unit:
	PATH=$(GOBIN):$$PATH ginkgo -p -r -skipPackage=integration ./...

# /*/ to match only dirs, realpath to strip the trailing /, and notdir to strip the leading dirname (integration/)
INTEGRATION_SUBDIRS=$(notdir $(realpath $(wildcard integration/*/)))

integration_targets=$(addprefix integration-,$(INTEGRATION_SUBDIRS))
gke_integration_targets=$(addprefix gke-integration-,$(INTEGRATION_SUBDIRS))

.PHONY: integration gke-integration
integration gke-integration:
	$(foreach i,$(INTEGRATION_SUBDIRS), \
		$(MAKE) $@-$(i) || exit 1; \
	)

integration-upgrade: image-check
gke-integration-upgrade: $(RELEASE_DIR)/greenplum-for-kubernetes-$(OLD_UPGRADE_VERSION)

$(integration_targets): INTEGRATION_TEST_DIR=integration/$(patsubst integration-%,%,$@)
$(gke_integration_targets): INTEGRATION_TEST_DIR=integration/$(patsubst gke-integration-%,%,$@)

.PHONY: $(integration_targets)
$(integration_targets):
	$(MAKE) integration-impl INTEGRATION_TEST_DIR=$(INTEGRATION_TEST_DIR)

.PHONY: $(gke_integration_targets)
$(gke_integration_targets):
	. /tmp/.gp-kubernetes_gke_image_tags && \
	OPERATOR_IMAGE_REPO="gcr.io/gp-kubernetes/greenplum-operator" \
	GREENPLUM_IMAGE_REPO="gcr.io/gp-kubernetes/greenplum-for-kubernetes" \
	$(MAKE) integration-impl INTEGRATION_TEST_DIR=$(INTEGRATION_TEST_DIR)

ifeq ($(CONCOURSE_BUILD),false)
LOCAL_SECRETS_FILE:=~/workspace/gp-continuous-integration/secrets/greenplum-for-kubernetes/pipeline-secrets.yml
FAKE_CONCOURSE_PARAMS=\
	ACCESS_KEY_ID=$$(yq r $(LOCAL_SECRETS_FILE) gp-kubernetes-s3.s3_access_key_id) \
	SECRET_ACCESS_KEY=$$(yq r $(LOCAL_SECRETS_FILE) gp-kubernetes-s3.s3_secret_access_key)
endif

.PHONY: integration-impl
integration-impl: psql
	PATH=$(GOBIN):$$PATH \
	$(FAKE_CONCOURSE_PARAMS) \
	ginkgo -v -failFast -r $(INTEGRATION_TEST_DIR) -- \
		--operatorImageRepository $${OPERATOR_IMAGE_REPO:-greenplum-operator} \
		--operatorImageTag $${OPERATOR_IMAGE_TAG:-latest} \
		--greenplumImageRepository $${GREENPLUM_IMAGE_REPO:-greenplum-for-kubernetes} \
		--greenplumImageTag $${GREENPLUM_IMAGE_TAG:-latest}

$(RELEASE_DIR)/greenplum-for-kubernetes-$(OLD_UPGRADE_VERSION):
	mkdir -p $(RELEASE_DIR)
	gsutil cp gs://greenplum-for-kubernetes-release/greenplum-for-kubernetes-$(OLD_UPGRADE_VERSION).tar.gz $(RELEASE_DIR)
	tar xvf $(RELEASE_DIR)/greenplum-for-kubernetes-$(OLD_UPGRADE_VERSION).tar.gz -C $(RELEASE_DIR)

cmd/greenplumReady/greenplumReady: cmd/greenplumReady/main.go ../pkg/integrationutils/kubewait/kubewait.go
	cd cmd/greenplumReady; go build

.PHONY: regsecret
regsecret:
	@if [ $$(kubectl get secret regsecret --no-headers | wc -l) == "1" ]; then \
		echo "INFO: regsecret already present. Skipping creating regsecret..."; \
	else \
		if [ "$$GCP_SVC_ACCT_KEY" != "" ] ; then \
			kubectl create secret docker-registry regsecret --docker-server=https://gcr.io --docker-username=_json_key --docker-password="$$GCP_SVC_ACCT_KEY";\
		 elif [ -f operator/key.json ]; then \
			kubectl create secret  docker-registry  regsecret --docker-server=https://gcr.io --docker-username=_json_key --docker-password="`cat operator/key.json`";\
		 else \
			echo "ERROR: could not find google service account key in env or operator dir" && exit 1; \
		fi \
	fi

.PHONY: deploy
deploy:
	$(MAKE) deploy-operator
	$(MAKE) deploy-gpdb

.PHONY: gke-deploy
gke-deploy:
	$(MAKE) gke-deploy-operator
	$(MAKE) deploy-gpdb

.PHONY: deploy-gpdb
deploy-gpdb:
	cat ../workspace/my-gp-instance.yaml | kubectl create -f -
	cmd/greenplumReady/greenplumReady

.PHONY: gke-deploy-gpdb
gke-deploy-gpdb: deploy-gpdb

.PHONY: deploy-operator
deploy-operator: regsecret cmd/greenplumReady/greenplumReady
	OPERATOR_IMAGE_REPO=$$(yq r operator/values.yaml operatorImageRepository) \
	OPERATOR_IMAGE_TAG=$$(yq r operator/values.yaml operatorImageTag) \
	GREENPLUM_IMAGE_REPO=$$(yq r operator/values.yaml greenplumImageRepository) \
	GREENPLUM_IMAGE_TAG=$$(yq r operator/values.yaml greenplumImageTag) \
	$(MAKE) deploy-operator-impl

.PHONY: gke-deploy-operator
gke-deploy-operator: regsecret cmd/greenplumReady/greenplumReady
	source /tmp/.gp-kubernetes_gke_image_tags && \
	OPERATOR_IMAGE_REPO="gcr.io/gp-kubernetes/greenplum-operator" \
	GREENPLUM_IMAGE_REPO="gcr.io/gp-kubernetes/greenplum-for-kubernetes" \
	$(MAKE) deploy-operator-impl

.PHONY: deploy-operator-impl
deploy-operator-impl:
	helm install greenplum-operator ./operator \
		--set operatorImageRepository=$$OPERATOR_IMAGE_REPO \
		--set operatorImageTag=$$OPERATOR_IMAGE_TAG \
		--set greenplumImageRepository=$$GREENPLUM_IMAGE_REPO \
		--set greenplumImageTag=$$GREENPLUM_IMAGE_TAG
	for i in {1..3}; \
		do kubectl wait --for condition=available deployment.apps/greenplum-operator && break || sleep 1; \
	done

.PHONY: deploy-clean
deploy-clean:
	kubectl delete -f ../workspace/my-gp-instance.yaml || true
	kubectl delete crd greenplumclusters.greenplum.pivotal.io || true
	kubectl delete crd greenplumpxfservices.greenplum.pivotal.io || true
	kubectl delete --wait all  -l app=greenplum > /dev/null 2>&1 || true
	kubectl delete pvc --all || true
	kubectl delete --wait configmap/greenplum-config secrets/ssh-secrets > /dev/null 2>&1 || true
	helm uninstall greenplum-operator || true
	kubectl	wait --for=delete deployment.apps/greenplum-operator || true
	kubectl delete service/greenplum-validating-webhook || true
	kubectl delete validatingwebhookconfigurations greenplum-validating-webhook-config || true
	kubectl delete csr greenplum-validating-webhook || true
	kubectl delete jobs --all || true
	kubectl delete --wait secrets/regsecret > /dev/null 2>&1 || true
	while [ "$$(kubectl get all 2>/dev/null | wc -l)" -gt 2 ]; do sleep 1; printf "."; done

.PHONY: gke-deploy-clean
gke-deploy-clean: deploy-clean

.PHONY: list
list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs

.PHONY: psql
psql:
	@if ! psql --version ; then \
		echo "please make sure psql is in $$PATH"; exit 1; \
	fi

.PHONY: image-check
image-check: $(RELEASE_DIR)/greenplum-for-kubernetes-$(OLD_UPGRADE_VERSION)
	@numOldOperatorImages=$$(docker images --filter="reference=greenplum-operator:$(OLD_UPGRADE_VERSION)" -q | wc -l) ; \
	if [ $$numOldOperatorImages -eq 0 ] ; then \
        echo "$(OLD_UPGRADE_VERSION) operator image is missing. Loading from release tarball" ; \
        docker load -i $(RELEASE_DIR)/greenplum-for-kubernetes-$(OLD_UPGRADE_VERSION)/images/greenplum-for-kubernetes ; \
        docker load -i $(RELEASE_DIR)/greenplum-for-kubernetes-$(OLD_UPGRADE_VERSION)/images/greenplum-operator ; \
     fi

### Kubebuilder Scaffolding:

# TODO: Add dependencies on `manifests` and `generate`. Perhaps restore some of the scaffolding that we deleted:
#   - `test`, `run`, `install`, `deploy` all depended on `manifests`
#   - `test`, `manager`, `run` all depended on `generate`

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
# Produce apiextensions v1 CRDs
CRD_OPTIONS ?= "crd:crdVersions=v1"

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	kubectl kustomize config/crd > operator/templates/greenplum-operator-crds.yaml

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./api/...

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go get -modfile ../tools/go.mod sigs.k8s.io/controller-tools/cmd/controller-gen@v0.10.0
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
