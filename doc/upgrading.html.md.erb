---
title: Upgrading Greenplum for Kubernetes
---

This topic describes how to upgrade  <%=vars.product_name_long %> from version 2.x to a subsequent version 2.y release. The upgrade process involves first deleting any existing Greenplum cluster deployments, and then upgrading the Greenplum Operator to the latest version. You then use the new Greenplum Operator to re-create earlier cluster deployments, using the same manifest files. During this process, you re-use any existing persistent volumes so that Greenplum cluster data is preserved.  

## Prerequisites

- This procedure assumes that you have installed a previous minor version Greenplum for Kubernetes (version 2.x). Greenplum for Kubernetes supports upgrades only from the prior minor version (for example, from version 2.0 to version 2.1). If your installed version of the product is more than one version prior, you will need to upgrade incrementally to reach your target version, using the upgrade instructions associated with each new version. 

- Before you can install the newer Greenplum for Kubernetes version, ensure that you have installed  all required software and prepared your Kubernetes environment as described in [Prerequisites](prepare-k8s.html).

## Procedure

Follow these steps to upgrade the Greenplum Operator resource:

1. Navigate to the `workspace` directory of the Greenplum for Kubernetes installation (or to the location of the Kubernetes manifest that you used to deploy the cluster). For example:

    ``` bash
    $ cd ./greenplum-for-kubernetes-*/workspace
    ```

2. Execute the `kubectl delete` command, specifying the manifest that you used to deploy the cluster. For example:

    ``` bash
    $ kubectl delete -f ./my-gp-instance.yaml
    ```

    `kubectl` stops the Greenplum for Kubernetes instance and deletes the kubernetes resources for the Greenplum deployment.

3. Use `kubectl` to monitor the progress of terminating Greenplum resources in your cluster. For example, if your cluster deployment was named `my-greenplum`:

    ``` bash
    $ kubectl get all -l greenplum-cluster=my-greenplum
    ```
    ``` bash
    NAME                                     READY     STATUS        RESTARTS   AGE
    pod/greenplum-operator-7b5ddcb79-vnwvc   1/1       Running       0          9m
    pod/master-0                             0/1       Terminating   0          5m
    pod/segment-a-0                          0/1       Terminating   0          5m
    pod/segment-a-1                          0/1       Terminating   0          5m
    pod/segment-b-0                          0/1       Terminating   0          5m

    NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
    service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   26m

    NAME                                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/greenplum-operator   1         1         1            1           9m

    NAME                                           DESIRED   CURRENT   READY     AGE
    replicaset.apps/greenplum-operator-7b5ddcb79   1         1         1         9m
    ```

4. The deletion process is complete when the segment pods are no longer available (and no resources remain):

    ``` bash
    $ kubectl get all -l greenplum-cluster=my-greenplum
    ```
    ``` bash
    No resources found in default namespace.
    ```

    The Greenplum Operator should remain:

    ``` bash
    $ kubectl get all
    ```
    ``` bash
    NAME                                     READY     STATUS    RESTARTS   AGE
    pod/greenplum-operator-7b5ddcb79-vnwvc   1/1       Running   0          34m
    
    NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
    service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   50m
    
    NAME                                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/greenplum-operator   1         1         1            1           34m
    
    NAME                                           DESIRED   CURRENT   READY     AGE
    replicaset.apps/greenplum-operator-7b5ddcb79   1         1         1         34m
    ```

5. Repeat the previous steps for each version cluster that you deployed.

