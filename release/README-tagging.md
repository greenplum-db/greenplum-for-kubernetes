## Release to pivnet

1. Before making a final tag, tag a release candidate commit with an "-rc" semantic version.

    ```bash
    git tag -a -m "v<version>-rc.1 Release Candidate" v<version>-rc.1  # message is up to you
    git push origin v<version>-rc.1
    ```

1. Wait for a new `tagged-release-tarball` to be automatically built and uploaded to the gcs bucket by the ["make-new-release-upon-tagging" job](https://k8s.ci.gpdb.pivotal.io/teams/main/pipelines/gp-kubernetes/jobs/make-new-release-upon-tagging/),

1. Check on the Security Notices.
    1. Once the receipts have been pushed to [github](https://github.com/pivotal/greenplum-for-kubernetes-receipts/releases) for the release tag, download them to your system.
    1. Export a `.csv` of the Security Notice stories on tracker.
    1. Check the package versions using that list against the Security Notice stories.

        ```bash
        go run security_check.go -receipt ./greenplum-for-kubernetes-receipt.txt -securityNotices ./gp_kubernetes_20200512_2127.csv
        ```

1. If the release process went well, and the security notices are resolved as expected, then tag a final release, tagging the same commit as before. (If it didn't go well, and fixes were required, then restart the process with "-rc.2", etc.)

    ```bash
    git tag -a -m "v<version> Release on Pivnet" v<version> v<version>-rc.<x>^{}  # message is up to you
    git push origin v<version>
    ```

1. Generate ODP file and upload it to Google Cloud Storage (`gs://greenplum-for-kubernetes-release-odp/VMware-greenplum-gp4k-<version>-ODP.tar.gz`). See [https://confluence.eng.vmware.com/pages/viewpage.action?spaceKey=CNA&title=%5BWIP%5D+OSM+Workflows+for+Pivotal-Alum+products](https://confluence.eng.vmware.com/pages/viewpage.action?spaceKey=CNA&title=%5BWIP%5D+OSM+Workflows+for+Pivotal-Alum+products) and [https://www.pivotaltracker.com/n/projects/2087405/stories/174712876](https://www.pivotaltracker.com/n/projects/2087405/stories/174712876) for ODP file generation instructions.

1. Upload OSL file to Google Cloud Storage (`gs://greenplum-for-kubernetes-release-osl/open-source-licenses-v<version>.txt`).

1. Manually trigger the ["push-to-pivnet-gp-kubernetes" job](https://k8s.ci.gpdb.pivotal.io/teams/main/pipelines/gp-kubernetes/jobs/push-to-pivnet-gp-kubernetes/).

1. After the ["push-to-pivnet-gp-kubernetes" job](https://k8s.ci.gpdb.pivotal.io/teams/main/pipelines/gp-kubernetes/jobs/push-to-pivnet-gp-kubernetes/) finishes, confirm that the release is available on PivNet as "Admins Only" and the associated files are uploaded with the appropriate metadata, including release date and file name.

1. Once the PM decides that the release is ready to ship, update the release by changing the availability from "Admins Only" to "All Users" on PivNet.  Note this action is irreversible, as it marks the release and associated files as immutable and thereby un-deletable.


### Notes
* The `tagged-release-tarball` is uploaded to the gcs bucket: `gs://greenplum-for-kubernetes-release/`, which can be downloaded via a command like

 ```bash
gsutil cp gs://greenplum-for-kubernetes-release/greenplum-for-kubernetes-v<VERSION OF RELEASE>.tar.gz  .
```