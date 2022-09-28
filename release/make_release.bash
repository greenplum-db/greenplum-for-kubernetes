#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="${SCRIPT_DIR}/.."

RELEASE_DIR="/tmp/greenplum-instance_release"
GREENPLUM_OPERATOR_IMAGE_NAME="greenplum-operator"
GREENPLUM_OPERATOR_IMAGE_REPO_PATH="greenplum-operator"
GREENPLUM_IMAGE_REPO_PATH="greenplum-instance-image"
GREENPLUM_IMAGE_NAME="greenplum-for-kubernetes"
GREENPLUM_IMAGE="${GREENPLUM_IMAGE_NAME}"
GREENPLUM_OPERATOR_IMAGE="${GREENPLUM_OPERATOR_IMAGE_NAME}"

function save_image() {
    REPO_PATH=$1
    IMAGE_NAME_AND_TAG=$2

    mkdir -p ${REPO_PATH}
    docker save -o ${REPO_PATH}/image ${IMAGE_NAME_AND_TAG}
    # TODO: We can use real digests here when https://github.com/docker/cli/issues/728 is fixed
    echo $(docker images --format {{.ID}} ${IMAGE_NAME_AND_TAG})> ${REPO_PATH}/image-id

}

# image manipulation here is for local script debugging only
function validate_and_save_local_images() {
    echo "Checking if docker images exist. "
    if [ -z "$(docker images -q ${GREENPLUM_OPERATOR_IMAGE}:${IMAGE_VERSION_TAG})" ] ; then
        echo "Image Check Failed: Run 'make docker' in repo: greenplum-for-kubernetes"
        exit 1
    fi
    if [ -z "$(docker images -q ${GREENPLUM_IMAGE}:${IMAGE_VERSION_TAG})" ] ; then
        echo "Image Check Failed: Run 'make docker' in repo: greenplum-for-kubernetes"
        exit 1
    fi

    RELEASE_SCRATCH_DIR=${RELEASE_DIR}/build
    GREENPLUM_OPERATOR_IMAGE_REPO_PATH="${RELEASE_SCRATCH_DIR}/${GREENPLUM_OPERATOR_IMAGE_NAME}"
    GREENPLUM_IMAGE_REPO_PATH="${RELEASE_SCRATCH_DIR}/${GREENPLUM_IMAGE_NAME}"
    pushd  ${RELEASE_DIR}
        save_image ${GREENPLUM_OPERATOR_IMAGE_REPO_PATH} ${GREENPLUM_OPERATOR_IMAGE}:${IMAGE_VERSION_TAG}
        save_image ${GREENPLUM_IMAGE_REPO_PATH} ${GREENPLUM_IMAGE}:${IMAGE_VERSION_TAG}
    popd
}

function copy_files() {
    cp -r ${PROJECT_DIR}/workspace ${RELEASE_ROOT_DIR}/

    install -d ${RELEASE_ROOT_DIR}/images
    cp ${GREENPLUM_IMAGE_REPO_PATH}/image ${RELEASE_ROOT_DIR}/images/${GREENPLUM_IMAGE_NAME}
    cp ${GREENPLUM_IMAGE_REPO_PATH}/image-id ${RELEASE_ROOT_DIR}/images/${GREENPLUM_IMAGE_NAME}-id
    echo ${IMAGE_VERSION_TAG} > ${RELEASE_ROOT_DIR}/images/${GREENPLUM_IMAGE_NAME}-tag

    cp ${GREENPLUM_OPERATOR_IMAGE_REPO_PATH}/image ${RELEASE_ROOT_DIR}/images/${GREENPLUM_OPERATOR_IMAGE_NAME}
    cp ${GREENPLUM_OPERATOR_IMAGE_REPO_PATH}/image-id ${RELEASE_ROOT_DIR}/images/${GREENPLUM_OPERATOR_IMAGE_NAME}-id
    echo ${IMAGE_VERSION_TAG} > ${RELEASE_ROOT_DIR}/images/${GREENPLUM_OPERATOR_IMAGE_NAME}-tag

    rsync -ar ${PROJECT_DIR}/greenplum-operator/operator ${RELEASE_ROOT_DIR}/ \
        --exclude .gitignore \
        --exclude key.json
}

