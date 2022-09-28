---
title: Master Segment Recovery
---
When a given Greenplum cluster was created without a standby (`standby=no`) and if the pod `master-0` (the active Greenplum master instance) fails or is deleted, 
the Greenplum `StatefulSet` restarts the pod. As part of the restart process, master-0 pod will run `gpstart -am && gpstop -ar` to automatically restart the greenplum cluster.

If a given Greenplum cluster was created with a standby, after the master pod is restarted, you should manually start the cluster.
 
## Failing Over to a Standby Master

If the pod `master-0` (the active Greenplum master instance) fails to restart, you can fail over to the standby master instance.  

However, the cluster remains unavailable:

``` bash
$ kubectl exec master-0  -- /bin/bash -c "source /usr/local/greenplum-db/greenplum_path.sh; gpstate"
```
``` bash
20181022:17:34:54:000073 gpstate:master-0:gpadmin-[INFO]:-Starting gpstate with args:
20181022:17:34:55:000073 gpstate:master-0:gpadmin-[INFO]:-local Greenplum Version: 'postgres (GreenplumDatabase) 5.11.3 build dev'
20181022:17:34:56:000073 gpstate:master-0:gpadmin-[CRITICAL]:-gpstate failed. (Reason='could not connect toserver: Connection refused
	Is the server running on host "localhost" (127.0.0.1) and accepting
	TCP/IP connections on port 5432?
could not connect to server: Cannot assign requested address
	Is the server running on host "localhost" (::1) and accepting
	TCP/IP connections on port 5432?
') exiting...
command terminated with exit code 2
```

## <a id="procedure"></a>Procedure

Follow these steps to fail over to a standby master instance in Kubernetes, should the active master instance fail or if the active master pod is deleted. 

1. Login to the standby master host and execute `gpactivatestandby` to activate the host as the standby master. This procedure uses `master-1` to indicate the standby master instance that is being promoted to operate as the active master instance:

    ```bash
    $ kubectl exec -it master-1 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; gpactivatestandby -d /greenplum/data-1  -f"
    ```
    ``` bash
    20181017:21:39:02:000721 gpactivatestandby:master-1:gpadmin-[INFO]:------------------------------------------------------
    20181017:21:39:02:000721 gpactivatestandby:master-1:gpadmin-[INFO]:-Standby data directory    = /greenplum/data-1
    20181017:21:39:02:000721 gpactivatestandby:master-1:gpadmin-[INFO]:-Standby port              = 5432
    20181017:21:39:02:000721 gpactivatestandby:master-1:gpadmin-[INFO]:-Standby running           = yes
    20181017:21:39:02:000721 gpactivatestandby:master-1:gpadmin-[INFO]:-Force standby activation  = no
    20181017:21:39:02:000721 gpactivatestandby:master-1:gpadmin-[INFO]:------------------------------------------------------
    Do you want to continue with standby master activation? Yy|Nn (default=N):
    > 
    ```

    Enter `Y` when prompted to activate the standby master.

2. After the container named `master-1` becomes master, you must use its external IP address to access the Greenplum service. Execute this command to update the load balancer:

    ```bash
    $ kubectl patch service greenplum -p '{"spec":{"selector":{"statefulset.kubernetes.io/pod-name": "master-1"}}}'
    ```
    ``` bash
    service "greenplum" patched
    ```

