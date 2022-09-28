#!/usr/bin/env bash

set -x

GREENPLUM_FOR_KUBERNETES_VERSION=$(./greenplum-for-kubernetes/getversion)

# concourse needs separate directories for input and output
git clone greenplum-for-kubernetes-receipts greenplum-for-kubernetes-receipts-output
cp gp-kubernetes-rc-release-receipts/greenplum-for-kubernetes-v*-receipt.txt greenplum-for-kubernetes-receipts-output/greenplum-for-kubernetes-receipt.txt
cd greenplum-for-kubernetes-receipts-output
git add .
git config user.email "greenplum-for-kubernetes-bot@example.com"
git config user.name "GREENPLUM_FOR_KUBERNETES_BOT"
git commit -m "Committing receipt version ${GREENPLUM_FOR_KUBERNETES_VERSION}"
git tag "${GREENPLUM_FOR_KUBERNETES_VERSION}"
