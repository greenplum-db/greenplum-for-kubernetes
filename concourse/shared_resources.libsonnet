
local timetriggers = {
    weekdays: [ "Monday", "Tuesday", "Wednesday", "Thursday", "Friday" ],
    morning: {start: "5:00 AM", stop: "6:00 AM"},
    evening: {start: "10:00 PM", stop: "11:00 PM"},
    weekday_trigger(name, days=self.weekdays, time_range=self.morning): {
        name: name,
        type: "time",
        source: {
            location: "America/Los_Angeles",
            days: days,
            start: time_range.start,
            stop: time_range.stop,
        },
    },
};

{
resource_types: [
    {
        name: "gcs-resource",
        type: "registry-image",
        source: {
            repository: "frodenas/gcs-resource",
            tag: "v0.4.1"
        }
    },
    {
        name: "pivnet",
        type: "registry-image",
        source: {
            repository: "pivotalcf/pivnet-resource",
            tag: "latest-final"
        }
    },
],
resources+: [
    # Time
    timetriggers.weekday_trigger("evening-delete-trigger", time_range=timetriggers.evening) + {dev_pipeline:: 'yes'},

    # Source repositories
    {
        name: "greenplum-for-kubernetes",
        dev_pipeline:: 'yes',
        type: "git",
        source: {
            uri: "git@github.com:pivotal/greenplum-for-kubernetes",
            branch: $.tla.git_branch,
            private_key: "((gpcloud-greenplum-for-kubernetes-git-deploy-key))",
            ignore_paths: [
                "doc/*",
                "README*"
            ]
        }
    },
    {
        name: "greenplum-for-kubernetes-release-tag-observer",
        type: "git",
        source: {
            uri: "git@github.com:pivotal/greenplum-for-kubernetes",
            branch: "master",
            private_key: "((gpcloud-greenplum-for-kubernetes-git-deploy-key))",
            # official tags must be annotated git tags, have prefix "v", and have just one digit for major version
            tag_filter: "v[0-9].*"
        }
    },
    {
        name: "greenplum-for-kubernetes-receipts",
        type: "git",
        source: {
            uri: "git@github.com:pivotal/greenplum-for-kubernetes-receipts",
            branch: "master",
            private_key: "((greenplum-for-kubernetes-receipts-git-deploy-key))"
        }
    },
    {
        name: "docker-in-concourse",
        dev_pipeline:: 'yes',
        type: "git",
        source: {
            uri: "git@github.com:pivotal/docker-in-concourse",
            private_key: "((docker-in-concourse-github-key))"
        }
    },
    {
        name: "gpdb_5X_src",
        type: "git",
        source: {
            branch: "5X_STABLE",
            uri: "((gpdb-git-remote))",
            ignore_paths: [
                "gpdb-doc/*",
                "README*"
            ]
        }
    },
    # Docker images
    {
        name: "gpdb-ubuntu1804-test-gp4k8s",
        dev_pipeline:: 'yes',
        type: "registry-image",
        source: {
            repository: "gcr.io/gp-kubernetes/gpdb-ubuntu18.04-test-gp4k8s",
            tag: "latest",
            username: "_json_key",
            password: "((gcp.svc_acct_key))"
        }
    },
    {
        name: "ubuntu16.04-test",
        type: "registry-image",
        source: {
            repository: "pivotaldata/ubuntu16.04-test",
            tag: "gpdb5-latest"
        }
    },
    # Blob storage
    {
        name: "gp-kubernetes-rc-release",
        dev_pipeline:: 'yes',
        type: "gcs-resource",
        source: {
            bucket: "gp-kubernetes-ci-release",
            json_key: "((gcp.svc_acct_key))",
            regexp: "greenplum-for-kubernetes-v(.*).tar.gz"
        }
    },
    {
        name: "gp-kubernetes-rc-release-receipts",
        dev_pipeline:: 'yes',
        type: "gcs-resource",
        source: {
            bucket: "greenplum-for-kubernetes-release-receipts",
            json_key: "((gcp.svc_acct_key))",
            regexp: "greenplum-for-kubernetes-v(.*)-receipt.txt"
        }
    },
    {
        name: "gp-kubernetes-green-release",
        type: "gcs-resource",
        source: {
            bucket: "gp-kubernetes-ci-release",
            json_key: "((gcp.svc_acct_key))",
            regexp: "green/greenplum-for-kubernetes-v(.*).tar.gz"
        }
    },
    {
        name: "gp-kubernetes-tagged-release-tarball",
        type: "gcs-resource",
        source: {
            bucket: "greenplum-for-kubernetes-release",
            json_key: "((gcp.svc_acct_key))",
            regexp: "greenplum-for-kubernetes-v(\\d+\\.\\d+\\.\\d+(-(alpha|beta|rc)\\.\\d+)?)\\.tar\\.gz"
        }
    },
    {
        name: "greenplum-debian-binary",
        dev_pipeline:: 'yes',
        type: "gcs-resource",
        source: {
            bucket: "pivotal-gpdb-concourse-resources-intermediates-prod",
            json_key: "((gpdb-prod-gcs.svc-acct-key))",
            versioned_file: "gpdb6-integration-testing/deb_gpdb_ubuntu18.04/greenplum-db-ubuntu18.04-amd64.deb",
        }
    },
    {
        name: "greenplum-debian-binary-stable",
        type: "gcs-resource",
        source: {
            bucket: "pivotal-gpdb-concourse-resources-prod",
            json_key: "((gpdb-prod-gcs.svc-acct-key))",
            regexp: "server/released/gpdb6/greenplum-db-(.*)-ubuntu18.04-amd64.deb",
        }
    },
    {
        name: "deb_package_open_source_ubuntu16",
        type: "s3",
        source: {
            access_key_id: "((bucket-access-key-id))",
            # NOTE: Hard-coding bleeding edge bucket name
            bucket: "gpdb5-stable-concourse-builds",
            region_name: "((aws-region))",
            secret_access_key: "((bucket-secret-access-key))",
            versioned_file: "((deb_package_open_source_ubuntu16_versioned_file))"
        }
    },
],
}
