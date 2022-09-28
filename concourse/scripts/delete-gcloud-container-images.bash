#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${SCRIPT_DIR}/cloud_login.bash

auth_gcloud

# Delete 6 days old gcloud container non-release images if they exist
function delete_6_day_old_images {
    IMAGE_REPO=$1
    for digest in $(gcloud container images list-tags \
      ${IMAGE_REPO} --limit=999999 \
      --sort-by=TIMESTAMP \
      --filter="timestamp.datetime<$(date --date="6 days ago" +%F), NOT(tags~'v\d+.\d+.\d+(-(alpha|beta|rc).\d+)?$')" \
      --format='get(digest)');
    do
        gcloud container images delete -q --force-delete-tags "${IMAGE_REPO}@${digest}"
    done
}

# Delete dangling images
function delete_dangling_images {
    IMAGE_REPO=$1
    for digest in $(gcloud container images list-tags \
      ${IMAGE_REPO} --limit=999999 \
      --sort-by=TIMESTAMP \
      --filter="-tags:*" \
      --format='get(digest)');
    do
        gcloud container images delete -q --force-delete-tags "${IMAGE_REPO}@${digest}"
    done
}

IMAGE_REPO=gcr.io/gp-kubernetes/${IMAGE}

delete_6_day_old_images ${IMAGE_REPO}
delete_dangling_images ${IMAGE_REPO}
