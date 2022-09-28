#!/usr/bin/env bash

# idempotent if individual creation lines are allowed to fail if already existing
set -euxo pipefail

# Roles assigned to policy bindings
NETWORK_ADMIN="roles/compute.networkAdmin"
COMPUTE_VIEWER="roles/compute.viewer"
STORAGE_OBJECT_CREATOR="roles/storage.objectCreator"
CLOUD_BUILDS_BUILDER="roles/cloudbuild.builds.builder"
KUBERNETES_ENGINE_ADMIN="roles/container.admin"
KUBERNETES_ENGINE_CLUSTER_ADMIN="roles/container.clusterAdmin"
KUBERNETES_ENGINE_DEVELOPER="roles/container.developer"
SERVICE_ACCOUNT_USER="roles/iam.serviceAccountUser"
COMPUTE_INSTANCE_ADMIN="roles/compute.instanceAdmin.v1"
DNS_ADMIN="roles/dns.admin"

#GCloud project details
ACCOUNT_NAME="k8s-gcr-auth-ro"
PROJECT_NAME=$(gcloud config list core/project --format='value(core.project)')
SA_EMAIL="$ACCOUNT_NAME@$PROJECT_NAME.iam.gserviceaccount.com"

function remove_policy_binding {
    local role=$1
    gcloud projects remove-iam-policy-binding ${PROJECT_NAME} --member serviceAccount:${SA_EMAIL} --role "${role}" > /dev/null
}

function remove_all_policy_bindings {
    local jq_filter=".bindings[] | select(.members[] == \"serviceAccount:${SA_EMAIL}\") | .role"
    local roles=(
        $(gcloud projects get-iam-policy gp-kubernetes --format json | jq -r "${jq_filter}")
    )
    for role in "${roles[@]}" ; do
        remove_policy_binding "$role"
    done
}

function add_policy_binding {
    local role=$1
    gcloud projects add-iam-policy-binding ${PROJECT_NAME} --member serviceAccount:${SA_EMAIL} --role "${role}" > /dev/null
}

function delete_service_account_if_exists {
    if gcloud iam service-accounts describe ${SA_EMAIL} 2>/dev/null ; then
        gcloud --quiet iam service-accounts delete ${SA_EMAIL}
    fi
}

function create_service_account {
    SA_EMAIL=$(gcloud iam service-accounts --format='value(email)' create ${ACCOUNT_NAME} --display-name "${PROJECT_NAME}-ci")
}

KEY_JSON_DIR=${1:-"/tmp"}
remove_all_policy_bindings
delete_service_account_if_exists
# create a GCP service account; format of account is email address
create_service_account

# create the json key file and associate it with the service account
gcloud iam service-accounts keys create ${KEY_JSON_DIR}/key.json --iam-account=${SA_EMAIL}
echo "Location of key.json: ${KEY_JSON_DIR}/key.json"

add_policy_binding ${NETWORK_ADMIN}
add_policy_binding ${COMPUTE_VIEWER}
add_policy_binding ${STORAGE_OBJECT_CREATOR}
add_policy_binding ${CLOUD_BUILDS_BUILDER}
add_policy_binding ${COMPUTE_INSTANCE_ADMIN}
add_policy_binding ${SERVICE_ACCOUNT_USER}
add_policy_binding ${KUBERNETES_ENGINE_DEVELOPER}
add_policy_binding ${KUBERNETES_ENGINE_CLUSTER_ADMIN}
add_policy_binding ${KUBERNETES_ENGINE_ADMIN}
add_policy_binding ${DNS_ADMIN}

# required for GCR container registry
gsutil iam ch serviceAccount:${SA_EMAIL}:objectViewer gs://artifacts.${PROJECT_NAME}.appspot.com

echo "created json key file at ./key.json ; please copy this into pks secrets file if you are replacing it."
