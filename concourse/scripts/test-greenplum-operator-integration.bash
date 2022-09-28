#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${SCRIPT_DIR}/cloud_login.bash

# shellcheck disable=SC2223
: ${INTEGRATION_TYPE:="default"}
# shellcheck disable=SC2223
: ${GP_INSTANCE_IMAGE_REPO:=gcr.io/${GCP_PROJECT}/greenplum-for-kubernetes}

RELEASE_DIR="gp-kubernetes-release/greenplum-for-kubernetes-v$(cat gp-kubernetes-release/version)"
OLD_RELEASE_DIR="greenplum-for-kubernetes-old-release/greenplum-for-kubernetes-v2.2.0"
TAG=$(cat "${RELEASE_DIR}"/images/greenplum-for-kubernetes-tag)

"${SCRIPT_DIR}"/login-k8s-cluster.bash

auth_gcloud
# fail fast if image is not available
gcloud container images describe "${GP_INSTANCE_IMAGE_REPO}":"${TAG}"

INTEGRATION_TYPE=${INTEGRATION_TYPE} \
    RELEASE_DIR=${RELEASE_DIR} \
    OLD_RELEASE_DIR=${OLD_RELEASE_DIR} \
    GCP_IMAGE_TAG=${TAG} \
    greenplum-for-kubernetes/test/operator_integration_test.bash
