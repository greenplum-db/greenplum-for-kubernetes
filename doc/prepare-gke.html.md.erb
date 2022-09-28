---
title: Google Kubernetes Engine (GKE) running on Google Cloud Platform (GCP)
---

This section describes the requirements for using <%=vars.product_name %> with Google Kubernetes Engine (GKE) deployments.

## <a id="softwarereq"></a>Required Software

To deploy <%=vars.product_name %> on Google Kubernetes Engine, you require the following software:

<%=partial 'partials/prerequisites-common' %>

* Google Kubernetes Engine (GKE) Kubernetes 1.16.7

## <a id="clusterreq"></a>Cluster Requirements

When creating the GKE cluster, ensure that you make the following selections on the **Create a Kubernetes cluster** screen of the Google Cloud Platform console:

- For the **Cluster Version** option, select the most recent version of Kubernetes.
- Scale the **Machine Type** option to at least **2 vCPUs / 7.5 GB memory**.
- For the **Node Image** option, you must select **Ubuntu**.  You cannot deploy Greenplum with the **Container-Optimized OS (cos)** image.
- Set the **Size** to 4 or more nodes.
- Set **Automatic node repair** to **Disabled**.

In addition to the above, the <%=vars.product_name %> deployment process requires the ability to map the host system's `/sys/fs/cgroup` directory onto each container's `/sys/fs/cgroup`. Ensure that no kernel security module (for example, AppArmor) uses a profile that disallows mounting `/sys/fs/cgroup`.

## <a id="context"></a>Setting the Kubernetes Context

After creating your GKE cluster, use the `gcloud` utility to login to GCP, and to set your current project and cluster context:

1. Log into GCP:

    ``` bash
    $ gcloud auth login
    ```

2. Set the current project to the project where you will deploy Greenplum:

    ``` bash
    $ gcloud config set project <project-name>
    ```

3. Set the context to the Kubernetes cluster that you created for Greenplum:

    1. Access GCP Console.
    2. Select **Kubernetes Engine > Clusters**.
    3. Click **Connect** next to the cluster that you configured for Greenplum, and copy the connection command.
    4. On your local client machine, paste the command to set the context to your cluster.  For example:

        ``` bash
        $ gcloud container clusters get-credentials <cluster-name> --zone us-central1-a --project <my-project>
        ```

        ``` bash
        Fetching cluster endpoint and auth data.
        kubeconfig entry generated for <cluster-name>.
        ```

## <a id="getkey"></a>Obtaining the Service Account Key

<%=partial 'partials/prerequisites-key-json' %>
