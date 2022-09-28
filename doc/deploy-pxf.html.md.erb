---
title: Deploying PXF with Greenplum
---

This section describes procedures for deploying a <%=vars.product_name %> cluster with the Greenplum platform extension framework (PXF) on Kubernetes.

## <a id="about"></a>About PXF with <%=vars.product_name_long %>

When you deploy PXF to <%=vars.product_name_long %>, the Greenplum Operator creates one or more dedicated pods, or _replicas_, to host the PXF server instances. This differs from Greenplum deployed to other platforms, where a PXF server instance is deployed to each Greenplum segment host. With <%=vars.product_name_long %>, you can choose to deploy as many PXF server _replicas_ as needed to provide redundancy should a PXF pod fail and to distribute load.

When deploying to Kubernetes, you store all PXF configuration files for a cluster externally, on an S3 data source. The deployment manifest file then specifies the S3 bucket-path to use for downloading the PXF configuration to all configured PXF servers.

When you install a new Greenplum cluster using the template PXF manifest file, `workspace/samples/my-gp-with-pxf-instance.yaml`, PXF is installed and initialized with a default (empty) PXF configuration directory. After deploying the cluster, you can customize the configuration by creating PXF server configurations for multiple data sources, and then redeploy with an updated manifest file to use the PXF configuration in your cluster.

