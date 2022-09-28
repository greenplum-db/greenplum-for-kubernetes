#!/usr/bin/env bash
set -euxo pipefail

SCRIPT_DIR=`dirname $0`
source ${SCRIPT_DIR}/cloud_login.bash

TAG_PREFIX=${PIPELINE_NAME}${PIPELINE_NAME:+-}
GREENPLUM_INSTANCE_IMG_TAG="greenplum-for-kubernetes:${TAG_PREFIX}$(greenplum-for-kubernetes/getversion --image)"
GREENPLUM_OPERATOR_IMG_TAG="greenplum-operator:${TAG_PREFIX}$(greenplum-for-kubernetes/getversion --image)"
echo "Building ${GREENPLUM_INSTANCE_IMG_TAG} and ${GREENPLUM_OPERATOR_IMG_TAG} ..."

####
# load script for start_docker
####
. docker-in-concourse/dind.bash
####
# start dockerd
####
max_concurrent_downloads=4
max_concurrent_uploads=4
start_docker ${max_concurrent_downloads} ${max_concurrent_uploads} "" ""

(
    set +x
    echo "Logging in to gcr.io"
    echo "${GCP_SVC_ACCT_KEY}" | docker login -u _json_key --password-stdin https://gcr.io
)

mkdir -p /.ssh
auth_gcloud

####
# make in docker
####
# GREENPLUM_VERSION determined from resource provided by concourse
# It can be 5.13.0+dev.16.gb10dbec for dev builds
# Or 5.13.0 for release
export GREENPLUM_VERSION=$(dpkg-deb -f greenplum-debian-binary/greenplum-db*.deb Version | sed 's/+/./g')
cd greenplum-for-kubernetes
make docker-check TAG_PREFIX="$TAG_PREFIX"

# Check that the version of GPDB in the image is not a dev release
# NB: We do this here, rather than in `make docker-check` so that developers can
# still use `make docker-check` to build & test images with development GPDB builds.
if ${DEV_BUILD}; then
    gpdb_version_label=$(docker image inspect ${GREENPLUM_INSTANCE_IMG_TAG} --format '{{.Config.Labels.GREENPLUM_VERSION}}')
    if echo "$gpdb_version_label" | grep -q dev ; then
        echo "Refusing to push: Image contains a development build of GPDB: $gpdb_version_label"
        exit 1
    fi
fi

docker tag ${GREENPLUM_INSTANCE_IMG_TAG} gcr.io/gp-kubernetes/${GREENPLUM_INSTANCE_IMG_TAG}
docker push gcr.io/gp-kubernetes/${GREENPLUM_INSTANCE_IMG_TAG}

docker tag ${GREENPLUM_OPERATOR_IMG_TAG} gcr.io/gp-kubernetes/${GREENPLUM_OPERATOR_IMG_TAG}
docker push gcr.io/gp-kubernetes/${GREENPLUM_OPERATOR_IMG_TAG}

make release TAG_PREFIX="$TAG_PREFIX"
make release-check TAG_PREFIX="$TAG_PREFIX"

cd ..
cp /tmp/greenplum-instance_release/greenplum-for-kubernetes-v*.tar.gz gp-kubernetes-rc-release-output/
cp /tmp/greenplum-instance_release/greenplum-for-kubernetes-v*-receipt.txt gp-kubernetes-rc-release-receipt-output/
