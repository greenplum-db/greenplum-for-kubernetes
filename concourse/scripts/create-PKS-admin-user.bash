#!/usr/bin/env bash

set -euo pipefail

apt-get update && \
  apt-get install -y \
    lsb-core \
    ruby-full && \
  gem install cf-uaac

# get the manifest and parse it
# manifest is from gcs bucket
# we upload the manifest file once when we create the PCF env
OPSMAN_URL=$(<greenplum-for-kubernetes-pcf-manifest/manifest.json jq -r '.ops_manager.url')
PKS_URL=$(<greenplum-for-kubernetes-pcf-manifest/manifest.json jq -r '.pks_api.url')
UAA_ADMIN_USER=$(<greenplum-for-kubernetes-pcf-manifest/manifest.json jq -r '.ops_manager.username')
UAA_ADMIN_PASSWORD=$(<greenplum-for-kubernetes-pcf-manifest/manifest.json jq -r '.ops_manager.password')
# Authenticate to opsman

uaac target ${OPSMAN_URL}/uaa --skip-ssl-validation
uaac token owner get opsman ${UAA_ADMIN_USER} -p ${UAA_ADMIN_PASSWORD} -s "" # These are in the manifest
TOKEN=$(yq r ~/.uaac.yml \"${OPSMAN_URL}/uaa\".contexts.pivotalcf.access_token) # in pivotalcf context we have access token

# Look for installed products of type pivotal-container-service via API
PKS_GUID=$(curl ${OPSMAN_URL}/api/v0/deployed/products -k -H "Authorization: Bearer ${TOKEN}" \
| jq -r '.[] | select(.type == "pivotal-container-service") | .guid')

ADMIN_CLIENT_SECRET=$(curl ${OPSMAN_URL}/api/v0/deployed/products/${PKS_GUID}/credentials/.properties.pks_uaa_management_admin_client \
-k -H "Authorization: Bearer ${TOKEN}" | jq -r '.credential.value.secret')

uaac target ${PKS_URL}:8443 --skip-ssl-validation
uaac token client get admin -s ${ADMIN_CLIENT_SECRET} # Pks Uaa Management Admin Client

uaac user add gpkubernetes --emails gpkubernetes@pivotal.io -p ${PKS_API_USER_PASSWORD}  # This comes from concourse as a secret
uaac member add pks.clusters.admin gpkubernetes
