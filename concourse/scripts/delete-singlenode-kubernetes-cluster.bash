#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${SCRIPT_DIR}/cloud_login.bash

auth_gcloud

# cleanup env
SINGLENODE_KUBE_CLUSTER_NAME="singlenode-kubernetes"
gcloud container clusters delete ${SINGLENODE_KUBE_CLUSTER_NAME} --quiet 2>/dev/null || true

