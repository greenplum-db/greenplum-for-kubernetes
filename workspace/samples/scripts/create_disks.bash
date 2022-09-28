#!/bin/bash

set -ux

if [ $# -lt 2 ]; then
    set +x
    echo "usage: $0 <prefix> <segment_count>"
    exit 1
fi

PREFIX=$1
SEGMENT_COUNT=$2

gcloud compute disks create --size=20GB master-0-${PREFIX}-disk
gcloud compute disks create --size=20GB master-1-${PREFIX}-disk

SEGMENT_LIMIT=$(($SEGMENT_COUNT-1))
for i in `seq 0 ${SEGMENT_LIMIT}`;
do
    gcloud compute disks create --size=20GB segment-a-${i}-${PREFIX}-disk
    gcloud compute disks create --size=20GB segment-b-${i}-${PREFIX}-disk
done

