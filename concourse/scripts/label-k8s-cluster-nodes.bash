#!/usr/bin/env bash

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${script_dir}/validate-env.bash"

# params (* = optional):
# * MASTER_LABEL: key/value label to apply to 2 nodes designated for gpdb masters
# * SEGMENT_LABEL: key/value label to apply to remaining nodes designated for gpdb segments
delete_k8s_cluster_node_labels() {
  if [ -n "${MASTER_LABEL}" ]; then
    echo "removing ${MASTER_LABEL} from all nodes"
    kubectl label nodes --all "$(echo "${MASTER_LABEL}" | awk -F'=' '{print $1}')-"
  fi
  if [ -n "${SEGMENT_LABEL}" ]; then
    echo "removing ${SEGMENT_LABEL} from all nodes"
    kubectl label nodes --all "$(echo "${SEGMENT_LABEL}" | awk -F'=' '{print $1}')-"
  fi
  return 0
}

# params (* = optional):
# * MASTER_LABEL: key/value label to apply to 2 nodes designated for gpdb masters
# * SEGMENT_LABEL: key/value label to apply to remaining nodes designated for gpdb segments
label_k8s_cluster_nodes() {
  NODES=$(kubectl get nodes -o json | jq -r '.items[].metadata.name')
  # convert NODES to an array
  SAVEIFS=$IFS
  IFS=$'\n'
  NODES=($NODES)
  IFS=$SAVEIFS

  i=0
  if [ -n "${MASTER_LABEL}" ]; then
    echo "labeling nodes with ${MASTER_LABEL}"
    for (( ; i<2 && i<${#NODES[@]}; i++ )); do
      kubectl label node "${NODES[i]}" "${MASTER_LABEL}"
    done
  fi
  if [ -n "${SEGMENT_LABEL}" ]; then
    echo "labeling nodes with ${SEGMENT_LABEL}"
    for (( ; i<${#NODES[@]}; i++ )); do
      kubectl label node "${NODES[i]}" "${SEGMENT_LABEL}"
    done
  fi
  return 0
}

function _main() {
  if [ "${SKIP_LOGIN}" != "TRUE" ] ; then
    case "${K8S_CLUSTER_TYPE}" in
      "gke" | "gke-private")
        required_gke_env_vars=(
          "K8S_CLUSTER_NAME"
          "GCP_SVC_ACCT_KEY"
          "GCP_PROJECT"
          "GCP_ZONE"
        )
        validate_env_vars "${required_gke_env_vars[@]}" || exit 1
        ;;
      "pks")
        required_pks_env_vars=(
          "K8S_CLUSTER_NAME"
          "PKS_USER"
          "PKS_PASSWORD"
        )
        validate_env_vars "${required_pks_env_vars[@]}" || exit 1
        ;;
      *)
        echo "K8S_CLUSTER_TYPE env var must be one of: {gke, gke-private, pks}"
        exit 1
        ;;
    esac

    # shellcheck disable=SC1090
    source "${script_dir}/login-k8s-cluster.bash"
  fi

  echo "labeling kubernetes cluster nodes..."
  delete_k8s_cluster_node_labels || exit 1
  label_k8s_cluster_nodes || exit 1
  echo "...kubernetes cluster node labeling successful"
}

_main
