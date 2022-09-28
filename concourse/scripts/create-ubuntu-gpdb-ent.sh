#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR=`dirname $0`
source ${SCRIPT_DIR}/cloud_login.bash
# GREENPLUM_VERSION determined from resource provided by concourse
# It can be 5.13.0+dev.16.gb10dbec for dev builds
# Or 5.13.0 for release
TAG_PREFIX=${PIPELINE_NAME}${PIPELINE_NAME:+-}
GPDB_VERSION=$(dpkg-deb -f greenplum-debian-binary/greenplum-db*.deb Version)
GPDB_TAG=$(echo ${GPDB_VERSION} | sed 's/+/\./g')

function start_docker_and_login() {
    . ./docker-in-concourse/dind.bash
    max_concurrent_downloads=4
    max_concurrent_uploads=4
    start_docker ${max_concurrent_downloads} ${max_concurrent_uploads} "" ""
    (
      set +x
      echo "Logging in to gcr.io"
      echo "${GCP_SVC_ACCT_KEY}" | docker login -u _json_key --password-stdin https://gcr.io
    )
}

function build_ubuntu_gpdb_ent_image() {
    make -C greenplum-for-kubernetes/greenplum-instance/ubuntu-gpdb-ent docker-check
}

function push_ubuntu_gpdb_ent_image() {
    docker tag "gcr.io/gp-kubernetes/ubuntu-gpdb-ent:latest" "gcr.io/gp-kubernetes/ubuntu-gpdb-ent:${TAG_PREFIX}${GPDB_TAG}"
    docker tag "gcr.io/gp-kubernetes/ubuntu-gpdb-ent:latest" "gcr.io/gp-kubernetes/ubuntu-gpdb-ent:${TAG_PREFIX}latest"
    docker push "gcr.io/gp-kubernetes/ubuntu-gpdb-ent:${TAG_PREFIX}${GPDB_TAG}"
    docker push "gcr.io/gp-kubernetes/ubuntu-gpdb-ent:${TAG_PREFIX}latest"
}

start_docker_and_login
auth_gcloud
build_ubuntu_gpdb_ent_image
push_ubuntu_gpdb_ent_image
