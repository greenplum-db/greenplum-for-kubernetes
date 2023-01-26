#!/bin/bash

set -euxo pipefail

: ${GP4K_VERSION:="latest-green"}
GP_INSTANCE_IMAGE_REPO=gcr.io/gp-kubernetes/greenplum-for-kubernetes
GP_OPERATOR_IMAGE_REPO=gcr.io/gp-kubernetes/greenplum-operator

: ${SEGMENT_COUNT:=1}
: ${NODE_COUNT:=$((2*${SEGMENT_COUNT} + 2))}
: ${CPU_RATE_LIMIT:=70}
: ${MEMORY_LIMIT:=70}
: ${GP_INSTANCE_NAME:="my-greenplum"}

source greenplum-for-kubernetes/concourse/scripts/cloud_login.bash

buildDefaultOperatorValuesYml() {
    echo "use default operator values.yaml as below"
    cat > greenplum-for-kubernetes/greenplum-operator/operator/values.yaml <<_EOF
---
# specify the url for the docker image for the operator, e.g. gcr.io/<my_project>/greenplum-operator
operatorImageRepository: ${GP_OPERATOR_IMAGE_REPO}
operatorImageTag: ${GP4K_VERSION}

# specify the docker image for greenplum instance, e.g. gcr.io/<my_project>/greenplum-instance
greenplumImageRepository: ${GP_INSTANCE_IMAGE_REPO}
greenplumImageTag: ${GP4K_VERSION}

_EOF
   cat greenplum-for-kubernetes/greenplum-operator/operator/values.yaml
}

createGKECluster() {
    if ! gcloud container clusters describe ${CLUSTER_NAME} > /dev/null 2>&1 ; then
        NETWORK=${NETWORK:-default}
        MACHINE_TYPE=${MACHINE_TYPE:-n1-standard-2}

        CIDR_PARAMS=""
        if [ ! -z "${CIDR_RANGE}" ]; then
            CIDR_PARAMS="--cluster-ipv4-cidr ${CIDR_RANGE}"
        fi

        gcloud container clusters create \
            "${CLUSTER_NAME}" \
            --num-nodes ${NODE_COUNT} \
            --cluster-version latest \
            --machine-type ${MACHINE_TYPE} \
            --image-type ubuntu \
            --network ${NETWORK} \
            --subnetwork ${NETWORK} \
            --zone "${GCP_ZONE}" \
            ${CIDR_PARAMS}

        kubectl wait --for condition=available --timeout=300s --all apiservice

        kubectl create -f greenplum-for-kubernetes/concourse/scripts/docker-cleanup-daemonset.yaml
    else
        echo "Skipping creating cluster. ${CLUSTER_NAME} already exists"
        gke_credentials
    fi
}

createPKSCluster() {
    RELEASE_DIR="gp-kubernetes-rc-release/greenplum-for-kubernetes-v$(cat gp-kubernetes-rc-release/version)"

    pks --version
    login_pks

    if ! pks cluster ${CLUSTER_NAME} 2>/dev/null 1>/dev/null ; then
        ${RELEASE_DIR}/workspace/samples/scripts/create_pks_cluster_on_gcp.bash ${CLUSTER_NAME} ${CLUSTER_LOAD_BALANCER} ${NODE_COUNT}

        kubectl create -f greenplum-for-kubernetes/concourse/scripts/docker-cleanup-daemonset.yaml
    else
        echo "Skipping creating cluster. ${CLUSTER_NAME} already exists"
        pks_credentials
    fi
}

_main() {

    auth_gcloud
    gcloud version

    if [ ${KUBEENV} == "GKE" ]; then
        createGKECluster
    else
        createPKSCluster
    fi

    mv ./key.json greenplum-for-kubernetes/greenplum-operator/operator/

    buildDefaultOperatorValuesYml

    if [ ! -z "${CLUSTER_CONFIG}" ]; then
        echo "overide my-gp-instance.yml with"
        echo "${CLUSTER_CONFIG}" > greenplum-for-kubernetes/workspace/my-gp-instance.yaml
    fi

    if [ ! -z "${GP_INSTANCE_NAME}" ]; then
        sed -i "s%  name:.*%  name: ${GP_INSTANCE_NAME}%" greenplum-for-kubernetes/workspace/my-gp-instance.yaml
    fi

    make -C greenplum-for-kubernetes/greenplum-operator deploy-clean
    make -C greenplum-for-kubernetes/greenplum-operator deploy

#    TODO Remove alter commands that workaround bug after resource group bug is fixed
    kubectl exec -it master-0 -- bash -c " source /usr/local/greenplum-db/greenplum_path.sh; psql -c 'alter resource group admin_group SET CPU_RATE_LIMIT ${CPU_RATE_LIMIT}'"
    kubectl exec -it master-0 -- bash -c " source /usr/local/greenplum-db/greenplum_path.sh; psql -c 'alter resource group admin_group SET MEMORY_LIMIT ${MEMORY_LIMIT}'"

}

_main "$@"
