#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

set -euxo pipefail

# clean-up default namespace
kubectl config set-context $(kubectl config current-context) --namespace=
make -C ${SCRIPT_DIR}/../greenplum-operator deploy-clean

# create a new namespace
NAMESPACE=new-namespace
kubectl config set-context $(kubectl config current-context) --namespace=${NAMESPACE}
kubectl delete namespace ${NAMESPACE} || true
kubectl create namespace ${NAMESPACE}

# deploy should succeed with new namespace in my-gp-instance.yaml
if ! make -C ${SCRIPT_DIR}/../greenplum-operator deploy ; then
    echo "Deploy to namespace '${NAMESPACE}' should be successful."
    exit 1
fi

# verify new resources deployed in the ${NAMESPACE}
# master-0, segment-a-0 and title line
kubectl wait --for=condition=ready pod/master-0 pod/segment-a-0 || true

if ! kubectl get pods -l app=greenplum -o name | wc -l | grep -q 2 ; then
    echo "Failed to deploy to namespace '${NAMESPACE}'."
    exit 1
fi

# cleanup
make -C ${SCRIPT_DIR}/../greenplum-operator deploy-clean
kubectl delete namespace ${NAMESPACE}
kubectl config set-context $(kubectl config current-context) --namespace=
