# syntax = docker/dockerfile:experimental
FROM golang:1.19 as build-in-docker

COPY greenplum-operator/ /greenplum-for-kubernetes/greenplum-operator/
COPY pkg/ /greenplum-for-kubernetes/pkg/
COPY go.mod go.sum /greenplum-for-kubernetes/

WORKDIR /greenplum-for-kubernetes/greenplum-operator
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd cmd/greenplumOperator && go build

# build greenplum-operator image from here
FROM cloudfoundry/run:base

COPY --from=build-in-docker \
    /greenplum-for-kubernetes/greenplum-operator/cmd/greenplumOperator/greenplumOperator \
    /usr/local/bin/greenplum-operator

RUN set -x && \
    apt-get update && \
    apt-get upgrade -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ARG VERSION=dev
ARG DATE=unknown
ARG LICENSE="https://network.pivotal.io/legal_documents/vmware_eula"
ARG GREENPLUM_VERSION=unknown

ENV GREENPLUM_VERSION=${GREENPLUM_VERSION}

LABEL name="Greenplum Operator" \
    vendor="Pivotal Software" \
    version="${VERSION}" \
    build_date="${DATE}" \
    license="${LICENSE}" \
    greenplum_version="${GREENPLUM_VERSION}"
