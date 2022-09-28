#!/usr/bin/env bash

if [ $# -lt 1 ]; then
    set +x
    echo "usage: cleanup.bash <cluster_name>"
    exit 1
fi

CLUSTER_NAME=$1
POOL_NAME="default-pool"

# cluster might have already been purged, so ok if this fails
helm uninstall ${CLUSTER_NAME}

