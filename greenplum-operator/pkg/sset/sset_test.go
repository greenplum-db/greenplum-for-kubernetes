package sset_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/sset"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/heapvalue"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	segmentCountNine              int32 = 9
	SecretVolumeSourceDefaultMode int32 = 0444
)

var _ = Describe("GreenplumCluster StatefulSet spec", func() {

	var (
		subject         *appsv1.StatefulSet
		greenplumParams *sset.GreenplumStatefulSetParams
	)

	BeforeEach(func() {
		greenplumParams = &sset.GreenplumStatefulSetParams{
			ClusterName:   "my-greenplum",
			Replicas:      segmentCountNine,
			InstanceImage: "my-repo:my-tag",
			GpPodSpec: greenplumv1.GreenplumPodSpec{
				StorageClassName: "fakeStorageClassName",
				Storage:          resource.MustParse("5G"),
			},
		}
		subject = &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test-namespace",
			},
		}
		sset.ModifyGreenplumStatefulSet(greenplumParams, subject)
	})

	It("has all the required metadata parameters", func() {
		Expect(subject.Name).To(Equal("test"))
		Expect(len(subject.Labels)).To(Equal(3))
		Expect(subject.Labels["app"]).To(Equal("greenplum"))
		Expect(subject.Labels["type"]).To(Equal("test"))
		Expect(subject.Labels["greenplum-cluster"]).To(Equal("my-greenplum"))
		Expect(subject.Namespace).To(Equal("test-namespace"))
		Expect(subject.Spec).ToNot(BeNil())
	})
	It("has the default spec parameters", func() {
		greenplumStatefulSetSpec := subject.Spec
		Expect(*greenplumStatefulSetSpec.Replicas).To(Equal(segmentCountNine))
		Expect(greenplumStatefulSetSpec.ServiceName).To(Equal("agent"))
		Expect(greenplumStatefulSetSpec.Selector.MatchLabels).To(Equal(map[string]string{
			"app":               "greenplum",
			"greenplum-cluster": "my-greenplum",
			"type":              "test",
		}))
		Expect(len(greenplumStatefulSetSpec.Template.ObjectMeta.Labels)).To(Equal(3))
		Expect(greenplumStatefulSetSpec.Template.ObjectMeta.Labels["app"]).To(Equal("greenplum"))
		Expect(greenplumStatefulSetSpec.Template.ObjectMeta.Labels["greenplum-cluster"]).To(Equal("my-greenplum"))
		Expect(greenplumStatefulSetSpec.Template.ObjectMeta.Labels["type"]).To(Equal("test"))
		Expect(greenplumStatefulSetSpec.Template.Spec).ToNot(BeNil())
		Expect(greenplumStatefulSetSpec.PodManagementPolicy).To(Equal(appsv1.ParallelPodManagement))
	})
	It("has all the required parameters in pod spec", func() {
		greenplumPodSpec := subject.Spec.Template.Spec
		Expect(greenplumPodSpec.ImagePullSecrets[0].Name).To(Equal("regsecret"))
		Expect(len(greenplumPodSpec.Containers)).ToNot(BeZero())
		Expect(len(greenplumPodSpec.Volumes)).ToNot(BeZero())
		Expect(greenplumPodSpec.DNSConfig).To(Equal(&corev1.PodDNSConfig{Searches: []string{"agent.test-namespace.svc.cluster.local"}}))
	})

	It("does not set NodeSelector by default", func() {
		Expect(subject.Spec.Template.Spec.NodeSelector).To(BeNil())
	})

	When("workerSelector is specified", func() {
		BeforeEach(func() {
			greenplumParams.GpPodSpec.WorkerSelector = map[string]string{
				"worker": "my-greenplum-master",
			}
			sset.ModifyGreenplumStatefulSet(greenplumParams, subject)
		})

		It("has WorkerSelector labels", func() {
			Expect(subject.Spec.Template.Spec.NodeSelector).To(HaveKeyWithValue("worker", "my-greenplum-master"))
		})
	})

	When("antiAffinity is specified", func() {
		BeforeEach(func() {
			greenplumParams.GpPodSpec.AntiAffinity = "yes"
		})
		It("gets an affinity object for a master pod", func() {
			greenplumParams.Type = sset.TypeMaster

			sset.ModifyGreenplumStatefulSet(greenplumParams, subject)
			Expect(subject.Spec.Template.Spec.Affinity).ToNot(BeNil())
			nodeSelectorMatchExpr := subject.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0]

			Expect(nodeSelectorMatchExpr.Key).To(Equal("greenplum-affinity-test-namespace-master"))
			Expect(nodeSelectorMatchExpr.Operator).To(Equal(corev1.NodeSelectorOpIn))

			Expect(nodeSelectorMatchExpr.Values).To(Equal([]string{"true"}))

			podAntiAffinitySelectorMatchExpr := subject.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchExpressions[0]
			Expect(subject.Spec.Template.Spec.Affinity.PodAntiAffinity).ToNot(BeNil())
			Expect(podAntiAffinitySelectorMatchExpr.Key).To(Equal("type"))
			Expect(podAntiAffinitySelectorMatchExpr.Operator).To(Equal(metav1.LabelSelectorOpIn))
			Expect(podAntiAffinitySelectorMatchExpr.Values).To(Equal([]string{"master"}))
		})
		It("gets a node affinity object for a segment-a pod", func() {
			greenplumParams.Type = sset.TypeSegmentA

			sset.ModifyGreenplumStatefulSet(greenplumParams, subject)
			Expect(subject.Spec.Template.Spec.Affinity).ToNot(BeNil())
			nodeSelectorMatchExpr := subject.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0]
			Expect(nodeSelectorMatchExpr.Key).To(Equal("greenplum-affinity-test-namespace-segment"))
			Expect(nodeSelectorMatchExpr.Operator).To(Equal(corev1.NodeSelectorOpIn))
			Expect(nodeSelectorMatchExpr.Values).To(Equal([]string{"a"}))
		})
		It("gets a node affinity object for a segment-b pod", func() {
			greenplumParams.Type = sset.TypeSegmentB

			sset.ModifyGreenplumStatefulSet(greenplumParams, subject)
			Expect(subject.Spec.Template.Spec.Affinity).ToNot(BeNil())
			nodeSelectorMatchExpr := subject.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0]
			Expect(nodeSelectorMatchExpr.Key).To(Equal("greenplum-affinity-test-namespace-segment"))
			Expect(nodeSelectorMatchExpr.Operator).To(Equal(corev1.NodeSelectorOpIn))
			Expect(nodeSelectorMatchExpr.Values).To(Equal([]string{"b"}))
		})
	})

	It("has container spec with correct parameters", func() {
		expectedPort := []corev1.ContainerPort{
			{
				ContainerPort: 22,
				Protocol:      corev1.ProtocolTCP,
			},
		}
		expectedProbe := &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(22),
				},
			},
			InitialDelaySeconds: 5,
		}
		expectedVolumeMounts := []corev1.VolumeMount{
			{
				Name:      "ssh-key-volume",
				MountPath: "/etc/ssh-key",
			},
			{
				Name:      "config-volume",
				MountPath: "/etc/config",
			},
			{
				Name:      "my-greenplum-pgdata",
				MountPath: "/greenplum",
			},
			{
				Name:      "cgroups",
				MountPath: "/sys/fs/cgroup",
			},
			{
				Name:      "podinfo",
				MountPath: "/etc/podinfo",
			},
		}
		expectedEnvVars := []corev1.EnvVar{
			{
				Name:  "MASTER_DATA_DIRECTORY",
				Value: "/greenplum/data-1",
			},
		}
		containerDef := subject.Spec.Template.Spec.Containers
		Expect(len(containerDef)).To(Equal(1))
		Expect(containerDef[0].Name).To(Equal("greenplum"))
		Expect(containerDef[0].Image).To(Equal("my-repo:my-tag"))
		Expect(containerDef[0].ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))
		Expect(len(containerDef[0].Ports)).To(Equal(1))
		Expect(containerDef[0].Ports).To(Equal(expectedPort))
		Expect(containerDef[0].ReadinessProbe).ToNot(BeNil())
		Expect(containerDef[0].ReadinessProbe).To(Equal(expectedProbe))
		Expect(containerDef[0].VolumeMounts).To(Equal(expectedVolumeMounts))
		Expect(containerDef[0].Env).To(Equal(expectedEnvVars))
		Expect(len(containerDef[0].Args)).To(Equal(1))
		Expect(containerDef[0].Args[0]).To(Equal("/home/gpadmin/tools/startGreenplumContainer"))
	})

	When("a readiness probe already exists", func() {
		BeforeEach(func() {
			subject.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec:    &corev1.ExecAction{Command: []string{"i shouldn't be here"}},
					HTTPGet: &corev1.HTTPGetAction{Host: "twinkie"},
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.FromString("Wrong Port"),
						Host: "twinkie", // twinkies are a bad host(ess)
					},
				},
				InitialDelaySeconds: 500,
				TimeoutSeconds:      10,
				PeriodSeconds:       11,
				SuccessThreshold:    12,
				FailureThreshold:    13,
			}
			sset.ModifyGreenplumStatefulSet(greenplumParams, subject)
		})
		It("reconciles only the fields we care about", func() {
			reconciledProbe := subject.Spec.Template.Spec.Containers[0].ReadinessProbe
			Expect(reconciledProbe.ProbeHandler.Exec).To(BeNil(), "should be deleted")
			Expect(reconciledProbe.ProbeHandler.HTTPGet).To(BeNil(), "should be deleted")
			Expect(reconciledProbe.ProbeHandler.TCPSocket).To(gstruct.PointTo(Equal(corev1.TCPSocketAction{
				Port: intstr.FromInt(22), // overwrite
			})))
			Expect(reconciledProbe.InitialDelaySeconds).To(BeNumerically("==", 5), "overwrite")
			Expect(reconciledProbe.TimeoutSeconds).To(BeNumerically("==", 10), "preserve")
			Expect(reconciledProbe.PeriodSeconds).To(BeNumerically("==", 11), "preserve")
			Expect(reconciledProbe.SuccessThreshold).To(BeNumerically("==", 12), "preserve")
			Expect(reconciledProbe.FailureThreshold).To(BeNumerically("==", 13), "preserve")
		})
	})

	It("creates all needed volume sources", func() {
		expectedVolumes := []corev1.Volume{
			{
				Name: "ssh-key-volume",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  "ssh-secrets",
						DefaultMode: heapvalue.NewInt32(SecretVolumeSourceDefaultMode),
					},
				},
			},
			{
				Name: "config-volume",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "greenplum-config",
						},
						DefaultMode: heapvalue.NewInt32(corev1.ConfigMapVolumeSourceDefaultMode),
					},
				},
			},
			{
				Name: "cgroups",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/sys/fs/cgroup",
						Type: heapvalue.NewHostPathType(corev1.HostPathUnset),
					},
				},
			},
			{
				Name: "podinfo",
				VolumeSource: corev1.VolumeSource{
					DownwardAPI: &corev1.DownwardAPIVolumeSource{
						Items: []corev1.DownwardAPIVolumeFile{
							{
								Path: "namespace",
								FieldRef: &corev1.ObjectFieldSelector{
									APIVersion: "v1",
									FieldPath:  "metadata.namespace",
								},
							},
							{
								Path: "greenplumClusterName",
								FieldRef: &corev1.ObjectFieldSelector{
									APIVersion: "v1",
									FieldPath:  "metadata.labels['greenplum-cluster']",
								},
							},
						},
						DefaultMode: heapvalue.NewInt32(corev1.DownwardAPIVolumeSourceDefaultMode),
					},
				},
			},
		}

		volumeDef := subject.Spec.Template.Spec.Volumes
		Expect(len(volumeDef)).To(Equal(4))
		Expect(volumeDef).To(Equal(expectedVolumes))
	})

	It("creates persistent volume claim for my-greenplum", func() {
		var name = "fakeStorageClassName"
		var storageSize = resource.MustParse("5G")
		expectedVolumeClaimTemplate := corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-greenplum-pgdata",
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceStorage: storageSize,
					},
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: storageSize,
					},
				},
				StorageClassName: &name,
			},
		}

		volumeClaimTemplate := subject.Spec.VolumeClaimTemplates[0]
		Expect(volumeClaimTemplate).To(Equal(expectedVolumeClaimTemplate))
	})

	Context("resource limits tests", func() {
		When("resource limits are not provided", func() {
			It("does not apply pod resource limits if none are provided", func() {
				resourceLimitsDef := subject.Spec.Template.Spec.Containers[0].Resources.Limits
				Expect(resourceLimitsDef.Cpu().String()).To(Equal("0"))
				Expect(resourceLimitsDef.Memory().String()).To(Equal("0"))

			})
		})

		When("both resource limits are provided", func() {
			BeforeEach(func() {
				greenplumParams.GpPodSpec.Memory = resource.MustParse("800Mi")
				greenplumParams.GpPodSpec.CPU = resource.MustParse("0.8")
				sset.ModifyGreenplumStatefulSet(greenplumParams, subject)
			})
			It("applies resource limits", func() {
				expectedContainerResourceLimits := corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("800Mi"),
					corev1.ResourceCPU:    resource.MustParse("0.8"),
				}
				Expect(subject.Name).To(Equal("test"))
				resourceLimitsDef := subject.Spec.Template.Spec.Containers[0].Resources.Limits
				Expect(resourceLimitsDef).To(Equal(expectedContainerResourceLimits))
			})
		})

		When("only CPU resource limit is provided", func() {
			BeforeEach(func() {
				greenplumParams.GpPodSpec.CPU = resource.MustParse("0.8")
				sset.ModifyGreenplumStatefulSet(greenplumParams, subject)
			})
			It("applies resource limits", func() {
				Expect(subject.Name).To(Equal("test"))
				resourceLimitsDef := subject.Spec.Template.Spec.Containers[0].Resources.Limits
				Expect(resourceLimitsDef.Cpu().String()).To(Equal("800m"))
				Expect(resourceLimitsDef.Memory().String()).To(Equal("0"))
			})
		})

		When("only Memory resource limit is provided", func() {
			BeforeEach(func() {
				greenplumParams.GpPodSpec.Memory = resource.MustParse("500Gi")
				sset.ModifyGreenplumStatefulSet(greenplumParams, subject)
			})
			It("applies resource limits", func() {
				Expect(subject.Name).To(Equal("test"))
				resourceLimitsDef := subject.Spec.Template.Spec.Containers[0].Resources.Limits
				Expect(resourceLimitsDef.Cpu().String()).To(Equal("0"))
				Expect(resourceLimitsDef.Memory().String()).To(Equal("500Gi"))
			})
		})
	})
})

