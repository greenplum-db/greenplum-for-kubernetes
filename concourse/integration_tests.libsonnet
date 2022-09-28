local concourse = import "concourse.libsonnet";
local creds = import "credentials.libsonnet";
local escape_regex = import "escape_regex.libsonnet";
local slack = import "slack.libsonnet";

{
local integration_test_job = {
    local this = self,
    creds:: error 'Must provide "creds"',
    cluster_name:: if $.tla.pipeline_name == "" then this.name+"-concourse"
        else "gke-" + std.substr($.tla.pipeline_name, 0, 26) + "-concourse",
    gp4k_source:: {},
    release_tarball:: error 'Must provide "release_tarball"',
    additional_inputs:: [],
    test_task_inputs:: [],
    gke_master_cidr:: {},

    local IntegrationTestTask(type) = {
        task: "operator integration "+type+" tests",
        image: "gpdb-ubuntu1804-test-gp4k8s",
        config: {
            platform: "linux",
            inputs: this.test_task_inputs + [
                { name: "gp-kubernetes-release" },
                {
                    name: "greenplum-for-kubernetes",
                },
            ],
            params: this.creds + {
                CONCOURSE_BUILD: true,
                INTEGRATION_TYPE: type,
                POLL_TIMEOUT: 300,
                K8S_CLUSTER_NAME:
                    if (std.length(this.cluster_name) > 40) then
                        error std.format('cluster name "%s" must be no longer than 40 chars for GKE', this.cluster_name)
                    else this.cluster_name,
            },
            run: {
                path: "greenplum-for-kubernetes/concourse/scripts/test-greenplum-operator-integration.bash"
            }
        }

    },

    name: error 'Must provide "name"',
    max_in_flight: 1,
    plan: [
        concourse.InParallel(this.additional_inputs + [
            {
                get: "gp-kubernetes-release",
                trigger: true,
                params: {
                    unpack: true
                },
            } + this.release_tarball,
            { get: "gpdb-ubuntu1804-test-gp4k8s" },
            {
                get: "greenplum-for-kubernetes-old-release",
                params: {
                    unpack: true
                },
            },
            {
                get: "greenplum-for-kubernetes",
            } + this.gp4k_source,
        ]),
        {
            task: "test-helm-lint",
            image: "gpdb-ubuntu1804-test-gp4k8s",
            config: {
                platform: "linux",
                inputs: [
                    { name: "greenplum-for-kubernetes" },
                ],
                run: {
                    path: "greenplum-for-kubernetes/test/helm_lint.bash"
                }
            }
        },
        {
            task: "provision-k8s-cluster",
            file: "greenplum-for-kubernetes/concourse/tasks/provision-k8s-cluster.yml",
            image: "gpdb-ubuntu1804-test-gp4k8s",
            params: this.creds + this.gke_master_cidr + {
                K8S_CLUSTER_NAME: this.cluster_name,
                K8S_CLUSTER_NODE_COUNT: 8,
            }
        },
        IntegrationTestTask(type="default"),
        IntegrationTestTask(type="ha"),
        IntegrationTestTask(type="upgrade") {
            config+: {
                inputs+: [
                    { name: "greenplum-for-kubernetes-old-release" },
                ]
            }
        },
        IntegrationTestTask(type="component") {
            config+: {
                params+: creds.gp_kubernetes_ci_s3,
            }
        },
    ],
    on_failure: slack.alert_and_sleep
},

local gke_job = integration_test_job {
    creds: creds.gke {
        K8S_CLUSTER_TYPE: "gke-private",
        GCP_NETWORK: "bosh-network",
    },
},

local dev_props = {
    group:: "Dev daily",
    gp4k_source: {
        passed: [ "create-rc-release" ],
    },
    release_tarball: {
        resource: "gp-kubernetes-rc-release",
        passed: [ "create-rc-release" ],
    },
},

local release_props = {
    group:: "Release",
    gp4k_source: {
        resource: "greenplum-for-kubernetes-release-tag-observer",
        passed: [ "make-new-release-upon-tagging" ],
    },
    release_tarball: {
        resource: "gp-kubernetes-tagged-release-tarball",
        passed: [ "make-new-release-upon-tagging" ],
    },
},

local integration_test_jobs = [
    gke_job + dev_props {
        local this = self,
        name: "gke-integration-test-dev",
        gke_master_cidr: {
            GCP_GKE_MASTER_CIDR: "172.16.0.0/28",
        },
        creds : gke_job.creds + {
            GCP_SUBNETWORK: this.name + "-concourse-network"
        }
    },
    integration_test_job + dev_props {
        name: "gke-integration-test-dev",
        dev_pipeline:: 'only',
        creds: creds.gke {
            GCP_NETWORK: "default",
            GCP_SUBNETWORK: "default",
        },
    },
    gke_job + release_props {
        local this = self,
        name: "gke-integration-test-release",
        gke_master_cidr: {
            GCP_GKE_MASTER_CIDR: "172.16.0.16/28",
        },
        creds : gke_job.creds + {
            GCP_SUBNETWORK: this.name + "-concourse-network"
        }
    },
],

local singlenode_integration_test_jobs = [
    {
        name: "singlenode-kubernetes",
        group:: "Dev daily",
        max_in_flight: 1,
        plan: [
            concourse.InParallel([
                { get: "greenplum-for-kubernetes" },
                { get: "gpdb-ubuntu1804-test-gp4k8s" },
                {
                    get: "gp-kubernetes-rc-release",
                    passed: [ "create-rc-release" ],
                    trigger: true,
                    params: { unpack: true },
                },
            ]),
            {
                task: "test-singlenode-kubernetes",
                image: "gpdb-ubuntu1804-test-gp4k8s",
                config: {
                    platform: "linux",
                    inputs: [
                        { name: "greenplum-for-kubernetes" },
                        { name: "gp-kubernetes-rc-release" },
                    ],
                    params: creds.gke,
                    run: {
                        path: "greenplum-for-kubernetes/concourse/scripts/test-singlenode-kubernetes.bash"
                    }
                },
                on_failure: slack.alert_and_sleep
            }
        ]
    }
],

local prod_cleanup_tasks = [
{
    task: "Delete singlenode kubernetes cluster",
    image: "gpdb-ubuntu1804-test-gp4k8s",
    config: {
        platform: "linux",
        inputs: [
            { name: "greenplum-for-kubernetes" },
        ],
        params: {
            IMAGE: "greenplum-for-kubernetes",
            GCP_SVC_ACCT_KEY: "((gcp.svc_acct_key))",
            GCP_PROJECT: "((gcp.project_name))"
        },
        run: {
            path: "greenplum-for-kubernetes/concourse/scripts/delete-singlenode-kubernetes-cluster.bash"
        }
    }
},
{
    task: "Delete old unused pvcs and disks",
    image: "gpdb-ubuntu1804-test-gp4k8s",
    config: {
        platform: "linux",
        inputs: [
            { name: "greenplum-for-kubernetes" },
        ],
        params: {
            GCP_SVC_ACCT_KEY: "((gcp.svc_acct_key))",
            GCP_PROJECT: "((gcp.project_name))"
        },
        run: {
            path: "greenplum-for-kubernetes/concourse/scripts/delete-old-unused-pvc-disks.bash"
        }
    }
},
],

    resources+: [
        {
            name: "greenplum-for-kubernetes-old-release",
            dev_pipeline:: 'yes',
            type: "gcs-resource",
            source: {
                bucket: "greenplum-for-kubernetes-release",
                json_key: "((gcp.svc_acct_key))",
                regexp: "greenplum-for-kubernetes-(" + escape_regex("v2.2.0") + ").tar.gz"
            }
        },
    ],
    jobs+:
        integration_test_jobs +
        singlenode_integration_test_jobs +
        [{
            name: "delete-k8s-cluster",
            group:: "Dev daily",
            dev_pipeline:: "yes",
            plan: [
                concourse.InParallel([
                    { get: "evening-delete-trigger", trigger: true },
                    { get: "greenplum-for-kubernetes" },
                    { get: "gpdb-ubuntu1804-test-gp4k8s" },
                ]),
            ] +
            // "resources" is not manifested (b/c of the ".jobs" on the end), so it's not required
            local active_integration_test_jobs = ({jobs: integration_test_jobs, tla:: $.tla}+concourse.FilterPipelineForProdOrDev).jobs;
            [
                {
                    task: "delete-"+job.cluster_name,
                    file: "greenplum-for-kubernetes/concourse/tasks/delete-k8s-cluster.yml",
                    image: "gpdb-ubuntu1804-test-gp4k8s",
                    params: job.creds + {CLUSTER_NAME: job.cluster_name},
                }
                for job in active_integration_test_jobs
            ] +
            (if $.tla.prod then prod_cleanup_tasks else [])
        },
    ]
}

