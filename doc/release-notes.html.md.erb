<% set_title(product_name_long, "Version", product_version, "Release Notes") %>

<%=vars.product_name_long %> 2.3 is a minor release that introduces new features and bug fixes.

Refer to the [<%=vars.product_name %> Database](https://gpdb.docs.pivotal.io) documentation for detailed information about <%=vars.product_name %> Database.

**Note:** Upgrading or migrating data from <%=vars.product_name %> version 1.x to version 2.x is not supported. You must uninstall <%=vars.product_name %> version 1.x using the instuctions in [Uninstalling <%=vars.product_name_long %>](uninstalling.html) before installing version 2.x. Persistent volumes that were created for a version 1.x cluster cannot be used with version 2.x.

## <a id="component_versions"></a>Components

<%=vars.product_name_long %> includes the <%=vars.product_name %> Database version identified below:

| <%=vars.product_name %> Version | <%=vars.product_name %> Database Version |
|----------------------------------|----------------------------|
| 2.3.0 | 6.11.1 |
| 2.2.0 | 6.10.1 |
| 2.1.0, 2.0.1 | 6.8.1 |
| 2.0.0 | 6.8.0 |

## <a id="supported_platforms"></a>Supported Platforms

This version of <%=vars.product_name_long %> is supported on the following platforms:

- VMware Tanzu Kubernetes Grid Integrated (TKGI) Edition 1.7.0 (contains Kubernetes 1.16.7)
- Google Kubernetes Engine (GKE) Kubernetes 1.16.7
- Amazon Elastic Kubernetes Service (Amazon EKS) 1.16

Additional Kubernetes environments, such as Minikube, can be used for testing or demonstration purposes.

## <a id="changedfeatures201"></a>Release 2.3.0 Changes
<%=vars.product_name %> 2.3.0 is a minor release that includes these changes:

- Greenplum Text is now a supported feature.
- Greenplum Command Center installation files are now installed by default with Greenplum for Kubernetes. To use Command Center you must execute the Command Center installation after deploying a new cluster. See [Using Greenplum Command Center](gpcc.html).
- `gpbackup` can back up Greenplum data either to local backup files, or to Minio S3 using the `gpbackup` S3 plugin. See [Using the S3 Storage Plugin with gpbackup and gprestore](https://gpdb.docs.pivotal.io/backup-restore/latest/admin_guide/managing/backup-s3-plugin.html) in the Greenplum Backup and Restore documentation for more information on configuring the plugin.

## <a id="knownissues"></a>Known Issues and Limitations for Kubernetes Deployments

- Currently <%=vars.product_name_long %> does not support automatic upgrades of Greenplum Text. If you installed an earlier version of Greenplum Text in <%=vars.product_name_long %> and then upgrade to version 2.3, you must manually complete the Greenplum Text upgrade by executing `gptext-installsql -c gpadmin && gptext-installsql gpadmin` after the upgraded cluster is fully running.
- The `zkManager` utility, used for checking the ZooKeeper cluster state and cluster management with Greenplum Text deployments, is not available with <%=vars.product_name_long %>. As a workaround, use `kubectl` to manage the ZooKeeper cluster.
- If you deploy a cluster with Greenplum Text, the `gptext-state` utility always throws the warning message: `object of type 'NoneType' has no len()`. This message can be safely ignored in most cases. However, note that the utility cannot detect if the ZooKeeper cluster is down.
- VMware does not support deployments that have been modified by adding layers to the packaged Docker images, or deployments that reference images other than the Greenplum Operator. VMware does not support changing the contents of the deployed containers and pods in any way.
- Greenplum Database connector functionality is not supported. This includes Greenplum Streaming Server (gpss) and its API, Greenplum-Informatica Connector, Greenplum-Spark Connector, and the Greenplum-Kafka Integration.
- <%=vars.product_name_long %> does not support the built-in SNMP features that are available in <%=vars.product_name %> Database.
- <%=vars.product_name_long %> does not support installing Greenplum extensions that use the `.gppkg` format (and `gppkg` utility). Future releases will include these extensions as part of the distribution, as with MADlib.
- The Greenplum Operator does not yet support changing all attribute values of a deployed Greenplum cluster. See [Greenplum Operator Manifest File](operator-reference.html) for details.
