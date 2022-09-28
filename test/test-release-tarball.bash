#!/usr/bin/env bash

verify_imageTag_in_Values_Common() {
    TARBALL_ABSOLUTE_PATH=$1
    mkdir -p /tmp/validate-tarball
    pushd /tmp/validate-tarball
        tar xf ${TARBALL_ABSOLUTE_PATH} ${VERSIONED_PKG}/operator/values.yaml
        cat ${VERSIONED_PKG}/operator/values.yaml | grep "operatorImageTag: ${IMAGE_VERSION}"
        cat ${VERSIONED_PKG}/operator/values.yaml | grep "greenplumImageTag: ${IMAGE_VERSION}"
    popd
    rm -rf /tmp/validate-tarball
}

verify_imageTag_in_docker_image() {
    TARBALL_ABSOLUTE_PATH=$1
    local extracted_dir=$(mktemp -d ./extracted_dir.XXXXXX)
    pushd ${extracted_dir}
        tar xf ${TARBALL_ABSOLUTE_PATH}
        tar xOf ${VERSIONED_PKG}/images/greenplum-for-kubernetes manifest.json \
            | grep -q "\"RepoTags\":\[\"greenplum-for-kubernetes:${IMAGE_VERSION}\"\]"
        tar xOf ${VERSIONED_PKG}/images/greenplum-operator manifest.json \
            | grep -q "\"RepoTags\":\[\"greenplum-operator:${IMAGE_VERSION}\"\]"
    popd
    rm -rf ${extracted_dir}
}

verify_appVersion_in_helm_chart() {
    TARBALL_ABSOLUTE_PATH=$1
    local extracted_dir=$(mktemp -d ./extracted_dir.XXXXXX)
    pushd ${extracted_dir}
        tar xf ${TARBALL_ABSOLUTE_PATH}
        cat ${VERSIONED_PKG}/operator/Chart.yaml \
            | grep -q "appVersion: ${IMAGE_VERSION}"
    popd
    rm -rf ${extracted_dir}
}

# NOTE: sorted by alphabet for easy lookup
file_list() {
    sort <<_EOF
${VERSIONED_PKG}/
${VERSIONED_PKG}/README.txt
${VERSIONED_PKG}/images/
${VERSIONED_PKG}/images/greenplum-for-kubernetes
${VERSIONED_PKG}/images/greenplum-for-kubernetes-id
${VERSIONED_PKG}/images/greenplum-for-kubernetes-tag
${VERSIONED_PKG}/images/greenplum-operator
${VERSIONED_PKG}/images/greenplum-operator-id
${VERSIONED_PKG}/images/greenplum-operator-tag
${VERSIONED_PKG}/operator/
${VERSIONED_PKG}/operator/Chart.yaml
${VERSIONED_PKG}/operator/templates/
${VERSIONED_PKG}/operator/templates/NOTES.txt
${VERSIONED_PKG}/operator/templates/greenplum-operator-cluster-role.yaml
${VERSIONED_PKG}/operator/templates/greenplum-operator-cluster-role-binding.yaml
${VERSIONED_PKG}/operator/templates/greenplum-operator-crds.yaml
${VERSIONED_PKG}/operator/templates/greenplum-operator-service-account.yaml
${VERSIONED_PKG}/operator/templates/greenplum-operator.yaml
${VERSIONED_PKG}/operator/values.yaml
${VERSIONED_PKG}/workspace/
${VERSIONED_PKG}/workspace/my-gp-instance.yaml
${VERSIONED_PKG}/workspace/samples/
${VERSIONED_PKG}/workspace/samples/my-gp-with-pxf-instance.yaml
${VERSIONED_PKG}/workspace/samples/openpolicyagent/
${VERSIONED_PKG}/workspace/samples/openpolicyagent/README.md
${VERSIONED_PKG}/workspace/samples/openpolicyagent/aws-load-balancer.rego
${VERSIONED_PKG}/workspace/samples/openpolicyagent/gpdb-config-opa.yaml
${VERSIONED_PKG}/workspace/samples/scripts/
${VERSIONED_PKG}/workspace/samples/scripts/create_disks.bash
${VERSIONED_PKG}/workspace/samples/scripts/create_pks_cluster_on_gcp.bash
${VERSIONED_PKG}/workspace/samples/scripts/regsecret-test.bash
${VERSIONED_PKG}/workspace/samples/scripts/regsecret-test.yaml
_EOF
}

main() {
    set -exuo pipefail

    SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
    VERSION=$(${SCRIPT_DIR}/../getversion)
    IMAGE_VERSION=${TAG_PREFIX:-}$(${SCRIPT_DIR}/../getversion --image)
    VERSIONED_PKG=greenplum-for-kubernetes-${VERSION}
    RELEASE_TARBALL=/tmp/greenplum-instance_release/${VERSIONED_PKG}.tar.gz

    verify_imageTag_in_Values_Common "${RELEASE_TARBALL}"
    verify_imageTag_in_docker_image "${RELEASE_TARBALL}"
    verify_appVersion_in_helm_chart "${RELEASE_TARBALL}"

    diff -u <(file_list) <(tar tf "${RELEASE_TARBALL}" | sort)
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
