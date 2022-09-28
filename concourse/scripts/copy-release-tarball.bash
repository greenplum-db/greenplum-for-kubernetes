#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${SCRIPT_DIR}/cloud_login.bash

TARBALL_NAME=$(ls gp-kubernetes-tagged-release-tarball/greenplum-for-kubernetes-*.tar.gz)
TARBALL_BASE_NAME=$(basename ${TARBALL_NAME})

auth_gcloud

gsutil cp gs://gp-kubernetes-ci-release/${TARBALL_BASE_NAME} gs://greenplum-for-kubernetes-release/

gsutil ls gs://greenplum-for-kubernetes-release/${TARBALL_BASE_NAME}
