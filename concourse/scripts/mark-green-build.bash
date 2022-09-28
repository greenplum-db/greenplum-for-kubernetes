#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ADDITIONAL_TAG=${ADDITIONAL_TAG:-"latest-green"}
source ${SCRIPT_DIR}/cloud_login.bash

RELEASE_NAME=$(basename gp-kubernetes-rc-release/*.tar.gz .tar.gz)

tar xf gp-kubernetes-rc-release/$RELEASE_NAME.tar.gz

GREENPLUM_INSTANCE_IMG_TAG="greenplum-for-kubernetes:$(cat $RELEASE_NAME/images/greenplum-for-kubernetes-tag)"
GREENPLUM_OPERATOR_IMG_TAG="greenplum-operator:$(cat $RELEASE_NAME/images/greenplum-operator-tag)"

cp gp-kubernetes-rc-release/* gp-kubernetes-green-output/

auth_gcloud
gcloud container images add-tag -q gcr.io/gp-kubernetes/${GREENPLUM_INSTANCE_IMG_TAG} gcr.io/gp-kubernetes/greenplum-for-kubernetes:${ADDITIONAL_TAG}
gcloud container images add-tag -q gcr.io/gp-kubernetes/${GREENPLUM_OPERATOR_IMG_TAG} gcr.io/gp-kubernetes/greenplum-operator:${ADDITIONAL_TAG}