function add_readme() {
    cat << EOF > ${RELEASE_ROOT_DIR}/README.txt
Please see documentation at:
http://greenplum-kubernetes.docs.pivotal.io/
EOF
}

function update_version_references() {
    ${SED} -i \
        -e "s/operatorImageTag:.*$/operatorImageTag: ${IMAGE_VERSION_TAG}/" \
        -e "s/greenplumImageTag:.*$/greenplumImageTag: ${IMAGE_VERSION_TAG}/" \
        ${RELEASE_ROOT_DIR}/operator/values.yaml

    ${SED} -i \
        -e "s/appVersion:.*$/appVersion: ${IMAGE_VERSION_TAG}/" \
        ${RELEASE_ROOT_DIR}/operator/Chart.yaml
}

function create_package() {
    local tarball="${RELEASE_DIR}/greenplum-for-kubernetes-${RELEASE_VERSION}.tar.gz"
    tar czf "$tarball" -C "${RELEASE_DIR}" $(basename ${RELEASE_ROOT_DIR})
    echo "Location of tarball is: ${tarball}"
}

function create_receipt() {
    local tarball="${RELEASE_DIR}/greenplum-for-kubernetes-${RELEASE_VERSION}.tar.gz"
    local dpkg_list=(dpkg-query -W -f '${db:Status-Abbrev;-3} ${binary:Package;-40} ${Version;-50} ${Architecture;-8} ${binary:Summary}\n')
    docker run --rm ${GREENPLUM_OPERATOR_IMAGE}:${IMAGE_VERSION_TAG} "${dpkg_list[@]}" | tee "${RELEASE_DIR}/packages-list-${GREENPLUM_OPERATOR_IMAGE_NAME}"
    # we remove greenplum-db and madlib from dpgk list because they won't scan properly
    docker run --rm ${GREENPLUM_IMAGE}:${IMAGE_VERSION_TAG} "${dpkg_list[@]}" | grep -v -e "greenplum-db" -e "madlib" | tee "${RELEASE_DIR}/packages-list-${GREENPLUM_IMAGE_NAME}"

    # create a receipt for the release
    echo "Release SHASUM: $(${SHASUM} "$tarball" | cut -d' ' -f1)" > "${RELEASE_DIR}/greenplum-for-kubernetes-${RELEASE_VERSION}-receipt.txt"
    echo "" >> "${RELEASE_DIR}/greenplum-for-kubernetes-${RELEASE_VERSION}-receipt.txt"
    sort -u "${RELEASE_DIR}/packages-list-${GREENPLUM_OPERATOR_IMAGE_NAME}" \
         "${RELEASE_DIR}/packages-list-${GREENPLUM_IMAGE_NAME}" >> "${RELEASE_DIR}/greenplum-for-kubernetes-${RELEASE_VERSION}-receipt.txt"
}

function main() {
    set -euxo pipefail

    if [ -n "${PKS_RELEASE:-}" ]; then
        RELEASE_DIR=$(pwd)/${PKS_RELEASE}
    fi

    RELEASE_VERSION=$(${PROJECT_DIR}/getversion)
    IMAGE_VERSION_TAG=${TAG_PREFIX:-}$(${PROJECT_DIR}/getversion --image)

    # pick proper sed/gsed based on OS
    case "$(uname -s)" in
        Darwin*)
            SED="/usr/local/bin/gsed"
            SHASUM="/usr/bin/shasum -a 256";;
        *)
            SED=$(which sed)
            SHASUM=$(which sha256sum);;
    esac

    if [ ! -f ${SED} ]; then
        echo "Please make sure ${SED} exists."
        echo "on MacOS, please run 'brew install gnu-sed'."
        exit 1
    fi

    rm -rf ${RELEASE_DIR}/*
    RELEASE_ROOT_DIR=${RELEASE_DIR}/greenplum-for-kubernetes-${RELEASE_VERSION}
    mkdir -p ${RELEASE_ROOT_DIR}

    validate_and_save_local_images

    copy_files
    add_readme
    update_version_references
    create_package
    create_receipt
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
