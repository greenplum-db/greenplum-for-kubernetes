---
title: Minikube
---

This section describes the software and configuration necessary to run <%=vars.product_name %> in Minikube. Using Minikube offers a quick way to demonstrate Greenplum on your local system.

## <a id="softwarereq"></a>Required Software

To deploy <%=vars.product_name %> on Minikube, you require the following software:

<%=partial 'partials/prerequisites-common' %>

* Minikube. See the [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) documentation to install the latest version of Minikube. As part of the Minikube installation you must install a compatible hypervisor to your system (if one is not already available) as well as a corresponding Minikube driver for the hypervisor. For example, on MacOS systems you can use the built-in Hyperkit hypervisor by installing the Minikube [Hyperkit driver](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver).

    <br/>**Note:** Do not install or re-install `kubectl` as part of the Minikube installation.

To validate that your system meets these prerequisites, ensure that the following commands execute without any errors, and that the output versions are similar to the ones shown:

```bash
$ kubectl version --client
Client Version: version.Info{Major:"1", Minor:"18", GitVersion:"v1.18.2", GitCommit:"52c56ce7a8272c798dbc29846288d7cd9fbae032", GitTreeState:"clean", BuildDate:"2020-04-16T11:56:40Z", GoVersion:"go1.13.9", Compiler:"gc", Platform:"linux/amd64"}
$ docker --version
Docker version 19.03.8, build afacb8b
$ minikube version
minikube version: v1.10.1
$ helm version --client
version.BuildInfo{Version:"v3.2.1", GitCommit:"fe51cd1e31e6a202cba7dead9552a6d418ded79a", GitTreeState:"clean", GoVersion:"go1.13.10"}
```

**Note:** The documented procedure builds the required Docker image locally and uploads it to Minikube. As an alternative you can use [Docker support for Kubernetes](https://www.docker.com/kubernetes).


## <a id="starting"></a>Configuring the Minikube Cluster

Follow this procedure to start your local Minikube cluster and configure Docker for installing <%=vars.product_name %>:

1. Start Docker if it is not already running on your system.

2. Start the Minikube cluster using enough resources to ensure reasonable response times for the Greenplum service. For example to run only a basic Greenplum Database cluster use:

    ```bash
    $ minikube start --memory 4096 --cpus 4 --kubernetes-version=v1.16.7 --vm-driver=<driver-name>
    ```

    If you want to run Greenplum database with PXF, additional resources are required for the sample deployment scripts:

    ```bash
    $ minikube start --memory 8192 --cpus 5 --kubernetes-version=v1.16.7 --vm-driver=<driver-name>
    ```

    Substitute the name of the hypervisor driver you want to use for `<driver-name>`. For example, if you are using MacOS and you installed the Minikube Hyperkit driver, you would execute:

    ```bash
    $ minikube start --memory 4096 --cpus 4 --kubernetes-version=v1.16.7 --vm-driver=hyperkit
    ```

    Minikube starts the cluster and displays its progress.

3. Confirm that the `kubectl` utility can access Minikube:

    ``` bash
    $ kubectl get node
    ```

    ``` bash
    NAME       STATUS   ROLES    AGE   VERSION
    minikube   Ready    master   78s   v1.16.7
    ```

    **Note:** If you have problems starting or connecting to Minikube, use `minikube delete` to remove the current Minikube and then recreate it.

4. Change the local docker environment to point to the Minikube docker, so that the local docker daemon interacts with images inside the Minikube docker environment:


    ``` bash
    $ eval $(minikube docker-env)
    ```

    **Note:** To undo this docker setting in the current shell, run `eval "$(docker-machine env -u)"`.


At this point, your system is available to install <%=vars.product_name %>. Follow the instructions in [Installing <%=vars.product_name %>](installing.html) to continue.  
