{
gke: {
    K8S_CLUSTER_TYPE: "gke",
    KUBEENV: "GKE", # TODO: get rid of this once the script conversion is finished
    GCP_SVC_ACCT_KEY: "((gcp.svc_acct_key))",
    GCP_PROJECT: "((gcp.project_name))",
    GCP_ZONE: "us-central1-f",
    GCP_GKE_VERSION: "latest"
},

gpdb_ci_s3: {
    AWS_BUCKET_ACCESS_KEY: "((bucket-access-key-id))",
    AWS_BUCKET_SECRET_ACCESS_KEY: "((bucket-secret-access-key))",
    AWS_BUCKET_REGION: "((aws-region))"
},

gp_kubernetes_ci_s3: {
    ACCESS_KEY_ID: "((gp-kubernetes-s3.s3_access_key_id))",
    SECRET_ACCESS_KEY: "((gp-kubernetes-s3.s3_secret_access_key))"
},
}
