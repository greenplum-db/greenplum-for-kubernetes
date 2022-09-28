---
title: Deleting a Greenplum Cluster
---

This section describes how to delete the pods and other resources that are created when you deploy a Greenplum cluster to Kubernetes. Note that deleting these cluster resources does not automatically delete the Persistenv Volume Claims (PVCs) that the cluster used to stored data. This enables you to re-deploy the same cluster at a later time, to pick up where you left off. You can optionally delete the PVCs if you to create an entirely new (empty) cluster at a later time.

## <a id="delpods"></a>Deleting Greenplum Pods and Resources

Follow these steps to delete the Greenplum pods, services, and other objects, leaving the Persistent Volumes intact:

1. Navigate to the `workspace` directory of the <%=vars.product_name %> distribution (or to the location of the Kubernetes manifest that you used to deploy the cluster). For example:

    ``` bash
    $ cd ./greenplum-for-kubernetes-*/workspace
    ```

2. Execute the `kubectl delete` command, specifying the manifest that you used to deploy the cluster. For example:

    ``` bash
    $ kubectl delete -f ./my-gp-instance.yaml --wait=false
    ```

    `kubectl` stops the <%=vars.product_name %> instance and deletes the Kubernetes resources for the Greenplum deployment.
   
    <br/>**Note:** Use the optional `--wait=false` flag to return immediately without waiting for the deletion to complete.

3. Use `kubectl` to describe the Greenplum cluster to verify `Status.Phase` and `Events`:

    ```bash
    $ kubectl describe greenplumcluster my-greenplum
     [...]
     Status:
       Instance Image:    greenplum-for-kubernetes:latest
       Operator Version:  greenplum-operator:latest
       Phase:             Deleting
     Events:
       Type    Reason                    Age   From               Message
       ----    ------                    ----  ----               -------
       Normal  CreatingGreenplumCluster  3m    greenplumOperator  Creating Greenplum cluster my-greenplum in default
       Normal  CreatedGreenplumCluster   1m    greenplumOperator  Successfully created Greenplum cluster my-greenplum in default
       Normal  DeletingGreenplumCluster  6s    greenplumOperator  Deleting Greenplum cluster my-greenplum in default
    ```
    <br/>If for any reason stopping the Greenplum instance fails, you should see a warning message in the greenplum-operator logs as shown below:
    
    ``` bash
    $ kubectl logs -l app=greenplum-operator
    [...]
    {"level":"INFO","ts":"2020-01-24T19:03:22.874Z","logger":"controllers.GreenplumCluster","msg":"DeletingGreenplumCluster","name":"my-greenplum","namespace":"default"}
    {"level":"INFO","ts":"2020-01-24T19:03:23.068Z","logger":"controllers.GreenplumCluster","msg":"initiating shutdown of the greenplum cluster"}
    {"level":"INFO","ts":"2020-01-24T19:03:31.971Z","logger":"controllers.GreenplumCluster","msg":"gpstop did not stop cleanly. Please check gpAdminLogs for more info."}
    [...]
    {"level":"INFO","ts":"2020-01-24T19:03:32.252Z","logger":"controllers.GreenplumCluster","msg":"DeletedGreenplumCluster","name":"my-greenplum","namespace":"default"}
    ```
    <br/>However, the Greenplum instance still gets deleted and all associated resources get cleaned up.

4. Use `kubectl` to monitor the progress of terminating Greenplum resources in your cluster. For example, if your cluster deployment was named `my-greenplum`:

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

5. The deletion process is complete when the segment pods are no longer available (and no resources remain):

    ``` bash
    $ kubectl get all -l greenplum-cluster=my-greenplum
    ```
    ``` bash
    No resources found in default namespace.
    ```

    If the Kubernetes resources remain and do not enter the "Terminating" status after 5 minutes, check the operator logs for error messages.
    ``` bash
    $ kubectl logs -l app=greenplum-operator
    ```

    The Greenplum Operator should remain for future deployments:

    ``` bash
    $ kubectl get all
    ```
    ``` bash
    NAME                                      READY   STATUS    RESTARTS   AGE
    pod/greenplum-operator-6ff95b6b79-kq9vr   1/1     Running   0          68m

    NAME                                                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
    service/greenplum-validating-webhook-service-6ff95b6b79-kq9vr   ClusterIP   10.106.60.103   <none>        443/TCP   68m
    service/kubernetes                                              ClusterIP   10.96.0.1       <none>        443/TCP   72m

    NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/greenplum-operator   1/1     1            1           68m

    NAME                                            DESIRED   CURRENT   READY   AGE
    replicaset.apps/greenplum-operator-6ff95b6b79   1         1         1       68m
    ```


