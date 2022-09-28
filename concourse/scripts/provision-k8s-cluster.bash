#!/usr/bin/env bash

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${script_dir}/validate-env.bash"
# shellcheck disable=SC1090
source "${script_dir}/cloud_login.bash"

K8S_CLUSTER_NODE_COUNT=${K8S_CLUSTER_NODE_COUNT:-4}
GCP_MACHINE_TYPE=${GCP_MACHINE_TYPE:-e2-standard-2}

# params (* = optional):
#   GCP_SVC_ACCT_KEY: gcp service account key for authentication
#   GCP_PROJECT: gcp project for k8s cluster
#   K8S_CLUSTER_NAME: k8s cluster name
#   GCP_NETWORK: gcp network for k8s cluster
#   GCP_SUBNETWORK: gcp subnetwork for k8s cluster
#   GCP_ZONE: gcp zone for k8s cluster
# * GCP_MACHINE_TYPE: gcp vm instance type for k8s cluster nodes (default: n1-standard-2)
# * K8S_CLUSTER_NODE_COUNT: number of nodes in k8s cluster (default: 4)
# * GCP_GKE_VERSION: gke version for k8s cluster (default: latest)
# * GCP_GKE_CLUSTER_CIDR: cidr range for k8s cluster pods (default: not specified)
function provision_gke_k8s_cluster() {
  auth_gcloud
  if ! gcloud container clusters describe "${K8S_CLUSTER_NAME}" > /dev/null 2>&1 ; then
    CIDR_PARAMS=""
    if [ ! -z "${GCP_GKE_CLUSTER_CIDR}" ]; then
        CIDR_PARAMS="--cluster-ipv4-cidr ${GCP_GKE_CLUSTER_CIDR}"
    fi
    GCP_GKE_VERSION="${GCP_GKE_VERSION:-latest}"
    echo "provisioning gke kubernetes cluster..."
    gcloud container clusters create \
      "${K8S_CLUSTER_NAME}" \
      --cluster-version "${GCP_GKE_VERSION}" \
      --num-nodes "${K8S_CLUSTER_NODE_COUNT}" \
      --machine-type "${GCP_MACHINE_TYPE}" \
      --image-type ubuntu \
      --network "${GCP_NETWORK}" \
      --subnetwork "${GCP_SUBNETWORK}" \
      --zone "${GCP_ZONE}" \
      ${CIDR_PARAMS}
    kubectl wait --for condition=available --timeout=300s --all apiservice
  else
    echo "skipping creating gke cluster. ${K8S_CLUSTER_NAME} already exists"
  fi
  return 0
}

# params (* = optional):
#   GCP_SVC_ACCT_KEY: gcp service account key for authentication
#   GCP_PROJECT: gcp project for k8s cluster
#   K8S_CLUSTER_NAME: k8s cluster name
#   GCP_NETWORK: gcp network for k8s cluster
#   GCP_SUBNETWORK: gcp subnetwork for k8s cluster
#   GCP_ZONE: gcp zone for k8s cluster
#   GCP_GKE_MASTER_CIDR: cidr range to use for master vm network
# * GCP_MACHINE_TYPE: gcp vm instance type for k8s cluster nodes (default: n1-standard-2)
# * K8S_CLUSTER_NODE_COUNT: number of nodes in k8s cluster (default: 4)
# * GCP_GKE_VERSION: gke version for k8s cluster (default: latest)
function provision_gke_private_k8s_cluster() {
  auth_gcloud
  if ! gcloud container clusters describe "${K8S_CLUSTER_NAME}" > /dev/null 2>&1 ; then
    GCP_GKE_VERSION="${GCP_GKE_VERSION:-latest}"
    echo "provisioning gke-private kubernetes cluster..."
    gcloud container clusters create \
      "${K8S_CLUSTER_NAME}" \
      --cluster-version "${GCP_GKE_VERSION}" \
      --num-nodes "${K8S_CLUSTER_NODE_COUNT}" \
      --machine-type "${GCP_MACHINE_TYPE}" \
      --image-type ubuntu \
      --network "${GCP_NETWORK}" \
      --subnetwork "${GCP_SUBNETWORK}" \
      --zone "${GCP_ZONE}" \
      --enable-master-authorized-networks \
      --master-authorized-networks 10.0.128.0/23 \
      --enable-ip-alias \
      --cluster-secondary-range-name pods \
      --services-secondary-range-name services \
      --enable-private-nodes \
      --master-ipv4-cidr "${GCP_GKE_MASTER_CIDR}" \
      --no-enable-basic-auth \
      --no-issue-client-certificate \
      --enable-private-endpoint \
      --scopes=logging-write,monitoring,service-management,service-control,trace
    kubectl wait --for condition=available --timeout=120s --all apiservice
  else
    echo "skipping creating gke cluster. ${K8S_CLUSTER_NAME} already exists"
  fi
  return 0
}

