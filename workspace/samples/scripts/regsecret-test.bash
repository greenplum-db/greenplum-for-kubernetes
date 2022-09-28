#! /usr/bin/env bash

GREENPLUM_OPERATOR_IMAGE=$1

die() {
    echo "$*"
    exit 1
}

if [ -z "${GREENPLUM_OPERATOR_IMAGE}" ]; then
    echo "Usage: regsecret-test.bash <greenplum-operator-image-repo:tag>"
    exit 1
fi

cd $(dirname "$0") || die "Couldn't change to script directory"

<regsecret-test.yaml sed -e "s#GREENPLUM_OPERATOR_IMAGE#${GREENPLUM_OPERATOR_IMAGE}#g" \
| kubectl create -f - || die "Failed to create Job"

for ((i=0; i<120; i++)); do
    sleep 1
    if kubectl logs job.batch/greenplum-operator-fetch-test 2>&1 | grep -q 'GREENPLUM-OPERATOR TEST OK\|trying and failing to pull image' ; then
        break
    fi
done

kubectl logs job.batch/greenplum-operator-fetch-test | grep "GREENPLUM-OPERATOR TEST OK"

kubectl delete job.batch/greenplum-operator-fetch-test
