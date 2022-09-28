---
title: Recovering Failed Nodes
---

Nodes may fail in a Kubernetes cluster for a variety of reasons, including drive failures, memory failures, and network failures. After a node fails, it is up to the Kubernetes cluster operator to recover the node and re-attach it to the cluster. During a node failure event, the Kubernetes cluster is operating in a degraded state, leading to potential resource constraints on a deployed Greenplum cluster. For example, segment pods previously scheduled on the failed node may not get re-scheduled on the remaining nodes. For these reasons, it's important to recover failed nodes in a timely fashion.

## About Reapplying Node Labels

After a failed node has been re-created and re-attached to the Kubernetes cluster, there may be manual steps necessary to incorporate it as part of the Greenplum cluster. Greenplum on Kubernetes relies on node labels for some functionality. Currently, these labels are not automatically reapplied to nodes upon node recreation.  Therefore, manual re-applications of node labels are necessary for the following features in the [Operator Manifest](operator-reference.html).

## <a id='operator-reference.html#workerSelector'></a>Reapply workerSelector labels

If `workerSelector` is not specified in the manifest, there are no steps required to re-apply `workerSelector` labels. If `workerSelector` is specified in the manifest, then you must reapply the appropriate `workerSelector` label to the new node to indicate whether it belongs in the `masterAndStandby` or `segments` `workerSelector` pool.

```bash
$ kubectl label node <node name> <key>=<value>
```

## <a id='operator-reference.html#antiAffinity'></a>Reapply antiAffinity labels 

 If `antiAffinity` is set to "no" (the default if the property is not configured), then there are no steps required to re-apply `antiAffinity` node labels. If `antiAffinity` is set to "yes" and a Greenplum cluster has been deployed, then you must re-apply the appropriate `antiAffinity` label(s) to a recovered node.  See the below chart to determine which `antiAffinity` label(s) to apply depending on different scenarios.

| Scenario | `antiAffinity` label to apply |
| -------- | ----------------------------- |
| There is no `masterAndStandby` `workerSelector` specified in the manifest OR <br/><br/> The node has the `masterAndStandby` `workerSelector` label applied | `masterAndStandby` `antiAffinity` label |
| There is no `segments` `workerSelector` specified in the manifest OR <br/><br/> The node has the `segments` `workerSelector` label applied | `segments` `antiAffinity` label |
| `antiAffinity` is explicitly set to "no" or the property is omitted | no `antiAffinity` labels are needed |

### <a id='masterlabel'></a>Label Master and Standby Nodes
To apply the `masterAndStandby` `antiAffinity` label, use the following command:

``` bash
$ kubectl label node <node name> greenplum-affinity-<namespace>-master=true
```

### <a id='segmentlabel'></a>Label Segment Nodes
To apply the segments `antiAffinity` label, first determine whether the recovered node should be an "a" or "b" node. Examine the number of existing nodes that are "a" vs. "b" nodes by running,

```bash
$ kubectl get nodes --show-labels | grep greenplum-affinity-default-segment=a | wc -l  # Number of "a" nodes
$ kubectl get nodes --show-labels | grep greenplum-affinity-default-segment=b | wc -l  # Number of "b" nodes
```

If there are the same number of "a" nodes and "b" nodes, the new node could be either an "a" node or a "b" node. To apply the label, run:

```bash
$ kubectl label node <node name> greenplum-affinity-<namespace>-segment=<a or b>
```

If there are fewer "a" nodes than "b" nodes, the new node should be labeled as "a". To apply the label, run:

```bash
$ kubectl label node <node name> greenplum-affinity-<namespace>-segment=a
```

If there are fewer "b" nodes than "a" nodes, the new node should be labeled as "b". To apply the label, run:

```bash
$ kubectl label node <node name> greenplum-affinity-<namespace>-segment=b
```

