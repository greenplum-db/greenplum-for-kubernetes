---
title: Recovering Failed Segments
---

## <a id="auto"></a>Automatic Segment Recovery for Mirrorless Deployments


If the pod that runs a Greenplum primary segment instance fails or is deleted, the Greenplum `StatefulSet` restarts the pod. The primary segment starts postgres on its startup and joins the Greenplum Cluster automatically.
However, all queries are aborted until the segment finishes restarting. Normal query processing resumes only after the pod restarts and the segment rejoins the cluster.

## <a id="mirroring"></a>Segment Recovery with Mirroring Enabled

If the pod that runs a Greenplum primary or mirror segment instance fails or is deleted, the Greenplum `StatefulSet` restarts the pod. However, the segment remains in a failed state until you take measures to recover it. For example, if a primary segment fails, the `gpstate -e` command will show that the roles for the primary and mirror segments have switched:

```bash
$ kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; gpstate -e"
```
``` bash
20181026:00:14:07:004894 gpstate:master-0:gpadmin-[INFO]:-Starting gpstate with args: -e
20181026:00:14:08:004894 gpstate:master-0:gpadmin-[INFO]:-local Greenplum Version: 'postgres (Greenplum Database) 5.12.0 build dev'
20181026:00:14:08:004894 gpstate:master-0:gpadmin-[INFO]:-master Greenplum Version: 'PostgreSQL 8.3.23 (Greenplum Database 5.12.0 build dev) on x86_64-pc-linux-gnu, compiled by GCC gcc (Ubuntu 6.4.0-17ubuntu1~16.04) 6.4.0 20180424, 64-bit compiled on Oct 19 2018 16:56:25'
20181026:00:14:09:004894 gpstate:master-0:gpadmin-[INFO]:-Obtaining Segment details from master...
20181026:00:14:09:004894 gpstate:master-0:gpadmin-[INFO]:-Gathering data from segments...
....... 
20181026:00:14:16:004894 gpstate:master-0:gpadmin-[INFO]:-----------------------------------------------------
20181026:00:14:16:004894 gpstate:master-0:gpadmin-[INFO]:-Segment Mirroring Status Report
20181026:00:14:16:004894 gpstate:master-0:gpadmin-[INFO]:-----------------------------------------------------
20181026:00:14:16:004894 gpstate:master-0:gpadmin-[INFO]:-Segments with Primary and Mirror Roles Switched
20181026:00:14:16:004894 gpstate:master-0:gpadmin-[INFO]:-   Current Primary                               Port    Mirror                                        Port
20181026:00:14:16:004894 gpstate:master-0:gpadmin-[INFO]:-   segment-b-0.agent.default.svc.cluster.local   50000   segment-a-0.agent.default.svc.cluster.local   40000
20181026:00:14:16:004894 gpstate:master-0:gpadmin-[INFO]:-----------------------------------------------------
20181026:00:14:16:004894 gpstate:master-0:gpadmin-[INFO]:-Primaries in Change Tracking
20181026:00:14:16:004894 gpstate:master-0:gpadmin-[INFO]:-   Current Primary                               Port    Change tracking size   Mirror                                        Port
20181026:00:14:16:004894 gpstate:master-0:gpadmin-[INFO]:-   segment-b-0.agent.default.svc.cluster.local   50000   128 bytes              segment-a-0.agent.default.svc.cluster.local   40000
command terminated with exit code 1
```failover.html.md

Mirror host failures appear in the output of `gpstate -m`:

