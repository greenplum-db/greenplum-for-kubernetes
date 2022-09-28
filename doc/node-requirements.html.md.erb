---
title: Kubernetes Node Configuration
---

This section describes the Linux kernel configuration requirements for each Kubernetes node that is used in a <%=vars.product_name %> cluster.

## <a id="mem-config"></a>About Memory Overcommit Configuration

Although the <%=vars.product_name %> documentation states that `vm.overcommit_memory` must be set to `2` (no overcommit) for production databases, this setting is not recommended with <%=vars.product_name %>, nor is it possible to configure overcommit in this way for Kubernetes nodes. Kubernetes (`kubelet`) uses `vm.overcommit_memory=1` (always overcommit) on its workers with the assumption that applications often request more memory than they actually touch (read or write).

The value of `vm.overcommit_memory` does not impact the ACID guarantees of Greenplum Database.

## <a id="net-config"></a>Network Configuration

If your Kubernetes deployment uses large nodes (for example, nodes with 40+ cores), and you want to create many segments on each node (more than 16 segments), then you must configure the ARP cache setting on each node to allow the numerous connections that are created by the interconnect process between Greenplum segments.

On Ubuntu systems, update the `/etc/sysctl.conf` to add the settings:

``` bash
net.ipv4.neigh.default.gc_thresh1=30000
net.ipv4.neigh.default.gc_thresh2=32000
net.ipv4.neigh.default.gc_thresh3=32768
net.ipv6.neigh.default.gc_thresh1=30000
net.ipv6.neigh.default.gc_thresh2=32000
net.ipv6.neigh.default.gc_thresh3=32768
```

If you fail to increase these thresholds but deploy a large number of segments on a node, queries that access numerous segments will generate interconnect errors such as:

``` bash
ERROR:  Interconnect error writing an outgoing packet: Invalid argument ...
```

## <a id="file-config"></a>File Configuration

Greenplum parallel query processing against multiple segments requires a high open file limit for each Kubernetes node.

On Ubuntu systems, update the `/etc/systemd/systemd.conf` file to include the setting:

``` bash
DefaultLimitNOFILE=65536
```
