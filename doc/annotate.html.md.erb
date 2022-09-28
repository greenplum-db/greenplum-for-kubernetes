---
title: Annotating the Greenplum Service
---

Open Policy Agent (OPA) can be used to annotate Greenplum services as they are deployed, in order to implement policy rules. <%=vars.product_name_long %> provides sample files to deploy a mutating webhook. The webhook uses OPA with a Rego policy file to configure Greenplum services to use an AWS internal load balancer. Annotations are added and modified to the Greenplum services as they are deployed to the target namespace.

## <a id="prereq"></a>Prerequisites

All sample files are provided in the `workspace/samples/openpolicyagent` directory of your installation.

The sample configuration uses a self-signed certificate that is created via a template in the `stable/opa` helm chart. If you want to cert-manager instead of a self-signed certificate, consult the helm chart documentation for configuration details.

## <a id="procedure"></a>Procedure

Follow these steps to deploy the sample OPA webhook and apply the AWS load balancer annotation rules:

1. Go to the `workspace` subdirectory where you unpacked the <%=vars.product_name %> distribution for Kubernetes:

    ``` bash
    $ cd ./greenplum-for-kubernetes-*/workspace
    ```

1. Create a namespace for the mutating webhook. For example:

    ``` bash
    $ kubectl create namespace greenplum-opa
    ```
    ``` bash
    namespace/greenplum-opa created
    ```

1. Add the stable repo to helm:

    ``` bash
    $ helm repo add stable https://kubernetes-charts.storage.googleapis.com
    ```
    ``` bash
    "stable" has been added to your repositories
    ```
    ``` bash
    $ helm repo update
    ```
    ``` bash
    Hang tight while we grab the latest from your chart repositories...
    ...Successfully got an update from the "stable" chart repository
    Update Complete. ⎈ Happy Helming!⎈
    ```

1. Deploy OPA from the helm chart:

    ``` bash
    $ helm install -n greenplum-opa greenplum-opa -f ./samples/openpolicyagent/gpdb-config-opa.yaml stable/opa
    ```
    ``` bash
    NAME: greenplum-opa
    LAST DEPLOYED: Wed Jun 17 09:58:53 2020
    NAMESPACE: greenplum-opa
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    NOTES:
    Please wait while the OPA is deployed on your cluster.
    
    For example policies that you can enforce with OPA see https://www.openpolicyagent.org.
    
    If you installed this chart with the default values, you can exercise the sample policy.
    [...]
    You can query OPA to see the policies it has loaded (you will need to turn off authz as described above):
    
    export OPA_POD_NAME=$(kubectl get pods --namespace greenplum-opa -l "app=greenplum-opa" -o jsonpath="    {.items[0].metadata.name}")
    
    kubectl port-forward $OPA_POD_NAME 8080:443 --namespace greenplum-opa
    
    curl -k -s https://localhost:8080/v1/policies | jq -r '.result[].raw'
    ```

1. The `./samples/openpolicyagent/aws-load-balancer.rego` file adds the annotation necessary to use an AWS load balancer, and sets the load balancer address.  Edit the file as necessary to modify the Greenplum namespace (if you intend to deploy Greenplum into a non-default namespace) and the load balancer address. By default this file contains:

    ``` 
    package system

    patch[content] {
    	is_create_or_update

    	is_kind("Service")
    	is_name("greenplum")
    	is_namespace("default")
    	has_label("app", "greenplum")

    	content := add_annotation("service.beta.kubernetes.io/aws-load-balancer-internal", "true")
    }

    patch[content] {
    	is_create_or_update

    	is_kind("Service")
    	is_name("greenplum")
    	is_namespace("default")
    	has_label("app", "greenplum")

    	content := replace_at_jsonpath("/spec", "loadBalancerSourceRanges", ["10.32.4.1/32"])
    }
    ```

1. Apply the load balancer policy to the OPA service:

    ``` bash
    $ kubectl create configmap greenplum-loadbalancer-policy --from-file=./samples/ openpolicyagent/aws-load-balancer.rego -n greenplum-opa
    ```
    ``` bash
    configmap/greenplum-loadbalancer-policy created
    ```

1. Annotate the namespace where you intend to deploy Greenplum with `{"opa-controlled": "true"}`. For example:

    ``` bash
    $ kubectl label namespace/default opa-controlled=true
    ```
    ``` bash
    namespace/default labeled
    ```

1. Deploy the Greenplum cluster as you normally would. See [Deploying or Redeploying a Greenplum Cluster](deploy-operator.html). Annotations are added or modified during the deployment process. 

## <a id="applications"></a>Additional Applications

Although the sample Rego file configures Greenplum to use an AWS load balancer, you can apply any OPA policy that adds annotations or modifies an existing JSON path. (Accepting or rejecting a manifest based on a policy is not currently supported.)

For example, the following Rego contents direct the Greenplum operator to create a `NodePort` instead of a load balancer:

``` pre
package system
patch[content] {
        is_create_or_update
        is_kind("Service")
        is_name("greenplum")
        is_namespace("default")
        has_label("app", "greenplum")
        content := replace_at_jsonpath("/spec", "type", "NodePort")
}
```
