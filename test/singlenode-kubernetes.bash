#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ ! $(kubectl get nodes --no-headers | wc -l) -eq 1 ]; then \
    echo "There are more than one nodes. Exiting..."
    exit -1
fi

# set antiAffinity to `no`
sed -i "s%  antiAffinity:.*%  antiAffinity: \"no\"%" ${SCRIPT_DIR}/../workspace/my-gp-instance.yaml

# clean up
make -C ${SCRIPT_DIR}/../greenplum-operator deploy-clean

# deploy gpdb cluster on single node kubernetes
make -C ${SCRIPT_DIR}/../greenplum-operator deploy

# test greenplum
kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; psql -c 'select * from gp_segment_configuration'"

