#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source ${SCRIPT_DIR}/cloud_login.bash

: ${GCP_PROJECT:="gp-kubernetes"}
auth_gcloud

# Variant that echos first
gcloud_e() {
  echo gcloud "$@"
  command gcloud "$@"
}

# Run gcloud queries and build the JSON database
run_queries() {
  # All TCP forwarding-rules created over a month ago
  fwdrules=$(
    gcloud compute forwarding-rules list \
      --format 'json(name, target, region.scope())' \
      --filter 'creationTimestamp < -P1W AND IPProtocol = "TCP" AND region != null AND portRange = "5432-5432"'
  )

  targetLinks=($(jq -r '.[].target' <<<"$fwdrules"))

  # Make map of the target pools by their selfLink
  # TODO: this is slow b/c of the potentially huge list in the filter
  targetPools=$(
    gcloud compute target-pools list \
      --format 'json(selfLink, name, region.scope(), instances, healthChecks.map().scope())' \
      --filter "selfLink =( ${targetLinks[*]} )" |
      jq 'map({(.selfLink): .}) | add'
  )

  # Join fwdrules and targetPools, and select those with zero instances in the target pool.
  fwdrules=$(
    jq '[ .[] | (.target = $targetPools[.target]) | select(.target.instances | length == 0) ]' \
      --argjson targetPools "$targetPools" <<<"$fwdrules"
  )

  # Unique regions of forwarding rules and target pools
  regions=($(jq -r 'map(.region, .target.region) | unique | .[]' <<<"$fwdrules"))
}

delete_resources() {
  for region in "${regions[@]}"; do
    echo "deleting resources in region: ${region}"
    deleteFwdrules=($(jq -r ".[] | select(.region == \"$region\") | .name" <<<"$fwdrules"))
    deleteTargetpools=($(jq -r ".[] | select(.region == \"$region\") | .target.name" <<<"$fwdrules"))
    deleteHealthchecks=($(jq -r ".[] | select(.region == \"$region\") | .target.healthChecks[] // empty" <<<"$fwdrules"))
    # TODO: Is there a way to determine if a firewall-rule is unused?
    #   Or, a better way to match a firewall-rule to a load-balancer other than the name coincidence?
    #   Also, nothing is done here to determine if they actually exist.
    deleteFirewallRules=($(jq -r ".[] | select(.region == \"$region\") | .name | select(test(\"a[0-9a-f]{31}\")) | \"k8s-fw-\"+. " <<<"$fwdrules"))

    # Continue even if some deletions fail (|| true).
    gcloud_e compute forwarding-rules delete -q --region=$region "${deleteFwdrules[@]}" || true
    gcloud_e compute target-pools delete -q --region=$region "${deleteTargetpools[@]}" || true
    gcloud_e compute health-checks delete -q "${deleteHealthchecks[@]}" || true
    gcloud_e compute firewall-rules delete -q "${deleteFirewallRules[@]}" || true
    cat <<_EOF
Deleted in region ${region}:
    ${#deleteFwdrules[@]} forwarding-rules,
    ${#deleteTargetpools[@]} target-pools,
    ${#deleteHealthchecks[@]} health-checks,
    and up to ${#deleteFirewallRules[@]} firewall-rules
_EOF
  done
}

delete_singlenode_firewall_rules() {
  # list and delete singlenode firewall rules
  firewallRules=($(gcloud compute firewall-rules list \
    --flatten="targetTags[]" \
    --format 'value(name)' \
    --filter "name ~ k8s-* AND targetTags: (gke-singlenode-*)"))

  echo "Deleting singlenode firewall rules..."
  gcloud_e compute firewall-rules delete -q "${firewallRules[@]}" || true
  echo "Deleted ${#firewallRules[@]} firewall-rules"
}

run_queries
delete_resources
delete_singlenode_firewall_rules
