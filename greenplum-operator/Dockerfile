# syntax = docker/dockerfile:experimental
FROM golang:1.19 as build-in-docker

COPY greenplum-operator/ /greenplum-for-kubernetes/greenplum-operator/
COPY pkg/ /greenplum-for-kubernetes/pkg/
COPY tools/ /greenplum-for-kubernetes/tools/
COPY go.mod go.sum /greenplum-for-kubernetes/

RUN curl -LO "https://dl.k8s.io/release/v1.26.3/bin/linux/amd64/kubectl" \
    && install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

WORKDIR /greenplum-for-kubernetes/greenplum-operator

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd /greenplum-for-kubernetes/greenplum-operator/cmd/greenplumOperator \
    && go build

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
ARG DATE="25 Jan 2023"
ARG LICENSE="https://www.apache.org/licenses/LICENSE-2.0"
ARG GREENPLUM_VERSION=7.00.0

ENV GREENPLUM_VERSION=${GREENPLUM_VERSION}

LABEL name="Greenplum Operator" \
    vendor="PlaidCloud" \
    version="${VERSION}" \
    build_date="${DATE}" \
    license="${LICENSE}" \
    greenplum_version="${GREENPLUM_VERSION}"
