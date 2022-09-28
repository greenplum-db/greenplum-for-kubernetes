#!/usr/bin/env bash

mkdir -p /home/gpadmin/.ssh
ssh-keyscan -H "$GPEXPAND_HOST" >> /home/gpadmin/.ssh/known_hosts
/usr/bin/ssh -i /etc/ssh-key/id_rsa "$GPEXPAND_HOST" /tools/runGpexpand --newPrimarySegmentCount "$NEW_SEG_COUNT"
