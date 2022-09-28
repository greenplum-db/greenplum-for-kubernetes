---
title: Greenplum PXF Service Properties
---

This section describes each of the properties that you can define for a `GreenplumPXFService` configuration in the <%=vars.product_name %> manifest file.

## <a id="synopsis"></a>Synopsis

``` yaml
apiVersion: "greenplum.pivotal.io/v1beta1"
kind: "GreenplumPXFService"
metadata:
  name: <string>
  namespace: <string>
spec:
  replicas: <integer>
  cpu: <cpu-limit>
  memory: <memory-limit>
  workerSelector: {
        <label>: "<value>"
        [ ... ]
  }
  pxfConf:
    s3Source:
      secret: <Secrets name string>
      endpoint: <valid URL string>
      protocol: <http|https>
      bucket: <string>
      folder: <string> [Optional]

```

## <a id="description"></a>Description

You specify Greenplum PXF configuration properties to the Greenplum Operator via the YAML-formatted Greenplum manifest file. A sample manifest file is provided in `workspace/samples/my-gp-with-pxf-instance.yaml`. The current version of the manifest supports configuring the cluster name, number of PXF replicas, and the memory, CPU, and remote PXF_CONF configs. See also [Deploying PXF with Greenplum](deploy-pxf.html) for information about deploying a new Greenplum cluster with PXF using a manifest file.

**Note:** As a best practice, keep the PXF configuration properties in the same manifest file as Greenplum Database, to simplify upgrades or changes to the related service objects.

## <a id="keywords"></a>Keywords and Values


### Cluster Metadata

<dt>`name: <string>`</dt>
<dd>(Required.) Sets the name of the Greenplum PXF instance resources. You can filter the output of `kubectl` commands using this name. </dd>
<dd><br/>This value cannot be dynamically changed for an existing cluster.  If you attempt to change this value and re-apply it to an existing cluster, the Operator will create a new deployment.</dd>

<dt>`namespace: <string>`</dt>
<dd>(Optional.) Specifies the namespace in which the Greenplum PXF resources are deployed. If this property is not specified, the current kubectl context's namespace is used for deployment. To set kubectl's current context to a specific namespace, use the command:

```bash
$ kubectl config set-context $(kubectl config current-context) --namespace=<NAMESPACE>
```
</dd>

<dd>This value cannot be dynamically changed for an existing cluster.  To deploy resources to an existing cluster but under a different namespace, first delete the cluster instance and then deploy it using the new `namespace` value.</dd>

### Greenplum PXF Configuration

<dt>`replicas: <int>`</dt>
<dd>(Optional.) The number of PXF replica pods to create in the Greenplum PXF cluster. The default is 2.</dd>
<dd><br/>You can increase this value and re-apply it to an existing cluster as necessary.</dd>

