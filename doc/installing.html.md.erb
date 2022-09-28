<% set_title("Installing", product_name_long) %>

This topic describes how to install <%=vars.product_name_long %>. The installation process involves loading the <%=vars.product_name %> container images into your container registry, and then using the `helm` package manager to install the Greenplum Operator resource in Kubernetes. After the Greenplum Operator resource is available, you can interact with it to deploy and manage Greenplum clusters in Kubernetes.

## Prerequisites

Before you install <%=vars.product_name %>, ensure that you have installed all required software and prepared your Kubernetes environment as described in [Prerequisites](prepare-k8s.html). Also, ensure that any previous <%=vars.product_name %> installation has been uninstalled as described in [Uninstalling <%=vars.product_name %>](uninstalling.html).

## Procedure

Follow these steps to download and install the <%=vars.product_name %> container images, and install the Greenplum Operator resource.

1. Download the <%=vars.product_name %> software from [VMware Tanzu Network](https://network.pivotal.io/products/greenplum-for-kubernetes). The download file has the name: `greenplum-for-kubernetes-<version>.tar.gz`.

2. Go to the directory where you downloaded the <%=vars.product_name_long %> distribution, and unpack the downloaded software. For example:

    ```bash
    $ cd ~/Downloads
    $ tar xzf greenplum-for-kubernetes-*.tar.gz
    ```

    The above command unpacks the distribution into a new directory named `greenplum-for-kubernetes-<version>`.

3. Go into the new `greenplum-for-kubernetes-<version>` directory:

    ``` bash
    $ cd ./greenplum-for-kubernetes-*/
    ```

4. **(This step is for Minikube deployments only.)** Ensure that the local docker daemon interacts with the Minikube docker container registry:

    ``` bash
    $ eval $(minikube docker-env)
    ```

    **Note:** To undo this docker setting in the current shell, run `eval "$(docker-machine env -u)"`.

5. Load the Greenplum Database Docker image to the local Docker registry:

    ``` bash
    $ docker load -i ./images/greenplum-for-kubernetes
    ```

    ``` bash
    b7f7d2967507: Loading layer [==================================================>]  65.58MB/65.58MB
    a6ebef4a95c3: Loading layer [==================================================>]  991.2kB/991.2kB
    838a37a24627: Loading layer [==================================================>]  15.87kB/15.87kB
    ...
    Loaded image: greenplum-for-kubernetes:v2.0.0
    ```

6. Load the Greenplum Operator Docker image to the Docker registry:

    ``` bash
    $ docker load -i ./images/greenplum-operator
    ```
    ```
    99c6db5a9ef4: Loading layer [==================================================>]  40.01MB/40.01MB
    4a8dbd663d0b: Loading layer [==================================================>]  1.143MB/1.143MB
    Loaded image: greenplum-operator:v2.0.0
    ```

7. Verify that both Docker images are now available:

    ```bash
    $ docker images "greenplum-*"
    ```

    ``` bash
    REPOSITORY                 TAG          IMAGE ID            CREATED             SIZE
    greenplum-operator         v2.0.0       ab4f09ba5cba        22 hours ago        112MB
    greenplum-for-kubernetes   v2.0.0       c508bc4f24f6        22 hours ago        2.33GB
    ```

8. **(Skip this step if you are using local docker images such as on Minikube.)** If you want to push the Greenplum docker images to a different container registry:

    1.  Set the project name and image repo name and then use Docker to push the images. For example, to push the images to Google Cloud Registry using the current Google Cloud project name:

        ```bash
        $ gcloud auth configure-docker
        
        $ PROJECT=$(gcloud config list core/project --format='value(core.project)')
        $ IMAGE_REPO="gcr.io/${PROJECT}"
        
        $ GREENPLUM_IMAGE_NAME="${IMAGE_REPO}/greenplum-for-kubernetes:$(cat ./images/greenplum-for-kubernetes-tag)"
        $ docker tag $(cat ./images/greenplum-for-kubernetes-id) ${GREENPLUM_IMAGE_NAME}
        $ docker push ${GREENPLUM_IMAGE_NAME}
        
        $ OPERATOR_IMAGE_NAME="${IMAGE_REPO}/greenplum-operator:$(cat ./images/greenplum-operator-tag)"
        $ docker tag $(cat ./images/greenplum-operator-id) ${OPERATOR_IMAGE_NAME}
        $ docker push ${OPERATOR_IMAGE_NAME}
        ```

    2. Create a docker-registry secret `regsecret` for pods to be able to fetch images from remote container registries. For example:

        ```bash
        # For GCR
        kubectl create secret  docker-registry  regsecret \
        --docker-server=https://gcr.io \
        --docker-username=_json_key \
        --docker-password="$(cat key.json)"

        #For ECR
        TOKEN=`aws ecr --region=$REGION get-authorization-token --output text --query authorizationData[].authorizationToken | base64 -d | cut -d: -f2`
        kubectl create secret docker-registry regsecret \
        --docker-server=https://${ACCOUNT}.dkr.ecr.${REGION}.amazonaws.com \
        --docker-username=AWS \
        --docker-password="${TOKEN}"

        #For Harbor
        kubectl create secret docker-registry regsecret \
        --docker-server=${HARBOR_URL} \
        --docker-username=${HARBOR_USER} \
        --docker-password="${HARBOR_PASSWORD}"
        ```

        **Note:** The `base64 -d` command in the above ECR section assumes you are using the GNU version of `base64`. If you are using a MacOS system, use the BSD syntax `base64 -D` instead.

    3. Verify the `regsecret`.

        ```bash
        ./workspace/samples/scripts/regsecret-test.bash ${OPERATOR_IMAGE_NAME}
        ```

        The output of above command should print `GREENPLUM-OPERATOR TEST OK`
        If this verification fails for Google Cloud Registry, then make sure the `key.json` file has the `roles/storage.objectViewer` role.

    4. If you pushed the Greenplum Operator and Greenplum Database Docker images to a container registry, create a new YAML file in the `workspace subdirectory` with two lines to indicate the registry where you pushed the images. For example, if you are using Google Cloud Registry you would add properties similar to:

        ```bash
        cat <<EOF >workspace/operator-values-overrides.yaml
        operatorImageRepository: ${IMAGE_REPO}/greenplum-operator
        greenplumImageRepository: ${IMAGE_REPO}/greenplum-for-kubernetes
        EOF
        ```

        Be sure to replace the project and repository names with the actual names used in your deployment.
        
        **Note:** If you did not tag the images with a container registry prefix or project name (for example, if you are using your own local Minikube deployment), then you can skip this step.
    
9. (Optional.) If you want to use a non-default logging level (for example, to enable `debug` logging), then follow the instructions in [Enabling Debug Logging](troubleshooting.html#debug).

10. (Optional.) If you want to specify a node for the operator to run on, first apply a label to the node.

    ```bash
    kubectl label node <node name> <key>=<value>
    ```

    Then, edit the `operator-values-overrides.yaml` file to include a matching set of key/value label selectors:

    ```
    operatorWorkerSelector: {
      <key>: "<value>"
      [ ... ]
    }
    ```
    See the documentation on the manifest's [workerSelector attribute](operator-reference.html#workerSelector) for more information on how <%=vars.product_name %> handles label selectors.

11. Use `helm` to create a new Greenplum Operator release, specifying the YAML configuration file if you created one. For example, to create a new release with the name "greenplum-operator":

    ```bash
    $ helm install greenplum-operator -f workspace/operator-values-overrides.yaml operator/
    ```

    If you did not create a YAML configuration file (as in the case with Minikube) omit the `-f` option:

    ``` bash
    $ helm install greenplum-operator operator/
    ```

    Helm begins installing the new release into the Kubernetes namespace specified in the current Kubernetes context. If you want to install into a different namespace, include the `--namespace` option in the `helm` command.
    
    <br/>The command displays the following message and concludes with a link to this documentation:

    ``` bash
    NAME: greenplum-operator
    LAST DEPLOYED: Wed May 13 09:55:06 2020
    NAMESPACE: default
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    NOTES:
    greenplum-operator has been installed.

    Please see documentation at:
    http://greenplum-kubernetes.docs.pivotal.io/
    ```

    Installing the operator creates a new service account named `greenplum-system-operator`.  It is for internal use, but it is visible if you use the `kubectl get serviceaccount` command:

    ``` bash
    $ kubectl get serviceaccount
    ```
    ``` bash
    NAME                        SECRETS   AGE
    default                     1         12m
    greenplum-system-operator   1         8m56s
    ```


12. Use `watch kubectl get all` to monitor the progress of the deployment. The deployment is complete when the Greenplum Operator pod is in the `Running` state and the replica set are available. For example:

    ``` bash
    $ watch kubectl get all
    ```
    ``` bash
    NAME                                      READY   STATUS    RESTARTS   AGE
    pod/greenplum-operator-6ff95b6b79-kq9vr   1/1     Running   0          24s

    NAME                                                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
    service/greenplum-validating-webhook-service-6ff95b6b79-kq9vr   ClusterIP   10.106.60.103   <none>        443/TCP   22s
    service/kubernetes                                              ClusterIP   10.96.0.1       <none>        443/TCP   4m14s

    NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/greenplum-operator   1/1     1            1           24s

    NAME                                            DESIRED   CURRENT   READY   AGE
    replicaset.apps/greenplum-operator-6ff95b6b79   1         1         1       24s
    ```

13. Check the logs of the operator to ensure that it is running properly.

    ``` bash
    $ kubectl logs -l app=greenplum-operator
    ```
    ``` bash
    {"level":"INFO","ts":"2020-05-13T18:17:32.173Z","logger":"controller-runtime.manager","msg":"starting metrics server","path":"/metrics"}
    {"level":"INFO","ts":"2020-05-13T18:17:32.173Z","logger":"admission","msg":"starting greenplum validating admission webhook server"}
    {"level":"INFO","ts":"2020-05-13T18:17:32.273Z","logger":"controller-runtime.controller","msg":"Starting Controller","controller":"greenplumcluster"}
    {"level":"INFO","ts":"2020-05-13T18:17:32.273Z","logger":"controller-runtime.controller","msg":"Starting Controller","controller":"greenplumpxfservice"}
    {"level":"INFO","ts":"2020-05-13T18:17:32.276Z","logger":"admission","msg":"CertificateSigningRequest: created"}
    {"level":"INFO","ts":"2020-05-13T18:17:32.291Z","logger":"admission","msg":"ValidatingWebhookConfiguration: created"}
    {"level":"INFO","ts":"2020-05-13T18:17:32.373Z","logger":"controller-runtime.controller","msg":"Starting workers","controller":"greenplumpxfservice","worker count":1}
    {"level":"INFO","ts":"2020-05-13T18:17:32.373Z","logger":"controller-runtime.controller","msg":"Starting workers","controller":"greenplumcluster","worker count":1}
    ```

At this point, you can interact with the Greenplum Operator to deploy new Greenplum clusters or manage existing Greenplum clusters. See [About the Greenplum Operator](using.html).
