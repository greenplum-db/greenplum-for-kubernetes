local tla = import "tla.libsonnet";
local concourse = import "concourse.libsonnet";
local creds = import "credentials.libsonnet";
local slack = import "slack.libsonnet";
local shared_resources = import "shared_resources.libsonnet";
local integration_tests = import "integration_tests.libsonnet";
local push_to_pivnet = import "push_to_pivnet.libsonnet";
local ubuntu_gpdb_ent = import "ubuntu_gpdb_ent.libsonnet";

local main_pipeline = {
jobs+: [
    {
        name: "create-rc-release",
        group:: "Dev daily",
        dev_pipeline:: 'yes',
        plan: [
            concourse.InParallel([
                {
                    get: "greenplum-for-kubernetes",
                    passed: [ "create-ubuntu-gpdb-ent-dev" ],
                    trigger: true,
                },
                { get: "docker-in-concourse" },
                { get: "gpdb-ubuntu1804-test-gp4k8s" },
                {
                    get: "greenplum-debian-binary",
                    passed: [ "create-ubuntu-gpdb-ent-dev" ],
                },
            ]),
            {
                task: "build and push greenplum-for-kubernetes image",
                image: "gpdb-ubuntu1804-test-gp4k8s",
                privileged: true,
                config: {
                    platform: "linux",
                    inputs: [
                        { name: "docker-in-concourse" },
                        { name: "greenplum-for-kubernetes" },
                        { name: "greenplum-debian-binary" },
                    ],
                    outputs: [
                        { name: "gp-kubernetes-rc-release-output" },
                        { name: "gp-kubernetes-rc-release-receipt-output" },
                    ],
                    run: {
                        path: "greenplum-for-kubernetes/concourse/scripts/build_and_push_release.sh"
                    },
                    params: {
                        GCP_SVC_ACCT_KEY: "((gcp.svc_acct_key))",
                        GCP_PROJECT: "((gcp.project_name))",
                        DEV_BUILD: true,
                        CONCOURSE_BUILD: true,
                        PIPELINE_NAME: $.tla.pipeline_name,
                    }
                }
            },
            {
                put: "gp-kubernetes-rc-release",
                params: {
                    file: "gp-kubernetes-rc-release-output/greenplum*.tar.gz"
                }
            },
            {
                put: "gp-kubernetes-rc-release-receipts",
                params: {
                    file: "gp-kubernetes-rc-release-receipt-output/greenplum-for-kubernetes-v*-receipt.txt"
                }
            }
        ],
        on_failure: slack.alert
    },
    {
        name: "validate_gpdb_5X_debian_oss",
        group:: "Misc",
        plan: [
            concourse.InParallel([
                { get: "greenplum-for-kubernetes" },
                { get: "ubuntu16.04-test" },
                { get: "gpdb_5X_src" },
                { get: "deb_package_open_source_ubuntu16", trigger: true },
            ]),
            {
                task: "oss-validation",
                file: "greenplum-for-kubernetes/concourse/tasks/oss-validation.yml",
                image: "ubuntu16.04-test",
                params: {
                    DEBIAN_PACKAGE: "((deb_package_open_source_ubuntu16_versioned_file))"
                }
            }
        ],
        on_failure: slack.alert
    },
    {
        name: "make-new-release-upon-tagging",
        group:: "Release",
        plan: [
            concourse.InParallel([
                {
                    get: "greenplum-for-kubernetes-release-tag-observer",
                    passed: [ "create-ubuntu-gpdb-ent-release" ],
                    trigger: true
                },
                { get: "gpdb-ubuntu1804-test-gp4k8s" },
                { get: "docker-in-concourse" },
                {
                    get: "greenplum-debian-binary-stable",
                    passed: [ "create-ubuntu-gpdb-ent-release" ],
                },
                {
                    get: "gp-kubernetes-rc-release",
                    passed: [ "mark-green-build" ],
                },
                { get: "greenplum-for-kubernetes-receipts" },
            ]),
            {
                task: "build and push greenplum-for-kubernetes image",
                image: "gpdb-ubuntu1804-test-gp4k8s",
                privileged: true,
                config: {
                    platform: "linux",
                    inputs: [
                        { name: "docker-in-concourse" },
                        {
                            name: "greenplum-for-kubernetes-release-tag-observer",
                            path: "greenplum-for-kubernetes",
                        },
                        {
                            name: "greenplum-debian-binary-stable",
                            path: "greenplum-debian-binary",
                        }
                    ],
                    outputs: [
                        { name: "gp-kubernetes-rc-release-output" },
                        { name: "gp-kubernetes-rc-release-receipt-output" },
                    ],
                    run: {
                        path: "greenplum-for-kubernetes/concourse/scripts/build_and_push_release.sh"
                    },
                    params: {
                        GCP_SVC_ACCT_KEY: "((gcp.svc_acct_key))",
                        GCP_PROJECT: "((gcp.project_name))",
                        DEV_BUILD: false,
                        CONCOURSE_BUILD: true,
                        PIPELINE_NAME: $.tla.pipeline_name,
                    }
                }
            },
            {
                task: "commit receipt",
                image: "gpdb-ubuntu1804-test-gp4k8s",
                config: {
                    platform: "linux",
                    inputs: [
                       {
                           name: "greenplum-for-kubernetes-release-tag-observer",
                           path: "greenplum-for-kubernetes",
                       },
                        { name: "greenplum-for-kubernetes-receipts" },
                        {
                            name: "gp-kubernetes-rc-release-receipt-output",
                            path: "gp-kubernetes-rc-release-receipts",
                        },
                    ],
                    outputs: [
                        { name: "greenplum-for-kubernetes-receipts-output" },
                    ],
                    run: {
                        path: "greenplum-for-kubernetes/concourse/scripts/commit-and-tag-receipt.bash"
                    }
                }
            },
            {
                put: "gp-kubernetes-tagged-release-tarball",
                params: {
                    file: "gp-kubernetes-rc-release-output/greenplum*.tar.gz"
                }
            },
            {
                put: "greenplum-for-kubernetes-receipts",
                params: {
                    repository: "greenplum-for-kubernetes-receipts-output",
                    rebase: true
                }
            },
        ],
        on_failure: slack.alert_and_sleep
    },
    {
        name: "mark-green-build",
        group:: "Dev daily",
        max_in_flight: 1,
        plan: [
            concourse.InParallel([
                { get: "greenplum-for-kubernetes" },
                {
                    get: "gp-kubernetes-rc-release",
                    passed: [
                        "gke-integration-test-dev",
                        "singlenode-kubernetes"
                    ],
                    trigger: true
                },
                { get: "gpdb-ubuntu1804-test-gp4k8s" },
            ]),
            {
                task: "mark tested rc release as green",
                image: "gpdb-ubuntu1804-test-gp4k8s",
                config: {
                    platform: "linux",
                    inputs: [
                        { name: "greenplum-for-kubernetes" },
                        { name: "gp-kubernetes-rc-release" },
                    ],
                    outputs: [
                        { name: "gp-kubernetes-green-output" },
                    ],
                    params: creds.gke + {
                        ADDITIONAL_TAG: "latest-green"
                    },
                    run: {
                        path: "greenplum-for-kubernetes/concourse/scripts/mark-green-build.bash"
                    }
                }
            },
            {
                put: "gp-kubernetes-green-release",
                params: {
                    file: "gp-kubernetes-green-output/greenplum*.tar.gz"
                }
            },
        ],
        on_failure: slack.alert
    },
    {
        name: "cleanup-resources",
        group:: "Dev daily",
        plan: [
            concourse.InParallel([
                { get: "evening-delete-trigger", trigger: true },
                { get: "greenplum-for-kubernetes" },
                { get: "gpdb-ubuntu1804-test-gp4k8s" },
            ]),
            local DeleteTask = {
                local task = self,
                script:: error 'Must provide script',
                params:: {},

                task: error 'Must provide "task"',
                image: "gpdb-ubuntu1804-test-gp4k8s",
                config: {
                    platform: "linux",
                    inputs: [
                        { name: "greenplum-for-kubernetes" },
                    ],
                    params: {
                        GCP_SVC_ACCT_KEY: "((gcp.svc_acct_key))",
                        GCP_PROJECT: "((gcp.project_name))"
                    } + task.params,
                    run: {
                        path: task.script
                    }
                }
            };
            concourse.InParallel([
                DeleteTask {
                    task: "Delete old greenplum-for-kubernetes gcloud container image",
                    script: "greenplum-for-kubernetes/concourse/scripts/delete-gcloud-container-images.bash",
                    params: {
                        IMAGE: "greenplum-for-kubernetes",
                    },
                },
                DeleteTask {
                    task: "Delete old greenplum-operator gcloud container image",
                    script: "greenplum-for-kubernetes/concourse/scripts/delete-gcloud-container-images.bash",
                    params: {
                        IMAGE: "greenplum-operator",
                    },
                },
                DeleteTask {
                    task: "Delete orphaned load balancers",
                    script: "greenplum-for-kubernetes/concourse/scripts/delete-orphaned-loadbalancers-gcp.bash",
                    params: {
                        REGION: "us-central1"
                    },
                }
            ]),
        ]
    },
    {
        name: "Delete perf k8s clusters",
        group:: "Dev daily",
        plan: [
            concourse.InParallel([
                { get: "evening-delete-trigger", trigger: true },
                { get: "greenplum-for-kubernetes" },
                { get: "gpdb-ubuntu1804-test-gp4k8s" },
            ]),
            {
                task: "delete concourse-benchmark-scale-1000-cluster",  # 1T perf pipeline
                file: "greenplum-for-kubernetes/concourse/tasks/delete-k8s-cluster.yml",
                image: "gpdb-ubuntu1804-test-gp4k8s",
                params: creds.gke + {CLUSTER_NAME: "concourse-benchmark-scale-1000-cluster"},
            },
        ]
    },
    {
        name: "validate_cross_team",
        group:: "Misc",
        plan: [
            concourse.InParallel([
                {
                    get: "gp-kubernetes-green-release",
                    trigger: true,
                    passed: [ "mark-green-build" ],
                },
                { get: "gpdb-ubuntu1804-test-gp4k8s" },
                { get: "greenplum-for-kubernetes" },
                {
                    get: "gp-kubernetes-rc-release",
                    params: { unpack: true },
                },
            ]),
            {
                task: "provision-and-deploy-greenplum-cluster",
                file: "greenplum-for-kubernetes/concourse/cross-team/provision-cluster.yml",
                image: "gpdb-ubuntu1804-test-gp4k8s",
                params: creds.gke { CLUSTER_NAME: "concourse-cross-team-cluster" } + {
                    GP_INSTANCE_NAME: "my-greenplum"
                }
            },
            {
                task: "cleanup-env",
                image: "gpdb-ubuntu1804-test-gp4k8s",
                config: {
                    platform: "linux",
                    inputs: [
                        { name: "greenplum-for-kubernetes" },
                    ],
                    params: creds.gke { CLUSTER_NAME: "concourse-cross-team-cluster" } + {
                        GP_INSTANCE_NAME: "my-greenplum"
                    },
                    run: {
                        path: "greenplum-for-kubernetes/concourse/cross-team/cleanup-env.bash"
                    }
                }
            }
        ],
        on_failure: slack.alert_and_sleep
    }
]
};

tla.Apply(
    // pipelines:
    main_pipeline +
    shared_resources +
    integration_tests +
    slack +
    push_to_pivnet +
    ubuntu_gpdb_ent +
    // mutators:
    concourse.AutoGroups(groupOrder=["Dev daily", "Release", "Misc"]) +
    concourse.FilterPipelineForProdOrDev
)