**Note:** By default, <%=vars.product_name_long %> configures PXF server JVMs with the `-XX:+MaxRAMPercentage=75.0` setting in `PXF_JVM_OPTS`. This enables PXF to use most of the memory available in its container. The setting differs from PXF in other deployment environments, where PXF servers are deployed alongside Greenplum Segments and a fixed JVM memory size (`-Xmx2g -Xms1g` by default) is used to avoid competing for memory resources. See [Starting, Stopping, and Restarting PXF](https://gpdb.docs.pivotal.io/5latest/pxf/cfginitstart_pxf.html) in the <%=vars.product_name %> documentation for information about other runtime configuration options.

## <a id="initialize"></a>Deploying a New Greenplum Cluster with PXF Enabled

Follow these steps to deploy a new <%=vars.product_name %> cluster on Kubernetes with PXF enabled. (To add the PXF service to an existing Greenplum cluster, see [Adding PXF to an Existing Greenplum Cluster](#add).)

You can deploy PXF servers either in their default, initialized state, or you can use an existing PXF configuration, stored in an S3 bucket location, to use as the PXF configuration for your cluster.

See also [Configuring PXF Servers](#configure) for information about how to create and apply PXF server configurations to a <%=vars.product_name %> cluster in Kubernetes. 

1. Use the procedure described in [Deploying or Redeploying a Greenplum Cluster](deploy-operator.html) to deploy the cluster, but use the `samples/my-gp-with-pxf-instance.yaml` as the basis for your deployment. Copy the file into your `/workspace` directory. For example:

    ``` bash
    $ cd ./greenplum-for-kubernetes-*/workspace
    $ cp ./samples/my-gp-with-pxf-instance.yaml .
    ```

2. Edit the file as necessary for your deployment. `my-gp-with-pxf-instance.yaml` includes properties to configure PXF in the basic Greenplum cluster:

    ``` 
    apiVersion: "greenplum.pivotal.io/v1"
    kind: "GreenplumCluster"
    metadata:
      name: my-greenplum
    spec:
      masterAndStandby:
        hostBasedAuthentication: |
          # host   all   gpadmin   0.0.0.0/0    trust
        memory: "800Mi"
        cpu: "0.5"
        storageClassName: standard
        storage: 1G
        workerSelector: {}
      segments:
        primarySegmentCount: 1
        memory: "800Mi"
        cpu: "0.5"
        storageClassName: standard
        storage: 2G
        workerSelector: {}
      pxf:
        serviceName: "my-greenplum-pxf"    
    ---
    apiVersion: "greenplum.pivotal.io/v1beta1"
    kind: "GreenplumPXFService"
    metadata:
      name: my-greenplum-pxf
    spec:
      replicas: 2
      cpu: "0.5"
      memory: "1Gi"
      workerSelector: {}
    #  pxfConf:
    #    s3Source:
    #      secret: "my-greenplum-pxf-configs"
    #      endpoint: "s3.amazonaws.com"
    #      bucket: "YOUR_S3_BUCKET_NAME"
    #      folder: "YOUR_S3_BUCKET_FOLDER-Optional"
    #
    # Note: If using pxfConf.s3Source, in addition to applying the above yaml be sure to create a secret using a command similar to:
    # kubectl create secret generic my-greenplum-pxf-configs --from-literal=‘access_key_id=XXX’ --from-literal=‘secret_access_key=XXX’
    ```

    The entry:

    ``` yaml
      pxf:
        serviceName: "my-greenplum-pxf"    
    ```

    Indicates that the cluster will use the PXF service configuration named `my-greenplum-pxf` that follows at the end of the yaml file. The sample configuration creates two PXF replica pods for redundancy with minimal settings for CPU and memory. You can customize these values as needed, as well as the `workerSelector` value if you want to constrain the replica pods to labeled nodes in your cluster. See [Greenplum PXF Service Properties](gp-pxf-reference.html) for information about each available property.

1. If you have an existing PXF configuration that you want to apply to the Greenplum cluster, follow these additional steps to edit your manifest file and provide access to the configuration:

    1. Uncomment the `pxfConf` configuration properties at the end of the template file:
    
        ``` yaml
        pxfConf:
          s3Source:
            secret: "my-greenplum-pxf-configs"
            endpoint: "s3.amazonaws.com"
            bucket: "YOUR_S3_BUCKET_NAME"
            folder: "YOUR_S3_BUCKET_FOLDER-Optional"
        ```
    
    1. Set the `endpoint:`, `bucket:`, and `folder:` properties to specify the full S3 location that contains your PXF configuration files. All directories and files located in the specified S3 bucket-folder are copied into the `PXF_CONF` directory on each PXF server in the cluster. See [Configuring PXF Servers](#configure) for an example configuration that uses MinIO.

    1. Create a secret that can be used to authenticate access to the S3 bucket and folder that contains the PXF configuration directory. The name of the secret must match the name specified in the manifest file (`secret: "my-greenplum-pxf-configs"` by default). For example:

        ``` bash
        $ kubectl create secret generic my-greenplum-pxf-configs --from-literal='access_key_id=AKIAIOSFODNN7EXAMPLE'
            --from-literal='secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY'
        ```
        ``` bash
        secret/my-greenplum-pxf-configs created
        ```

        The above command creates a secret named `my-greenplum-pxf-configs` using the S3 access and secret keys that you provide. Replace the access and secret key values with the actual values for your system. If necessary, use your S3 implementation documentation to generate a secret access key. 

1. Use the `kubectl apply` command with your modified Greenplum manifest file to send the deployment request to the Greenplum Operator. For example:

    ``` bash
    $ kubectl apply -f ./my-gp-with-pxf-instance.yaml 
    ```
    ```bash
    greenplumcluster.greenplum.pivotal.io/my-greenplum created
    greenplumpxfservice.greenplum.pivotal.io/my-greenplum-pxf created
    ```

    If you are deploying another instance of a Greenplum cluster, specify the Kubernetes namespace where you want to deploy the new cluster. For example, if you previously deployed a cluster in the namespace gpinstance-1, you could deploy a second Greenplum cluster in the gpinstance-2 namespace using the command:

    ```bash
    $ kubectl apply -f ./my-gp-with-pxf-instance.yaml -n gpinstance-2
    ```

    The Greenplum Operator deploys the necessary Greenplum and PXF resources according to your specification, and also initializes the Greenplum cluster.


1. Execute the following command to monitor the deployment of the Greenplum cluster and PXF service. While the Greenplum cluster is initializing its status will be `Pending`. While the PXF service is initializing its status will be `Pending` or `Degraded`. When there are zero ready pods in the PXF service, the status will be `Pending`. When the PXF service has at least one ready pod, but the desired state is not yet reached, the status will be `Degraded`:

    ``` bash
    $ watch kubectl get all
    ```
    ``` bash
    NAME                                      READY   STATUS    RESTARTS   AGE
    pod/greenplum-operator-6ff95b6b79-kq9vr   1/1     Running   0          90m
    pod/master-0                              1/1     Running   0          2m16s
    pod/my-greenplum-pxf-7c947668c6-bhwlk     1/1     Running   0          2m19s
    pod/my-greenplum-pxf-7c947668c6-hb2r5     1/1     Running   0          2m19s
    pod/segment-a-0                           1/1     Running   0          2m16s

    NAME                                                            TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
    service/agent                                                   ClusterIP      None             <none>        22/TCP           2m16s
    service/greenplum                                               LoadBalancer   10.104.136.176   <pending>     5432:30938/TCP   2m16s
    service/greenplum-validating-webhook-service-6ff95b6b79-kq9vr   ClusterIP      10.106.60.103    <none>        443/TCP          90m
    service/kubernetes                                              ClusterIP      10.96.0.1        <none>        443/TCP          94m
    service/my-greenplum-pxf                                        ClusterIP      10.102.160.31    <none>        5888/TCP         2m19s

    NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/greenplum-operator   1/1     1            1           90m
    deployment.apps/my-greenplum-pxf     2/2     2            2           2m19s

    NAME                                            DESIRED   CURRENT   READY   AGE
    replicaset.apps/greenplum-operator-6ff95b6b79   1         1         1       90m
    replicaset.apps/my-greenplum-pxf-7c947668c6     2         2         2       2m19s

    NAME                         READY   AGE
    statefulset.apps/master      1/1     2m16s
    statefulset.apps/segment-a   1/1     2m16s

    NAME                                                 STATUS    AGE
    greenplumcluster.greenplum.pivotal.io/my-greenplum   Running   2m19s

    NAME                                                        STATUS
    greenplumpxfservice.greenplum.pivotal.io/my-greenplum-pxf   Running
    ```

    Note that the Greenplum PXF service, deployment, and replicas are created in addition to the Greenplum cluster.

1. Describe your Greenplum cluster to verify that it was created successfully. The Phase should eventually transition to `Running`:

    ``` bash
    $ kubectl describe greenplumClusters/my-greenplum
    ```

    ``` bash
    Name:         my-greenplum
    Namespace:    default
    Labels:       <none>
    Annotations:  API Version:  greenplum.pivotal.io/v1
    Kind:         GreenplumCluster
    Metadata:
      Creation Timestamp:  2020-05-13T19:45:32Z
      Finalizers:
        stopcluster.greenplumcluster.pivotal.io
      Generation:        3
      Resource Version:  7840
      Self Link:         /apis/greenplum.pivotal.io/v1/namespaces/default/greenplumclusters/my-greenplum
      UID:               bf247e0e-9af6-42c0-9371-9f9dcef688e9
    Spec:
      Master And Standby:
        Cpu:                        0.5
        Host Based Authentication:  # host   all   gpadmin   0.0.0.0/0   trust

        Memory:              800Mi
        Storage:             1G
        Storage Class Name:  standard
        Worker Selector:
      Pxf:
        Service Name:  my-greenplum-pxf
      Segments:
        Cpu:                    0.5
        Memory:                 800Mi
        Primary Segment Count:  1
        Storage:                2G
        Storage Class Name:     standard
        Worker Selector:
    Status:
      Instance Image:    greenplum-for-kubernetes:v2.0.0
      Operator Version:  greenplum-operator:v2.0.0
      Phase:             Running
    Events:              <none>
    ```

    If you are deploying a brand new cluster, the Greenplum Operator automatically initializes the Greenplum cluster. The `Phase` should eventually transition from `Pending` to `Running` and the Events should match the output above.

    <br/>**Note:** If you redeployed a previously-deployed Greenplum cluster, the phase will begin at `Pending`. The cluster uses its existing Persistent Volume Claims if they are available. In this case, the master and segment data directories will already exist in their former state. The master-0 pod automatically starts the Greenplum Cluster, after which the phase transitions to `Running`.

1. Describe your PXF service to verify that it was created successfully. The Phase should eventually transition to `Running`:

    ``` bash
    kubectl describe greenplumpxfservice.greenplum.pivotal.io/my-greenplum-pxf
    ```
    ``` bash
    Name:         my-greenplum-pxf
    Namespace:    default
    Labels:       <none>
    Annotations:  API Version:  greenplum.pivotal.io/v1beta1
    Kind:         GreenplumPXFService
    Metadata:
      Creation Timestamp:  2020-05-13T19:45:32Z
      Generation:          4
      Resource Version:    7799
      Self Link:           /apis/greenplum.pivotal.io/v1beta1/namespaces/default/greenplumpxfservices/my-greenplum-pxf
      UID:                 b2cae1bf-7cd9-4f1e-a002-d9e74c5f5ca6
    Spec:
      Cpu:       0.5
      Memory:    1Gi
      Replicas:  2
      Worker Selector:
    Status:
      Phase:  Running
    Events:   <none>
    ```

    The PXF service should automatically initialize itself. The `Phase` should eventually transition to `Running`.

1. At this point, you can work with the deployed Greenplum cluster by executing Greenplum utilities from within Kubernetes, or by using a locally-installed tool, such as `psql`, to access the Greenplum instance running in Kubernetes. Examine the `PXF_CONF` directory on master:

    ``` bash
    $ kubectl exec -it master-0 -- bash -c "ls -R /etc/pxf"
    ```
    ``` bash
    /etc/pxf:
    conf  keytabs  lib  logs  servers  templates

    /etc/pxf/conf:
    pxf-env.sh  pxf-log4j.properties  pxf-profiles.xml

    /etc/pxf/keytabs:

    /etc/pxf/lib:

    /etc/pxf/logs:

    /etc/pxf/servers:
    default

    /etc/pxf/servers/default:

    /etc/pxf/templates:
    adl-site.xml   hbase-site.xml  jdbc-site.xml	s3-site.xml
    core-site.xml  hdfs-site.xml   mapred-site.xml	wasbs-site.xml
    gs-site.xml    hive-site.xml   minio-site.xml	yarn-site.xml
    ```

    The above output shows a default PXF service has just been initialized, where the `PXF_CONF` directory (`/etc/pxf`) contains only the default subdirectories and template configuration files. If you applied an existing PXF configuration, verify that your custom PXF server configuration files are present. If you did not apply an existing PXF configuration, continue with the instructions in [Configuring PXF Servers](#configure) to verify basic PXF functionality in the new cluster.

## <a id="add"></a>Adding PXF to an Existing Greenplum Cluster

Follow these steps to deploy a PXF Service and associate it with an existing Greenplum cluster deployment. You can deploy PXF servers either in their default, initialized state, or you can use an existing PXF configuration, stored in an S3 bucket location, to use as the PXF configuration for your cluster.

See also [Configuring PXF Servers](#configure) for information about how to create and apply PXF server configurations to a Greenplum cluster deployment.

1. Delete the existing Greenplum cluster, but leave the existing PVCs so that the data in your existing Greenplum cluster is preserved. For example:

    ```bash
    $ kubectl delete -f workspace/my-gp-instance.yaml
    ```

    Ensure that the `GreenplumCluster` and its associated resources are completely gone before continuing:

    ```bash
    $ watch kubectl get all
    ```

1. Edit your manifest file to add a `GreenplumPXFService`. Associate the new `GreenplumPXFService` with the previously-existing `GreenplumCluster` resource by setting the PXF `serviceName` value. See the [PXF Reference](https://greenplum-kubernetes.docs.pivotal.io/latest/gp-pxf-reference.html) and [GreenplumCluster Reference](https://greenplum-kubernetes.docs.pivotal.io/latest/operator-reference.html) pages for more information and descriptions of the various configuration options. Below is an example manifest file:

    ```bash
    $ cat workspace/my-gp-instance.yaml
    ```
    ```yaml
    ---
    apiVersion: "greenplum.pivotal.io/v1"
    kind: "GreenplumCluster"
    metadata:
      name: my-greenplum
    spec:
      masterAndStandby:
        hostBasedAuthentication: |
          # host   all   gpadmin   0.0.0.0/0    trust
        memory: "800Mi"
        cpu: "0.5"
        storageClassName: standard
        storage: 1G
        workerSelector: {}
      segments:
        primarySegmentCount: 1
        memory: "800Mi"
        cpu: "0.5"
        storageClassName: standard
        storage: 2G
        workerSelector: {}
      pxf:
        serviceName: "my-greenplum-pxf"
    ---
    apiVersion: "greenplum.pivotal.io/v1beta1"
    kind: "GreenplumPXFService"
    metadata:
      name: my-greenplum-pxf
    spec:
      replicas: 2
      cpu: "0.5"
      memory: "1Gi"
      workerSelector: {}
    ```

1. Apply your updated manifest file to create the `GreenplumCluster` and `GreenplumPXFService` resources:

    ```bash
    $ kubectl apply -f workspace/my-gp-instance.yaml
    ```

1. Describe your Greenplum cluster to verify that it was created successfully.

    ``` bash
    $ kubectl describe greenplumClusters/my-greenplum
    ```

    ``` bash
    Name:         my-greenplum
    Namespace:    default
    Labels:       <none>
    Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                   {"apiVersion":"greenplum.pivotal.io/v1","kind":"GreenplumCluster", "metadata":{"annotations":{},"name":"my-greenplum", "namespace":"default"...
    API Version:  greenplum.pivotal.io/v1
    Kind:         GreenplumCluster
    Metadata:
     Creation Timestamp:  2019-04-01T15:19:17Z
     Generation:          1
     Resource Version:    1469567
     Self Link:           /apis/greenplum.pivotal.io/v1/namespaces/default/greenplumclusters/my-greenplum
     UID:                 83e0bdfd-5491-11e9-a268-c28bb5ff3d1c
    Spec:
     Master And Standby:
       Anti Affinity:              yes
       Cpu:                        0.5
       Host Based Authentication:  # host   all   gpadmin   0.0.0.0/0    trust

       Memory:              800Mi
       Storage:             1G
       Storage Class Name:  standard
       Worker Selector:
     Segments:
       Anti Affinity:          yes
       Cpu:                    0.5
       Memory:                 800Mi
       Primary Segment Count:  1
       Storage:                2G
       Storage Class Name:     standard
       Worker Selector:
    Status:
     Instance Image:    greenplum-for-kubernetes:latest
     Operator Version:  greenplum-operator:latest
     Phase:             Running
    Events:
     Type    Reason                    Age   From               Message
     ----    ------                    ----  ----               -------
     Normal  CreatingGreenplumCluster  2m    greenplumOperator  Creating Greenplum cluster my-greenplum in default
     Normal  CreatedGreenplumCluster   8s    greenplumOperator  Successfully created Greenplum cluster my-greenplum in default
    ```

    <br/>**Note:** If your Greenplum cluster is configured to use a standby master, the `Phase` will stay at `Pending` until after you perform the next step. Then, the `Phase` transitions to `Running`.

1. The following steps are required only if your cluster is configured to use a standby master. (If you do not use a standby master, skip to the next step.)

    1. Connect to the `master-0` pod and execute the `gpstart` command manually. For example:

        ``` bash
        $ kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; gpstart"
        ```
        ``` bash
        20200212:19:45:55:000517 gpstart:master-0:gpadmin-[INFO]:-Starting gpstart with args: 
        20200212:19:45:55:000517 gpstart:master-0:gpadmin-[INFO]:-Gathering information and validating the environment...
        20200212:19:45:55:000517 gpstart:master-0:gpadmin-[INFO]:-Greenplum Binary Version: 'postgres (Greenplum Database) 5.24.2     build dev'
        20200212:19:45:55:000517 gpstart:master-0:gpadmin-[INFO]:-Greenplum Catalog Version: '301705051'
        20200212:19:45:55:000517 gpstart:master-0:gpadmin-[INFO]:-Starting Master instance in admin mode
        20200212:19:45:56:000517 gpstart:master-0:gpadmin-[INFO]:-Obtaining Greenplum Master catalog information
        20200212:19:45:56:000517 gpstart:master-0:gpadmin-[INFO]:-Obtaining Segment details from master...
        20200212:19:45:56:000517 gpstart:master-0:gpadmin-[INFO]:-Setting new master era
        20200212:19:45:56:000517 gpstart:master-0:gpadmin-[INFO]:-Master Started...
        20200212:19:45:56:000517 gpstart:master-0:gpadmin-[INFO]:-Shutting down master
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:---------------------------
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:-Master instance parameters
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:---------------------------
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:-Database                 = template1
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:-Master Port              = 5432
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:-Master directory         = /greenplum/data-1
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:-Timeout                  = 600 seconds
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:-Master standby start     = On
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:---------------------------------------
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:-Segment instances that will be started
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:---------------------------------------
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:-   Host          Datadir                  Port    Role
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:-   segment-a-0   /greenplum/data          40000   Primary
        20200212:19:45:58:000517 gpstart:master-0:gpadmin-[INFO]:-   segment-b-0   /greenplum/mirror/data   50000   Mirror
    
        Continue with Greenplum instance startup Yy|Nn (default=N):
        ```

        Press `Y` to continue the startup.

    1. Describe your Greenplum PXF Service to verify that it was created successfully.
    
        ``` bash
        $ kubectl describe greenplumpxfservices/my-greenplum-pxf
        ```
        ``` bash
        Name:         my-greenplum-pxf
        Namespace:    default
        Labels:       <none>
        Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                       {"apiVersion":"greenplum.pivotal.io/v1beta1","kind":"GreenplumPXFService","metadata":{"annotations":{},    "name":"my-greenplum-pxf","namespac...
        API Version:  greenplum.pivotal.io/v1beta1
        Kind:         GreenplumPXFService
        Metadata:
         Creation Timestamp:  2020-03-12T21:44:49Z
         Generation:          4
         Resource Version:    47534
         Self Link:           /apis/greenplum.pivotal.io/v1beta1/namespaces/default/greenplumpxfservices/my-greenplum-pxf
         UID:                 3aafe6f1-5e6f-4bca-bb53-f97987debf9e
        Spec:
         Cpu:       0.5
         Memory:    1Gi
         Replicas:  4
         Worker Selector:
        Status:
         Phase:  Running
        Events:   <none>
        ```

        Eventually, the phase changes to `Running`, indicating that the PXF service is ready for use.

    1. Enable the PXF extension for your Greenplum cluster by issuing the following commands.
    
        ```bash
        $ kubectl exec -it master-0 -- bash
        $ psql -d gpadmin -c 'CREATE EXTENSION IF NOT EXISTS pxf;'
        ```

1. Test the PXF service to ensure that it works:

    <!-- TODO: there will soon be a new method for smoke testing pxf. See https://www.pivotaltracker.com/story/show/168672648/comments/212389966 -->
    ``` bash
    $ kubectl exec -it master-0 -- bash
    $ psql -d gpadmin
    ```
    ```sql
    postgres=# CREATE EXTERNAL TABLE pxf_read_test (a TEXT, b TEXT, c TEXT)
        LOCATION ('pxf://tmp/dummy1'
        '?FRAGMENTER=org.greenplum.pxf.api.examples.DemoFragmenter'
        '&ACCESSOR=org.greenplum.pxf.api.examples.DemoAccessor'
        '&RESOLVER=org.greenplum.pxf.api.examples.DemoTextResolver')
        FORMAT 'TEXT' (DELIMITER ',');
    SELECT * FROM pxf_read_test;
    postgres=#     SELECT * FROM pxf_read_test;
           a        |   b    |   c    
    ----------------+--------+--------
     fragment1 row1 | value1 | value2
     fragment1 row2 | value1 | value2
     fragment2 row1 | value1 | value2
     fragment2 row2 | value1 | value2
     fragment3 row1 | value1 | value2
     fragment3 row2 | value1 | value2
    (6 rows)
    ```

## <a id="configure"></a>Configuring PXF Servers

With <%=vars.product_name_long %>, all PXF configuration files for a cluster are stored externally, on an S3 data source. The Greenplum manifest file then specifies the S3 bucket-path to use for downloading the PXF configuration to all configured PXF servers. Any directories and files at the specified bucket-path are copied as-is to all PXF Servers configured for the cluster.

This procedure describes how to add or modify a PXF configuration in a <%=vars.product_name %> cluster on Kubernetes.

### <a id="prereq"></a>Prerequisites

This procedure uses MinIO as an example data source both for storing the PXF server configuration and for accessing remote data via PXF. If you want to follow along using the MinIO example, install the MinIO client, `mc` to your local system. See the [MinIO Client Quickstart Guide](https://docs.min.io/docs/minio-client-quickstart-guide.html) for installation instructions.

You should also have access to a <%=vars.product_name %> deployment on Kubernetes that includes PXF.  See [Deploying a Cluster with PXF Enabled](#initialize).

### <a id="procedure"></a>Procedure

1. To use MinIO as a sample data source, first install a standalone MinIO server to your cluster using `helm`. For example:

    ``` bash
    $ helm install my-minio stable/minio
    ```
    ``` bash
    NAME: my-minio
    LAST DEPLOYED: Wed May 13 13:36:37 2020
    NAMESPACE: default
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    NOTES:
    Minio can be accessed via port 9000 on the following DNS name from within your cluster:
    my-minio.default.svc.cluster.local

    To access Minio from localhost, run the below commands:

      1. export POD_NAME=$(kubectl get pods --namespace default -l "release=my-minio" -o jsonpath="{.items[0].metadata.name}")

      2. kubectl port-forward $POD_NAME 9000 --namespace default

    Read more about port forwarding here: http://kubernetes.io/docs/user-guide/kubectl/kubectl_port-forward/

    You can now access Minio server on http://localhost:9000. Follow the below steps to connect to Minio server with mc client:

      1. Download the Minio mc client - https://docs.minio.io/docs/minio-client-quickstart-guide

      2. mc config host add my-minio-local http://localhost:9000 AKIAIOSFODNN7EXAMPLE wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY S3v4

      3. mc ls my-minio-local

    Alternately, you can use your browser or the Minio SDK to access the server - https://docs.minio.io/categories/17
    ```

2. Execute the commands shown at the end of the MinIO deployment output to make the MinIO server accessible from the local host. Using the above output as an example:

    ``` bash
    $ export POD_NAME=$(kubectl get pods --namespace default -l "release=my-minio" -o jsonpath="{.items[0].metadata.name}")
    $ kubectl port-forward $POD_NAME 9000 --namespace default
    ```
    ``` bash
    Forwarding from 127.0.0.1:9000 -> 9000
    Forwarding from [::1]:9000 -> 9000
    ```

3. Follow these steps to create the sample data file and copy it to MinIO:

    1. Configure the `mc` client to use the MinIO server you just deployed:

        ``` bash
        $ mc config host add my-minio-local http://localhost:9000 AKIAIOSFODNN7EXAMPLE wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY S3v4
        ```
        ``` bash
        Added `my-minio-local` successfully.
        ```

        The `accessKey` and `secretKey` in the above command are used in the default `helm` chart deployment.

    1. Make two new buckets named to store the sample data and PXF configuration:

        ``` bash
        $ mc mb minio/pxf-config
        $ mc mb minio/pxf-data
        ```
        ``` bash
        Bucket created successfully `minio/pxf-config`.
        Bucket created successfully `minio/pxf-data`.
        ```

    1. Create a delimited plain text data file named `pxf_s3_simple.txt` to provide the sample data:

        ``` bash
        $ echo 'Prague,Jan,101,4875.33
        Rome,Mar,87,1557.39
        Bangalore,May,317,8936.99
        Beijing,Jul,411,11600.67' > ./pxf_s3_simple.txt
        ```

    1. Copy the sample data file to the MinIO bucket you created:

        ``` bash
        $  mc cp ./pxf_s3_simple.txt minio/pxf-data
        ```
        ``` bash
        ./pxf_s3_simple.txt:                      192 B / 192 B [===============] 100.00% 6.46 KiB/s 0s
        ```

4. Follow these steps to create the PXF server configuration file to access MinIO, and store it on the MinIO server:

    1. Copy the template PXF MinIO configuration file from the Greenplum master pod to your local host:

        ``` bash
        $ kubectl cp master-0:/etc/pxf/templates/minio-site.xml ./minio-site.xml
        ```

    1. Open the copied template file in a text editor, and edit the file entries to access the MinIO server that you deployed with `helm`. The file contents should be similar to:

        ``` xml
        <?xml version="1.0" encoding="UTF-8"?>
        <configuration>
            <property>
                <name>fs.s3a.endpoint</name>
                <value>http://my-minio:9000</value>
            </property>
            <property>
                <name>fs.s3a.access.key</name>
                <value>AKIAIOSFODNN7EXAMPLE</value>
            </property>
            <property>
                <name>fs.s3a.secret.key</name>
                <value>wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY</value>
            </property>
            <property>
                <name>fs.s3a.fast.upload</name>
                <value>true</value>
            </property>
            <property>
                <name>fs.s3a.path.style.access</name>
                <value>true</value>
            </property>
        </configuration>
        ```

        Be sure to change the first property value, `fs.s3a.endpoint`, to the URL of the MinIO service that was deployed in your cluster.  The `access.key` and `secret.key` values in the above output are used in the default `helm` chart deployment. All other property values are the defaults provided in the template file.
    
    1. Save the file and exit your text editor.

    1. Copy your modified `minio-site.xml` file for use as the default PXF server configuration on all PXF pods deployed in your server. To do this, you will place it in the example `pxf-config` bucket under the `/servers/default` directory. For example:

        ``` bash
        $ mc cp ./minio-site.xml minio/pxf-config/servers/default/minio-site.xml
        ```
        ``` bash
        ./minio-site.xml:       643 B / 643 B [===============] 100.00% 6.46 KiB/    s 0s
        ```

        At this point, you have deployed a MinIO server with sample data, and placed a sample PXF server configuration for the minio server in a location where it will be copied and used as the default PXF server configuration. 
    
5. Follow these steps to update your Greenplum cluster to use the new PXF server configuration file that you created and staged in MinIO:

    1. Move to the <%=vars.product_name %> `workspace` directory you used to deploy the Greenplum cluster.

    1. Edit the manifest file for your cluster (for example, `my-gp-with-pxf-instance.yaml`) in a text editor.

    1. Uncomment and edit the `pxfConf` configuration properties at the end of the template file to describe the MinIO location where you copied the PXF configuration file.  For example:

        ``` yaml
        pxfConf:
          s3Source:
            secret: "my-greenplum-pxf-configs"
            endpoint: "my-minio:9000"
            protocol:  "http"
            bucket: "pxf-config"
        ```

    1. Create a secret that can be used to authenticate access to the S3 bucket and folder that contains the PXF configuration directory. The name of the secret must match the name specified in the manifest file (`secret: "my-greenplum-pxf-configs"` by default). For example:

        ``` bash
        $ kubectl create secret generic my-greenplum-pxf-configs --from-literal='access_key_id=AKIAIOSFODNN7EXAMPLE' \
            --from-literal='secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY'
        ```
        ``` bash
        secret/my-greenplum-pxf-configs created
        ```

    1.  Apply the modified configuration:

        ``` bash
        $ kubectl apply -f ./my-gp-with-pxf-instance.yaml
        ```
        ``` bash
        greenplumcluster.greenplum.pivotal.io/my-greenplum unchanged
        greenplumpxfservice.greenplum.pivotal.io/my-greenplum-pxf configured
        ```

    1. If you encounter any errors while setting up `pxfConf`, refer to [Troubleshooting GreenplumPXFService pod startup](troubleshooting.html#troubleshootingpxf).

6. Perform the remaining steps on the Greenplum master pod to create and query an external table that references the sample MinIO data:

    1. Open a bash shell on the `master-0` pod:

        ``` bash
        $ kubectl exec -it master-0 -- bash
        ```

    1. Start the `psql` subsystem:

        ``` bash
        $ psql -d gpadmin
        ```
        ``` bash
        psql (9.4.24)
        Type "help" for help.
    
        postgres=#
        ```

    1. Create the PXF extension in the database:

        ``` sql
        postgres=# create extension pxf;
        ```
        ``` sql
        CREATE EXTENSION
        ```

    1. Use the PXF `s3:text` profile to create a Greenplum Database external table that references the `pxf_s3_simple.txt` file that you just created and added to MinIO. This command omits the typical `&SERVER=<server_name>` option in the PXF location URL, because the procedure created only the default server configuration:

        ``` sql
        postgres=# CREATE EXTERNAL TABLE pxf_s3_textsimple(location text, month text, num_orders int, total_sales float8) 
                   LOCATION ('pxf://pxf-data/pxf_s3_simple.txt?PROFILE=s3:text') 
                   FORMAT 'TEXT' (delimiter=E',');
        ```
        ``` sql
        CREATE EXTERNAL TABLE
        ```

    1. Query the external table to access the sample data stored on MinIO:

        ``` sql
        postgres=# SELECT * FROM pxf_s3_textsimple;
        ```
        ``` sql
         location  | month | num_orders | total_sales
        -----------+-------+------------+-------------
         Prague    | Jan   |        101 |     4875.33
         Rome      | Mar   |         87 |     1557.39
         Bangalore | May   |        317 |     8936.99
         Beijing   | Jul   |        411 |    11600.67
        (4 rows)
        ```

    If you receive any errors when querying the external table, verify the contents of the `/etc/pxf/servers/default/minio-site.xml` file on each PXF server in the cluster. Also use the `mc` client to verify the contents and location of the sample data file on MinIO.
    
    Further PXF troubleshooting information is available in the Greenplum Database documentation at [Troubleshooting PXF](https://gpdb.docs.pivotal.io/latest/pxf/troubleshooting_pxf.html).
