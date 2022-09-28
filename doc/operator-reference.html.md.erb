---
title: Greenplum Database Properties
---

This section describes each of the properties that you can define in a Greenplum Operator manifest file.

## <a id="synopsis"></a>Synopsis

``` yaml
apiVersion: "greenplum.pivotal.io/v1"
kind: "GreenplumCluster"
metadata:
  name: <string>
  namespace: <string>
spec:
  masterAndStandby:
    standby: <yes|no>
    hostBasedAuthentication: |
      [ host  <database>  <role>  <address>  <authentication-method> ]
      [ ... ]
    memory: <memory-limit>
    cpu: <cpu-limit>
    storageClassName: <storage-class>
    storageSize: <size>
    workerSelector: {
      <label>: "<value>"
      [ ... ]
    }
    antiAffinity: <yes|no>
  segments:
    primarySegmentCount: <int>
    memory: <memory-limit>
    cpu: <cpu-limit>
    storageClassName: <storage-class>
    storageSize: <size>
    workerSelector: {
      <label>: "<value>"
      [ ... ]
    }
    antiAffinity: <yes|no>
    mirrors: <yes|no>
  pxf:
    serviceName: "<pxf-service-name>" 
```

## <a id="description"></a>Description

You specify Greenplum cluster configuration properties to the Greenplum Operator via a YAML-formatted manifest file. A sample manifest file is provided in `workspace/my-greenplum-cluster.yaml`. The current version of the manifest supports configuring the cluster name, number of segments, and the memory, cpu, and storage settings for master and segment pods. See also [Deploying or Redeploying a Greenplum Cluster](deploy-operator.html) for information about deploying a new cluster using a manifest file.

## <a id="keywords"></a>Keywords and Values


### <a id="metadata"></a>Cluster Metadata

<dt>`name: <string>`</dt>
<dd>(Required) Sets the name of the Greenplum cluster instance resources. You can filter the output of `kubectl` commands using this name.  For example, if you set the name to `my-greenplum` then you can use commands like `kubectl get all -l greenplum-cluster=my-greenplum` or `kubectl get pvc -l greenplum-cluster=my-greenplum` to get all resources or PVCs associated with the Greenplum cluster instance, respectively. </dd>
<dd><br/>This value cannot be dynamically changed for an existing cluster.  If you attempt to change this value and re-apply it to an existing cluster, the Operator will interpret the change as a new deployment and reject it as only one Greenplum cluster instance is allowed per namespace.</dd>

<dt>`namespace: <string>`</dt>
<dd>(Optional) Specifies the namespace in which the Greenplum cluster is deployed. If not specified, the current kubectl context's namespace will be used for cluster deployment. To set kubectl's current context to a specific namespace, use the command:

```bash
$ kubectl config set-context $(kubectl config current-context) --namespace=<NAMESPACE>
```
</dd>
<dd>This value cannot be dynamically changed for an existing cluster.  To deploy an existing cluster to a different namespace, first delete the cluster instance and then deploy it using the new `namespace` value.</dd>

### <a id="segment"></a>Segment Configuration

<dt>`masterAndStandby:`, `segments:`</dt>
<dd>These sections share many of the same properties to configure memory, CPU, and storage for Greenplum segment pods. `masterAndStandby:` settings apply only to both the master and standby master pods. All <%=vars.product_name %> clusters include a standby master. The `segments:` section applies to each primary segment and optional mirror segment pod.</dd>

