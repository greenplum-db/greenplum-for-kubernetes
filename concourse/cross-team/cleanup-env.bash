#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source ${SCRIPT_DIR}/../scripts/cloud_login.bash
auth_gcloud

# TODO: come back and remove KUBEENV when conversion is finished
if [ ${KUBEENV} == "GKE" ]; then
    echo "logging into GKE..."
    gcloud version
    gke_credentials
elif [ ${KUBEENV} == "PKS" ]; then
    echo "logging into PKS..."
    pks --version
    login_pks
    pks_credentials
else
    echo "Not supported kube environment. Bailing..."
    exit 1
fi

kubectl get greenplumclusters.greenplum.pivotal.io ${GP_INSTANCE_NAME} -o yaml > ${SCRIPT_DIR}/../../workspace/my-gp-instance.yaml

echo "Cleaning Greenplum Cluster K8s resources, including ssets, pods, crd, secretes, etc..."
make -C ${SCRIPT_DIR}/../../greenplum-operator deploy-clean

echo "Deleting the Kubernetes cluster..."
${SCRIPT_DIR}/../scripts/delete-k8s-cluster.bash

echo "Deleting any remaining cloud resources, including disks, firewall, forwarding-rules and target-pools..."
set +e
    DISKS="$(gcloud compute disks list --format='value(name)' --filter=${CLUSTER_NAME})"
    if [[ ! -z ${DISKS} ]] ; then
        gcloud compute "--project=${GCP_PROJECT}" -q disks delete ${DISKS} || true
    fi

    FIREWALL_RULES="$(gcloud compute firewall-rules list --format='value(name)' --filter=${CLUSTER_NAME})"
    if [[ ! -z ${FIREWALL_RULES} ]] ; then
        gcloud compute "--project=${GCP_PROJECT}" -q firewall-rules delete ${FIREWALL_RULES} || true
    fi

    FORWARDING_RULES="$(gcloud compute forwarding-rules list --format='value(name)' --filter=${CLUSTER_NAME})"
    if [[ ! -z ${FORWARDING_RULES} ]] ; then
        gcloud compute "--project=${GCP_PROJECT}" -q forwarding-rules delete ${FORWARDING_RULES} || true
    fi

    TARGET_POOLS="$(gcloud compute target-pools list --format='value(name)' --filter=${CLUSTER_NAME})"
    if [[ ! -z ${TARGET_POOLS} ]] ; then
        gcloud compute "--project=${GCP_PROJECT}" -q target-pools delete ${TARGET_POOLS} || true
    fi
set -e
