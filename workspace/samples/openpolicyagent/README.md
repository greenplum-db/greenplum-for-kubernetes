# OPA - Example mutating webhook

This is a sample mutating webhook using Open Policy Agent to annotate the
service/greenplum service and set the loadBalancerSourceRanges field.

This uses a self-signed certificate that is created via the sprig template as
part of the stable/opa helm chart, but the helm chart itself optionally
supports using cert-manager. Consult the helm chart documentation for details
on how to utilize cert-manager certificates.

## Deployment

1. Create a namespace for the mutating webhook
   ```bash
   kubectl create namespace greenplum-opa
   ```
1. Deploy OPA from the helm chart
   ```bash
   helm repo add stable https://kubernetes-charts.storage.googleapis.com
   helm repo update
   helm install -n greenplum-opa greenplum-opa -f ./workspace/samples/openpolicyagent/gpdb-config-opa.yaml stable/opa
   ```
1. Edit './workspace/samples/openpolicyagent/aws-load-balancer.rego' with the desired namespace, annototion, and loadBalancerSourceRanges
1. Apply the load balancer policy to the OPA service
   ```bash
   kubectl create configmap greenplum-loadbalancer-policy --from-file=./workspace/samples/openpolicyagent/aws-load-balancer.rego -n greenplum-opa
   ```
1. Annotate the namespace where a Greenplum cluster will be deployed with '{"opa-controlled": "true"}'
   ```bash
   kubectl label namespace/default opa-controlled=true
   ```
1. Deploy the Greenplum cluster
   ```bash
   kubectl apply -f ./workspace/my-gp-instance.yaml
   ```

## Links

* [OPA helm chart documentation](https://github.com/helm/charts/blob/master/stable/opa/README.md)
* [OPA Rego policy language documentation](https://www.openpolicyagent.org/docs/v0.15.1/policy-language/)
