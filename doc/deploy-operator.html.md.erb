---
title: Deploying or Redeploying a Greenplum Cluster
---

This section describes how to use the Greenplum Operator deploy a Greenplum cluster to your Kubernetes system. You can use these instructions either to deploy a brand new cluster (provisioning new, empty Persistent Volume Claims in Kubernetes), or to re-deploy an earlier cluster, re-using existing Persistent Volumes if available.

## Prerequisites

This procedure requires that you first install the <%=vars.product_name %> docker images and create the Greenplum Operator in your Kubernetes system. See [Installing <%=vars.product_name_long %>](installing.html) for more information.

Verify that the Greenplum Operator is installed and running in your system before you continue:

``` bash
$ helm list
```
``` bash
NAME              	NAMESPACE	REVISION	UPDATED                                	STATUS  	CHART           APP VERSION  
greenplum-operator	default  	1       	2020-05-13 11:17:30.640971495 -0700 PDT	deployed	operator-1.4.0  v2.0.0
```

To deploy multiple Greenplum cluster instances, you require multiple namespaces in your Kubernetes environment to target each cluster. If you need to create a new Kubernetes namespace, use the `kubectl create namespace` command. For example:

```bash
$ kubectl create namespace gpinstance-1
```

```bash
namespace/gpinstance-1 created
```

Verify that you have a namespace for the new Greenplum cluster instances that you want to deploy. For example:

```bash
$ kubectl get namespaces
```

```bash
NAME          STATUS    AGE
default       Active    50d
gpinstance-1  Active    50d 
gpinstance-2  Active    50d 
kube-public   Active    50d
kube-system   Active    50d
```

In the above output, gpinstance-1 and gpinstance-2 can be used as namespaces for deploying two different Greenplum cluster.

## Procedure

1. Go to the `workspace` subdirectory where you unpacked the <%=vars.product_name %> distribution for Kubernetes:

    ``` bash
    $ cd ./greenplum-for-kubernetes-*/workspace
    ```

1. If necessary, create a Kubernetes manifest file to specify the configuration of your Greenplum cluster. A sample file is provided in `workspace/my-gp-instance.yaml`. `my-gp-instance.yaml` contains the minimal set of instructions necessary to create a demonstration cluster named "my-greenplum" with a single segment and default storage, memory, and CPU settings:

    ``` yaml
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

    ```

    Most non-trivial clusters will require configuration changes to specify additional segments, cpu, memory, `pg_hba.conf` entries, and Storage Class resources. See [Greenplum Database Properties](operator-reference.html) for information about these configuration parameters and change them as necessary before you continue.

    <br/>If you want to re-deploy a Greenplum cluster that you previously deployed, simply locate and use the existing configuration file.

    <br/>If you want to deploy another Greenplum cluster (in a separate Kubernetes namespace), copy the `workspace/my-gp-instance.yaml` or a another deployment manifest file, and edit it as necessary to meet your cluster configuration requirements.

    <%=vars.product_name_long %> provides one additional sample manifest file that you can use as a template (copy to the `/workspace` directory before modifying):
    - `samples/my-gp-with-pxf-instance.yaml` contains the minimal configuration for a cluster that includes the Platform Extension Framework (PXF) deployed. PXF provides connectors that enable you to access data stored in sources external to your Greenplum Database deployment. These external sources include Hadoop (HDFS, Hive, HBase), object stores (Azure, Google Cloud Storage, Minio, S3), and SQL databases (via JDBC). See [Deploying PXF with Greenplum](deploy-pxf.html) for more information about deploying PXF with <%=vars.product_name %>.

    
1. (Optional) If you have specified `workerSelector` in your manifest file, then you need to apply the specified labels to the nodes that belong in the `masterAndStandby` and `segments` pools by using the following command:

    ```bash
    $ kubectl label node <node name> <key>=<value>
    ```

