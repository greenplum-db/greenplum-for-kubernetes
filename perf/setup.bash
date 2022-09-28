#!/usr/bin/env bash

set -euxo pipefail

if [ $# -lt 1 ]; then
    set +x
    echo "usage: setup.bash <segment_count>"
    exit 1
fi

function setup_container()  {
    CONTAINER_NAME=$1
    kubectl cp ~/workspace/greenplum-for-kubernetes/perf/disk_perf.py ${CONTAINER_NAME}:/home/gpadmin/
    kubectl cp ~/workspace/greenplum-for-kubernetes/perf/disk_perf.bash ${CONTAINER_NAME}:/home/gpadmin/
    kubectl cp ~/workspace/gpdb/gpMgmt/bin/gpcheckperf ${CONTAINER_NAME}:/usr/local/greenplum-db/bin/
    kubectl cp ~/workspace/gpdb/gpMgmt/bin/lib/multidd ${CONTAINER_NAME}:/usr/local/greenplum-db/bin/lib/

#    cannot run any process in detached mode. exec will wait for *any* child process, no matter what.
#    kubectl exec ${CONTAINER_NAME} -- /home/gpadmin/disk_perf.bash 10
}


setup_container "master-0"
setup_container "master-1"

SEGMENT_COUNT=$1

SEGMENT_LIMIT=$(($SEGMENT_COUNT-1))
for i in `seq 0 ${SEGMENT_LIMIT}`;
do
    setup_container "segment-a-${i}"
    setup_container "segment-b-${i}"
done
