<% set_title("Uninstalling", product_name_long) %>


This topic describes how to uninstall <%=vars.product_name_long %>. This process involves deleting previously-deployed Greenplum clusters, removing container images into your container registry, and using the `helm` package manager to delete the Greenplum Operator resource from Kubernetes. After you finish the uninstallation process, you will not be able to use the Greenplum Operator to manage existing Greenplum cluster deployments.

## Procedures

Follow these steps to uninstall <%=vars.product_name_long %>:

1. Delete any existing Greenplum clusters previously created with the operator. Follow the steps from [Deleting a Greenplum Cluster](deleting.html) to delete any deployed Greenplum clusters.

1. Use the `helm delete` command to delete the `greenplum-operator` release:

    ``` bash
    $ helm delete greenplum-operator
    ```

1. To delete the local Greenplum Docker images, first determine the image IDs. For example:

    ``` bash
    $ docker images "greenplum-*"
    ```
    ``` bash
    REPOSITORY                 TAG                     IMAGE ID            CREATED             SIZE
    greenplum-operator         v0.4.1.dev.9.g60effb8   d6ed470f5096        3 days ago          216MB
    greenplum-for-kubernetes   v0.4.1.dev.9.g60effb8   6e2eae3efc7b        3 days ago          785MB
    ```

1. After determining the image IDs to remove, use `docker rmi` to delete them:

    ``` bash
    $ docker rmi d6ed470f5096
    ```
    ``` bash
    Untagged: greenplum-operator:v0.4.1.dev.9.g60effb8
    Deleted: sha256:d6ed470f509641671f18c8b14a7fe2a90d21f42899091418f9f74bb2e580407a
    Deleted: sha256:20cc7faf223e061b83e8b0e46df5b30ef59e921aa30fb8c731b48bf061db5090
    Deleted: sha256:7f359f53c875b7f940cc218cd8db7ffba6c33a0e21b00e59b521d06f4be0f752
    Deleted: sha256:31d8cdb416145ad1dd5ae9ba852f8979b5f39e75ff0d79ff2418d6ac7984a34c
    ```
    ``` bash
    $ docker rmi 6e2eae3efc7b
    ```
    ``` bash
    Untagged: greenplum-for-kubernetes:v0.4.1.dev.9.g60effb8
    Deleted: sha256:6e2eae3efc7b225a417afbc04f9d6d00e96d2f412cd4b0f044965f2f0b607894
    Deleted: sha256:c888ec4c22e495a3534ec3c12dec1b3852171bd378be4bb14b08326f45705996
    Deleted: sha256:69aff71216917d3bcc33d132f3c5633d7f94bd62c1cac7aa37412203c0632e41
    Deleted: sha256:ea26fb50e841c71c7471d7071cfe3b8ef631c4589be413c168dda88367d0af10
    Deleted: sha256:f21f3641b5ceba2319de8d79251cee777af23974bf037b571e3090af1bad8db3
    Deleted: sha256:eaaac4da7046f6f0d8dad2e469ec3982137467f20b8b84c5fc202a104b74b31e
    Deleted: sha256:d02bb8400906504396c46d0cb2111a532d3e20001d4940fa3adf68f51c41d7d1
    Deleted: sha256:9fb0e3075a99c4b58ee2d599425ae920ef389178811400645d79a38364da848e
    Deleted: sha256:083cb21977d768ff10f2eb54a3674e7507b5640530a0e7553c7a6546c38063fa
    Deleted: sha256:bff984ca1b417796e8087e3619ccf78b28c857e88ddffb58ce02a85219643002
    Deleted: sha256:c76d65c498aca29a51bab36e10424c0e87e9a9526931b26a4df3b978a8ebf5bd
    Deleted: sha256:7570f0f44c3f21c89e0101db713a104f6d0b44441db79f0ef9c0ea347dd8610f
    Deleted: sha256:7a65049e912f4fb31073d221ed0f76656891aa16d070961b2fe2c68ce6341231
    Deleted: sha256:f430714203443a72c308d86ef4b5fb6bb1ae83abef9037fd48fd57bb43277d68
    Deleted: sha256:cce92fda689ab9033f0b8db214bc63edd1ae3e05831a0f3a9418976d7dc7ccdd
    Deleted: sha256:d22094bbd65447c59a42c580eaa3a44cee9cd855f00905f59409be21bcefc745
    Deleted: sha256:b8976847450013f3eb5e9a81a5778f73ed7bef67e6393049712ef17102b4b7b7
    Deleted: sha256:b8c891f0ffec910a12757d733b178e3f62d81dbbde2b31d3b754071c416108ed
    ```

1. If you pushed the Greenplum Docker images to a remote container registry, use the registry-specific instructions to delete those images. For example, if you pushed the images to Google Cloud Registry, follow the instructions in [Deleting Images](https://cloud.google.com/container-registry/docs/managing#deleting_images) in the Google Cloud documentation to remove them.