```bash
$ kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; gpstate -m"
```
``` bash
20181025:23:18:31:003178 gpstate:master-0:gpadmin-[INFO]:-Starting gpstate with args: -m
20181025:23:18:32:003178 gpstate:master-0:gpadmin-[INFO]:-local Greenplum Version: 'postgres (Greenplum Database) 5.12.0 build dev'
20181025:23:18:33:003178 gpstate:master-0:gpadmin-[INFO]:-master Greenplum Version: 'PostgreSQL 8.3.23 (Greenplum Database 5.12.0 build dev) on x86_64-pc-linux-gnu, compiled by GCC gcc (Ubuntu 6.4.0-17ubuntu1~16.04) 6.4.0 20180424, 64-bit compiled on Oct 19 2018 16:56:25'
20181025:23:18:33:003178 gpstate:master-0:gpadmin-[INFO]:-Obtaining Segment details from master...
20181025:23:18:33:003178 gpstate:master-0:gpadmin-[INFO]:--------------------------------------------------------------
20181025:23:18:33:003178 gpstate:master-0:gpadmin-[INFO]:--Current GPDB mirror list and status
20181025:23:18:33:003178 gpstate:master-0:gpadmin-[INFO]:--Type = Spread
20181025:23:18:33:003178 gpstate:master-0:gpadmin-[INFO]:--------------------------------------------------------------
20181025:23:18:33:003178 gpstate:master-0:gpadmin-[INFO]:-   Mirror                                        Datadir                  Port    Status   Data Status   
20181025:23:18:33:003178 gpstate:master-0:gpadmin-[WARNING]:-segment-b-0.agent.default.svc.cluster.local   /greenplum/mirror/data   50000   Failed                 <<<<<<<<
20181025:23:18:33:003178 gpstate:master-0:gpadmin-[INFO]:--------------------------------------------------------------
20181025:23:18:33:003178 gpstate:master-0:gpadmin-[WARNING]:-1 segment(s) configured as mirror(s) have failed
```

## <a id="procedure"></a>Procedure

For either primary or mirror segment failures, follow these steps to recover failed segments:

1. Login to the master host and execute `gprecoverseg`:

    ```bash
    $ kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; gprecoverseg"
    ```
    ``` bash
    20181025:23:18:45:003227 gprecoverseg:master-0:gpadmin-[INFO]:-Starting gprecoverseg with args: 
    20181025:23:18:45:003227 gprecoverseg:master-0:gpadmin-[INFO]:-local Greenplum Version: 'postgres (Greenplum     Database) 5.12.0 build dev'
    20181025:23:18:46:003227 gprecoverseg:master-0:gpadmin-[INFO]:-master Greenplum Version: 'PostgreSQL 8.3.23     (Greenplum Database 5.12.0 build dev) on x86_64-pc-linux-gnu, compiled by GCC gcc (Ubuntu 6.4.0-17ubuntu1~16.04)     6.4.0 20180424, 64-bit compiled on Oct 19 2018 16:56:25'
    20181025:23:18:47:003227 gprecoverseg:master-0:gpadmin-[INFO]:-Checking if segments are ready to connect
    20181025:23:18:47:003227 gprecoverseg:master-0:gpadmin-[INFO]:-Obtaining Segment details from master...
    20181025:23:18:48:003227 gprecoverseg:master-0:gpadmin-[INFO]:-Obtaining Segment details from master...
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-Heap checksum setting is consistent between     master and the segments that are candidates for recoverseg
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-Greenplum instance recovery parameters
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]    :----------------------------------------------------------
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-Recovery type              = Standard
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]    :----------------------------------------------------------
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-Recovery 1 of 1
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]    :----------------------------------------------------------
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-   Synchronization mode                        =     Incremental
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-   Failed instance host                        =     segment-b-0
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-   Failed instance address                     =     segment-b-0.agent.default.svc.cluster.local
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-   Failed instance directory                   =     /greenplum/mirror/data
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-   Failed instance port                        =     50000
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-   Failed instance replication port            =     6001
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-   Recovery Source instance host               =     segment-a-0
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-   Recovery Source instance address            =     segment-a-0.agent.default.svc.cluster.local
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-   Recovery Source instance directory          =     /greenplum/data
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-   Recovery Source instance port               =     40000
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-   Recovery Source instance replication port   =     6000
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]:-   Recovery Target                             =     in-place
    20181025:23:18:51:003227 gprecoverseg:master-0:gpadmin-[INFO]    :----------------------------------------------------------
    
    Continue with segment recovery procedure Yy|Nn (default=N):
    >
    ```

    Enter `Y` when prompted to initiate the recovery procedure.