6. Download the new Greenplum for Kubernetes software from [VMware Tanzu Network](https://network.pivotal.io/products/greenplum-for-kubernetes). The download file has the name: `greenplum-for-kubernetes-<version>.tar.gz`.

8. Go to the directory where you downloaded the new version of Greenplum for Kubernetes, and unpack the downloaded software. For example:

    ```bash
    $ cd ~/Downloads
    $ tar xzf greenplum-for-kubernetes-*.tar.gz
    ```

    The above command unpacks the distribution into a new directory named `greenplum-for-kubernetes-<version>`.

9. Go into the new `greenplum-for-kubernetes-<version>` directory:

    ``` bash
    $ cd ./greenplum-for-kubernetes-*
    ```

10. For Minikube deployments only, ensure that the local docker daemon interacts with the Minikube docker container registry:

    ``` bash
    $ eval $(minikube docker-env)
    ```

    **Note:** To undo this docker setting in the current shell, run `eval "$(docker-machine env -u)"`.

11. Load the new Greenplum Operator Docker image to the Docker registry. For example:

    ``` bash
    $ docker load -i ./images/greenplum-operator
    ```
    ```
    96eda0f553ba: Loading layer [==================================================>]  127.6MB/127.6MB
    bd95983a8d99: Loading layer [==================================================>]  11.78kB/11.78kB
    293b479c17a5: Loading layer [==================================================>]  15.87kB/15.87kB
    fa1693d66d0b: Loading layer [==================================================>]  3.072kB/3.072kB
    5d5f706821ef: Loading layer [==================================================>]  92.66MB/92.66MB
    3c59653fa97d: Loading layer [==================================================>]   40.5MB/40.5MB
    4457bbd6a302: Loading layer [==================================================>]  40.85MB/40.85MB
    Loaded image: greenplum-operator:v1.12.0
    ```

12. Load the Greenplum for Kubernetes Docker image to the local Docker registry as well. For example:

    ``` bash
    $ docker load -i ./images/greenplum-for-kubernetes
    ```

    ``` bash
    b04caf3e714d: Loading layer [==================================================>]  114.7MB/114.7MB
    5d990f175f75: Loading layer [==================================================>]  12.31MB/12.31MB
    029e8dc80732: Loading layer [==================================================>]  7.226MB/7.226MB
    e7a72a352b50: Loading layer [==================================================>]  744.3MB/744.3MB
    ...
    Loaded image: greenplum-for-kubernetes:v1.12.0
    ```

13. Verify that the new Docker images are now available:

    ```bash
    $ docker images "greenplum-*"
    ```

    ``` bash
    REPOSITORY                 TAG                 IMAGE ID            CREATED             SIZE
    greenplum-operator         v1.12.0             e82305e017a4        45 hours ago        295MB
    greenplum-for-kubernetes   v1.12.0             647cef39bbd8        45 hours ago        3.15GB
    greenplum-operator         v1.11.0             852cafa7ac90        7 weeks ago         269MB
    greenplum-for-kubernetes   v1.11.0             3819d17a577a        7 weeks ago         3.05GB
    ```

14. **(Skip this step if you are using local docker images such as on Minikube.)** If you push the Greenplum for Kubernetes docker images to a different container registry:

    1.  If you want to push the Greenplum for Kubernetes docker images to a different container registry, set the project name and image repo name and then use Docker to push the images. For example, to push the images to Google Cloud Registry using the current Google Cloud project name:

        ```bash
        $ gcloud auth configure-docker
        
        $ PROJECT=$(gcloud config list core/project --format='value(core.project)')
        $ IMAGE_REPO="gcr.io/${PROJECT}"
        
        $ GREENPLUM_IMAGE_NAME="${IMAGE_REPO}/greenplum-for-kubernetes:$(cat ./images/greenplum-for-kubernetes-tag)"
        $ docker tag $(cat ./images/greenplum-for-kubernetes-id) ${GREENPLUM_IMAGE_NAME}
        $ docker push ${GREENPLUM_IMAGE_NAME}
        
        $ OPERATOR_IMAGE_NAME="${IMAGE_REPO}/greenplum-operator:$(cat ./images/greenplum-operator-tag)"
        $ docker tag $(cat ./images/greenplum-operator-id) ${OPERATOR_IMAGE_NAME}
        $ docker push ${OPERATOR_IMAGE_NAME}
        ```

    2. Copy the values yaml file from your existing deployment, or create a new YAML file in the workspace subdirectory as shown below:
    
        ``` bash
        $ touch workspace/operator-values-overrides.yaml
        ```

    3. If you pushed the Greenplum Operator and Greenplum Database Docker images to a container registry, add two additional lines to the configuration file to indicate the registry where you pushed the images. For example, if you are using Google Cloud Registry with a project named "gp-kubernetes", you would add the properties:

        ```yaml
        operatorImageRepository: gcr.io/gp-kubernetes/greenplum-operator
        greenplumImageRepository: gcr.io/gp-kubernetes/greenplum-for-kubernetes
        ```

        **Note:** If you did not tag the images with a container registry prefix or project name (for example, if you are using your own local Minikube deployment), then you can skip this step.

15. Use `helm` to upgrade the Greenplum Operator to the latest release, specifying a customized YAML configuration file if you created one. For example:

    ```bash
    $ helm upgrade greenplum-operator -f workspace/operator-values-overrides.yaml operator/
    ```

    If you did not create a YAML configuration file (as in the case with Minikube) omit the `-f` option:

    ``` bash
    $ helm upgrade greenplum-operator operator/
    ```

    Helm upgrades to the new release in the Kubernetes namespace specified in the current Kubernetes context. If you want to upgrade a Greenplum Operator from a different namespace, include the `--namespace` option in the `helm` command.
    
    <br/>The command displays the following message and concludes with a link to this documentation:

    ``` bash
    Release "greenplum-operator" has been upgraded. Happy Helming!
    NAME: greenplum-operator
    LAST DEPLOYED: Wed Feb 12 11:41:47 2020
    NAMESPACE: default
    STATUS: deployed
    REVISION: 2
    TEST SUITE: None
    NOTES:
    greenplum-operator has been installed.

    Please see documentation at:
    http://greenplum-kubernetes.docs.pivotal.io/
    ```

17. Use `watch kubectl get all -l app=greenplum-operator` to ensure that the Greenplum Operator pod is in the `Running` state. For example:

    ``` bash
    $ watch kubectl get all -l app=greenplum-operator
    ```
    ``` bash
    NAME                                    READY   STATUS    RESTARTS   AGE
    pod/greenplum-operator-ddbc89dc-v67ng   1/1     Running   0          116s

    NAME                                                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
    service/greenplum-validating-webhook-service-ddbc89dc-v67ng   ClusterIP   10.100.181.61   <none>        443/TCP   115s

    NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/greenplum-operator   1/1     1            1           64m

    NAME                                            DESIRED   CURRENT   READY   AGE
    replicaset.apps/greenplum-operator-667ccc59fd   0         0         0       64m
    replicaset.apps/greenplum-operator-ddbc89dc     1         1         1       116s
    ```

18. Return to the `workspace` directory that contains the manifest files you used to deploy your clusters (or copy those manifest files to the new Greenplum for Kubernetes version `workspace` directory.

19. Use `kubectl apply` and specify your manifest file to send the deployment request to the Greenplum Operator. For example, to use the sample `my-gp-instance.yaml` file:

    ``` bash
    $ kubectl apply -f ./my-gp-instance.yaml 
    ```
    ```bash
    greenplumcluster.greenplum.pivotal.io/my-greenplum created
    ```

    The Greenplum Operator deploys the necessary Greenplum resources according to your specification, using the existing Persistent Volume Claims as-is with their available data.

20. Use `kubectl get all -l greenplum-cluster=<your Greenplum cluster instance name>` and wait until all Greenplum cluster pods have the status `Running`:

    ``` bash
    $ watch kubectl get all -l greenplum-cluster=my-greenplum
    ```
    ``` bash
    NAME              READY   STATUS    RESTARTS   AGE
    pod/master-0      1/1     Running   0          34s
    pod/master-1      1/1     Running   0          34s
    pod/segment-a-0   1/1     Running   0          34s
    pod/segment-b-0   1/1     Running   0          34s

    NAME                TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
    service/agent       ClusterIP      None            <none>        22/TCP           34s
    service/greenplum   LoadBalancer   10.107.25.227   <pending>     5432:32149/TCP   34s

    NAME                         READY   AGE
    statefulset.apps/master      2/2     34s
    statefulset.apps/segment-a   1/1     34s
    statefulset.apps/segment-b   1/1     34s
    ```

    At this point, the upgraded cluster is available. If you are using a persistent `storageClass`, the updated cluster is created with the same Persistent Volume Claims (PVCs) and data.

21. _If your cluster is configured to use a standby master_, connect to the `master-0` pod and execute the `gpstart` command manually. For example:

    ``` bash
    kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; gpstart"
    ```
    ``` bash
    20200212:19:45:55:000517 gpstart:master-0:gpadmin-[INFO]:-Starting gpstart with args: 
    20200212:19:45:55:000517 gpstart:master-0:gpadmin-[INFO]:-Gathering information and validating the environment...
    20200212:19:45:55:000517 gpstart:master-0:gpadmin-[INFO]:-Greenplum Binary Version: 'postgres (Greenplum Database) 5.24.2 build dev'
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

22. Describe your Greenplum cluster to observe that the update succeeded.

    ``` bash
    $ kubectl describe greenplumClusters/my-greenplum
    ```
    ``` bash
    Name:         my-greenplum
    Namespace:    default
    Labels:       <none>
    Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                    {"apiVersion":"greenplum.pivotal.io/v1","kind":"GreenplumCluster","metadata":{"annotations":{},"name":"my-greenplum",   "namespace":"default"...
    API Version:  greenplum.pivotal.io/v1
    Kind:         GreenplumCluster
    Metadata:
      Creation Timestamp:  2020-02-12T19:44:19Z
      Finalizers:
        stopcluster.greenplumcluster.pivotal.io
      Generation:        3
      Resource Version:  6167
      Self Link:         /apis/greenplum.pivotal.io/v1/namespaces/default/greenplumclusters/my-greenplum
      UID:               204c1dc7-f7b8-4fae-a784-3cc06579211a
    Spec:
      Master And Standby:
        Anti Affinity:              no
        Cpu:                        0.5
        Host Based Authentication:  # host   all   gpadmin   1.2.3.4/32   trust
    # host   all   gpuser    0.0.0.0/0   md5

        Memory:              800Mi
        Storage:             1G
        Storage Class Name:  standard
        Worker Selector:
      Segments:
        Anti Affinity:          no
        Cpu:                    0.5
        Memory:                 800Mi
        Primary Segment Count:  1
        Storage:                2G
        Storage Class Name:     standard
        Worker Selector:
    Status:
      Instance Image:    greenplum-for-kubernetes:v1.12.0
      Operator Version:  greenplum-operator:v1.12.0
      Phase:             Running
    Events:              <none>
    ```

    The `Phase` should be `Running` and the images should show the latest version.