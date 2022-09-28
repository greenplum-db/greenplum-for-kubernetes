---
title: Expanding a Greenplum Deployment
---

To expand a Greenplum cluster, you first use the Greenplum Operator to apply an updated Greenplum cluster configuration that increases the number of segments. The Greenplum Operator automatically creates the new segment pods in Kubernetes and starts a job to run `gpexpand`  and initialize the new segments. You can optionally run manual commands to redistribute data to the new segments, and to remove the `gpexpand` schema that is created during the expansion process.

**Note:** You cannot resize a cluster to use a lower number of segments; you must delete and re-create the cluster to reduce the number of segments.

Follow these steps to expand a <%=vars.product_name %> cluster on Kubernetes:

1. Go to the `workspace` subdirectory where you unpacked the <%=vars.product_name %> distribution, or to the directory where you created your Greenplum cluster deployment manifest. For example:

    ``` bash
    $ cd ./greenplum-for-kubernetes-*/workspace
    ```

2. Edit the manifest file that you used to deploy your Greenplum cluster. Edit the file to increase the `primarySegmentCount` value. This example increases the number of segments defined in the default Greenplum deployment (`my-gp-instance.yaml`) to 6:

    ``` yaml
    primarySegmentCount: 6
    ```

3. After modifying the file, use `kubectl` to apply the change. For example:

    ``` bash
    $ kubectl apply -f my-gp-instance.yaml
    ```
    ``` bash
    greenplumcluster.greenplum.pivotal.io/my-greenplum configured
    ```

4. Execute `watch kubectl get all` and wait until the new Greenplum pods reach the `Running` state:. Also observe the progress of the expansion job (`job.batch/my-greenplum-gpexpand-job`) and wait for it to complete:

    ``` bash
    $ watch kubectl get all
    ```
    ``` bash
    NAME                                      READY   STATUS    RESTARTS   AGE
    pod/greenplum-operator-6ff95b6b79-kq9vr   1/1     Running   0          60m
    pod/master-0                              1/1     Running   0          43m
    pod/my-greenplum-gpexpand-job-52g4q       1/1     Running   0          30s
    pod/segment-a-0                           1/1     Running   0          43m
    pod/segment-a-1                           1/1     Running   0          31s
    pod/segment-a-2                           1/1     Running   0          31s
    pod/segment-a-3                           1/1     Running   0          31s
    pod/segment-a-4                           1/1     Running   0          31s
    pod/segment-a-5                           1/1     Running   0          31s

    NAME                                                            TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
    service/agent                                                   ClusterIP      None             <none>        22/TCP           43m
    service/greenplum                                               LoadBalancer   10.102.131.136   <pending>     5432:32753/TCP   43m
    service/greenplum-validating-webhook-service-6ff95b6b79-kq9vr   ClusterIP      10.106.60.103    <none>        443/TCP          60m
    service/kubernetes                                              ClusterIP      10.96.0.1        <none>        443/TCP          64m

    NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/greenplum-operator   1/1     1            1           60m

    NAME                                            DESIRED   CURRENT   READY   AGE
    replicaset.apps/greenplum-operator-6ff95b6b79   1         1         1       60m

    NAME                         READY   AGE
    statefulset.apps/master      1/1     43m
    statefulset.apps/segment-a   6/6     43m

    NAME                                  COMPLETIONS   DURATION   AGE
    job.batch/my-greenplum-gpexpand-job   0/1           30s        30s

    NAME                                                 STATUS    AGE
    greenplumcluster.greenplum.pivotal.io/my-greenplum   Running   43m
    ```

    In the unlikely case that the update fails, the Greenplum cluster's status will be `UpdateFailed`. Should that occur, investigate the logs (for example, `kubectl logs pod/my-greenplum-gpexpand-job-52g4q`) to see what happened, address the underlying problem, and use kubectl to re-apply the change.

    <br/>The expansion process is complete after all pods' expansion jobs are marked `Complete`, and `job.batch/my-greenplum-gpexpand-job`, shows 1/1 Completions. At that point, you can either use the cluster with the new segment resources as-is, or continue with the optional steps below to redistribute data to the new segment pods and/or remove the expansion schema.