3. At this point, executing `gpstate` shows that no standby master instance is currently configured:

    ``` bash
    $ kubectl exec -it master-1 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; gpstate"
    ```
    ``` bash
    20181017:21:51:31:001142 gpstate:master-1:gpadmin-[INFO]:-Starting gpstate with args:
    20181017:21:51:32:001142 gpstate:master-1:gpadmin-[INFO]:-local Greenplum Version: 'postgres (Greenplum Database) 5.11.3 build dev'
    20181017:21:51:33:001142 gpstate:master-1:gpadmin-[INFO]:-master Greenplum Version: 'PostgreSQL 8.3.23 (Greenplum Database 5.11.3 build dev) on x86_64-pc-linux-gnu, compiled by    GCC gcc (Ubuntu 6.4.0-17ubuntu1~16.04) 6.4.0 20180424, 64-bit compiled on Oct 10 2018 22:25:23'
    20181017:21:51:33:001142 gpstate:master-1:gpadmin-[INFO]:-Obtaining Segment details from master...
    20181017:21:51:33:001142 gpstate:master-1:gpadmin-[INFO]:-Gathering data from segments...
    .........
    20181017:21:51:42:001142 gpstate:master-1:gpadmin-[INFO]:-Greenplum instance status summary
    20181017:21:51:43:001142 gpstate:master-1:gpadmin-[INFO]:-----------------------------------------------------
    20181017:21:51:43:001142 gpstate:master-1:gpadmin-[INFO]:-   Master instance                                           = Active
    20181017:21:51:43:001142 gpstate:master-1:gpadmin-[INFO]:-   Master standby                                            = No master standby configured
    ...
    ```

    Continue following the remaining steps to initialize the previous master instance as the standby master.

4. Remove the master data directory from the inactive master instance (`master-0` in this example):

    ```bash
    $ kubectl exec master-0  -- /bin/bash -c 'source /usr/local/greenplum-db/greenplum_path.sh; rm -rf ${MASTER_DATA_DIRECTORY}'
    ```

5. On the active master host, execute the following two commands to prepare and initialize the new standby master host:

    ```bash
    $ kubectl exec master-1  -- /bin/bash -c "source /usr/local/greenplum-db/greenplum_path.sh; /home/gpadmin/tools/sshKeyScan"
    ```
    ``` bash
    Key scanning started
    ```
    ``` bash
    $ kubectl exec master-1  -- /bin/bash -c "source /usr/local/greenplum-db/greenplum_path.sh; gpinitstandby -a -s master-0"
    ```
    ``` bash
    20181022:17:44:11:000595 gpinitstandby:master-1:gpadmin-[INFO]:-Validating environment and parameters for standby initialization...
    20181022:17:44:12:000595 gpinitstandby:master-1:gpadmin-[INFO]:-Checking for filespace directory /greenplum/data-1 on master-0
    20181022:17:44:13:000595 gpinitstandby:master-1:gpadmin-[INFO]:------------------------------------------------------
    20181022:17:44:13:000595 gpinitstandby:master-1:gpadmin-[INFO]:-Greenplum standby master initialization parameters
    20181022:17:44:13:000595 gpinitstandby:master-1:gpadmin-[INFO]:------------------------------------------------------
    20181022:17:44:13:000595 gpinitstandby:master-1:gpadmin-[INFO]:-Greenplum master hostname               = master-1
    20181022:17:44:13:000595 gpinitstandby:master-1:gpadmin-[INFO]:-Greenplum master data directory         = /greenplum/data-1
    20181022:17:44:13:000595 gpinitstandby:master-1:gpadmin-[INFO]:-Greenplum master port                   = 5432
    20181022:17:44:13:000595 gpinitstandby:master-1:gpadmin-[INFO]:-Greenplum standby master hostname       = master-0
    20181022:17:44:13:000595 gpinitstandby:master-1:gpadmin-[INFO]:-Greenplum standby master port           = 5432
    20181022:17:44:13:000595 gpinitstandby:master-1:gpadmin-[INFO]:-Greenplum standby master data directory = /greenplum/data-1
    20181022:17:44:13:000595 gpinitstandby:master-1:gpadmin-[INFO]:-Greenplum update system catalog         = On
    ```

6. At this point the active master runs on the pod named "master-1" and the standby master runs on the pod named "master-0." Verify the role of each segment host:

    ``` bash
    $ kubectl exec -it master-1 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; psql -c 'select hostname, role from gp_segment_configuration;'"
    ```
    ``` sql
      hostname   | role
    -------------+------
     segment-a-0 | p
     segment-b-0 | m
     master-1    | p
     master-0    | m
    (4 rows)
    ```

## <a id="moreinfo"></a>Getting More Information

For more information about failing over to a standby master, see [Recovering a Failed Master](http://greenplum.docs.pivotal.io/5120/admin_guide/highavail/topics/g-recovering-a-failed-master.html) in the <%=vars.product_name %> documentation.

