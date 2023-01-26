
SHELL:=/bin/bash
VERSION=$(shell ./getversion)
export GOBIN=${PWD}/bin
CONCOURSE_BUILD?=false
DEST=.
GOFLAGS:=

SUBDIRS=greenplum-instance greenplum-operator

# default target
.PHONY: build
build: $(SUBDIRS)

.PHONY: $(SUBDIRS)
$(SUBDIRS):
	$(MAKE) -C $@

.PHONY: check
check:
	if $(CONCOURSE_BUILD) ; then \
		docker run --rm \
			-v $$PWD:/greenplum-for-kubernetes \
			-w /greenplum-for-kubernetes \
			golang:1.19 \
			make race lint; \
	else \
		make unit lint; \
	fi

# TODO: Do we still need this for Concourse?
unit race lint: PATH:=$(GOBIN):$(PATH)

unit: tools
	ginkgo -p -r -skipPackage=integration ./...
race: tools
	ginkgo -p -r -skipPackage=integration -race ./...
lint: tools
	golangci-lint run -c .golangci.yml --deadline 5m

docker-check: check docker

# Targets that just run the targets in $(SUBDIRS)
.PHONY: docker docker-check
docker docker-check:
	$(info $@ ...)
	for dir in $(SUBDIRS) ; do \
		$(MAKE) -C $$dir $@ || exit 1; \
	done

.PHONY: gke-docker
gke-docker: docker
	TAG_PREFIX=dev-$${TAG_PREFIX}$${TAG_PREFIX:+-}; \
	OPERATOR_IMAGE_SHA=$$(docker images --filter=reference=greenplum-operator:latest --format "{{.ID}}"); \
	OPERATOR_IMAGE_TAG=$$TAG_PREFIX$$OPERATOR_IMAGE_SHA; \
	echo "export OPERATOR_IMAGE_TAG=$$OPERATOR_IMAGE_TAG" > /tmp/.gp-kubernetes_gke_image_tags; \
	docker tag greenplum-operator:latest gcr.io/gp-kubernetes/greenplum-operator:$$OPERATOR_IMAGE_TAG; \
	docker push gcr.io/gp-kubernetes/greenplum-operator:$$OPERATOR_IMAGE_TAG; \
	GREENPLUM_IMAGE_SHA=$$(docker images --filter=reference=greenplum-for-kubernetes:latest --format "{{.ID}}"); \
	GREENPLUM_IMAGE_TAG=$$TAG_PREFIX$$GREENPLUM_IMAGE_SHA; \
	echo "export GREENPLUM_IMAGE_TAG=$$GREENPLUM_IMAGE_TAG" >> /tmp/.gp-kubernetes_gke_image_tags; \
	docker tag greenplum-for-kubernetes:latest gcr.io/gp-kubernetes/greenplum-for-kubernetes:$$GREENPLUM_IMAGE_TAG; \
	docker push gcr.io/gp-kubernetes/greenplum-for-kubernetes:$$GREENPLUM_IMAGE_TAG

.PHONY: tools
tools: ${GOBIN}/goimports ${GOBIN}/ginkgo ${GOBIN}/golangci-lint ${GOBIN}/loadmaster ${GOBIN}/kustomize ${GOBIN}/jsonnet controller-gen

${GOBIN}/goimports:
	go install -modfile tools/go.mod golang.org/x/tools/cmd/goimports

${GOBIN}/ginkgo:
	go install -modfile tools/go.mod github.com/onsi/ginkgo/ginkgo

.PHONY: ${GOBIN}/golangci-lint
${GOBIN}/golangci-lint:
	go install -modfile tools/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint

${GOBIN}/loadmaster:
	go install -modfile tools/go.mod github.com/bradfordboyle/loadmaster

${GOBIN}/kustomize:
	go install -modfile tools/go.mod sigs.k8s.io/kustomize/kustomize/v4

${GOBIN}/jsonnet:
	go install -modfile tools/go.mod github.com/google/go-jsonnet/cmd/jsonnet

.PHONY: controller-gen
controller-gen:
	make -C greenplum-operator controller-gen

.PHONY: release
release /tmp/greenplum-instance_release/greenplum-for-kubernetes-$(VERSION).tar.gz: docker
	TAG_PREFIX=${TAG_PREFIX} release/make_release.bash

.PHONY: release-check
release-check: /tmp/greenplum-instance_release/greenplum-for-kubernetes-$(VERSION).tar.gz
	TAG_PREFIX=${TAG_PREFIX} test/test-release-tarball.bash

.PHONY: pipeline-unit-tests
pipeline-unit-tests:
	jsonnet concourse/tests.jsonnet | awk 'BEGIN{fail=1} NR==1 && $$0 == "\"PASS\"" { fail=0 } {print} END{exit fail}' # unit tests

.PHONY: fly-pipeline
fly-pipeline: PIPELINE_NAME:=gp-kubernetes

.PHONY: fly-dev-pipeline
fly-dev-pipeline: GIT_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD)
fly-dev-pipeline: PIPELINE_NAME:=dev-$(shell echo -n ${GIT_BRANCH} | tr '[:upper:]' '[:lower:]' | tr -C '[:alnum:]' '-')
fly-dev-pipeline: JSONNET_ARGS:=--tla-code prod=false --tla-str git_branch=${GIT_BRANCH} --tla-str pipeline_name=${PIPELINE_NAME}

fly-pipeline fly-dev-pipeline: pipeline-unit-tests
	jsonnet concourse/pipeline.jsonnet ${JSONNET_ARGS} >/dev/null # check for jsonnet errors
	echo "setting pipeline ${PIPELINE_NAME}"
	fly -t k8s set-pipeline -p ${PIPELINE_NAME} -c <(jsonnet concourse/pipeline.jsonnet ${JSONNET_ARGS}) \
		--check-creds \
		-l ~/workspace/gp-continuous-integration/secrets/gpdb-5X-release-secrets.prod.yml \
		-l ~/workspace/gp-continuous-integration/secrets/greenplum-for-kubernetes/pipeline-secrets.yml \
		-l ~/workspace/gp-continuous-integration/secrets/greenplum-for-kubernetes/vsphere-secrets.yml \
		-l ~/workspace/gp-continuous-integration/secrets/gpdb_common-ci-secrets.yml

.PHONY: clean
clean:
	$(info $@ ...)
	for dir in $(SUBDIRS) ; do \
		$(MAKE) -C $$dir $@ || exit 1; \
	done

.PHONY: docker-clean
docker-clean:
	docker image prune -f --filter="dangling=true"

.PHONY: list
list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs

.PHONY: docker-debug
docker-debug:
	docker run -it --rm $$(docker images -q | head -n 1) bash

.PHONY: smoke
smoke:
	# container related
	$(MAKE) check
	$(MAKE) docker -B
	$(MAKE) docker-check
	$(MAKE) release
	$(MAKE) release-check

	test/singlenode-kubernetes.bash
	test/non-default-namespace.bash

	# integration
	$(MAKE) -C greenplum-operator integration

	# clean
	$(MAKE) deploy-clean

.PHONY: fire
fire:
	./test/fire.py

deploy gke-deploy deploy-operator gke-deploy-operator deploy-gpdb gke-deploy-gpdb deploy-clean gke-deploy-clean integration gke-integration:
	$(MAKE) -C greenplum-operator $@
