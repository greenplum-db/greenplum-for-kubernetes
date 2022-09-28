#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${SCRIPT_DIR}/cloud_login.bash

auth_gcloud

# Delete unused old pvcs & disks

pvc_json=$(gcloud compute disks list --filter="name:(pvc) AND NOT name:($(gcloud container clusters list --format="value(name)" | cut -c1-10))" --format=json)
disk_json=$(gcloud compute disks list --filter="name:(disk)" --format=json)

echo "$pvc_json" "$disk_json" | jq -r '.[] | ("gcloud compute disks delete " + .name + " --zone=" + .zone + " --quiet || true")' | while read -r command
do
    eval "${command}"
done
