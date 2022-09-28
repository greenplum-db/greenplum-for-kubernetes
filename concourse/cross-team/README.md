## Cross Team scripts

### Concourse

#### Provision cluster and deploy GPDB
We provide the `provision-cluster` task yaml and associated bash script for you to consume in your pipeline.  This task will provision a cluster and deploy GPDB, provided that you request access to the Greenplum For Kubernetes team gcr.io registry.  Feel free to reach out the gp-kubernetes slack channel for help.  The task yaml is shown below:


```yaml
platform: linux

image_resource:
  type: registry-image

inputs:
- name: greenplum-for-kubernetes
- name: greenplum-for-kubernetes-pcf-manifest  # Only needed if provisioning a PKS cluster
- name: gp-kubernetes-rc-release  # Only needed if provisioning a PKS cluster

run:
  path: greenplum-for-kubernetes/concourse/cross-team/provision-cluster.bash

params:
  MACHINE_TYPE: # Defaults to n1-standard-2
  NODE_COUNT: # Defaults to 2xSEGMENT_COUNT + 2
  SEGMENT_COUNT: # Defaults to 1 primary segment pod & 1 mirror segment pod
  GP4K_VERSION: # Defaults to latest green
  CPU_RATE_LIMIT: # Defaults CPU_RATE_LIMIT for default resource_group to 70
  MEMORY_LIMIT: # Defaults MEMORY_LIMIT for default resource_group to 70
  CIDR_RANGE: # Defaults to empty
  CLUSTER_LOAD_BALANCER: # Required for PKS clusters; Optional for GKE
  NETWORK: # Uses the default network if not specified
  MASTERS_LABEL: # Does not label master nodes if not specified
  SEGMENTS_LABEL: # Does not label segment nodes if not specified
  GP_INSTANCE_NAME: required # Overrides the greenplum cluster name  cluster_config if specified
  KUBEENV: required # either GKE or PKS
  CLUSTER_NAME: required # GKE cluster name
  GCP_SVC_ACCT_KEY: required
  GCP_PROJECT: required
  GCP_ZONE:
  CLUSTER_CONFIG: |
    apiVersion: "greenplum.pivotal.io/v1"
    kind: "GreenplumCluster"
    metadata:
      name: # Use the `GP_INSTANCE_NAME` parameter to set this
    spec:
      masterAndStandby:
        hostBasedAuthentication: |
          # host   all   gpadmin   1.2.3.4/32   trust
          # host   all   gpuser    0.0.0.0/0   md5
        memory: "800Mi"
        cpu: "0.5"
        storageClassName: standard
        storage: 1G
        antiAffinity: "yes"
        workerSelector: {}
      segments:
        primarySegmentCount: 1
        memory: "800Mi"
        cpu: "0.5"
        storageClassName: standard
        storage: 2G
        antiAffinity: "yes"
        workerSelector: {}
```

Using the `CLUSTER_CONFIG` parameter, you can specify the Greenplum Cluster manifest that is currently supported by this release. You may also specify the Kubernetes cluster configurations to be provisioned using parameters such as `MACHINE_TYPE` and `NODE_COUNT`.  If you do not specify `GP4K_VERSION`, this task will deploy a Greenplum Cluster using the latest green development images from the Greenplum for Kubernetes team's concourse.

Note the parameter `CIDR_RANGE` is optional unless you need to connect to a host in a different gcp-project/network than the project default.

The parameters `MASTERS_LABEL` and `SEGMENTS_LABEL`, if specified, will automatically label the nodes in the cluster. These parameters allow you to take advantage of the `workerSelector` feature in the manifest. If `MASTERS_LABEL` is specified, the first two nodes will be labeled with the specified key/value. An example `MASTERS_LABEL` would be `"worker=master"`. This will apply the key `worker` and the value `master` to the first two nodes. The `SEGMENTS_LABEL` works the same way as the `MASTERS_LABEL` except it will label all the remaining nodes in the cluster besides the first two (those labeled with the `MASTERS_LABEL`).

This task also support deploying to both GKE and PKS clusters when provided with the necessary credentials. You can interpolate variables in concourse pipelines and inject them into the `CLUSTER_CONFIG` yaml as shown below:

