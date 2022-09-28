---
title: Creating Local Persistent Volumes for Greenplum
---

Greenplum requires a Kubernetes Storage Class to use for provisioning Persistent Volumes (PVs) for Greenplum segment pods. This topic describes how to use a storage class that is backed by manually-provisioned Kubernetes [local volumes](https://kubernetes.io/docs/concepts/storage/volumes/#local). A local volume is a disk device, partition, or directory that is directly mounted to a Kubernetes node.

**Note:** Persistent Volume Claims (PVCs) that were created for a version 1.x cluster cannot be used with version 2.x. <%=vars.product_name_long %> version 2.x now tags PVCs with the Greenplum Database major version, for example `greenplum-major-version=6`.

Local volumes can provide high performance in Greenplum clusters when they are used with fast SSDs or RAID systems. However, keep in mind that using local volumes introduces an additional dependency between each Greenplum pod and the node that hosts the local volume; if the local volume node fails, then the associated pod will also fail. 

## Prerequisites

Each Greenplum segment instance (the master, standby, and each primary and mirror segment instance) requires its own, dedicated data storage area.  This means configuring one Kubernetes storage volume per segment instance. To satisfy this requirement with local persistent volumes, you may need to create multiple directories, partitions, or logical volumes (using a tool like [LVM](https://en.wikipedia.org/wiki/Logical_Volume_Manager_(Linux))) to create the required number of mount points.  

For example, consider a case where Greenplum is deployed using 3 large servers, each having 2 SSDs for local storage. In order to deploy a Greenplum cluster of 20 segments (primary/mirror segment pairs) with a master and standby master, you require 42 logical volumes (segment instances = 20 * 2 + 1 + 1 ). With LVM, each pair of physical SSDs could form a logical volume group that is then used to create 14 logical volumes. Other configuration options are possible, but keep in mind that any required directories, partitions, or logical volumes must be created before you can create the associated local volume. 

For best performance, use persistent volumes that are backed by a local SSD with the XFS filesystem, using `readahead` cache for best performance. Use the mount options `rw,nodev,noatime,nobarrier,inode64` to mount the volume. See [Optimizing Persistent Disk and Local SSD Performance](https://cloud.google.com/compute/docs/disks/performance) in the Google Cloud documentation for more information.

## Procedure

1. Create the directory, partition, or logical volume that you want to use as a Kubernetes local volume.

2. Create a `PersistentVolume` definition, specifying the local volume and the required `NodeAffinity` field.  For example:

    ``` yaml
    apiVersion: v1
    kind: PersistentVolume
    metadata:
      name: greenplum-local-pv
    spec:
      capacity:
        storage: 500Gi
      accessModes:
      - ReadWriteOnce
      persistentVolumeReclaimPolicy: Retain
      storageClassName: local-storage
      local:
        path: /mnt/disks/greenplum-vol1
      nodeAffinity:
        required:
          nodeSelectorTerms:
          - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
              - my-greenplum-node
    ```

    See [local](https://kubernetes.io/docs/concepts/storage/volumes/#local) in the Kubernetes documentation for [Volumes](https://kubernetes.io/docs/concepts/storage/volumes/) for more information about configuring local persistent volumes.

3. Repeat the previous step for each `PersistentVolume` required for your cluster. Remember that each Greenplum segment host requires a dedicated storage volume.

4. Create the `StorageClass` definition, specifying `no-provisioner` in order to manually provision local persistent volumes. Using `volumeBindingMode: WaitForFirstConsumer` is also recommended to delay binding the local `PersistenVolume` until a pod requires it. For example:

    ```yaml
    kind: StorageClass
    apiVersion: storage.k8s.io/v1
    metadata:
      name: local-storage
    provisioner: kubernetes.io/no-provisioner
    volumeBindingMode: WaitForFirstConsumer
    ```

5. Use `kubectl` to apply the `StorageClass` and `PersistentVolume` configurations that you created.

6. Specify the local storage `StorageClass` name when you deploy a new Greenplum cluster. See [Deploying or Redeploying a Greenplum Cluster](deploy-operator.html) for instructions.