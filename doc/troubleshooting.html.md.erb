---
title: Troubleshooting Common Problems
---

## <a id='debug'></a>Enabling Debug Logging

By default, <%=vars.product_name_long %> logs `info` level messages. You can obtain more detailed log messages for certain problems by changing the log level to `debug`. Note that changes to the logging level must be applied before the Greenplum Operator is installed.

To change the log level:

1. Go to the `operator` subdirectory of your <%=vars.product_name %> software directory. For example:

    ``` bash
    $ cd ~/greenplum-for-kubernetes-*/operator
    ```
  
2. Open the `values.yaml` file in a text editor.

3. To change the default log level to `debug`, add the following line to the end of the file:

    ``` yaml
    logLevel: debug
    ```

    To revert to default logging, either remove this line or change it to read `logLevel: info`

4. Install the Greenplum Operator to use the new logging level.

## <a id='nocgroups'></a>Read-Only File System Error

**Symptom:**

The command `kubectl logs <pod-name>` shows the error: 

``` bash
install: cannot create directory '/sys/fs/cgroup/devices/kubepods': Read-only file system
```

**Resolution:**

The <%=vars.product_name %> deployment process requires the ability to map the host system's `/sys/fs/cgroup` directory onto each container's `/sys/fs/cgroup`. Ensure that no kernel security module (for example, AppArmor) uses a profile that disallows mounting `/sys/fs/cgroup`.

## <a id='nodenotlabeled'></a>Node Not Labeled

**Symptom:**

```
node "gke-gpdb-test-default-pool-20a900ca-3trh" not labeled
```

**Resolution:**

This is a common output from GCP. It indicates that the node is *already* labeled correctly, so no label action was necessary.

## <a id='exenotfound'></a>Executable Not Found

**Symptom:**

```
executable not found in $PATH
```

**Resolution:**

This error appears on the events tab of a container to indicate that xfs is not supported on Container OS (COS). To resolve this, use Ubuntu OS on the node.

## <a id='sandboxchanged'></a>Sandbox Has Changed

**Symptom:**

```
Sandbox has changed
```

**Resolution:**

This error appears on the events tab of a container to indicate that `sysctl` settings have become corrupted or failed. To resolve this, remove the `sysctl` settings from pod YAML file.

## <a id='connectiontimed'></a>Connection Timed

**Symptom:**

```
 getsocket() connection timed 
```

**Resolution:**

This error can occur when accessing http://localhost:8001/ui, a `kubectl` proxy address. Make sure that there is a connection between the master  and worker nodes where kubernetes-dashboard is running. A network tag on the nodes like `gpcloud-internal` can establish a route among the nodes.

## <a id='unabletoconnect'></a>Unable to Connect to the Server

**Symptom:**

```bash
kubectl get nodes -w
Unable to connect to the server: x509: certificate is valid for 10.100.200.1, 35.199.191.209, not 35.197.83.225
```

**Resolution:**

This error indicates that you have chosen to update the wrong Load Balancer.  Each cluster has its own load balancer for the Kubernetes master, with a certificate for access.  Refer to the `workspace/samples/scripts/create_pks_cluster_on_gcp.bash` script for Bash commands that help to determine the master IP address for a given cluster name, and the commands used to attach to a Load Balancer.

## <a id='permissiondeniedwhengpstop'></a>Permission Denied Error when Stopping Greenplum

**Symptom:**

```bash
kubectl exec -it master bash
$ gpstop -u

#
# Exits with Permission denied error
#
.
20180828:14:34:53:002448 gpstop:master:gpadmin-[CRITICAL]:-Error occurred: Error Executing Command:
 Command was: 'ssh -o 'StrictHostKeyChecking no' master ". /usr/local/greenplum-db/greenplum_path.sh; $GPHOME/bin/pg_ctl reload -D /greenplum/data-1"'
rc=1, stdout='', stderr='pg_ctl: could not send reload signal (PID: 2137): Permission denied'
```

**Resolution:**

This error occurs because of the ssh context that Docker uses. Commands that are issued to a process have to use the same context as the originator of the process. This issue is fixed in recent Docker versions, but the fixes have not reached the latest kubernetes release. To avoid this issue, use the same ssh context that you used to initialize the Greenplum cluster. For example, if you used a `kubectl` session to initialize Greenplum, then use another `kubectl` session to run `gpstop` and stop Greenplum.

