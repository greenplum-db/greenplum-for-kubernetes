#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
VERSION=$(${SCRIPT_DIR}/../getversion)
: ${RELEASE_DIR:="/tmp/greenplum-instance_release/greenplum-for-kubernetes-${VERSION}"}
TAG=$(cat ${RELEASE_DIR}/images/greenplum-for-kubernetes-tag)

# load images from RELEASE_DIR
for img in greenplum-for-kubernetes greenplum-operator ; do
    docker load -i ${RELEASE_DIR}/images/${img}
    docker tag ${img}:${TAG} ${img}:latest
    docker images | grep ${img} | grep ${TAG}
done