<dt>`memory: <memory-limit>`</dt>
<dd>(Optional.) The amount of memory allocated to each Greenplum PXF pod. This value defines a memory limit; if a pod tries to exceed the limit it is removed and replaced by a new pod. You can specify a suffix to define the memory units (for example, `4.5Gi`.). If omitted, the pod has no upper bound on the memory resource it can use or inherits the default limit if one is specified in its deployed namespace.  See [Assign Memory Resources to Containers and Pods](https://kubernetes.io/docs/tasks/configure-pod-container/assign-memory-resource) in the Kubernetes documentation for more information. </dd>
<dd><br/>If you attempt to make changes to this value and re-apply it to an existing cluster, it immediately re-creates existing pods causing service interruptions.</dd>
<dd><br/>**Note:** If you do not want to specify a memory limit, comment-out or remove the `memory:` keyword from the YAML file.</dd>

<dt>`cpu: <cpu-limit>`</dt>
<dd>(Optional.) The amount of CPU resources allocated to each Greenplum PXF pod, specified as a Kubernetes CPU unit (for example, `cpu: "1.2"`). If omitted, the pod has no upper bound on the CPU resource it can use or inherits the default limit if one is specified in its deployed namespace.  See [Assign CPU Resources to Containers and Pods](https://kubernetes.io/docs/tasks/configure-pod-container/assign-cpu-resource/) in the Kubernetes documentation for more information.</dd>
<dd><br/>If you attempt to make changes to this value and re-apply it to an existing cluster, it re-creates existing pods causing service interruptions.</dd>
<dd><br/>**Note:** If you do not want to specify a cpu limit, comment-out or remove the `cpu:` keyword from the YAML file.</dd>

<dt><a id="workerSelector"></a>`workerSelector: <map of key-value pairs>`</dt>
<dd>(Optional.) One or more [selector labels](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) to use for choosing Greenplum PXF pods. Specify one or more label-value pairs to constrain Greenplum PXF pods to nodes having the matching labels. Define the selector labels as you would for a pod's `nodeSelector` attribute. If a `workerSelector` is not desired, remove the `workerSelector` attribute from the manifest file. </dd>
<dd><br/>For example, consider the case where you assign the label `worker=gpdb-pxf` to one or more pods using the command:

``` bash
$ kubectl label node <node_name> worker=gpdb-pxf
```

With the above labels present in your cluster, you would edit the Greenplum Operator manifest file to specify the same key-value pairs in the `workerSelector` attribute. This shows the relevant excerpt from the manifest file:

``` yaml
    ...
    workerSelector: {
      worker: "gpdb-pxf"
    }
    ...
```
</dd>
<dd><br/>This value cannot be dynamically changed for an existing cluster.  If you update this value, it recreates the Greenplum PXF cluster for the new value to take effect.</dd>

<dt><a id="pxfConf"></a>`pxfConf: <s3Source>`</dt>
<dd>(Optional.) Specifies an S3 location (endpoint, bucket, and folder) and secrets file to use for downloading an existing PXF configuration for use with a new <%=vars.product_name %> cluster deployment. The Greenplum Operator copies the contents of the S3 location to each Greenplum segment host for use as the `PXF_CONF` directory (`/etc/pxf`). You must ensure that the bucket-folder path contains the complete directory structure and customized files for one or more PXF server configurations. See [Deploying PXF with the Default Configuration](deploy-pxf.html#initialize) for information about deploying Greenplum with a default, initialized PXF configuration directory that you can customize for accessing your data sources.</dd>

<dt>`s3Source`</dt>
<dd>This section contains all of the S3-related attributes required to access the PXF configuration directory.</dd>

<dt>`secret: <string>`</dt>
<dd>The name of a secret file the Greenplum Operator uses to access the S3 location for copying the PXF configuration. For example:

``` bash
$ kubectl create secret generic my-greenplum-pxf-configs --from-literal=‘access_key_id=<accessKey>’ --from-literal=‘secret_access_key=<secretKey>’
```
</dd>
<dd><br/>The above command creates a secret file named `my-greenplum-pxf-configs` using the S3 access and secret keys that you provide. Replace `<accessKey>` and `<secretKey>` with the actual S3 access and secret key values for your system. If necessary, use your S3 implementation documentation to generate a secret access key. 
</dd>

<dt>`endpoint: <string>`</dt>
<dd>The URL of the S3 provider. For example, if you are pulling the PXF configuration from AWS S3, use the endpoint "s3.amazonaws.com".</dd>

<dt>`protocol: <http|https>`</dt>
<dd>(Optional.) The protocol to use for connecting to the specified S3 endpoint when downloading files from the bucket and folder. The default is `https`.</dd>

<dt>`bucket: <string>`</dt>
<dd>The S3 bucket name that contains the PXF configuration.</dd>

<dt>`folder: <string>`</dt>
<dd>(Optional.) The folder name in the S3 bucket where the PXF configuration directory is stored.</dd>

## <a id="examples"></a>Examples

See the `workspace/samples/my-gp-with-pxf-instance.yaml` file for an example manifest that configures the PXF resource.

## <a id="seealso"></a>See Also

[Deploying PXF with Greenplum](deploy-pxf.html)
