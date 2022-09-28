---
title: VMware Tanzu Kubernetes Grid Integrated (TKGI) Edition running on Google Cloud Platform (GCP)
---

Follow this procedure to deploy <%=vars.product_name %> to VMware Tanzu Kubernetes Grid Integrated (TKGI) Edition.

## <a id="softwarereq"></a>Required Software

To deploy <%=vars.product_name %> on VMware Tanzu Kubernetes Grid Integrated Edition, you require the following software:

<%=partial 'partials/prerequisites-common' %>

* VMware Tanzu Kubernetes Grid Integrated (TKGI) Edition 1.7.0 (contains Kubernetes 1.16.7)


## <a id="softwarereq"></a>Cluster Requirements

This procedure requires that VMware Tanzu Kubernetes Grid Integrated (TKGI) Edition is installed and running, along with all prerequisite
software and configuration. See [Installing VMware Tanzu Kubernetes Grid Integrated Edition for <%=vars.product_name %>, Using Google Cloud Platform (GCP)](prerequisites.html) for information.

In the TKGI tile under **Kubernetes Cloud Provider**, ensure that the service accounts that are listed under **GCP Master Service Account ID** and **GCP Worker Service Account ID** have permission to pull images from the GCS bucket named `artifacts.<project-name>.appspot.com`.

## <a id="getkey"></a>Obtaining the Service Account Key

<%=partial 'partials/prerequisites-key-json' %>

Before you attempt to deploy Greenplum, ensure that the target cluster is available. Execute the following command to make sure that the target cluster displays in the output:

```bash
pks list-clusters
```

**Note:** The `pks` login cookie typically expires after a day or two.

The <%=vars.product_name %> deployment process requires the ability to map the host system's `/sys/fs/cgroup` directory onto each container's `/sys/fs/cgroup`. Ensure that no kernel security module (for example, AppArmor) uses a profile that disallows mounting `/sys/fs/cgroup`.

To use pre-created disks with TKGI instead of (default) automatically-managed persistent volumes, follow the instructions in [(Optional) Preparing Pre-Created Disks](#pre_created_disks) before continuing with the procedure.

**Note:** If any problems occur during deployment, retry deploying Greenplum by first removing the previous deployment.

<%=vars.product_name %> requires the following recommended settings:

1. Ability to increase the open file limit on the TKGI cluster nodes

In order to accomplish this, do the following on each node using bosh:

```bash

#!/usr/bin/env bash

# fail fast if bosh isn't working
which bosh

deployment=$1
bosh deployment -d ${deployment}

# change limits file
bosh ssh master -d ${deployment} -c "sudo sed -i 's/#DefaultLimitNOFILE=/DefaultLimitNOFILE=65535/' /etc/systemd/system.conf"

# change running processes by getting pid for both monit and the kube-apiserver processes
bosh ssh master -d ${deployment} -c 'sudo bash -c "prlimit --pid $(pgrep monit) --nofile=65535:65535"'
bosh ssh master -d ${deployment} -c 'sudo bash -c "prlimit --pid $(pgrep kube-apiserver) --nofile=65535:65535"'
# repeat above for other nodes besides master in the deployment
```
