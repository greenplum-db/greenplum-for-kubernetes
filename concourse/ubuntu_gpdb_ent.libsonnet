local concourse = import "concourse.libsonnet";
local creds = import "credentials.libsonnet";
local slack = import "slack.libsonnet";

{
local create_ubuntu_gpdb_ent = {
    local this = self,
    name: error 'Must provide "name"',

    additional_inputs:: [],

    plan: [
        concourse.InParallel(this.additional_inputs + [
            { get: "docker-in-concourse" },
            { get: "gpdb-ubuntu1804-test-gp4k8s" },
            { get: "madlib-deb-ubuntu18" },
            { get: "pxf-gp6-deb-ubuntu18" },
            {
                get: "gpbackup-ubuntu18",
                params: {
                    globs: [
                        "pivotal_greenplum_backup_restore-*.tar.gz",
                        "version",
                    ],
                },
            },
        ]),
        {
            task: this.name,
            image: "gpdb-ubuntu1804-test-gp4k8s",
            privileged: true,
            config: {
                platform: "linux",
                inputs: [
                    { name: "docker-in-concourse" },
                    { name: "greenplum-for-kubernetes" },
                    { name: "greenplum-debian-binary" },
                    { name: "madlib-deb-ubuntu18" },
                    { name: "pxf-gp6-deb-ubuntu18" },
                    { name: "gpbackup-ubuntu18" },
                ],
                params: creds.gke + creds.gpdb_ci_s3 + {
                    CONCOURSE_BUILD: true,
                    PIPELINE_NAME: $.tla.pipeline_name,
                },
                run: {
                    path: "greenplum-for-kubernetes/concourse/scripts/create-ubuntu-gpdb-ent.sh"
                }
            }
        }
    ],
    on_failure: slack.alert
},

# TODO: this resource type should be removed once we no longer use the madlib oss build
resource_types+: [
    {
        name: "static",
        type: "registry-image",
        source: {
            repository: "eugenmayer/concourse-static-resource",
            tag: "latest"
        }
    },
],

resources+: [
    {
        name: "madlib-deb-ubuntu18",
        dev_pipeline:: 'yes',
        type: "static",
        source: {
            uri: "https://dist.apache.org/repos/dist/release/madlib/1.17.0/apache-madlib-1.17.0-bin-Linux.deb",
            version_static: "1.17.0"
        }
    },
    {
        name: "pxf-gp6-deb-ubuntu18",
        dev_pipeline:: 'yes',
        type: "gcs-resource",
        source: {
            bucket: "pivotal-gpdb-concourse-resources-prod",
            json_key: "((gpdb-prod-gcs.svc-acct-key))",
            regexp: "pxf/published/gpdb6/pxf-gp6-(.*)-ubuntu18.04-amd64.deb",
        }
    },
    {
        name: "gpbackup-ubuntu18",
        dev_pipeline:: 'yes',
        type: "pivnet",
        source: {
            api_token: "((pivnet.api_token))",
            product_slug: "pivotal-gpdb-backup-restore",
            endpoint: "https://network.pivotal.io",
            sort_by: "semver"
        }
    },
],

jobs+: [
    create_ubuntu_gpdb_ent {
        name: "create-ubuntu-gpdb-ent-dev",
        group:: "Dev daily",
        dev_pipeline:: 'yes',
        additional_inputs:: [
            { get: "greenplum-for-kubernetes", trigger: true },
            { get: "greenplum-debian-binary", trigger: true },
        ],
    },
    create_ubuntu_gpdb_ent {
        name: "create-ubuntu-gpdb-ent-release",
        group:: "Release",
        additional_inputs:: [
            {
                get: "greenplum-for-kubernetes",
                resource: "greenplum-for-kubernetes-release-tag-observer",
                trigger: true
            },
            {
                get: "greenplum-debian-binary",
                resource: "greenplum-debian-binary-stable",
                trigger: true
            },
        ],
    },
],
}
