
# Greenplum for Kubernetes

This repo houses the Helm charts as well as code and resources that go into the
Docker images for a Greenplum on Kubernetes release.

```
git clone git@github.com:greenplum-db/greenplum-for-kubernetes.git
```

## Table of Contents

- [Run Unit and Lint Tests](#run-unit-and-lint-tests)
- [Deploy Greenplum for Kubernetes on Minikube](#deploy-greenplum-for-kubernetes-on-minikube)
  * [Setup](#setup)
  * [Build](#build)
  * [Deploy](#deploy)
  * [Integration Test](#integration-test)
- [Deploy Greenplum for Kubernetes on GKE](#deploy-greenplum-for-kubernetes-on-gke)
  * [Setup](#setup-1)
  * [Build](#build-1)
  * [Deploy](#deploy-1)
  * [Integration Test](#integration-test-1)

## Run Unit and Lint Tests

```
# runs only unit tests
make unit

# runs only linting
make lint

# runs unit and lint
make check
```

## Deploy Greenplum for Kubernetes on Minikube

### Setup

Install Minikube and it's pre-requisites by following instructions
[here](https://kubernetes.io/docs/tasks/tools/install-minikube/).
(If there are any problems, use `minikube delete` to delete the current minikube and then recreate it.)

Use HyperKit with minikube. It's much faster:

```bash
brew install docker-machine-driver-hyperkit
# Make sure to follow the `sudo chown` commands that brew outputs
minikube config set vm-driver hyperkit
```

Start minikube as below. You can adjust the resources to be lower if needed, but we recommend these resource settings
for improved performance and enabling larger clusters.

```bash
minikube start --memory 8192 --cpus 8 --kubernetes-version=v1.16.7 --disk-size 80g
```

Confirm kubectl can access the minikube.

```bash
kubectl get nodes

NAME       STATUS    ROLES     AGE       VERSION
minikube   Ready     master    21d       v1.10.0
```

Authenticate with gcloud to pull required images from gcr.io
```bash
gcloud auth login
gcloud auth configure-docker
```

### Build

Run the following command to change the local docker environment to point to the minikube docker.
You only need to run this command once (per shell).

```bash
eval $(minikube docker-env)
```

(Note: to undo this docker setting in current shell, run `eval "$(docker-machine env -u)"`)

Then, to build and upload the Greenplum images to minikube's docker registry.

```bash
make docker
```

The image should now have the tag `greenplum-for-kubernetes:<version>` as shown below:

```bash
$ docker images | grep greenplum
greenplum-operator                                  container-structure-test-in-docker                  2de1e0fa79b1        8 minutes ago       928MB
greenplum-operator                                  latest                                              24ca53fb90aa        8 minutes ago       111MB
greenplum-operator                                  v2.0.0-alpha.2.dev.7.gee6a7552                      24ca53fb90aa        8 minutes ago       111MB
greenplum-instance                                  container-structure-test-in-docker                  b90a82a6a746        10 minutes ago      928MB
greenplum-for-kubernetes                            latest                                              a2076cf10832        10 minutes ago      2.13GB
greenplum-for-kubernetes                            v2.0.0-alpha.2.dev.7.gee6a7552                      a2076cf10832        10 minutes ago      2.13GB
```

To clean up images:

```bash
# clean files and images directly created by our Makefile
make clean

# clean all the dangling docker images
make docker-clean
```

### Deploy

Our `make deploy` target creates a regsecret used by the pods to download images.
This requires a service account key for GCP. You must either place the key in a file named `key.json` inside the
`operator` directory or a set the GCP_SVC_ACCT_KEY environment variable for this to work.

To install Greenplum cluster in the minikube environment, run the command below:

```bash
make deploy
```

You can access the Greenplum cluster with `psql` running in master-0:

```bash
kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; psql"
```

If you want to access the Greenplum service outside the minikube and
you have a compatible "psql" executable in your path, you can do:

```bash
PGPORT=$(minikube service --format "{{.Port}}" --url greenplum) \
  PGHOST=$(minikube service --format "{{.IP}}" --url greenplum) \
  bash -c 'psql -U gpadmin -h $PGHOST -p $PGPORT'
```

To remove the Greenplum deployment:

```bash
make deploy-clean
```

### Integration Test

The integration tests work by deploying Greenplum for Kubernetes and running checks against the cluster.
Make sure you have performed the `Build` step before attempting to run the integration tests.

To run integration tests on minikube, execute the following command:

```bash
make integration
```

To run a specific integration test suite (from `greenplum-operator/integration/`), run:

```bash
make -C greenplum-operator integration-<suite> # e.g. `integration-ha`
```

## Deploy Greenplum for Kubernetes on GKE

### Setup

Create a cluster in GKE either with the web console or on the command line.

Run the following command to set the GKE cluster as your Kubernetes context for the command line

```bash
gcloud container clusters get-credentials <your-cluster-name> --project <your-project> --zone <cluster-zone>
```

Confirm kubectl can access the GKE cluster.

```bash
kubectl get nodes

NAME                                          STATUS   ROLES    AGE   VERSION
gke-test-default-pool-dba9314c-6xkg   Ready    <none>   20h   v1.15.9-gke.24
gke-test-default-pool-dba9314c-gjxc   Ready    <none>   20h   v1.15.9-gke.24
gke-test-default-pool-dba9314c-s83j   Ready    <none>   20h   v1.15.9-gke.24
gke-test-default-pool-dba9314c-tz4v   Ready    <none>   20h   v1.15.9-gke.24
gke-test-default-pool-dba9314c-wwc6   Ready    <none>   20h   v1.15.9-gke.24
gke-test-default-pool-dba9314c-x1cq   Ready    <none>   20h   v1.15.9-gke.24
```

Authenticate with gcloud to pull required images from gcr.io
```bash
gcloud auth login
gcloud auth configure-docker
```

### Build

The following command will build the docker images for the Greenplum instance locally
and then upload them to GCR so they can be used on your GKE cluster.

```bash
make gke-docker
```

The image should now have the tag `greenplum-for-kubernetes:<version>` as shown below:

```bash
$ docker images | grep greenplum
greenplum-operator                                  container-structure-test-in-docker                  2de1e0fa79b1        8 minutes ago       928MB
greenplum-operator                                  latest                                              24ca53fb90aa        8 minutes ago       111MB
greenplum-operator                                  v2.0.0-alpha.2.dev.7.gee6a7552                      24ca53fb90aa        8 minutes ago       111MB
gcr.io/gp-kubernetes/greenplum-operator             dev-24ca53fb90aa                                    24ca53fb90aa        8 minutes ago       111MB
greenplum-instance                                  container-structure-test-in-docker                  b90a82a6a746        10 minutes ago      928MB
greenplum-for-kubernetes                            latest                                              a2076cf10832        10 minutes ago      2.13GB
greenplum-for-kubernetes                            v2.0.0-alpha.2.dev.7.gee6a7552                      a2076cf10832        10 minutes ago      2.13GB
gcr.io/gp-kubernetes/greenplum-for-kubernetes       dev-a2076cf10832                                    a2076cf10832        10 minutes ago      2.13GB
```

A file will be created at `/tmp/.gp-kubernetes_gke_image_tags` containing the tags of the built images. This file will be used by the `make gke-deploy` target to reference the correct images.

To clean up images:

```bash
# clean files and images directly created by our Makefile
make clean

# clean all the dangling docker images
make docker-clean
```

### Deploy

Our `make gke-deploy` target creates a regsecret used by the pods to download images.
This requires a service account key for GCP. You must either place the key in a file named `key.json` inside the
`operator` directory or a set the GCP_SVC_ACCT_KEY environment variable for this to work.

To install Greenplum cluster in the GKE environment, run the command below. This will set up parameters for the operator
to use the docker images that were built with the `gke-docker` target.

```bash
make gke-deploy
```

You can access the Greenplum cluster with `psql` directly through the master pod with:

```bash
kubectl exec -it master-0 -- bash -c "source /usr/local/greenplum-db/greenplum_path.sh; psql"
```

To remove the Greenplum deployment:

```bash
make gke-deploy-clean
```

### Integration Test

The integration tests work by deploying Greenplum for Kubernetes and running checks against the cluster.
Make sure you have performed the `Build` step before attempting to run the integration tests.

To run integration tests on GKE, run the following command:

```bash
make gke-integration
```

To run a specific integration test suite (from `greenplum-operator/integration/`), run:

```bash
make -C greenplum-operator gke-integration-<suite> # e.g. `gke-integration-ha`
```
