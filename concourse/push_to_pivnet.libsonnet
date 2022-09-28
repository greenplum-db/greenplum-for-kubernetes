local concourse = import "concourse.libsonnet";

local staging_resource = {
    name: "pivnet-staging-gp-kubernetes",
    type: "pivnet",
    source: {
        api_token: "((pivnet.api_token))",
        product_slug: "greenplum-for-kubernetes",
        endpoint: "https://pivnet-integration.cfapps.io",
        bucket: "pivotal-network-staging",
        access_key_id: "((pivnet.bucket_access_key_id))",
        secret_access_key: "((pivnet.bucket_secret_access_key))",
        region: "eu-west-1",
        sort_by: "semver"
    }
};

local production_resource = staging_resource {
    name: "pivnet-gp-kubernetes",
    source+: {
        endpoint: "https://network.pivotal.io",
        bucket: "pivotalnetwork",
    }
};

local license_resources = [
    {
        name: "greenplum-for-kubernetes-release-osl",
        type: "gcs-resource",
        source: {
            bucket: "greenplum-for-kubernetes-release-osl",
            json_key: "((gcp.svc_acct_key))",
            regexp: "open-source-licenses-v(.*).txt"
        }
    },
    {
        name: "greenplum-for-kubernetes-release-odp",
        type: "gcs-resource",
        source: {
            bucket: "greenplum-for-kubernetes-release-odp",
            json_key: "((gcp.svc_acct_key))",
            regexp: "VMware-greenplum-gp4k-(.*)-ODP.tar.gz"
        }
    },
    {
        name: "gpdb_open_source_osl_licenses",
        type: "gcs-resource",
        source: {
            access_key_id: "((bucket-access-key-id))",
            # NOTE: Hard-coding bleeding edge bucket name
            bucket: "pivotal-gpdb-concourse-resources-prod",
            json_key: "((gpdb-prod-gcs.svc-acct-key))",
            regexp: "osl/released/gpdb6/open_source_license_pivotal-gpdb-(6.0.0)-1787e8a-1566480846.txt"
        }
    },
];

local release_job(name, pivnet_resource, trigger_on_tarball=false) = {
    name: name,
    group:: "Release",
    plan: [
        concourse.InParallel([
            {
                get: "gp-kubernetes-tagged-release-tarball",
                passed: [
                    "gke-integration-test-release",
                ],
                trigger: trigger_on_tarball,
            },
            {
                get: "greenplum-for-kubernetes",
                passed: [ "create-rc-release" ],
            },
            { get: "greenplum-for-kubernetes-release-osl" },
            { get: "greenplum-for-kubernetes-release-odp" },
            { get: "gpdb_open_source_osl_licenses" },
        ]),
        {
            task: "update metadata.yml",
            config: {
                platform: "linux",
                image_resource: {
                    type: "registry-image",
                    source: {
                        repository: "bash"
                    }
                },
                inputs: [
                    { name: "gp-kubernetes-tagged-release-tarball" },
                    { name: "greenplum-for-kubernetes" },
                    { name: "greenplum-for-kubernetes-release-osl" },
                    { name: "greenplum-for-kubernetes-release-odp" },
                    { name: "gpdb_open_source_osl_licenses" },
                ],
                outputs: [
                    { name: "workspace" },
                ],
                run: {
                    path: "bash",
                    args: [
                        "-exc",
                        |||
                            RELEASE_VERSION=$(echo gp-kubernetes-tagged-release-tarball/greenplum-for-kubernetes-v*.tar.gz | sed -n 's#gp-kubernetes-tagged-release-tarball/greenplum-for-kubernetes-v\(.*\)\.tar\.gz#\1#p')
                            OSL_VERSION=$(echo greenplum-for-kubernetes-release-osl/open-source-licenses-v*.txt | sed -n 's#greenplum-for-kubernetes-release-osl/open-source-licenses-v\(.*\)\.txt#\1#p')
                            ODP_VERSION=$(echo greenplum-for-kubernetes-release-odp/VMware-greenplum-gp4k-*-ODP.tar.gz | sed -n 's#greenplum-for-kubernetes-release-odp/VMware-greenplum-gp4k-\(.*\)-ODP\.tar\.gz#\1#p')

                            echo "RELEASE_VERSION: ${RELEASE_VERSION}"
                            echo "OSL_VERSION: ${OSL_VERSION}"
                            echo "ODP_VERSION: ${ODP_VERSION}"

                            if [ "${RELEASE_VERSION}" != "${OSL_VERSION}" ]; then
                                echo "error: RELEASE_VERSION != OSL_VERSION"
                                exit 1
                            fi
                            if [ "${RELEASE_VERSION}" != "${ODP_VERSION}" ]; then
                                echo "error: RELEASE_VERSION != ODP_VERSION"
                                exit 1
                            fi

                            tar -xzf "gp-kubernetes-tagged-release-tarball/greenplum-for-kubernetes-v${RELEASE_VERSION}.tar.gz"
                            cp "greenplum-for-kubernetes-release-osl/open-source-licenses-v${RELEASE_VERSION}.txt" "greenplum-for-kubernetes-v${RELEASE_VERSION}/"
                            tar -czf "greenplum-for-kubernetes-v${RELEASE_VERSION}.tar.gz" "greenplum-for-kubernetes-v${RELEASE_VERSION}/"

                            mkdir workspace/files-to-upload

                            cp greenplum-for-kubernetes/release/metadata.yml workspace/
                            sed -i "s/<RELEASE_VERSION>/${RELEASE_VERSION}/g" workspace/metadata.yml

                            cp "greenplum-for-kubernetes-v${RELEASE_VERSION}.tar.gz" workspace/files-to-upload/
                            cp "greenplum-for-kubernetes-release-osl/open-source-licenses-v${RELEASE_VERSION}.txt" workspace/files-to-upload/
                            cp "greenplum-for-kubernetes-release-odp/VMware-greenplum-gp4k-${RELEASE_VERSION}-ODP.tar.gz" workspace/files-to-upload/
                            cp gpdb_open_source_osl_licenses/open_source_license_pivotal-gpdb-6.0.0-1787e8a-1566480846.txt workspace/files-to-upload/

                            cat workspace/metadata.yml

                            ls -l workspace/files-to-upload
                        |||
                    ]
                }
            }
        },
        {
            put: pivnet_resource,
            params: {
                metadata_file: "workspace/metadata.yml",
                file_glob: "workspace/files-to-upload/*",
                s3_filepath_prefix: "product-files/greenplum-for-kubernetes"
            }
        }
    ]
};

{
resources+: [
    production_resource,
    staging_resource,
] + license_resources,
jobs+: [
    release_job(name="push-to-pivnet-gp-kubernetes", pivnet_resource=production_resource.name),
    release_job(name="push-to-pivnet-staging-gp-kubernetes", pivnet_resource=staging_resource.name, trigger_on_tarball=true),
],
}