2. Execute `gpstate -s` to monitor the resynchronization process for the segment:

    ```bash
    $ kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; gpstate -s"
    ```
    ``` bash
    20181026:01:04:16:005887 gprecoverseg:master-0:gpadmin-[INFO]   :-******************************************************************
    gpadmin@master-0:~$ gpstate -s
    20181026:01:04:23:005959 gpstate:master-0:gpadmin-[INFO]:-Starting gpstate with args: -s
    20181026:01:04:23:005959 gpstate:master-0:gpadmin-[INFO]:-local Greenplum Version: 'postgres (Greenplum     Database) 5.12.0 build dev'
    20181026:01:04:24:005959 gpstate:master-0:gpadmin-[INFO]:-master Greenplum Version: 'PostgreSQL 8.3.23  (Greenplum Database 5.12.0 build dev) on x86_64-pc-linux-gnu, compiled by GCC gcc (Ubuntu 6.4.0-17ubuntu1~16.04)  6.4.0 20180424, 64-bit compiled on Oct 19 2018 16:56:25'
    20181026:01:04:25:005959 gpstate:master-0:gpadmin-[INFO]:-Obtaining Segment details from master...
    20181026:01:04:25:005959 gpstate:master-0:gpadmin-[INFO]:-Gathering data from segments...
    ....... 
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-----------------------------------------------------
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:--Master Configuration & Status
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-----------------------------------------------------
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-   Master host                    = master-0
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-   Master postgres process ID     = 2017
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-   Master data directory          = /greenplum/data-1
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-   Master port                    = 5432
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-   Master current role            = dispatch
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-   Greenplum initsystem version   = 5.12.0 build dev
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-   Greenplum current version      = PostgreSQL 8.3.23     (Greenplum Database 5.12.0 build dev) on x86_64-pc-linux-gnu, compiled by GCC gcc (Ubuntu 6.4.0-17ubuntu1~16.04)     6.4.0 20180424, 64-bit compiled on Oct 19 2018 16:56:25
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-   Postgres version               = 8.3.23
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-   Master standby                 =   master-1.agent.default.svc.cluster.local
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-   Standby master state           = Standby host  passive
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-----------------------------------------------------
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-Segment Instance Status Report
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-----------------------------------------------------
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-   Segment Info
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-      Hostname                                =   segment-b-0
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-      Address                                 =   segment-b-0.agent.default.svc.cluster.local
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-      Datadir                                 =   /greenplum/mirror/data
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-      Port                                    = 50000
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-   Mirroring Info
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-      Current role                            =   Primary
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-      Preferred role                          = Mirror
    20181026:01:04:32:005959 gpstate:master-0:gpadmin-[INFO]:-      Mirror status                           =   Resynchronizing
    ```

    The recovery process is complete when the `Data Status` changes from `Resynchronizing` to `Synchronized`.

3. If a primary segment originally failed, the running segment instances will have changed their roles in the cluster. You can optionally verify the role of each segment host with the command:

    ``` bash
    $ kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; psql -c 'select hostname, role, preferred_role from gp_segment_configuration;'"

    ```
    ``` sql
                     hostname                 | role | preferred_role 
    ------------------------------------------+------+----------------
     master-0                                 | p    | p
     master-1.agent.default.svc.cluster.local | m    | m
     segment-a-0                              | p    | m
     segment-b-0                              | m    | p
    (4 rows)
    ```

    The above output shows that each segment is operating in its preferred role.

4. If, after segment recovery, a segment is not operating in its preferred role, execute `gprecoverseg -r` to return segments to their preferred roles:

    ```bash
    $ kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; gprecoverseg -r"
    ```

    Enter `Y` when prompted to initiate the procedure.

5. Use `gpstate -e` to monitor the resynchronization process. The process is complete when the command output completes with:

    ``` bash
    20181026:01:11:10:006463 gpstate:master-0:gpadmin-[INFO]:-----------------------------------------------------
    20181026:01:11:10:006463 gpstate:master-0:gpadmin-[INFO]:-Segment Mirroring Status Report
    20181026:01:11:10:006463 gpstate:master-0:gpadmin-[INFO]:-----------------------------------------------------
    20181026:01:11:10:006463 gpstate:master-0:gpadmin-[INFO]:-All segments are running normally
    ```

## <a id="moreinfo"></a>Getting More Information

For more information about recovering failed segments, see these links in the <%=vars.product_name %> documentation:

* [gprecoverseg](http://greenplum.docs.pivotal.io/5120/utility_guide/admin_utilities/gprecoverseg.html) Reference Page
* [Detecting a Failed Segment](http://greenplum.docs.pivotal.io/5120/admin_guide/highavail/topics/g-detecting-a-failed-segment.html)
* [Recovering a Failed Segment](http://greenplum.docs.pivotal.io/5120/admin_guide/highavail/topics/g-recovering-a-failed-segment.html)
