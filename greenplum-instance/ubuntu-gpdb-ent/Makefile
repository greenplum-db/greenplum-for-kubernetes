SHELL:=/bin/bash
CONCOURSE_BUILD?=false

# TODO: need to fix a lot of thngs related to components
# Should get PXF version from greenplum debian (or decide to install PXF separately ourselves)

# If minikube is around, don't use it when running loadmaster.
ifeq (${MINIKUBE_ACTIVE_DOCKERD},minikube)
unset_minikube=(eval $$(minikube docker-env -u); $(1))
else
unset_minikube=$(1)
endif

.PHONY: docker
docker: export DOCKER_BUILDKIT=1
docker: fly-deps
	GREENPLUM_VERSION=$$( $(call unset_minikube,docker run --rm -v ${PWD}/greenplum-debian-binary:/greenplum-debian-binary gcr.io/gp-kubernetes/gpdb6-ubuntu18.04-base-gp4k8s bash -c 'dpkg-deb -f /greenplum-debian-binary/greenplum-db*.deb Version')); \
	MADLIB_VERSION=$$( $(call unset_minikube,docker run --rm -v ${PWD}/madlib-deb-ubuntu18:/madlib-deb-ubuntu18 gcr.io/gp-kubernetes/gpdb6-ubuntu18.04-base-gp4k8s bash -c 'dpkg-deb -f /madlib-deb-ubuntu18/apache-madlib*.deb Version')); \
	PXF_VERSION=$$( $(call unset_minikube,docker run --rm -v ${PWD}/pxf-gp6-deb-ubuntu18:/pxf-gp6-deb-ubuntu18 gcr.io/gp-kubernetes/gpdb6-ubuntu18.04-base-gp4k8s bash -c 'dpkg-deb -f /pxf-gp6-deb-ubuntu18/pxf-gp6-*.deb Version')); \
	GPBACKUP_VERSION=$$(cat ${PWD}/gpbackup-ubuntu18/version | awk -F'#' '{print $$1}'); \
	GPDB_TAG=$$(echo $${GREENPLUM_VERSION} | sed 's/+/\./g'); \
	echo "Using GREENPLUM_VERSION: $${GPDB_TAG}" ; \
	echo "Using MADLIB_VERSION: $${MADLIB_VERSION}" ; \
	echo "Using GPBACKUP_VERSION: $${GPBACKUP_VERSION}" ; \
	docker build --no-cache --tag gcr.io/gp-kubernetes/ubuntu-gpdb-ent:latest \
	    --build-arg GREENPLUM_VERSION=$${GPDB_TAG} \
	    --build-arg MADLIB_VERSION=$${MADLIB_VERSION} \
	    --build-arg PXF_VERSION=$${PXF_VERSION} \
	    --build-arg GPBACKUP_VERSION=$${GPBACKUP_VERSION} \
	    -f ./Dockerfile \
	     ${PWD} && \
	docker build --no-cache --tag ubuntu-gpdb-ent:container-structure-test-in-docker \
	    --iidfile container-structure-test-in-docker \
	    --build-arg GREENPLUM_VERSION=$${GPDB_TAG} \
	    --build-arg MADLIB_VERSION=$${MADLIB_VERSION} \
	    --build-arg PXF_VERSION=$${PXF_VERSION} \
	    --build-arg GPBACKUP_VERSION=$${GPBACKUP_VERSION} \
	    --build-arg IMAGE_TAG="gcr.io/gp-kubernetes/ubuntu-gpdb-ent:latest" \
	    -f Dockerfile_container_structure_test \
	    .

.PHONY: docker-check
docker-check: docker
	docker run --tty --rm -v /var/run/docker.sock:/var/run/docker.sock ubuntu-gpdb-ent:container-structure-test-in-docker

.PHONY: loadmaster
loadmaster:
	@if ! $(CONCOURSE_BUILD) && ! which loadmaster ; then \
	    echo "Please install loadmaster"; exit 1; \
	 fi

.PHONY: fly-deps
fly-deps: loadmaster \
  ${PWD}/greenplum-debian-binary \
  ${PWD}/madlib-deb-ubuntu18 \
  ${PWD}/pxf-gp6-deb-ubuntu18 \
  ${PWD}/gpbackup-ubuntu18 \

${PWD}/greenplum-debian-binary:
	cd ${PWD}; fly -t k8s gp -p gp-kubernetes | $(call unset_minikube,loadmaster -i greenplum-debian-binary)

${PWD}/madlib-deb-ubuntu18:
	cd ${PWD}; fly -t k8s gp -p gp-kubernetes | $(call unset_minikube,loadmaster -i madlib-deb-ubuntu18)

${PWD}/pxf-gp6-deb-ubuntu18:
	cd ${PWD}; fly -t k8s gp -p gp-kubernetes | $(call unset_minikube,loadmaster -i pxf-gp6-deb-ubuntu18)

${PWD}/gpbackup-ubuntu18:
	cd ${PWD}; fly -t k8s gp -p gp-kubernetes | $(call unset_minikube,loadmaster -i gpbackup-ubuntu18)

# TODO: clean up components when we add them back
.PHONY: clean
clean:
	rm -rf ${PWD}/greenplum-debian-binary \
	       ${PWD}/madlib-deb-ubuntu18 \
	       ${PWD}/pxf-gp6-deb-ubuntu18 \
	       ${PWD}/gpbackup-ubuntu18 \
