#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${SCRIPT_DIR}/cloud_login.bash

SINGLENODE_KUBE_CLUSTER_NAME="singlenode-kubernetes"
RELEASE_DIR="$PWD/gp-kubernetes-rc-release/greenplum-for-kubernetes-v$(cat gp-kubernetes-rc-release/version)"

auth_gcloud

# cleanup env
gcloud container clusters delete ${SINGLENODE_KUBE_CLUSTER_NAME} --quiet 2>/dev/null || true

# create a 1 node cluster in GKE
gcloud beta container clusters create \
    ${SINGLENODE_KUBE_CLUSTER_NAME} \
    --num-nodes 1 \
    --cluster-version ${GCP_GKE_VERSION} \
    --machine-type n1-standard-4

node_hostname=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(.type == "Hostname")].address}' | awk -F ".c.gp-kubernetes" '{ print $1 }')

# exchange keys first time on SSH
yes | gcloud compute ssh \
    pivotal@${node_hostname} \
    --command="echo SSH established" \
    --strict-host-key-checking=no || true
# use docker in GKE worker node
gcloud compute ssh \
    pivotal@${node_hostname} -- \
    -N -o StreamLocalBindUnlink=yes \
    -L $HOME/gke.sock:/var/run/docker.sock &
DOCKER_PID=${!}

export DOCKER_HOST=unix://$HOME/gke.sock

# wait for GKE socket
while [ ! -S ${HOME}/gke.sock ]; do sleep 1; printf "."; done

# build greenplumReady
pushd greenplum-for-kubernetes/greenplum-operator/cmd/greenplumReady && go build && popd

RELEASE_DIR=${RELEASE_DIR} ${PWD}/greenplum-for-kubernetes/test/load-release.bash

${PWD}/greenplum-for-kubernetes/test/singlenode-kubernetes.bash
${PWD}/greenplum-for-kubernetes/test/non-default-namespace.bash

kill ${DOCKER_PID}
wait ${DOCKER_PID}
gcloud container clusters delete ${SINGLENODE_KUBE_CLUSTER_NAME} --quiet