```yaml
  CLUSTER_CONFIG: |
    apiVersion: "greenplum.pivotal.io/v1"
    kind: "GreenplumCluster"
    metadata:
      name: my-greenplum
    spec:
      masterAndStandby:
        hostBasedAuthentication: |
          # host   all   gpadmin   1.2.3.4/32   trust
          # host   all   gpuser    0.0.0.0/0   md5
        memory: "((pod-memory))"
        cpu: "((pod-cpu))"
        storageClassName: standard
        storage: "((pod-storage))"
      segments:
        primarySegmentCount: ((segment-count))
        memory: "((pod-memory))"
        cpu: "((pod-cpu))"
        storageClassName: standard
        storage: "((pod-storage))"
```
Concourse can only interpolate integers and strings so when using floats use \" \" e.g: "0.5" for relevant parameters.

**Note** Deleting the GKE cluster without deleting the Greenplum instance that was deployed on the cluster leaves behind orphaned cloud resources like load-balancers and disks that were being used by that particular Greenplum instance. In order to clean those up, refer to the [cleanup section](#cleanup) that can be used in a nightly cleanup job.

#### <a id="cleanup"/>Cleanup the deployed k8s resources

We provide the `cleanup-env` task yaml and associated bash script for you to consume in your pipeline to perform cleanup.  The task yaml is shown below:

```yaml
platform: linux

image_resource:
  type: registry-image

inputs:
- name: greenplum-for-kubernetes

run:
  path: greenplum-for-kubernetes/concourse/cross-team/cleanup-env.bash

params:
  # should be same as that of provision-cluster.yml
  KUBEENV: required
  CLUSTER_NAME: required # GKE cluster name
  GCP_SVC_ACCT_KEY: required
  GCP_PROJECT: required
  GP_INSTANCE_NAME: required

```
This task deletes the following:

1. Greenplum instance
1. All unused PVCs
1. Greenplum-operator
1. K8s cluster

### Custom Script

Note: The use of the custom script approach is discouraged due to the fact the behavior of GPDB deployment may change in future releases breaking your script.  Instead, we encourage you to use the concourse approach. 

#### Create cluster

1. Create a k8s cluster in your GCP account. Look at `greenplum-for-kubernetes/concourse/scripts/create-cluster.bash` as reference.

#### Deploy gpdb instance

1. Login and credentials need to be provided. For e.g. you could run below:

```bash
source greenplum-for-kubernetes/concourse/scripts/cloud_login.bash && \
auth_gcloud && \
gke_credentials
```
1. Edit the image repository in `greenplum-for-kubernetes/greenplum-operator/operator/values.yaml`. For e.g.

```bash
echo 'greenplumImageRepository: gcr.io/<your gcp project>/greenplum-for-kubernetes' >> greenplum-for-kubernetes/greenplum-operator/operator/values.yaml && \
echo 'operatorImageRepository: gcr.io/<your gcp project>/greenplum-operator' >> greenplum-for-kubernetes/greenplum-operator/operator/values.yaml
```
1. Install the operator with helm `helm install --wait greenplum-operator greenplum-for-kubernetes/greenplum-operator/operator`
1. edit the `greenplum-for-kubernetes/workspace/my-gp-instance.yaml` to get the desired cluster spec
1. Create a gpdb instance `kubectl create -f <path to edited my-gp-instance.yaml>`
1. Check if gpdb instance pods are running `kubectl get po`

#### <a id="cleanup"/>Cleanup custom deployment and  k8s resources

The gp-kubernetes team runs this in an `ensure` block in the concourse pipeline so that if jobs/tasks get interrupted, cleanup always occurs.

1. Delete the gpdb instance `kubectl delete -f <path to edited my-gp-instance.yaml>`
1. When the gpdb instance is deleted, it leaves behind the pvc used (by design -- this is so you can spin up another instance and re-use the pvc/pv). If you are done using the pvcs, you should clean up the pvcs and pvs as below:

```bash
kubectl delete pvc --all
kubectl delete pv --all
```

1. Delete the greenplum-operator `helm uninstall greenplum-operator`
1. Delete the k8s cluster once your pipeline tasks are complete  `greenplum-for-kubernetes/concourse/scripts/delete-k8s-cluster.bash`
1. We also run `greenplum-for-kubernetes/concourse/scripts/delete-old-unused-pvc-disks.bash`  and `greenplum-for-kubernetes/concourse/scripts/delete-orphaned-loadbalancers-gcp.bash` daily to cleanup any other resources that might escape the cleanup.