# params (* = optional):
#   GCP_SVC_ACCT_KEY: gcp service account key for authentication
#   GCP_PROJECT: gcp project for k8s cluster
#   K8S_CLUSTER_NAME: k8s cluster name
#   PKS_USER: pks username
#   PKS_PASSWORD: pks password
#   PKS_CLUSTER_LOAD_BALANCER: load balancer address for pks k8s cluster
#   PKS_PLAN: pks plan name for k8s cluster
# * PKS_API_URL: pks api address
# * K8S_CLUSTER_NODE_COUNT: number of nodes in k8s cluster (default: 4)
function provision_pks_k8s_cluster() {
  auth_gcloud
  pks --version
  login_pks || exit 1
  if ! pks cluster "${K8S_CLUSTER_NAME}" 2>/dev/null 1>/dev/null ; then
    echo "provisioning pks kubernetes cluster..."
    "${script_dir}/../../workspace/samples/scripts/create_pks_cluster_on_gcp.bash" \
      "${K8S_CLUSTER_NAME}" \
      "${PKS_CLUSTER_LOAD_BALANCER}" \
      "${K8S_CLUSTER_NODE_COUNT}"
  else
    echo "skipping creating pks cluster. ${K8S_CLUSTER_NAME} already exists"
  fi
  return 0
}

function _main() {
  required_common_env_vars=(
    "K8S_CLUSTER_TYPE"
    "GCP_SVC_ACCT_KEY"
    "GCP_PROJECT"
    "K8S_CLUSTER_NAME"
  )
  validate_env_vars "${required_common_env_vars[@]}" || exit 1

  case "${K8S_CLUSTER_TYPE}" in
    "gke")
      required_gke_env_vars=(
        "GCP_NETWORK"
        "GCP_SUBNETWORK"
        "GCP_ZONE"
      )
      validate_env_vars "${required_gke_env_vars[@]}" || exit 1
      provision_gke_k8s_cluster || exit 1
      ;;
    "gke-private")
      required_gke_private_env_vars=(
        "GCP_MACHINE_TYPE"
        "GCP_NETWORK"
        "GCP_SUBNETWORK"
        "GCP_ZONE"
        "GCP_GKE_MASTER_CIDR"
      )
      validate_env_vars "${required_gke_private_env_vars[@]}" || exit 1
      provision_gke_private_k8s_cluster || exit 1
      ;;
    "pks")
      required_pks_env_vars=(
        "PKS_USER"
        "PKS_PASSWORD"
        "PKS_CLUSTER_LOAD_BALANCER"
        "PKS_PLAN"
      )
      validate_env_vars "${required_pks_env_vars[@]}" || exit 1
      provision_pks_k8s_cluster || exit 1
      ;;
    *)
      echo "K8S_CLUSTER_TYPE env var must be one of: {gke, gke-private, pks}"
      exit 1
      ;;
  esac

  echo "...kubernetes cluster provision successful"

  # shellcheck disable=SC1090
  source "${script_dir}/login-k8s-cluster.bash"
}

_main
