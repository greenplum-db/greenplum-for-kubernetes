#!/usr/bin/env bash

: ${GCP_ZONE:=us-central1-f}

function login_pks() {
    (
        set +x
        : "${PKS_API_URL:=$(<greenplum-for-kubernetes-pcf-manifest/manifest.json jq -r '.pks_api.url')}"
        echo "Logging in to PKS (${PKS_API_URL})..."
        pks login -a "${PKS_API_URL}" -u "${PKS_USER}" -p "${PKS_PASSWORD}" -k
    )
}

function auth_gcloud() {
    (
        set +x
        echo ${GCP_SVC_ACCT_KEY} > key.json
        gcloud auth activate-service-account --key-file=./key.json
        gcloud config set project ${GCP_PROJECT}
        gcloud config set compute/zone ${GCP_ZONE}
    )
}

function gke_credentials() {
    gcloud container clusters get-credentials ${CLUSTER_NAME}
}

function pks_credentials() {
    pks get-credentials ${CLUSTER_NAME}
}
