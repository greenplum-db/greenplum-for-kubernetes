#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${SCRIPT_DIR}/cloud_login.bash

# TODO: come back and remove KUBEENV when conversion is finished
if [ ${KUBEENV} == "GKE" ]; then
    auth_gcloud
    yes | gcloud container clusters delete ${CLUSTER_NAME} || true
else
    login_pks
    pks delete-cluster ${CLUSTER_NAME} --non-interactive || true  # ok to fail if there is nothing to delete
fi