5. (Optional.) If you want to redistribute existing data to use the new segment pods, perform these steps:

    1. Open a bash shell to the Greenplum master pod:

        ``` bash
        $ kubectl exec -it master-0 -- bash
        ```
        ``` bash
        gpadmin@master-0:~$ 
        ```

    1. Execute the `gpexpand` command with the `-d` or `-e` options to specify the maximum duration or end time after which redistribution stops, respectively.  Also include `-D gpadmin` to indicate that the expansion schema is stored in the gpadmin database.  For example, to redistribute tables for a maximum of 10 hours, enter:

        ``` bash
        $ gpexpand -d 10:00:00
        ```
        ``` bash
        20200513:19:20:32:002601 gpexpand:master-0:gpadmin-[INFO]:-local Greenplum Version: 'postgres (Greenplum Database) 6.8.0 build commit:a21de286045072d8d1df64fa48752b7dfac8c1b7'
        20200513:19:20:32:002601 gpexpand:master-0:gpadmin-[INFO]:-master Greenplum Version: 'PostgreSQL 9.4.24 (Greenplum Database 6.8.0 build commit:a21de286045072d8d1df64fa48752b7dfac8c1b7) on x86_64-unknown-linux-gnu, compiled by gcc (Ubuntu 7.5.      0-3ubuntu1~18.04) 7.5.0, 64-bit compiled on Apr 30 2020 00:14:35'
        20200513:19:20:32:002601 gpexpand:master-0:gpadmin-[INFO]:-Querying gpexpand schema for current expansion state
        20200513:19:20:32:002601 gpexpand:master-0:gpadmin-[INFO]:-Expanding gpadmin.madlib.migrationhistory
        20200513:19:20:32:002601 gpexpand:master-0:gpadmin-[INFO]:-Finished expanding gpadmin.madlib.migrationhistory
        20200513:19:20:37:002601 gpexpand:master-0:gpadmin-[INFO]:-EXPANSION COMPLETED SUCCESSFULLY
        20200513:19:20:37:002601 gpexpand:master-0:gpadmin-[INFO]:-Exiting...
        ```

        If you do not specify the `-d` or `-e` options, redistribution is performed until all tables in the expansion schema are redistributed. If you specify a duration or end time and redistribution stops before all tables are redistributed, you can continue redistributing tables at a later time.

6. (Optional.) Remove the expansion schema if you have finished redistributing tables to the new segments, or if you never intend to redistribute tables to the new segments. 

    <br/>**Note:** You _must_ remove the expansion schema before you can expand the Greenplum cluster again.

    1. Open a bash shell to the Greenplum master pod:

        ``` bash
        $ kubectl exec -it master-0 bash
        ```
        ``` bash
        gpadmin@master-0:~$ 
        ```

    1. Execute the `gpexpand` command with the `-c` option, and enter `y` when prompted to delete the expansion schema:

        ``` bash
        $ gpexpand -c
        ```
        ``` bash
        20200513:19:22:19:002637 gpexpand:master-0:gpadmin-[INFO]:-local Greenplum Version: 'postgres (Greenplum Database) 6.8.0 build commit:a21de286045072d8d1df64fa48752b7dfac8c1b7'
        20200513:19:22:19:002637 gpexpand:master-0:gpadmin-[INFO]:-master Greenplum Version: 'PostgreSQL 9.4.24 (Greenplum Database 6.8.0 build commit:a21de286045072d8d1df64fa48752b7dfac8c1b7) on x86_64-unknown-linux-gnu, compiled by gcc (Ubuntu 7.5.      0-3ubuntu1~18.04) 7.5.0, 64-bit compiled on Apr 30 2020 00:14:35'
        20200513:19:22:19:002637 gpexpand:master-0:gpadmin-[INFO]:-Querying gpexpand schema for current expansion state
        
        
        Do you want to dump the gpexpand.status_detail table to file? Yy|Nn (default=Y):
        > y
        20200513:19:22:33:002637 gpexpand:master-0:gpadmin-[INFO]:-Dumping gpexpand.status_detail to /greenplum/data-1/gpexpand.status_detail
        20200513:19:22:33:002637 gpexpand:master-0:gpadmin-[INFO]:-Removing gpexpand schema
        20200513:19:22:33:002637 gpexpand:master-0:gpadmin-[INFO]:-Cleanup Finished.  exiting...
        ```


