---
title: Accessing a Greenplum Cluster in Kubernetes
---

After you deploy a new Greenplum cluster to Kubernetes, you can access the cluster either by executing Greenplum utilities from within Kubernetes, or by using a locally-installed tool, such as `psql`, to access the Greenplum instance running in Kubernetes.

## <a id="ssh"></a>Accessing a Pod via Kubectl

Use the `kubectl` tool to run utilities directly in a Greenplum pod. For example, to execute `psql`:

``` bash
$ kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; psql"
```
```
psql (9.4.24)
Type "help" for help.
```

You can also simply execute a bash shell and then execute multiple Greenplum utilities as necessary. For example:

``` bash
$ kubectl exec -it master-0 -- /bin/bash
gpadmin@master-0:~$ gpstate
20200513:18:47:55:001929 gpstate:master-0:gpadmin-[INFO]:-Starting gpstate with args: 
20200513:18:47:55:001929 gpstate:master-0:gpadmin-[INFO]:-local Greenplum Version: 'postgres (Greenplum Database) 6.8.0 build commit:a21de286045072d8d1df64fa48752b7dfac8c1b7'
20200513:18:47:55:001929 gpstate:master-0:gpadmin-[INFO]:-master Greenplum Version: 'PostgreSQL 9.4.24 (Greenplum Database 6.8.0 build commit:a21de286045072d8d1df64fa48752b7dfac8c1b7) on x86_64-unknown-linux-gnu, compiled by gcc (Ubuntu 7.5.0-3ubuntu1~18.04) 7.5.0, 64-bit compiled on Apr 30 2020 00:14:35'
20200513:18:47:55:001929 gpstate:master-0:gpadmin-[INFO]:-Obtaining Segment details from master...
20200513:18:47:55:001929 gpstate:master-0:gpadmin-[INFO]:-Gathering data from segments...
... 

```

Substitute the name of a different pod to access, for example, the standby master or an individual segment instance.

## <a id="external"></a>Accessing Greenplum from External Clients

If you have installed `psql` or another client application outside of Kubernetes (for example, on your local client machine), then you need to obtain the address and port number at which you can reach the Greenplum cluster. Follow these steps:

1. For VMware Tanzu Kubernetes Grid Integrated (TKGI) Edition or GKE deployments, the Greenplum load balancer provides the external address and port you can use to reach the cluster:

    ``` bash
    $ kubectl get service/greenplum
    ```
    ``` bash
    NAME        TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
    greenplum   LoadBalancer   10.105.210.228   192.168.39.52 5432:31643/TCP   13m
    ```

2. For Minikube deployments, the Greenplum load balancer is not used. Instead, look for the `greenplum` entry in the output from `minikube service list`:

    ``` bash
    $ minikube service list
    ```
    ``` bash
    |-------------|-------------------------------------------------------|--------------|-----------------------------|
    |  NAMESPACE  |                         NAME                          | TARGET PORT  |             URL             |
    |-------------|-------------------------------------------------------|--------------|-----------------------------|
    | default     | agent                                                 | No node port |
    | default     | greenplum                                             |         5432 | http://192.168.39.204:32753 |
    | default     | greenplum-validating-webhook-service-6ff95b6b79-kq9vr | No node port |
    | default     | kubernetes                                            | No node port |
    | kube-system | kube-dns                                              | No node port |
    |-------------|-------------------------------------------------------|--------------|-----------------------------|
    ```

3. After you obtain the correct IP address and port, use them to connect to the Greenplum instance. For example, with a local `psql` client:

    ``` bash
    $ psql -h 192.168.39.204 -p 32753 -U gpadmin
    ```
