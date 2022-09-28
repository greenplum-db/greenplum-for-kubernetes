#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function get_status {
  pks show-cluster "$1" --json | jq .last_action_state --raw-output
}

function wait_for_loadbalancer() (
    set +xe

    echo ""
    echo "Waiting for loadbalancer to be up and running on ${LB_IP_ADDRESS} "
    local lb_status="not running"
    local time_in_seconds=0
    while [[ "$time_in_seconds" -lt 1800 && "$lb_status" != "Running" ]]; do
        local time_in_seconds=$(( time_in_seconds+8 ))
        sleep 8
        printf "."
        timeout 1 bash -c "(echo > /dev/tcp/${LB_IP_ADDRESS}/8443) >/dev/null 2>&1"
        RESULT=$?
        if [[ "$RESULT" == "0" ]]; then
            lb_status="Running"
            echo "LoadBalancer is up and running on ${LB_IP_ADDRESS}"
        fi
    done

    if [[ "$lb_status" != "Running" ]]; then
        echo "Failed to establish load balancer"
        exit 1
    fi
)

function get_pks_cluster_uuid {
  local cluster_name=$1
  # quiet stderr report if cluster not found. json will parse to empty string, no problem
  pks cluster ${cluster_name} --json  2> /dev/null | jq -r '.uuid'
}

function attach_lb_to_kube_master() {
    CLUSTER_NAME=$1
    LB_IP_ADDRESS=$2
    pks_cluster_uuid=$(get_pks_cluster_uuid ${CLUSTER_NAME})
    LB_BACKEND_INSTANCE_NAME=$(gcloud compute instances list \
        --filter="labels.instance_group=master AND labels.deployment=service-instance-$pks_cluster_uuid" \
        --format="value[terminator=','](name)" | sed 's/,$//')
    POOL=$(gcloud compute forwarding-rules list --filter="IP_ADDRESS: $LB_IP_ADDRESS" --format="value(target)")

    if [ "$POOL" == "" ]; then
        echo "cannot find load balancer associated with IP address $LB_IP_ADDRESS"
        echo "cannot attach load balancer, so kubectl will not work; please delete cluster and try again."
        exit 1
    fi

    # TODO: POOL can be a list, semi-colon delimited. We are assuming one instance for now
    OLD_INSTANCE_LIST=$(gcloud compute target-pools describe ${POOL} --format="value(instances)")

    if [ ! "$OLD_INSTANCE_LIST" == "" ]; then
        gcloud compute target-pools remove-instances ${POOL} --instances=${OLD_INSTANCE_LIST}
    fi
    gcloud compute target-pools add-instances ${POOL} --instances=${LB_BACKEND_INSTANCE_NAME}
}

function main {
    set -euo pipefail

    if [ $# -lt 3 ]; then
        echo "usage: create_pks_cluster_on_gcp.bash <cluster_name> <load balancer address (the front end)> <node_count>"
        exit 1
    fi

    CLUSTER_NAME=$1
    LB_IP_ADDRESS=$2
    NODE_COUNT=$3
    : ${PKS_PLAN:=medium}

    set -x

    pks create-cluster ${CLUSTER_NAME} --external-hostname ${LB_IP_ADDRESS} --num-nodes ${NODE_COUNT} -p ${PKS_PLAN} --wait
    attach_lb_to_kube_master ${CLUSTER_NAME} ${LB_IP_ADDRESS}
    pks get-credentials ${CLUSTER_NAME}
    time wait_for_loadbalancer ${LB_IP_ADDRESS}
}

# run main unless this file is just being "source"d for access to functions
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
