#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${SCRIPT_DIR}/kube_environment.bash

GCP_REGISTRY_PROJECT=gcr.io/gp-kubernetes
: ${GCP_GP_OP_REPO:=${GCP_REGISTRY_PROJECT}/greenplum-operator}
: ${GCP_GP_FOR_K8_REPO:=${GCP_REGISTRY_PROJECT}/greenplum-for-kubernetes}
: ${GCP_IMAGE_TAG:=$(${SCRIPT_DIR}/../getversion --image)}
: ${INTEGRATION_TYPE:=ha}
: ${RELEASE_DIR:=}
: ${OLD_RELEASE_DIR:=}

# fail fast if kubectl is not responsive
kubectl version

if is_minikube ; then
    make -C ${HOME}/workspace/greenplum-for-kubernetes/greenplum-operator integration-${INTEGRATION_TYPE}
else
    RELEASE_DIR_PATH=${PWD}/${RELEASE_DIR}
    OLD_RELEASE_DIR_PATH=${PWD}/${OLD_RELEASE_DIR}
    pushd greenplum-for-kubernetes
        export GOBIN=${PWD}/bin
        export PATH=${PATH}:${GOBIN}
        make tools
        ginkgo -v -failFast -r greenplum-operator/integration/${INTEGRATION_TYPE} -- \
            --release-dir=${RELEASE_DIR_PATH} \
            ${OLD_RELEASE_DIR:+--old-release-dir=${OLD_RELEASE_DIR_PATH}} \
            --greenplumImageRepository ${GCP_GP_FOR_K8_REPO} \
            --greenplumImageTag ${GCP_IMAGE_TAG} \
            --operatorImageRepository ${GCP_GP_OP_REPO} \
            --operatorImageTag ${GCP_IMAGE_TAG}
    popd
fi