## <a id="delpvs"></a>Deleting Greenplum Persistent Volume Claims

Deleting the Greenplum pods and other resources does not delete the associated persistent volume claims that were created for it. This is expected behavior for a Kubernetes cluster, as it gives you the opportunity to access or back up the data. If you no longer have any use for the Greenplum volumes (for example, if you want to install a brand new cluster), then follow this procedure to delete the Persistent Volume Claims (PVCs) and Persistent Volumes (PVs). 

You may also need to delete Greenplum PVCs if you want to change the storage size or deploy an existing cluster to a different storage class; the Greenplum Operator maintains an association between the cluster name and its PVCs, so you cannot redeploy a cluster to a different storage class or change the storage size without first deleting both the cluster deployment and its PVCs.

<br/>**Caution:** If the Persistent Volumes were created using dynamic provisioning, then deleting the PVCs will also delete the associated PVs. In this case, do not delete the PVCs unless you are certain that you no longer need the data.

1. Verify that the PVCs are present for your cluster. For example, to show the Persistent Volume Claims created for a cluster named `my-greenplum`:

    ``` bash
    $ kubectl get pvc -l greenplum-cluster=my-greenplum
    ```
    ``` bash
    NAME                              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
    my-greenplum-pgdata-master-0      Bound    pvc-5f7efa56-671b-470c-8a96-9b876e5c6207   1G         RWO            standard       51m
    my-greenplum-pgdata-segment-a-0   Bound    pvc-7611d47e-e125-4a43-b017-631721c63cec   2G         RWO            standard       51m
    my-greenplum-pgdata-segment-a-1   Bound    pvc-0ef8c1cc-3d23-4e56-9a07-892d9878f758   2G         RWO            standard       8m52s
    my-greenplum-pgdata-segment-a-2   Bound    pvc-9a793fd7-8e00-4180-b377-3aaefffbeed3   2G         RWO            standard       8m52s
    my-greenplum-pgdata-segment-a-3   Bound    pvc-131cb849-5d2e-4a08-a906-4eebc34de8ea   2G         RWO            standard       8m52s
    my-greenplum-pgdata-segment-a-4   Bound    pvc-9a2c4772-4af1-49b1-956e-919e939ffb7c   2G         RWO            standard       8m52s
    my-greenplum-pgdata-segment-a-5   Bound    pvc-7dd51842-b1bc-4eb8-9e25-7ba75da1836a   2G         RWO            standard       8m52s
    ```

2. Use `kubectl` to delete the PVCs associated with the cluster. For example, to delete all PersistentVolume Claims created for a cluster named `my-greenplum`:

    ``` bash
    $ kubectl delete pvc -l greenplum-cluster=my-greenplum
    ```
    ``` bash
    persistentvolumeclaim "my-greenplum-pgdata-master-0" deleted
    persistentvolumeclaim "my-greenplum-pgdata-segment-a-0" deleted
    persistentvolumeclaim "my-greenplum-pgdata-segment-a-1" deleted
    persistentvolumeclaim "my-greenplum-pgdata-segment-a-2" deleted
    persistentvolumeclaim "my-greenplum-pgdata-segment-a-3" deleted
    persistentvolumeclaim "my-greenplum-pgdata-segment-a-4" deleted
    persistentvolumeclaim "my-greenplum-pgdata-segment-a-5" deleted
    ```

3. If the Persistent Volumes were provisioned manually, then deleting the PVCs does not delete the associated PVs. (You can check for the PVs using `kubectl get pv`.) To delete any remaining Persistent Volumes, execute the command:

    ``` bash
    $ kubectl delete pv -l greenplum-cluster=my-greenplum
    ```

See [Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) in the Kubernetes documentation for more information.

## <a id="deloper"></a>Deleting Greenplum Operator

If you also want to remove the Greenplum Operator, follow the instructions in [Uninstalling <%=vars.product_name_long %>](uninstalling.html).
