package gpexpandjob

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("GenerateJob", func() {
	It("sets properties on the job", func() {
		job := GenerateJob("greenplum-for-kubernetes:magic", "master-0.agent.default.svc.cluster.local", 2)
		Expect(job.Spec.BackoffLimit).To(gstruct.PointTo(Equal(int32(0))))

		gpexpandPod := job.Spec.Template.Spec
		Expect(gpexpandPod.RestartPolicy).To(Equal(corev1.RestartPolicyNever))

		sshSecretVolume := gpexpandPod.Volumes[0]
		Expect(sshSecretVolume.Name).To(Equal("ssh-key"))
		Expect(sshSecretVolume.VolumeSource.Secret.SecretName).To(Equal("ssh-secrets"))
		Expect(sshSecretVolume.VolumeSource.Secret.DefaultMode).To(gstruct.PointTo(Equal(int32(0444))))

		Expect(gpexpandPod.ImagePullSecrets[0].Name).To(Equal("regsecret"))
		gpexpandContainer := gpexpandPod.Containers[0]
		Expect(gpexpandContainer.Name).To(Equal("gpexpand"))
		Expect(gpexpandContainer.Env[0].Name).To(Equal("GPEXPAND_HOST"))
		Expect(gpexpandContainer.Env[0].Value).To(Equal("master-0.agent.default.svc.cluster.local"))
		Expect(gpexpandContainer.Env[1].Name).To(Equal("NEW_SEG_COUNT"))
		Expect(gpexpandContainer.Env[1].Value).To(Equal("2"))
		Expect(gpexpandContainer.Image).To(Equal("greenplum-for-kubernetes:magic"))
		Expect(gpexpandContainer.ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))
		Expect(gpexpandContainer.Command).To(Equal([]string{
			"/home/gpadmin/tools/gpexpand_job.sh",
		}))

		sshSecretVolumeMount := gpexpandContainer.VolumeMounts[0]
		Expect(sshSecretVolumeMount.Name).To(Equal("ssh-key"))
		Expect(sshSecretVolumeMount.MountPath).To(Equal("/etc/ssh-key"))
	})
})
