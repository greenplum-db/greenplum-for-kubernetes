<% set_title("Installing VMware Tanzu Kubernetes Grid Integrated (TKGI) Edition for ", product_name_long, "(GCP)") %>

This release of <%=vars.product_name_long %> can be deployed with TKGI running on Google Cloud Platform (GCP). Management scripts are provided to help you configure required GCP and Kubernetes resources, as well as to deploy a <%=vars.product_name %> cluster.

Follow each of the steps in this section to install required software and configure resources **before** you attempt to deploy <%=vars.product_name %>.

## <a id='pks_install'></a>Step 1: Install VMware Tanzu Kubernetes Grid Integrated (TKGI) Edition

Follow the instructions in [Google Cloud Platform (GCP)](https://docs.vmware.com/en/VMware-Enterprise-PKS/1.7/vmware-enterprise-pks-17/GUID-gcp-index.html) in the VMware Tanzu Kubernetes Grid Integrated (TKGI) Edition documentation. Completing these instructions creates a new TKGI installation including an Operations Manager URL, which will be referenced in later steps as <OPSMANAGER_IP_ADDRESS_OR_URL>.

## <a id='client_tool_install'></a>Step 2: Install Client Tools

<%=vars.product_name_long %> provides a number of scripts to automate configuration tasks in Google Cloud Platform, TKGI, and Kubernetes. These scripts require your client system to have a number of tools installed. Follow these steps:

1. On MacOS systems, use `homebrew` to install required client system tools:

    ```bash
    $ brew install jq yq kubernetes-helm kubectl
    $ brew cask install docker
    ```
    
    The above command installs several tools that are required for for the <%=vars.product_name %> scripts:
    - `jq` - for parsing json on the command line
    - `yq` - for parsing yaml on the command line
    - `helm` - the CLI for the [Helm](https://helm.sh/) package manager for Kubernetes
    - `kubectl` - the CLI for deploying and managing applciations on Kubernetes
    - `docker` - software container platform and related tools

2. Download and install the [Google cloud tools](https://cloud.google.com/sdk/). Ensure that the `gcloud` utility is available before you continue.

3. Install the `PKS CLI` tool from the VMware Tanzu Network using the instructions at [Installing the PKS CLI](https://docs.vmware.com/en/VMware-Enterprise-PKS/1.7/vmware-enterprise-pks-17/GUID-installing-pks-cli.html?hWord=N4IghgNiBcIA4GsDOACAxhAliAvkA).

## <a id='configuring_pks_gpdb'></a>Step 3: Configure the GCP Service Account

1. Login to GCP with the `gcloud` tool and set a new project:

    ```bash
    $ gcloud auth login
    $ gcloud projects list
    $ gcloud config set project <project_name>
    ```

2. As a best practice, create a separate service account to use for deploying <%=vars.product_name %>:

    The <%=vars.product_name %> service account must have the following permissions, which are automatically set by the script:
    * Compute Viewer
    * Compute Network Admin
    * Storage Object Creator

    You can optionally grant these permissions to a different service account using [IAM & admin](https://console.cloud.google.com/iam-admin/iam) in the GCP console.

3. Activate the service account using the key file `key.json`:

    ```bash
    $ gcloud auth activate-service-account --key-file=./key.json
    ```

4. Set a preferred availability zone with a command similar to:

    ```bash
    $ gcloud config set compute/zone us-west1-a
    ```

    Update the zone name as necessary for your configuration.

5. Set up Google Container Registry to host Docker images

    [Enable the `Container Registry API`](https://console.cloud.google.com/flows/enableapi?apiid=containerregistry.googleapis.com)

    Configure docker to use the gcloud command-line tool as a credential helper for the activated service-account:

    ```bash
    $ gcloud auth configure-docker
    ```

   For additional details see [instructions](https://cloud.google.com/container-registry/docs/quickstart).

## <a id='firewall'></a>Step 4: Configure the GCP Firewall

In order to provide communication between the nodes and containers running Greenplum, you must create a firewall rule and add it to the default deployment tag. This creates routes between nodes and their containers. Follow these steps:
 
1. Access the Operations Manager configuration page at
`https://<OPSMANAGER_IP_ADDRESS_OR_URL>/infrastructure/iaas_configurations`. Look for the field named
`Default Deployment Tag` and record its value. If it is empty, record the value, `greenplum-tag`. 

2. Export a variable for the `Default Deployment Tag` value, as in:

    ```bash
    $ export DEFAULT_DEPLOYMENT_TAG="greenplum-tag"
    ```

    The `Default Deployment Tag` controls the default firewall rules for every VM that is created.

3. Determine the `Google Parent Network` name. To do so, access the Operations Manager page at `https://<OPSMANAGER_IP_ADDRESS_OR_URL>/infrastructure/networks/edit` and disclose one of the networks that you created for TKGI in the `Networks` section.  The `Google Network Name` value uses the format `<parent-network-name>/<subnet-name>/<region-name>`. Because all of the networks are based on the underlying Google Network, the first part of the full network name, `<parent-network-name>` corresponds to the `Google Parent Network`.

    For example, if the value of the `Google Network Name` reads `gpcloud-pks-net/foo/us-west1` then `gpcloud-pks-net` corresponds to the Google Parent network. 
    
4. Export an environment for the Google parent network:

    ```bash
    $ export GOOGLE_PARENT_NETWORK="gpcloud-pks-net"
    ``` 

5. Finally, create a firewall rule to allow inter-container traffic using `DEFAULT_DEPLOYMENT_TAG` and `GOOGLE_PARENT_NETWORK` variables that you exported:

    ```bash
    $ gcloud compute firewall-rules create greenplum-rule \
    --network=${GOOGLE_PARENT_NETWORK} \
    --action=ALLOW \
    --rules=tcp:1024-65535,tcp:22,icmp,udp \
    --source-ranges=0.0.0.0/0 \
    --target-tags ${DEFAULT_DEPLOYMENT_TAG}
    ```

## <a id='load_balancer'></a>Step 5: Configure the Kubernetes Load Balancer

Create or re-use a Load Balancer for a new Kubernetes cluster. If a TCP load balancer from a previous kubernetes cluster exists, reuse it by specifying the front-end IP address below in the [Step 6: Deploy a Kubernetes Cluster](#deploy_pks).

If no load balancer exists, use these commands to create one: 

```bash
$ export LB_NAME=greenplum-cluster1-lb
$ gcloud compute target-pools create ${LB_NAME}
$ REGION=$(gcloud config get-value compute/region)
$ gcloud compute forwarding-rules create ${LB_NAME}-forwarding-rule \
    --target-pool ${LB_NAME} --ports 8443 --ip-protocol TCP --region ${REGION}
$ echo " Front-end IP Address of load balancer for ${LB_NAME}:"
$ gcloud compute forwarding-rules describe ${LB_NAME}-forwarding-rule --region ${REGION} --format="value(IP_ADDRESS)"
```

You will reset the backend instances to attach to the new Kubernetes cluster in the next section. 

## <a id='deploy_pks'></a>Step 6: Deploy a Kubernetes Cluster

Follow these steps to deploy a new TKGI cluster using the firewall rules and load balancer you configured for Kubernetes:

1. Follow the instructions to [Logging in to Enterprise PKS](https://docs.vmware.com/en/VMware-Enterprise-PKS/1.7/vmware-enterprise-pks-17/GUID-login.html?hWord=N4IghgNiBcIQ9gcwAQEsB2yAu9kAcBrAZ2QGMJUQBfIA).

2. Execute the `create_pks_cluster_on_gcp.bash` script, specifying the IP address of the Load Balancer's front end. The command uses the syntax `create_cluster.bash <my_cluster-name>  <IP_address_of_Load_Balancer>`. For example:

    ```bash
    $ cd workspace
    $ samples/scripts/create_pks_cluster_on_gcp.bash gpdb 1.1.1.1
    ```

    This creates a Kubernetes cluster and assigns it to be accessed via the specified load balancer. You can use either an existing Load Balancer or a newly-created one, as described in [Step 5: Set Up the Kubernetes Load Balancer](#load_balancer).

3. Use the `kubectl` command to show the system containers and newly-created nodes available for deploying <%=vars.product_name %>:

    ```bash
    $ kubectl get pods --namespace kube-system
    $ kubectl get nodes 
    ```

## <a id='faq'></a>FAQ

### <a id='google_perms'></a>What permissions are required for the Google Service Account?

The service account used by TKGI must have the `compute.admin` role to create a new service. 
As part of creating a service, it must create firewall rules and create an external ip for the load balancer.

### <a id='bosh_shell'></a>How can I use bosh to obtain a shell?

In order to get a Bash shell on a particular node, go through Operations Manager to use it as a jump box:

1. In the Google Cloud web interface, look for "compute instances" and identify Ops Manager. 

2. Click the "ssh" button to launch an ssh shell on that instance.

3. After launching the ssh shell, use the BOSH CLI command from the Ops Manager Interface Credentials Tab. 
Look for a `Bosh Commandline Credentials` line that looks like:

    ```bash
    BOSH_CLIENT=<SOME_CLIENT> BOSH_CLIENT_SECRET=<SOME_SECRET> BOSH_CA_CERT=/var/tempest/workspaces/default/root_ca_certificate BOSH_ENVIRONMENT=<SOME_IP> bosh
    ```

4. Create an alias named `pks` and use it in commands below:

    ```bash
    BOSH_CLIENT=<SOME_CLIENT> BOSH_CLIENT_SECRET=<SOME_SECRET> BOSH_CA_CERT=/var/tempest/workspaces/default/root_ca_certificate bosh alias-env pks -e <SOME_IP>
    ```

5. For convenience, export the above two variables in the current shell:

    ```bash
    export BOSH_CLIENT=<SOME_CLIENT>
    export BOSH_CLIENT_SECRET=<SOME_SECRET>
    export BOSH_CA_CERT=/var/tempest/workspaces/default/root_ca_certificate
    ```

    After exporting those variables, the following command should work:

    ``` bash
    bosh -e pks deployments
    ```

6. Determine the IP address of the node you want by listing the nodes for each pod:

    ```bash
    kubectl get pods -o wide
    ```

    ```bash
    bosh -e pks vms
    bosh -e pks ssh <VM_NAME_FROM_COMMAND_ABOVE>
    ```

### <a id='macos_perms'></a>How do I set permissions for MacOS HomeBrew?

If necessary, execute the command:

```bash
sudo chown -R $USER:admin /usr/local
```

