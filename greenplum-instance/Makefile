SHELL:=/bin/bash
RELEASE_VERSION:=$(shell ../getversion)
IMAGE_REPO:=greenplum-for-kubernetes
#TAG_PREFIX empty by default
IMAGE_TAG:=$(IMAGE_REPO):$(TAG_PREFIX)$(shell ../getversion --image)
GOBIN:=$(PWD)/bin
DATE:=$(shell date +%FT%H:%M:%S)
# TODO: rename this.
DOCKERTAG_PREFIX:=greenplum-instance
CONCOURSE_BUILD?=false

UNAME:=$(shell uname -s)
GREP_WITH_REGEX:=grep -Pq
ifeq ($(UNAME),Darwin)
    GREP_WITH_REGEX:=grep -q
endif

# default target
.PHONY: build
build: build-in-docker

sources := $(shell find . -name "*.go" -print)
build-in-docker: $(sources) Dockerfile
	docker build --target build-in-docker --iidfile build-in-docker --tag $(DOCKERTAG_PREFIX):build-in-docker --file Dockerfile ..

.PHONY: docker-check
docker-check: package-in-docker container-structure-test-in-docker
	docker run --tty --rm -v /var/run/docker.sock:/var/run/docker.sock $(DOCKERTAG_PREFIX):container-structure-test-in-docker

.PHONY: docker
docker: package-in-docker

package-in-docker: export DOCKER_BUILDKIT=1
container-structure-test-in-docker: export DOCKER_BUILDKIT=1
package-in-docker container-structure-test-in-docker: $(sources) Dockerfile Dockerfile_container_structure_test Dockerfile_test.yml scripts/* scripts/build/*
	@if [ -z "${GREENPLUM_VERSION}" ]; then \
		GREENPLUM_VERSION=$$(gcloud container images list-tags gcr.io/gp-kubernetes/ubuntu-gpdb-ent --filter="tags=$${TAG_PREFIX}latest" --format="value(tags.filter(NOT \"$${TAG_PREFIX}latest\"))" | cut -c $$(($${#TAG_PREFIX}+1))-) ; \
    fi; \
	echo "Using GREENPLUM_VERSION: $${GREENPLUM_VERSION}" ; \
	echo "Using CONCOURSE_BUILD: $(CONCOURSE_BUILD)" ; \
	docker build \
		--iidfile package-in-docker \
		-t "$(IMAGE_TAG)" \
		-t "$(IMAGE_REPO):latest" \
		-f Dockerfile \
		--build-arg CONCOURSE_BUILD=$(CONCOURSE_BUILD) \
		--build-arg VERSION="$(RELEASE_VERSION)" \
		--build-arg DATE="$(DATE)" \
		--build-arg TAG_PREFIX="$(TAG_PREFIX)" \
		--build-arg GREENPLUM_VERSION="$${GREENPLUM_VERSION}" \
		.. && \
	docker build \
		--iidfile container-structure-test-in-docker \
		--build-arg GREENPLUM_VERSION="$${GREENPLUM_VERSION}" \
		--build-arg VERSION="$(RELEASE_VERSION)" \
		--build-arg DATE="$(DATE)" \
		--build-arg IMAGE_TAG="$(IMAGE_TAG)" \
		-t $(DOCKERTAG_PREFIX):container-structure-test-in-docker \
		-f Dockerfile_container_structure_test \
		..

.PHONY: lint
lint:
	PATH=$(GOBIN):$$PATH golangci-lint run -c ../.golangci.yml

.PHONY: check
check: unit lint

.PHONY: unit
unit:
	PATH=$(GOBIN):$$PATH ginkgo -p -r -skipPackage=integration -race ./...

.PHONY: clean
clean:
	for tag in \
			package-in-docker \
			container-structure-test-in-docker \
			build-in-docker \
			; do \
		if [ -f $$tag ] ; then \
			docker rmi -f $$(cat $$tag) ; \
			rm -f $$tag ; \
		fi ; \
	done