1. Use `kubectl apply` command and specify your manifest file to send the deployment request to the Greenplum Operator. For example, to use the sample `my-gp-instance.yaml` file:

    ``` bash
    $ kubectl apply -f ./my-gp-instance.yaml 
    ```
    ```bash
    greenplumcluster.greenplum.pivotal.io/my-greenplum created
    ```

    If you are deploying another instance of a Greenplum cluster, specify the Kubernetes namespace where you want to deploy the new cluster. For example, if you previously deployed a cluster in the namespace gpinstance-1, you could deploy a second Greenplum cluster in the gpinstance-2 namespace using the command:

    ```bash
    $ kubectl apply -f ./my-gp-instance.yaml -n gpinstance-2
    ```
    ```bash
    greenplumcluster.greenplum.pivotal.io/my-greenplum created
    ```

    The Greenplum Operator deploys the necessary Greenplum resources according to your specification, and also initializes the Greenplum cluster. If there are no existing Persistent Volume Claims for the cluster, new PVCs are created and used for the deployment.  If PVCs for the cluster already exist, they are used as-is with the available data.

    <br/>Deploying the cluster also creates a new service account named `greenplum-system-pod`.  Greenplum pods use this account internally to label their persistent volume claims (PVCs), but it is also visible if you use the `kubectl get serviceaccount` command:

    ```bash
    $ kubectl get serviceaccount
    ```
    ``` bash
    NAME                        SECRETS   AGE
    default                     1         39m
    greenplum-system-operator   1         36m
    greenplum-system-pod        1         18m
    ```

    <br/>If you enabled `antiAffinity` in your cluster configuration, individual nodes are labeled with `greenplum-affinity-<namespace>-segment=a`, `greenplum-affinity-<namespace>-segment=b`, and/or `greenplum-affinity-<namespace>-master=true`, as shown below:

    ```bash
    $ kubectl get nodes --show-labels
    NAME                                      STATUS   ROLES    AGE   VERSION   LABELS
    vm-4b50d90e-5e00-411f-5516-588711f0a618   Ready    <none>   11h   v1.16.7   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=custom-1-2048,beta.kubernetes.io/os=linux,bosh.id=3b3a6b47-8a1d-4a82-a06b-5349a241397e,bosh.zone=us-central1-f,failure-domain.beta.kubernetes.io/region=us-central1,failure-domain.beta.kubernetes.io/zone=us-central1-f,greenplum-affinity-default-master=true,greenplum-affinity-default-segment=a,kubernetes.io/hostname=vm-4b50d90e-5e00-411f-5516-588711f0a618,spec.ip=10.0.11.11,worker=my-gp-masters
    vm-50da037c-0c00-46f8-5968-2a51cf17e426   Ready    <none>   11h   v1.16.7   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=custom-1-2048,beta.kubernetes.io/os=linux,bosh.id=e6440a8d-8b75-4a0e-acc9-b210e81d59dc,bosh.zone=us-central1-f,failure-domain.beta.kubernetes.io/region=us-central1,failure-domain.beta.kubernetes.io/zone=us-central1-f,greenplum-affinity-default-master=true,greenplum-affinity-default-segment=b,kubernetes.io/hostname=vm-50da037c-0c00-46f8-5968-2a51cf17e426,spec.ip=10.0.11.16,worker=my-gp-masters
    vm-73e119aa-da79-4686-58df-1e9d7a9eff18   Ready    <none>   11h   v1.16.7   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=custom-1-2048,beta.kubernetes.io/os=linux,bosh.id=7e68ad80-6401-431b-8187-0ffc9c45dd69,bosh.zone=us-central1-f,failure-domain.beta.kubernetes.io/region=us-central1,failure-domain.beta.kubernetes.io/zone=us-central1-f,greenplum-affinity-default-master=true,greenplum-affinity-default-segment=a,kubernetes.io/hostname=vm-73e119aa-da79-4686-58df-1e9d7a9eff18,spec.ip=10.0.11.15,worker=my-gp-segments
    vm-8e43e0c6-6fd5-4bff-5c3a-150cbca76781   Ready    <none>   11h   v1.16.7   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=custom-1-2048,beta.kubernetes.io/os=linux,bosh.id=2bfd5222-96c5-47d7-98c2-52af11ea3854,bosh.zone=us-central1-f,failure-domain.beta.kubernetes.io/region=us-central1,failure-domain.beta.kubernetes.io/zone=us-central1-f,greenplum-affinity-default-master=true,greenplum-affinity-default-segment=b,kubernetes.io/hostname=vm-8e43e0c6-6fd5-4bff-5c3a-150cbca76781,spec.ip=10.0.11.13,worker=my-gp-segments
    vm-cf9fcef9-2557-43ca-43fa-01b21618e9ba   Ready    <none>   11h   v1.16.7   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=custom-1-2048,beta.kubernetes.io/os=linux,bosh.id=5a757d0f-d312-4fee-9c3f-52bd82c225f7,bosh.zone=us-central1-f,failure-domain.beta.kubernetes.io/region=us-central1,failure-domain.beta.kubernetes.io/zone=us-central1-f,greenplum-affinity-default-master=true,greenplum-affinity-default-segment=a,kubernetes.io/hostname=vm-cf9fcef9-2557-43ca-43fa-01b21618e9ba,spec.ip=10.0.11.14,worker=my-gp-segments
    vm-fb806a3c-8198-4608-671e-4659c940d2a4   Ready    <none>   11h   v1.16.7   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=custom-1-2048,beta.kubernetes.io/os=linux,bosh.id=18f8435d-be48-4445-b822-e0733ac7eced,bosh.zone=us-central1-f,failure-domain.beta.kubernetes.io/region=us-central1,failure-domain.beta.kubernetes.io/zone=us-central1-f,greenplum-affinity-default-master=true,greenplum-affinity-default-segment=b,kubernetes.io/hostname=vm-fb806a3c-8198-4608-671e-4659c940d2a4,spec.ip=10.0.11.12,worker=my-gp-segments
    ```

    <br/>Do not modify these labels, as they are used by the Operator for enforcing the `antiAffinity` setting.