## <a id='toomanyopenfiles'></a>Socket: too many open files

**Symptom:**
Executing any `kubectl` command yields an error similar to:
```
dial udp 1.2.3.4:53: socket: too many open files
```

**Resolution:**

Configure the underlying node to support a larger number of files. See [Files](prepare-k8s.html#files) in the Node Requirements documentation.

## <a id='helm-no-matches-for-crd'></a>No matches for CustomResourceDefinition on helm install/upgrade

**Symptom**

When using a version of Greenplum for Kubernetes that is v2.0.0 or above with Kubernetes version earlier than 1.16, helm install/upgrade fails with the following error:
```bash
unable to recognize "": no matches for kind "CustomResourceDefinition" in version "apiextensions.k8s.io/v1"
```

**Resolution**

Upgrade Kubernetes to version 1.16 or higher and redeploy the operator.

## <a id='regsecret-imagepullerror'></a>Cannot pull images

**Symptom**

When you try to deploy the Greenplum Operator using `helm` you see an error similar to:

``` bash
$ helm install greenplum-operator -f workspace/operator-values-overrides.yaml operator/
$ kubectl describe pod -l app=greenplum-operator
```
``` bash
...
Events:
  Type     Reason          Age              From                                 Message
  ----     ------          ----             ----                                 -------
  Normal   Scheduled       1m               default-scheduler                    Successfully assigned default/greenplum-operator-79bd8ccbc4-4lbxx to gke-oz-acceptance-default-pool-c7870f59-6h3f
  Normal   Pulling         1m (x2 over 1m)  kubelet, default-pool-c7870f59-6h3f  pulling image "greenplum-operator:v0.6.0.dev.103.gadfb9a1"
  Warning  Failed          1m (x2 over 1m)  kubelet, default-pool-c7870f59-6h3f  Failed to pull image "greenplum-operator:v0.6.0.dev.103.gadfb9a1": rpc error: code = Unknown desc = Error response from daemon: repository greenplum-operator not found: does not exist or no pull access
  Warning  Failed          1m (x2 over 1m)  kubelet, default-pool-c7870f59-6h3f  Error: ErrImagePull
  Normal   SandboxChanged  1m (x7 over 1m)  kubelet, default-pool-c7870f59-6h3f  Pod sandbox changed, it will be killed and re-created.
  Normal   BackOff         1m (x6 over 1m)  kubelet, default-pool-c7870f59-6h3f  Back-off pulling image "greenplum-operator:v0.6.0.dev.103.gadfb9a1"
  Warning  Failed          1m (x6 over 1m)  kubelet, default-pool-c7870f59-6h3f  Error: ImagePullBackOff
```

**Resolution**

This error indicates that the created `regsecret` does not have permission to pull images from the specified container registry or the secret was never created.

Make sure the `regsecret` docker-registry secret is created and contains appropriate privileges to fetch images.

## <a id='gpinitsystem-failure'></a>Failures During Initialization (Pending State)

**Symptom**

When creating a Greenplum instance, initialization can fail due to unavailable resources. For example:

- The number of available nodes may not be able to accommodate the specified segment count.
- Required labels may not be available on nodes, as specified in the configuration.

In either of the above cases, the segment pods will remain in `Pending` state until the required resources become available.

**Resolution**

- Try to increase number of nodes in your existing Kubernetes cluster. For example, add more node pools in your GKE cluster.
- If you cannot increase the node size, then reduce the CPU or memory resources allocated to each pod in the manifest file. Then, delete the existing GreenplumCluster, including PVCs, and re-apply the configuration.
- If enough nodes are available but pods remain in `Pending` state, verify that your nodes are properly labeled to match your deployment specification.

## <a id='gpexpand-failure'></a>Failures During Expansion (Pending State)

**Symptom**

When you try to expand a Greenplum instance, the expansion can fail due to unavailable resources. For example:

- The number of available nodes may not be able to accommodate the new segment count.
- Required labels may not be available on nodes, as specified in the configuration.

In either of the above cases, the segment pods will remain in `Pending` state until the required resources become available.

**Resolution**

- Try to increase number of nodes in your existing Kubernetes cluster. For example, add more node pools in your GKE cluster.
- If enough nodes are available but pods remain in `Pending` state, verify that your nodes are properly labeled to match your deployment specification.
- After the new or reconfigured segment pods are up and `Running`, perform these steps to recover from the failed expansion process:
    1. Execute `kubectl exec -it master-0 bash` to log into the master pod.
    2. Execute the rollback command `gpexpand -r -D gpadmin` to rollback any changes.
    3. Restart the cluster with `gpstart -a` if the Greenplum cluster has already stopped.
    4. Execute the command `gpexpand -D gpadmin -i /tmp/gpexpand_config` to perform the Greenplum expansion.
    5. Logout of `master-0` and run `kubectl delete job.batch/my-greenplum-gpexpand-job` to delete the failed `gpexpand` jobs.

## <a id='gpexpand-failure'></a>Failures During Installation After An Incomplete Uninstallation

**Symptom**

After reinstalling the greenplum operator and attempting to create a greenplum cluster, you see this error:

```
error: unable to recognize "workspace/my-gp-instance.yaml": no matches for kind "GreenplumCluster" in version "greenplum.pivotal.io/v1"
```

This might happen if you uninstalled the greenplum operator without deleting all previously deployed greenplum cluster resources. To avoid this problem in the future, make sure to follow all the steps from [Uninstalling  <%=vars.product_name_long %>](uninstalling.html)

**Resolution**

Uninstall the greenplum operator helm chart and reinstall it.

## <a id='troubleshootingpks'></a>VMware Tanzu Kubernetes Grid Integrated (TKGI) Edition Deployment Errors

### <a id='pks-packet'></a>Greenplum Query Fails to Write an Outgoing Packet

**Symptom:**

The Greenplum cluster is initialized and running, but a query returns an error similar to:

``` SQL
ERROR: Interconnect error writing an outgoing packet: Invalid argument (seg0 slice1 <ip>:<port> pid=1193)
```

**Resolution:**

This error occurs when ports are not garbage collected quickly enough. The problem is commonplace in systems that have many containers on a single kubernetes node, and the containers heavily use different ports to communicate with one another (as is the case with Greenplum segments).

To work around this problem, set the following `sysctl` attribute on the worker nodes:

``` bash
net.ipv4.neigh.default.gc_thresh1 = 30000
```

### <a id='pks-401'></a>Authorization Errors

**Symptom**

After certificate change (after any URL change for UAA domain name), you may see a `401` error from BOSH similar to:

```bash
bosh -e pks vms
Using environment '192.168.101.10' as anonymous user

Finding deployments:
  Director responded with non-successful status code '401' response 'Not authorized: '/deployments'
'

Exit code 1
```

**Resolution**

Go to the credentials web page (similar to `https://<ops manager>/infrastructure/director/credentials`) and look for Bosh command line credentials. The credentials look similar to:

```json
{"credential":"BOSH_CLIENT=<Some User> BOSH_CLIENT_SECRET=<Some Secret> BOSH_CA_CERT=/var/tempest/workspaces/default/root_ca_certificate BOSH_ENVIRONMENT=192.168.101.10 bosh "}
```

In the command for `uaac token owner get`, use:
 * BOSH_CLIENT as the "Client ID" 
 * BOSH_CLIENT_SECRET as "Client secret" 
 * "Admin" for User name 
 * associated password from Uaa Admin User Credentials
 
For example:

```bash
$ uaac token owner get
Client ID:  ops_manager
Client secret:  ********************************
User name:  Admin
Password:  ********************************
```

### <a id='pks-uaa'></a>Cannot Access UAA

**Symptom**

You can access Ops Manager, but you have problems accessing UAA. For example:

```bash
pks login -a https://pks-0.gpcloud.gpdb.pivotal.io:9021 -u dummy -p <some password> -k

Error: Post https://pks-0.gpcloud.gpdb.pivotal.io:8443/oauth/token: dial tcp 35.197.67.138:8443: getsockopt: connection refused
```

**Resolution**

This problem can be a symptom of having recycled the VM running the TKGI API, such that the external IP address defined in the
domain name is old.  Use `gcloud` or Google Cloud Console to determine the current VM for the VMware Tanzu Kubernetes Grid Integrated (TKGI) Edition API.  You can
distinguish it because it has two labels, "job" and "instance-group", which both have values "pivotal container service".
Get the *external* IP address for this and change the DNS definition to that external IP address. Use a command like:

```bash
$ watch dig '<my domain name>'
```

Wait to see that the DNS entry is updated for your local workstation.  When it updates, try the `pks login` command again.

## <a id='troubleshootingpxf'></a>Troubleshooting GreenplumPXFService pod startup

**Note:** When diagnosing errors during GreenplumPXFService startup, it may help to resize the deployment to 1 replica to make it easier to find the most recent pod. Be sure to set the replicas back to the desired number (2 by default) once the issue has been resolved.

### <a id='pxf-createcontainerconfigerror'></a>CreateContainerConfigError state

**Symptom**

When you attempt to deploy PXF, the GreenplumPXFService pods go into state `CreateContainerConfigError` and you see an error similar to:

``` bash
kubectl describe pod -l app=greenplum-pxf
```
``` bash
Events:
  Type     Reason     Age                   From               Message
  ----     ------     ----                  ----               -------
  Normal   Scheduled  2m48s                 default-scheduler  Successfully assigned default/my-greenplum-pxf-8487655d9b-sxxsh to minikube
  Warning  Failed     49s (x12 over 2m47s)  kubelet, minikube  Error: secret "my-greenplum-pxf-config" not found
  Normal   Pulled     38s (x13 over 2m47s)  kubelet, minikube  Container image "greenplum-for-kubernetes:latest" already present on machine

Warning  Failed     49s (x12 over 2m47s)  kubelet, minikube  Error: secret "my-greenplum-pxf-config" not found
```

**Resolution**

This is likely caused by one of two issues:

- The secret name is not specified correctly in the GreenplumPXFService manifest file.
- The secret does not exist in the Kubernetes API.

If it is the first issue, correct the manifest file and re-apply.

If it is the second, create the necessary secret. In this case you will need to delete the GreenplumPXFService before re-applying to ensure pods will be able to locate the new secret.

### <a id='pxf-conf-download-failure'></a>Failure to download PXF configuration

**Symptom**

When you attempt to deploy PXF, GreenplumPXFService pods go into state `Error` or `CrashLoopBackOff`

When you check the logs for the GreenplumPXFService pod, you will see an error similar to

``` bash
{"level":"INFO","ts":"2020-03-07T00:19:58.539Z","logger":"startPXF","msg":"downloading PXF configuration", "bucket", "my-bucket", "folder", "my-folder"}
{"level":"ERROR","ts":"2020-03-07T00:19:59.089Z","logger":"startPXF","msg":"PXF failed to start","error", <error message>
```

the exact error message will vary.

**Resolution**

This indicates that the PXF pod was unable to download configuration objects from S3

- Check the error message to discover what went wrong. Some example error messages are:
> "error":"s3 list objects: Access denied."
>
> "error":"s3 list objects: The request signature we calculated does not match the signature you provided. Check your Google secret key and signing method."
>
> "error":"s3 list objects: The specified bucket does not exist."
>
> "error":"no objects found in pxfConf.s3Source location: <bucket/folder>"
- Check to make sure you are using the correct key to access your S3 data.
- Check the permissions on your bucket and files.
- Check your manifest for your GreenplumPXFService. Make sure the `pxfConf.s3Source` field contains the correct values for `endpoint`, `bucket`, and `folder`.

If the solution to the error involved editing the GreenplumPXFService manifest, simply re-apply the file and the pods will be recreated.

If the solution did not involve editing the manifest, such as changing the secret or making changes in S3, delete the GreenplumPXFService and re-apply to see the changes.

### <a id='pxf-conf-warning'></a>Warning from PXF conf directory structure

**Symptom**

When you check the logs of a GreenplumPXFService pod, you see the following warning:

> warning: PXF S3 conf data doesn't look like a PXF conf directory. Please check pxfConf.s3Source setting


**Resolution**

This warning is emitted when the PXF pod is able to download objects from S3, but they did not follow the expected
structure of a PXF configuration directory, as described [in the PXF documentation](https://gpdb.docs.pivotal.io/6-3/pxf/about_pxf_dir.html#usercfg)

- Check your manifest for your GreenplumPXFService. Make sure the `pxfConf.s3Source` field contains the correct values for `bucket` and `folder`.
- Check the contents of S3. They should contain at least one of the following directories:
  - `conf/`
  - `keytabs/`
  - `lib/`
  - `logs/`
  - `servers/`
  - `templates/`
