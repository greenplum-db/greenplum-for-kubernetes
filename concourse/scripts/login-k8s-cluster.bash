#!/usr/bin/env bash

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${script_dir}/validate-env.bash"
# shellcheck disable=SC1090
source "${script_dir}/cloud_login.bash"

# params:
#   K8S_CLUSTER_NAME: k8s cluster name
#   GCP_SVC_ACCT_KEY: gcp service account key for authentication
#   GCP_PROJECT: gcp project for k8s cluster
#   GCP_ZONE: gcp zone for k8s cluster
function login_gke_k8s_cluster() {
  echo "login to gke kubernetes cluster..."
  auth_gcloud
  gcloud container clusters get-credentials "${K8S_CLUSTER_NAME}"
  return 0
}

# params (* = optional):
#   K8S_CLUSTER_NAME: k8s cluster name
#   PKS_USER: pks username
#   PKS_PASSWORD: pks password
# * PKS_API_URL: pks api address
function login_pks_k8s_cluster() {
  echo "login to pks kubernetes cluster..."
  login_pks
  pks get-credentials "${K8S_CLUSTER_NAME}"
  return 0
}

function _main() {
  required_common_env_vars=(
    "K8S_CLUSTER_TYPE"
    "K8S_CLUSTER_NAME"
  )
  validate_env_vars "${required_common_env_vars[@]}" || exit 1

  case "${K8S_CLUSTER_TYPE}" in
    "gke" | "gke-private")
      required_gke_env_vars=(
        "GCP_SVC_ACCT_KEY"
        "GCP_PROJECT"
        "GCP_ZONE"
      )
      validate_env_vars "${required_gke_env_vars[@]}" || exit 1
      login_gke_k8s_cluster || exit 1
      ;;
    "pks")
      required_pks_env_vars=(
        "PKS_USER"
        "PKS_PASSWORD"
      )
      validate_env_vars "${required_pks_env_vars[@]}" || exit 1
      login_pks_k8s_cluster || exit 1
      ;;
    *)
      echo "K8S_CLUSTER_TYPE env var must be one of: {gke, gke-private, pks}"
      exit 1
      ;;
  esac

  echo "...kubernetes cluster login successful"
}

_main