1. While the cluster is initializing the status will be `Pending`:

    ``` bash
    $ watch kubectl get all
    ```
    ``` bash
    NAME                                      READY   STATUS    RESTARTS   AGE
    pod/greenplum-operator-6ff95b6b79-kq9vr   1/1     Running   0          20m
    pod/master-0                              1/1     Running   0          2m47s
    pod/segment-a-0                           1/1     Running   0          2m47s

    NAME                                                            TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
    service/agent                                                   ClusterIP      None             <none>        22/TCP           2m48s
    service/greenplum                                               LoadBalancer   10.102.131.136   <pending>     5432:32753/TCP   2m48s
    service/greenplum-validating-webhook-service-6ff95b6b79-kq9vr   ClusterIP      10.106.60.103    <none>        443/TCP          20m
    service/kubernetes                                              ClusterIP      10.96.0.1        <none>        443/TCP          24m

    NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/greenplum-operator   1/1     1            1           20m

    NAME                                            DESIRED   CURRENT   READY   AGE
    replicaset.apps/greenplum-operator-6ff95b6b79   1         1         1       20m

    NAME                         READY   AGE
    statefulset.apps/master      1/1     2m47s
    statefulset.apps/segment-a   1/1     2m47s

    NAME                                                 STATUS    AGE
    greenplumcluster.greenplum.pivotal.io/my-greenplum   Running   2m49s
    ```

1. _If you are redeploying a cluster that was configured to use a standby master_, wait until all pods reach the `Running` status. Then connect to the `master-0` pod and execute the `gpstart` command manually. For example:

    ``` bash
    kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; gpstart"
    ```

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
      Creation Timestamp:  2020-05-13T18:34:54Z
      Finalizers:
        stopcluster.greenplumcluster.pivotal.io
      Generation:        3
      Resource Version:  2196
      Self Link:         /apis/greenplum.pivotal.io/v1/namespaces/default/greenplumclusters/my-greenplum
      UID:               247daddf-b7e3-4479-a175-8c03e53f910f
    Spec:
      Master And Standby:
        Cpu:                        0.5
        Host Based Authentication:  # host   all   gpadmin   0.0.0.0/0   trust

        Memory:              800Mi
        Storage:             1G
        Storage Class Name:  standard
        Worker Selector:
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

1. At this point, you can work with the deployed Greenplum cluster by executing Greenplum utilities from within Kubernetes, or by using a locally-installed tool, such as `psql`, to access the Greenplum instance running in Kubernetes. For example, to run the `psql` utility on the `master-0` pod:

    ``` bash
    $ kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; psql"
    ```
    ```
    psql (9.4.24)
    Type "help" for help.
    ```
    ``` sql
    gpadmin=# select * from gp_segment_configuration;
    ```
    ``` sql
     dbid | content | role | preferred_role | mode | status | port  |  hostname   |                   address                   |  
        datadir      
    ------+---------+------+----------------+------+--------+-------+-------------+---------------------------------------------+--
    -----------------
        1 |      -1 | p    | p              | n    | u      |  5432 | master-0    | master-0.agent.default.svc.cluster.local    | /
    greenplum/data-1
        2 |       0 | p    | p              | n    | u      | 40000 | segment-a-0 | segment-a-0.agent.default.svc.cluster.local | /
    greenplum/data
    (2 rows)
    ```

    (Enter `\q` to exit the `psql utility`.)  
    
    See also [Accessing a Greenplum Cluster in Kubernetes](accessing.html).