var _ = Describe("GenerateStatefulSetParams", func() {
	var (
		cluster       *greenplumv1.GreenplumCluster
		instanceImage = "test-image:version"
	)
	BeforeEach(func() {
		cluster = &greenplumv1.GreenplumCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-greenplum",
				Namespace: "test-ns",
			},
			Spec: greenplumv1.GreenplumClusterSpec{
				MasterAndStandby: greenplumv1.GreenplumMasterAndStandbySpec{
					GreenplumPodSpec: greenplumv1.GreenplumPodSpec{
						Storage: resource.MustParse("10Gi"),
					},
				},
				Segments: greenplumv1.GreenplumSegmentsSpec{
					GreenplumPodSpec: greenplumv1.GreenplumPodSpec{
						Storage: resource.MustParse("20Gi"),
					},
					PrimarySegmentCount: 3,
				},
			},
		}
	})
	When("generating params for master statefulset", func() {
		It("sets the passed-in properties", func() {
			params := sset.GenerateStatefulSetParams(sset.TypeMaster, cluster, instanceImage)

			Expect(params.Type).To(Equal(sset.TypeMaster))
			Expect(params.ClusterName).To(Equal("my-greenplum"))
			Expect(params.InstanceImage).To(Equal(instanceImage))
		})
		It("gets the masterAndStandby pod spec", func() {
			params := sset.GenerateStatefulSetParams(sset.TypeMaster, cluster, instanceImage)

			Expect(params.GpPodSpec.Storage).To(Equal(resource.MustParse("10Gi")))
		})
		When("standby = yes", func() {
			BeforeEach(func() {
				cluster.Spec.MasterAndStandby.Standby = "yes"
			})
			It("sets replicas to 2", func() {
				params := sset.GenerateStatefulSetParams(sset.TypeMaster, cluster, instanceImage)

				Expect(params.Replicas).To(Equal(int32(2)))
			})
		})
		When("standby = no", func() {
			BeforeEach(func() {
				cluster.Spec.MasterAndStandby.Standby = "no"
			})
			It("sets replicas to 1", func() {
				params := sset.GenerateStatefulSetParams(sset.TypeMaster, cluster, instanceImage)

				Expect(params.Replicas).To(Equal(int32(1)))
			})
		})
	})
	When("generating params for segment statefulset", func() {
		It("sets the passed-in properties", func() {
			params := sset.GenerateStatefulSetParams(sset.TypeSegmentA, cluster, instanceImage)

			Expect(params.Type).To(Equal(sset.TypeSegmentA))
			Expect(params.ClusterName).To(Equal("my-greenplum"))
			Expect(params.InstanceImage).To(Equal(instanceImage))
		})
		It("gets the segments pod spec", func() {
			params := sset.GenerateStatefulSetParams(sset.TypeSegmentA, cluster, instanceImage)

			Expect(params.GpPodSpec.Storage).To(Equal(resource.MustParse("20Gi")))
		})
		It("sets replicas to primarySegmentCount", func() {
			params := sset.GenerateStatefulSetParams(sset.TypeSegmentA, cluster, instanceImage)

			Expect(params.Replicas).To(Equal(int32(3)))
		})
	})
})