<dt>`standby: <yes or no>`</dt>
<dd>(Optional) Enables or disables the use of standby when deploying a Greenplum cluster. Defaults to "no" if omitted or left empty. This value cannot be dynamically changed for an existing cluster.</dd>
<dd><br/>**Note:** You cannot change this value for an existing cluster unless you first delete both the deployed cluster *and* the PVCs that were created for that cluster. This will result in a new, empty Greenplum cluster. See [Deleting Greenplum Persistent Volume Claims](deleting.html#delpvs).</dd>
<dd><br/>**Note:** If standby/mirrors is set to "no" (the default), antiAffinity must also be set to "no".</dd>

<dt>`hostBasedAuthentication:`</dt>
<dd>(Optional) Entries to add to the `pg_hba.conf` file generated for the Greenplum cluster. Each entry (multiple entries are possible) must include the items `host  <database>  <role>  <address>  <authentication-method>` in that order, to enable a role to access the indicated database (or `all` databases) from the specified CIDR and authentication method. See [Allowing Connections to Greenplum Database](http://gpdb.docs.pivotal.io/5110/admin_guide/client_auth.html#topic2) in the Greenplum Database documentation for more information about `pg_hba.conf` file entries.</dd>
<dd><br/>This value cannot be dynamically changed for an existing cluster.  The Operator only uses this value to populate the initial `pg_hba.conf` file that is created with a new cluster. You cannot use this property to change the existing generated file; instead, modify it directly on the `master` pod using a text editor. See the section on `Editing the pg_hba.conf File` in [Allowing Connections to Greenplum Database](http://gpdb.docs.pivotal.io/5110/admin_guide/client_auth.html#topic2).</dd>

<dt>`primarySegmentCount: <int>`</dt>
<dd>(Required) The number of primary/mirror segment pod pairs to create in the Greenplum cluster.  Segment pods use the naming format `segment-<type>-<number>` where the segment `<type>` is either `a` for primary segments or `b` for mirror segments. Segment numbering starts at zero.  If you omit this property, the Operator will fail to create a Greenplum cluster because it requires at least 1 primary segment.</dd>
<dd><br/>You can increase this value and re-apply it to an existing cluster, and the Greenplum operator automatically creates the new segment pods and initializes the Greenplum segment instances. You can optionally redistribute existing data to the new segments and/or delete the expansion schema that is created during this process.  See [Expanding a Greenplum Deployment](expanding.html).</dd>
<dd><br/>**Note:** You cannot decrease this value for an existing cluster unless you first delete both the deployed cluster *and* the PVCs that were created for that cluster. This will result in a new, empty Greenplum cluster. See [Deleting Greenplum Persistent Volume Claims](deleting.html#delpvs).</dd>

<dt>`memory: <memory-limit>`</dt>
<dd>(Optional) The amount of memory allocated to a Greenplum pod. This value defines a memory limit; if a pod tries to exceed the limit it is removed and replaced by a new pod. You can specify a suffix to define the memory units (for example, `4.5Gi`.). If omitted or left empty, the pod has no upper bound on the memory resource it can use or inherits the default limit if one is specified in its deployed namespace.  See [Assign Memory Resources to Containers and Pods](https://kubernetes.io/docs/tasks/configure-pod-container/assign-memory-resource) in the Kubernetes documentation for more information. </dd>
<dd><br/>This value cannot be dynamically changed for an existing cluster.  If you attempt to make changes to this value and re-apply it to an existing cluster, the change will be rejected.  If you wish to update this value, you must delete the existing cluster and recreate the cluster for the new value to take effect. </dd>
<dd><br/>**Note:** If you do not want to specify a memory limit, comment-out or remove the `memory:` keyword from the YAML file, or specify an empty string for its value (`memory: ""`). If the keyword appears in the YAML file, you must assign a valid string value to it.</dd>

<dt>`cpu: <cpu-limit>`</dt>
<dd>(Optional) The amount of CPU resources allocated to a Greenplum pod, specified as a Kubernetes CPU unit (for example, `cpu: "1.2"`). If omitted or left empty, the pod has no upper bound on the CPU resource it can use or inherits the default limit if one is specified in its deployed namespace.  See [Assign CPU Resources to Containers and Pods](https://kubernetes.io/docs/tasks/configure-pod-container/assign-cpu-resource/) in the Kubernetes documentation for more information.</dd>
<dd><br/>This value cannot be dynamically changed for an existing cluster.  If you attempt to make changes to this value and re-apply it to an existing cluster, the change will be rejected.  If you wish to update this value, you must delete the existing cluster and recreate the cluster for the new value to take effect.</dd>
<dd><br/>**Note:** If you do not want to specify a cpu limit, comment-out or remove the `cpu:` keyword from the YAML file, or specify an empty string for its value (`cpu: ""`). If the keyword appears in the YAML file, you must assign a valid string value to it.</dd>

<dt>`storageClassName: <storage-class>`</dt>
<dd>(Required) The Storage Class name to use for dynamically provisioning Persistent Volumes (PVs) for a Greenplum pod. If the PVs already exist, either from a previous deployment of the Greenplum instance or because you manually provisioned the PVs, then the Greenplum Operator uses the existing PVs. You can configure the Storage Class according to your performance needs. See [Storage Classes](https://kubernetes.io/docs/concepts/storage/storage-classes/) in the Kubernetes documentation to understand the different configuration options.</dd>
<dd><br/>For best performance, use persistent volumes that are backed by a local SSD with the XFS filesystem, using `readahead` cache for best performance. Use the mount options `rw,nodev,noatime,nobarrier,inode64` to mount the volume. See [Creating Local Persistent Volumes for Greenplum](create-local-pv.html) for information about manually provisioning local persistent volumes. See [Optimizing Persistent Disk and Local SSD Performance](https://cloud.google.com/compute/docs/disks/performance) in the Google Cloud documentation for information about the performance characteristics of different storage types.</dd>
<dd><br/>You cannot change this value for an existing cluster unless you first delete both the deployed cluster *and* the PVCs that were created for that cluster. See [Deleting Greenplum Persistent Volume Claims](deleting.html#delpvs).</dd>

<dt>`storageSize: <size>`</dt>
<dd>(Required) The storage size of the Persistent Volume Claim (PVC) for a Greenplum pod. Specify a suffix for the units (for example: `100G`, `1T`).</dd>
<dd><br/>You cannot change this value for an existing cluster unless you first delete both the deployed cluster *and* the PVCs that were created for that cluster. This will result in a new, empty Greenplum cluster. See [Deleting Greenplum Persistent Volume Claims](deleting.html#delpvs).</dd>

<dt><a id="workerSelector"></a>`workerSelector: <map of key-value pairs>`</dt>
<dd>(Optional) One or more [selector labels](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) to use for choosing Greenplum pods. Specify one or more label-value pairs to constrain Greenplum pods to nodes having the matching labels. Define the selector labels as you would for a pod's `nodeSelector` attribute. You can define the `workerSelector` attribute for Greenplum master and standby pods and/or for segment pods. If a `workerSelector` is not desired, remove the `workerSelector` attribute from the manifest file. </dd>
<dd><br/>For example, consider the case where you assign the label `gpdb-worker=master` to one or more pods using the command:

``` bash
$ kubectl label node <node_name> gpdb-worker=master
```
   
Similarly, pods reserved for Greenplum segments might be assigned the `gpdb-worker=segments` label:

  ``` bash
  $ kubectl label node <node_name> gpdb-worker=segments
  ```

With the above labels present in your cluster, you would edit the Greenplum Operator manifest file to specify the same key-value pairs in the `workerSelector` attribute. This shows the relevant excerpt from the manifest file:

``` yaml
  masterAndStandby:
    storageClassName: gpdb-storage
    storageSize: 5G
    workerSelector: {
      gpdb-worker: "master"
    }
  segments:
    primarySegmentCount: 6
    storageClassName: gpdb-storage
    storageSize: 50G
    workerSelector: {
      gpdb-worker: "segments"
    }
```
</dd>
<dd><br/>This value cannot be dynamically changed for an existing cluster.  If you want to update this value, you must first delete the existing cluster and then recreate the cluster for the new value to take effect.</dd>

<dt><a id="antiAffinity"></a>`antiAffinity: <yes or no>`</dt>
<dd>(Optional) Enables or disables the anti-affinity property when deploying a Greenplum cluster with mirror segments. Specifying "yes" means that the Operator guarantees scheduling a mirror segment to a different worker node than its corresponding primary segment (or similarly for master and standby).  If anti-affinity cannot be achieved, then the Operator aborts the Greenplum cluster deployment.  Specifying "no" means that the Operator uses the native Kubernetes system scheduler to schedule primary or mirror segments based on available worker nodes indiscriminately (similarly for master and standby).  Defaults to "no" if omitted or left empty. </dd>
<dd><br/><b>Notes:</b></dd>
<dd><ul>
  <li>If you are using virtual machines, always ensure that there is one Kubernetes worker per server for to ensure high availability.</li>
  <li>Set <code>antiAffinity</code> to "yes" when mirroring is enabled unless the deployment is to a single Kubernetes worker node.</li>
  <li>The <code>antiAffinity</code> values for <code>masterAndStandby</code> and <code>segments</code> must be the same. Otherwise, the cluster will fail to deploy.</li>
</ul></dd>
<dd><br/>This value cannot be dynamically changed for an existing cluster.  If you wish to update this value, you must delete the existing cluster and recreate the cluster for the new value to take effect.</dd>

<dt>`mirrors: <yes or no>`</dt>
<dd>(Optional) Enables or disables the use of segment mirroring when deploying a Greenplum cluster. Defaults to "no" if omitted or left empty. This value cannot be dynamically changed for an existing cluster.</dd>
<dd><br/>**Note:** You cannot change this value for an existing cluster unless you first delete both the deployed cluster *and* the PVCs that were created for that cluster. This will result in a new, empty Greenplum cluster. See [Deleting Greenplum Persistent Volume Claims](deleting.html#delpvs).</dd>
<dd><br/>**Note:** If standby/mirrors is set to "no", antiAffinity must also be set to "no" (the default).</dd>

### <a id="pxf"></a>PXF Configuration

<dt>`pxf.serviceName: "<pxf_service_name>"`</dt>
<dd>(Optional) Specifies the name of the Greenplum PXF service to which this GreenplumCluster connects. If you include a `GreenplumPXFService` configuration in your manifest file, specify its name here. As a best practice, keep the PXF service configuration properties in the same manifest file as Greenplum Database, as shown in the `workspace/samples/my-gp-with-pxf-instance.yaml` file. This simplifies upgrades or changes to the related service objects. When `pxf.serviceName` is set, the PXF extension is automatically created in the `gpadmin` database.</dd>
<dd>See [PXF Service Properties](gp-pxf-reference.html) for information about the properties used to configure the PXF service.</dd>

## <a id="examples"></a>Examples

See the `workspace/my-greenplum-cluster.yaml` for an example manifest.

## <a id="seealso"></a>See Also

[Deploying or Redeploying a Greenplum Cluster](deploy-operator.html), [Deleting a Greenplum Cluster](deleting.html), [Installing <%=vars.product_name_long %>](installing.html).
