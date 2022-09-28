package pxf_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	greenplumv1beta1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1beta1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/pxf"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PXF K8s resources", func() {
	var (
		greenplumPXF greenplumv1beta1.GreenplumPXFService

		labels = map[string]string{
			"app":           greenplumv1beta1.PXFAppName,
			"greenplum-pxf": "my-greenplum-pxf",
		}
	)

	BeforeEach(func() {
		greenplumPXF = greenplumv1beta1.GreenplumPXFService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-greenplum-pxf",
				Namespace: "test-ns",
			},
			Spec: greenplumv1beta1.GreenplumPXFServiceSpec{
				Replicas: 2,
				CPU:      resource.MustParse("2"),
				Memory:   resource.MustParse("2Gi"),
				WorkerSelector: map[string]string{
					"worker": "my-greenplum-pxf-worker",
				},
			},
		}
	})

	Describe("ModifyDeployment", func() {
		var deployment appsv1.Deployment
		BeforeEach(func() {
			deployment = appsv1.Deployment{}
		})
		It("sets properties from manifest", func() {
			pxf.ModifyDeployment(greenplumPXF, &deployment, "greenplum-for-kubernetes:v1.7.5")

			Expect(deployment.Spec.Replicas).To(gstruct.PointTo(Equal(int32(2))))
			pxfContainer := deployment.Spec.Template.Spec.Containers[0]
			cpu := pxfContainer.Resources.Limits[corev1.ResourceCPU]
			Expect(cpu.Cmp(resource.MustParse("2"))).To(Equal(0))
			mem := pxfContainer.Resources.Limits[corev1.ResourceMemory]
			Expect(mem.Cmp(resource.MustParse("2Gi"))).To(Equal(0))
			Expect(pxfContainer.Image).To(Equal("greenplum-for-kubernetes:v1.7.5"))
			Expect(deployment.Labels).To(Equal(labels))
			Expect(deployment.Spec.Template.Labels).To(Equal(labels))
			Expect(deployment.Spec.Selector.MatchLabels).To(Equal(labels))
			Expect(deployment.Spec.Template.Spec.ImagePullSecrets[0].Name).To(Equal("regsecret"))
			Expect(deployment.Spec.Template.Spec.NodeSelector).To(HaveKeyWithValue("worker", "my-greenplum-pxf-worker"))

			// default
			Expect(pxfContainer.Name).To(Equal("pxf"))
			Expect(pxfContainer.ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))
			Expect(pxfContainer.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(pxfContainer.Ports[0].ContainerPort).To(Equal(int32(5888)))
			Expect(pxfContainer.ReadinessProbe.Exec.Command).To(Equal([]string{"/usr/local/pxf-gp6/bin/pxf", "status"}))
			Expect(pxfContainer.ReadinessProbe.InitialDelaySeconds).To(Equal(int32(30)))
			Expect(pxfContainer.ReadinessProbe.TimeoutSeconds).To(Equal(int32(5)))
			Expect(pxfContainer.Args).To(Equal([]string{"/home/gpadmin/tools/startPXF"}))
		})
		It("sets PXF_JVM_OPTS with -XX:MaxRAMPercentage", func() {
			pxf.ModifyDeployment(greenplumPXF, &deployment, "greenplum-for-kubernetes:v1.7.5")

			pxfContainer := deployment.Spec.Template.Spec.Containers[0]
			Expect(pxfContainer.Env).To(ContainElement(gstruct.MatchAllFields(gstruct.Fields{
				"Name":      Equal("PXF_JVM_OPTS"),
				"Value":     Equal("-XX:MaxRAMPercentage=75.0"),
				"ValueFrom": BeNil(),
			})))
		})
		It("does not modify resourceVersion, name, or namespace", func() {
			deployment.Name = "my-greenplum-pxf"
			deployment.Namespace = "test-ns"
			deployment.ResourceVersion = "test-resource-version"

			pxf.ModifyDeployment(greenplumPXF, &deployment, "gcr.io/gp-kubernetes/greenplum-for-kubernetes:latest-green")

			Expect(deployment.Name).To(Equal("my-greenplum-pxf"))
			Expect(deployment.Namespace).To(Equal("test-ns"))
			Expect(deployment.ResourceVersion).To(Equal("test-resource-version"))
		})
		When("pxfConf is configured", func() {
			BeforeEach(func() {
				greenplumPXF.Spec.PXFConf = &greenplumv1beta1.GreenplumPXFConf{
					S3Source: greenplumv1beta1.S3Source{
						Secret:   "test-secret",
						Bucket:   "test-bucket",
						EndPoint: "test-endpoint",
						Folder:   "test-folder",
					},
				}
			})

			It("sets environment variables in the deployment", func() {
				pxf.ModifyDeployment(greenplumPXF, &deployment, "greenplum-for-kubernetes:v1.7.5")

				container := deployment.Spec.Template.Spec.Containers[0]
				Expect(container.Env).To(ContainElement(corev1.EnvVar{
					Name: "S3_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "access_key_id",
						},
					},
				}))
				Expect(container.Env).To(ContainElement(corev1.EnvVar{
					Name: "S3_SECRET_ACCESS_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-secret",
							},
							Key: "secret_access_key",
						},
					},
				}))
				Expect(container.Env).To(ContainElement(corev1.EnvVar{
					Name:  "S3_ENDPOINT",
					Value: "test-endpoint",
				}))
				Expect(container.Env).To(ContainElement(corev1.EnvVar{
					Name:  "S3_BUCKET",
					Value: "test-bucket",
				}))
				Expect(container.Env).To(ContainElement(corev1.EnvVar{
					Name:  "S3_FOLDER",
					Value: "test-folder",
				}))
			})
			Describe("s3source.protocol", func() {
				When("s3source.protocol is https", func() {
					BeforeEach(func() {
						greenplumPXF.Spec.PXFConf.S3Source.Protocol = "https"
					})

					It("sets S3_ENDPOINT_IS_SECURE to true", func() {
						pxf.ModifyDeployment(greenplumPXF, &deployment, "greenplum-for-kubernetes:v1.7.5")

						container := deployment.Spec.Template.Spec.Containers[0]
						Expect(container.Env).To(ContainElement(corev1.EnvVar{
							Name:  "S3_ENDPOINT_IS_SECURE",
							Value: "true",
						}))
					})
				})
				When("s3source.protocol is http", func() {
					BeforeEach(func() {
						greenplumPXF.Spec.PXFConf.S3Source.Protocol = "http"
					})

					It("sets S3_ENDPOINT_IS_SECURE to false", func() {
						pxf.ModifyDeployment(greenplumPXF, &deployment, "greenplum-for-kubernetes:v1.7.5")

						container := deployment.Spec.Template.Spec.Containers[0]
						Expect(container.Env).To(ContainElement(corev1.EnvVar{
							Name:  "S3_ENDPOINT_IS_SECURE",
							Value: "false",
						}))
					})
				})
				When("s3source.protocol is not specified", func() {
					BeforeEach(func() {
						greenplumPXF.Spec.PXFConf.S3Source.Protocol = ""
					})

					It("sets S3_ENDPOINT_IS_SECURE to true", func() {
						pxf.ModifyDeployment(greenplumPXF, &deployment, "greenplum-for-kubernetes:v1.7.5")

						container := deployment.Spec.Template.Spec.Containers[0]
						Expect(container.Env).To(ContainElement(corev1.EnvVar{
							Name:  "S3_ENDPOINT_IS_SECURE",
							Value: "true",
						}))
					})
				})
			})
		})
		When("pxfConf is empty", func() {
			BeforeEach(func() {
				greenplumPXF.Spec.PXFConf = nil
			})

			It("does not set S3 environment variables in the deployment", func() {
				pxf.ModifyDeployment(greenplumPXF, &deployment, "greenplum-for-kubernetes:v1.7.5")

				container := deployment.Spec.Template.Spec.Containers[0]
				Expect(container.Env).To(HaveLen(1))
				Expect(container.Env).To(ContainElement(gstruct.MatchAllFields(gstruct.Fields{
					"Name":      Equal("PXF_JVM_OPTS"),
					"Value":     Equal("-XX:MaxRAMPercentage=75.0"),
					"ValueFrom": BeNil(),
				})))
			})
		})
		When("the deployment is already populated (during resync update)", func() {
			var oldDeployment *appsv1.Deployment
			BeforeEach(func() {
				pxf.ModifyDeployment(greenplumPXF, &deployment, "gcr.io/gp-kubernetes/greenplum-for-kubernetes:latest-green")
				oldDeployment = deployment.DeepCopy()
			})
			It("should not change any of the existing information", func() {
				pxf.ModifyDeployment(greenplumPXF, &deployment, "gcr.io/gp-kubernetes/greenplum-for-kubernetes:latest-green")
				Expect(reflect.DeepEqual(oldDeployment, &deployment)).To(BeTrue())
			})
		})
	})

	Describe("ModifyService", func() {
		var service corev1.Service

		It("sets properties from manifest", func() {
			pxf.ModifyService(greenplumPXF, &service)

			Expect(len(service.Spec.Ports)).To(Equal(1))
			Expect(service.Spec.Ports[0].Port).To(Equal(int32(5888)))
			Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(5888)))
			Expect(service.Labels).To(Equal(labels))
			Expect(service.Spec.Selector).To(Equal(labels))
		})
		It("does not modify resourceVersion, name, or namespace", func() {
			service.Name = "my-greenplum-pxf"
			service.Namespace = "test-ns"
			service.ResourceVersion = "test-resource-version"

			pxf.ModifyService(greenplumPXF, &service)

			Expect(service.Name).To(Equal("my-greenplum-pxf"))
			Expect(service.Namespace).To(Equal("test-ns"))
			Expect(service.ResourceVersion).To(Equal("test-resource-version"))
		})
		When("the service is already populated (during resync update)", func() {
			var oldService *corev1.Service
			BeforeEach(func() {
				pxf.ModifyService(greenplumPXF, &service)
				oldService = service.DeepCopy()
			})
			It("should not change any of the existing information", func() {
				pxf.ModifyService(greenplumPXF, &service)
				Expect(reflect.DeepEqual(oldService, &service)).To(BeTrue())
			})
		})
	})

})
