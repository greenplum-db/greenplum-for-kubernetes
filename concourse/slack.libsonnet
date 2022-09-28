local concourse = import "concourse.libsonnet";

{
local slack = self,
resource_types+: [
    {
        name: "slack-notification",
        type: "registry-image",
        source: {
            repository: "cfcommunity/slack-notification-resource",
            tag: "latest"
        }
    },
],
resources+: [
    {
        name: "slack-alert",
        type: "slack-notification",
        source: {
            url: "((gp4k_slack_webhook))"
        }
    },
],
alert:: {
    dev_pipeline:: "no",
    put: "slack-alert",
    params: {
        text: "[gp-kubernetes/$BUILD_JOB_NAME] failed:\nhttps://k8s.ci.gpdb.pivotal.io/teams/main/pipelines/gp-kubernetes/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME\n"
    }
},

alert_and_sleep:: concourse.Do([
    slack.alert,
    {
        task: "sleep",
        config: {
            platform: "linux",
            run: { path: "sleep", args: [ "3600" ] },
            image_resource: {
                type: "registry-image",
                source: { repository: "busybox" }
            }
        }
    }
]),
}
