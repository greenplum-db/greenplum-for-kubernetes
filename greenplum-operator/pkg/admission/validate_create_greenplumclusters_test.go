package admission_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers/greenplumcluster"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/admission"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/reactive"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/gplog/testing"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/heapvalue"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("validateCreateGreenplumCluster", func() {

	var (
		subject admission.Handler
		logBuf  *gbytes.Buffer
	)

	BeforeEach(func() {
		reactiveClient := reactive.NewClient(fakeClient.NewFakeClientWithScheme(scheme.Scheme))
		subject = admission.Handler{
			KubeClient: reactiveClient,
		}
		logBuf = gbytes.NewBuffer()
		admission.Log = gplog.ForTest(logBuf)
	})

	ContainDisallowedGreenplumClusterEntry := func(expectedMessage string) types.GomegaMatcher {
		return ContainLogEntry(Keys{
			"msg":       Equal("/validate"),
			"GVK":       Equal("greenplum.pivotal.io/v1, Kind=GreenplumCluster"),
			"Name":      Equal("my-gp-instance"),
			"Namespace": Equal("test-ns"),
			"UID":       Equal("my-gp-instance-uid"),
			"Operation": Equal("CREATE"),
			"Allowed":   BeFalse(),
			"Message":   Equal(expectedMessage),
		})
	}

	ContainAllowedGreenplumClusterEntry := func() types.GomegaMatcher {
		return ContainLogEntry(Keys{
			"msg":       Equal("/validate"),
			"GVK":       Equal("greenplum.pivotal.io/v1, Kind=GreenplumCluster"),
			"Name":      Equal("my-gp-instance"),
			"Namespace": Equal("test-ns"),
			"UID":       Equal("my-gp-instance-uid"),
			"Operation": Equal("CREATE"),
			"Allowed":   BeTrue(),
		})
	}

	When("a cluster with both standby and mirrors has been deleted, but PVCs still exist", func() {
		When("PVCs exist", func() {

			BeforeEach(func() {
				createGPDBTestPVCs(subject.KubeClient, 2, 2, 2,
					map[string]string{"greenplum-major-version": greenplumcluster.SupportedGreenplumMajorVersion})
			})
			It("disallows requests when masterAndStandby storage differs from the existing PVC storage", func() {
				newGreenplum := exampleGreenplum.DeepCopy()
				newGreenplum.Spec.MasterAndStandby.Storage = resource.MustParse("20G") // matches size of the wrong PVC

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal("storage cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("storage cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"))
			})
			It("disallows requests when segments storage differs from the existing PVC storage", func() {
				newGreenplum := exampleGreenplum.DeepCopy()

				newGreenplum.Spec.Segments.Storage = resource.MustParse("10G") // matches size of the wrong PVC

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal("storage cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("storage cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"))
			})
			It("allows requests that match the existing PVC storage", func() {
				newGreenplum := exampleGreenplum.DeepCopy()

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
				Expect(outputReview.Response.Allowed).To(BeTrue(), "should be allowed")
				Expect(outputReview.Response.Result).To(BeNil())
				Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			})
			It("disallows requests when masterAndStandby storageClassName differs from the existing PVC", func() {
				newGreenplum := exampleGreenplum.DeepCopy()
				newGreenplum.Spec.MasterAndStandby.StorageClassName = "new-storage-class"

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal("storageClassName cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("storageClassName cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"))
			})
			It("disallows requests when segments storageClassName differs from the existing PVC", func() {
				newGreenplum := exampleGreenplum.DeepCopy()
				newGreenplum.Spec.Segments.StorageClassName = "new-storage-class"

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal("storageClassName cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("storageClassName cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"))
			})
			It("disallows requests that decrease primarySegmentCount", func() {
				newGreenplum := exampleGreenplum.DeepCopy()
				newGreenplum.Spec.Segments.PrimarySegmentCount = 1

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal("my-gp-instance has PVCs for 2 segments. segments.primarySegmentCount cannot be decreased without first deleting PVCs. This will result in a new, empty Greenplum cluster"),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("my-gp-instance has PVCs for 2 segments. segments.primarySegmentCount cannot be decreased without first deleting PVCs. This will result in a new, empty Greenplum cluster"))
			})
			It("disallows requests that change standby", func() {
				newGreenplum := exampleGreenplum.DeepCopy()
				newGreenplum.Spec.MasterAndStandby.Standby = "no"
				// Standby=no is invalid unless AntiAffinity=no
				newGreenplum.Spec.MasterAndStandby.AntiAffinity = "no"
				newGreenplum.Spec.Segments.AntiAffinity = "no"

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal("my-gp-instance has PVCs for 2 masters. masterAndStandby.standby cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("my-gp-instance has PVCs for 2 masters. masterAndStandby.standby cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"))
			})
			It("disallows requests that change mirrors", func() {
				newGreenplum := exampleGreenplum.DeepCopy()
				newGreenplum.Spec.Segments.Mirrors = "no"
				// Standby=no is invalid unless AntiAffinity=no
				newGreenplum.Spec.MasterAndStandby.AntiAffinity = "no"
				newGreenplum.Spec.Segments.AntiAffinity = "no"

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal("my-gp-instance has PVCs for 2 mirrors. segments.mirrors cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("my-gp-instance has PVCs for 2 mirrors. segments.mirrors cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"))
			})
			It("allows requests that change the value for standby without changing its meaning", func() {
				newGreenplum := exampleGreenplum.DeepCopy()
				newGreenplum.Spec.MasterAndStandby.Standby = "YES"

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeTrue(), "should be allowed")
				Expect(outputReview.Response.Result).To(BeNil())
				Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			})
			It(fmt.Sprintf("allows requests when greenplum-major-version label = %s", greenplumcluster.SupportedGreenplumMajorVersion), func() {
				outputReview := postValidateReview(subject.Handler(), &exampleGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeTrue(), "should be allowed")
				Expect(outputReview.Response.Result).To(BeNil())
				Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			})
			When("greenplum-major-version label does not match the expected value on segment-a sset PVCs", func() {
				BeforeEach(func() {
					reactiveClient := reactive.NewClient(fakeClient.NewFakeClientWithScheme(scheme.Scheme))
					subject.KubeClient = reactiveClient

					correctPVCLabels := generateGPDBLabels(map[string]string{"greenplum-major-version": greenplumcluster.SupportedGreenplumMajorVersion})
					correctPVCTemplate := generateTestPVCTemplate(correctPVCLabels)
					incorrectPVCLabels := generateGPDBLabels(map[string]string{"greenplum-major-version": "4"})
					incorrectPVCTemplate := generateTestPVCTemplate(incorrectPVCLabels)
					createTestMasterPVCs(subject.KubeClient, correctPVCTemplate, 2)
					createTestSegmentPVCs(subject.KubeClient, incorrectPVCTemplate, 2, "segment-a")
					createTestSegmentPVCs(subject.KubeClient, correctPVCTemplate, 2, "segment-b")
				})
				It("disallows requests", func() {
					outputReview := postValidateReview(subject.Handler(), &exampleGreenplum, nil)

					Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
					Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"Message": Equal("the existing PVCs for my-gp-instance are not compatible with this controller. Expected PVCs to have greenplum-major-version=" + greenplumcluster.SupportedGreenplumMajorVersion + "; found greenplum-major-version=4"),
					})))
					Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("the existing PVCs for my-gp-instance are not compatible with this controller. Expected PVCs to have greenplum-major-version=" + greenplumcluster.SupportedGreenplumMajorVersion + "; found greenplum-major-version=4"))
				})
			})
			When("greenplum-major-version label does not match the expected value on segment-b sset PVCs", func() {
				BeforeEach(func() {
					reactiveClient := reactive.NewClient(fakeClient.NewFakeClientWithScheme(scheme.Scheme))
					subject.KubeClient = reactiveClient

					correctPVCLabels := generateGPDBLabels(map[string]string{"greenplum-major-version": greenplumcluster.SupportedGreenplumMajorVersion})
					correctPVCTemplate := generateTestPVCTemplate(correctPVCLabels)
					incorrectPVCLabels := generateGPDBLabels(map[string]string{"greenplum-major-version": "4"})
					incorrectPVCTemplate := generateTestPVCTemplate(incorrectPVCLabels)
					createTestMasterPVCs(subject.KubeClient, correctPVCTemplate, 2)
					createTestSegmentPVCs(subject.KubeClient, correctPVCTemplate, 2, "segment-a")
					createTestSegmentPVCs(subject.KubeClient, incorrectPVCTemplate, 2, "segment-b")
				})
				It("disallows requests", func() {
					outputReview := postValidateReview(subject.Handler(), &exampleGreenplum, nil)

					Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
					Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"Message": Equal("the existing PVCs for my-gp-instance are not compatible with this controller. Expected PVCs to have greenplum-major-version=" + greenplumcluster.SupportedGreenplumMajorVersion + "; found greenplum-major-version=4"),
					})))
					Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("the existing PVCs for my-gp-instance are not compatible with this controller. Expected PVCs to have greenplum-major-version=" + greenplumcluster.SupportedGreenplumMajorVersion + "; found greenplum-major-version=4"))
				})
			})
			When("greenplum-major-version label is present, but does not match the expected value", func() {
				BeforeEach(func() {
					reactiveClient := reactive.NewClient(fakeClient.NewFakeClientWithScheme(scheme.Scheme))
					subject.KubeClient = reactiveClient
					createGPDBTestPVCs(reactiveClient, 2, 2, 2,
						map[string]string{"greenplum-major-version": "4"})
				})
				It("disallows requests", func() {
					outputReview := postValidateReview(subject.Handler(), &exampleGreenplum, nil)

					Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
					Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"Message": Equal("the existing PVCs for my-gp-instance are not compatible with this controller. Expected PVCs to have greenplum-major-version=" + greenplumcluster.SupportedGreenplumMajorVersion + "; found greenplum-major-version=4"),
					})))
					Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("the existing PVCs for my-gp-instance are not compatible with this controller. Expected PVCs to have greenplum-major-version=" + greenplumcluster.SupportedGreenplumMajorVersion + "; found greenplum-major-version=4"))
				})
			})
			When("greenplum-major-version label is not present", func() {
				BeforeEach(func() {
					reactiveClient := reactive.NewClient(fakeClient.NewFakeClientWithScheme(scheme.Scheme))
					subject.KubeClient = reactiveClient
					createGPDBTestPVCs(reactiveClient, 2, 2, 2, nil)
				})
				It("disallows requests", func() {
					outputReview := postValidateReview(subject.Handler(), &exampleGreenplum, nil)

					Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
					Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"Message": Equal("the existing PVCs for my-gp-instance are not compatible with this controller. Expected PVCs to have greenplum-major-version=" + greenplumcluster.SupportedGreenplumMajorVersion + "; found no label"),
					})))
					Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("the existing PVCs for my-gp-instance are not compatible with this controller. Expected PVCs to have greenplum-major-version=" + greenplumcluster.SupportedGreenplumMajorVersion + "; found no label"))
				})
			})
		})

		When("no PVCs exist", func() {
			It("allows requests", func() {
				newGreenplum := exampleGreenplum.DeepCopy()

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeTrue(), "should be allowed")
				Expect(outputReview.Response.Result).To(BeNil())
				Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			})
		})
	})
	When("a cluster with no standby or mirrors has been deleted, but PVCs still exist", func() {
		When("PVCs exist", func() {
			var newGreenplum *greenplumv1.GreenplumCluster
			BeforeEach(func() {
				newGreenplum = exampleGreenplum.DeepCopy()
				newGreenplum.Spec.MasterAndStandby.Standby = "no"
				newGreenplum.Spec.MasterAndStandby.AntiAffinity = "no"
				newGreenplum.Spec.Segments.Mirrors = "no"
				newGreenplum.Spec.Segments.AntiAffinity = "no"

				createGPDBTestPVCs(subject.KubeClient, 1, 2, 0,
					map[string]string{"greenplum-major-version": greenplumcluster.SupportedGreenplumMajorVersion})
			})
			It("disallows requests that change standby", func() {
				newGreenplum.Spec.MasterAndStandby.Standby = "yes"

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				expectedMessage := "my-gp-instance has PVCs for 1 masters. masterAndStandby.standby cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal(expectedMessage),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(expectedMessage))
			})
			It("disallows requests that change mirrors", func() {
				newGreenplum.Spec.Segments.Mirrors = "yes"

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				expectedMessage := "my-gp-instance has PVCs for 0 mirrors. segments.mirrors cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal(expectedMessage),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(expectedMessage))
			})
			It("allows requests that change the value for standby without changing its meaning", func() {
				newGreenplum.Spec.MasterAndStandby.Standby = "NO"

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeTrue(), "should be allowed")
				Expect(outputReview.Response.Result).To(BeNil())
				Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			})
		})

		When("no PVCs exist", func() {
			It("allows requests", func() {
				newGreenplum := exampleGreenplum.DeepCopy()

				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeTrue(), "should be allowed")
				Expect(outputReview.Response.Result).To(BeNil())
				Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			})
		})
	})

	When("mirrors=no", func() {
		var (
			newGreenplum *greenplumv1.GreenplumCluster
		)
		BeforeEach(func() {
			newGreenplum = exampleGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.Mirrors = "no"
		})

		When("antiAffinity is not specified", func() {
			It("rejects the request", func() {
				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal(`when mirrors is set to "no", antiAffinity must also be set to "no"`),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(`when mirrors is set to "no", antiAffinity must also be set to "no"`))
			})
		})

		When("antiAffinity=yes", func() {
			BeforeEach(func() {
				newGreenplum.Spec.MasterAndStandby.AntiAffinity = "yes"
				newGreenplum.Spec.Segments.AntiAffinity = "yes"
			})
			It("rejects the request", func() {
				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal(`when mirrors is set to "no", antiAffinity must also be set to "no"`),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(`when mirrors is set to "no", antiAffinity must also be set to "no"`))
			})
		})

		When("antiAffinity=no", func() {
			BeforeEach(func() {
				newGreenplum.Spec.MasterAndStandby.AntiAffinity = "no"
				newGreenplum.Spec.Segments.AntiAffinity = "no"
			})
			It("approves the request", func() {
				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeTrue(), "should be allowed")
				Expect(outputReview.Response.Result).To(BeNil())
				Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			})
		})
	})

	When("standby=no", func() {
		var newGreenplum *greenplumv1.GreenplumCluster
		BeforeEach(func() {
			newGreenplum = exampleGreenplum.DeepCopy()
			newGreenplum.Spec.MasterAndStandby.Standby = "no"
		})

		When("antiAffinity is not specified", func() {
			It("rejects the request", func() {
				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal(`when standby is set to "no", antiAffinity must also be set to "no"`),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(`when standby is set to "no", antiAffinity must also be set to "no"`))
			})
		})

		When("antiAffinity=yes", func() {
			BeforeEach(func() {
				newGreenplum.Spec.MasterAndStandby.AntiAffinity = "yes"
				newGreenplum.Spec.Segments.AntiAffinity = "yes"
			})
			It("rejects the request", func() {
				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal(`when standby is set to "no", antiAffinity must also be set to "no"`),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(`when standby is set to "no", antiAffinity must also be set to "no"`))
			})
		})

		When("antiAffinity=no", func() {
			BeforeEach(func() {
				newGreenplum.Spec.MasterAndStandby.AntiAffinity = "no"
				newGreenplum.Spec.Segments.AntiAffinity = "no"
			})
			It("approves the request", func() {
				outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

				Expect(outputReview.Response.Allowed).To(BeTrue(), "should be allowed")
				Expect(outputReview.Response.Result).To(BeNil())
				Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			})
		})
	})

	When("clusterExistsInNamespace check fails", func() {
		BeforeEach(func() {
			reactiveClient := reactive.NewClient(fakeClient.NewFakeClientWithScheme(scheme.Scheme))
			reactiveClient.PrependReactor("list", "greenplumclusters", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.New("custom statefulset error")
			})
			subject.KubeClient = reactiveClient
		})
		It("rejects the CREATE request with a message", func() {
			newGreenplum := exampleGreenplum.DeepCopy()

			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

			Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal("could not check if a cluster exists in namespace test-ns. custom statefulset error"),
			})))
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("could not check if a cluster exists in namespace test-ns. custom statefulset error"))
		})
	})

	When("a gpinstance exists", func() {
		BeforeEach(func() {
			reactiveClient := reactive.NewClient(fakeClient.NewFakeClientWithScheme(scheme.Scheme))
			reactiveClient.PrependReactor("list", "greenplumclusters", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				list := &greenplumv1.GreenplumClusterList{Items: []greenplumv1.GreenplumCluster{{}}}
				return true, list, nil
			})
			subject.KubeClient = reactiveClient
		})
		It("rejects the CREATE request for it ", func() {
			newGreenplum := exampleGreenplum.DeepCopy()

			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)

			Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal("only one GreenplumCluster is allowed in namespace test-ns"),
			})))
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry("only one GreenplumCluster is allowed in namespace test-ns"))
		})
	})

	DescribeTable("rejects invalid masterAndStandby workerSelector key/value",
		func(workerSelectorMap map[string]string) {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.MasterAndStandby.WorkerSelector = workerSelectorMap
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeFalse(), "did not match expected allowed value")

			expectedMessage := "masterAndStandby workerSelector key/value is longer than 63 characters"
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(expectedMessage))
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal(expectedMessage),
			})))
		},
		Entry("workerselector key is > 63 chars",
			map[string]string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": "labelValue"}),
		Entry("workerselector value is > 63 chars",
			map[string]string{"labelName": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}),
		Entry("workerselector key and value are both > 63 chars",
			map[string]string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}),
	)

	When("masterAndStandby workerSelector key/value are both < 63 characters", func() {
		It("allows the request", func() {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.MasterAndStandby.WorkerSelector = map[string]string{
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			}
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeTrue(), "did not match expected allowed value")
			Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			Expect(outputReview.Response.Result).To(BeNil())
		})
	})

	DescribeTable("rejects invalid segments workerSelector key/value",
		func(workerSelectorMap map[string]string) {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.WorkerSelector = workerSelectorMap
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeFalse(), "did not match expected allowed value")

			expectedMessage := "segments workerSelector key/value is longer than 63 characters"
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(expectedMessage))
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal(expectedMessage),
			})))
		},
		Entry("workerselector key is > 63 chars",
			map[string]string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": "labelValue"}),
		Entry("workerselector value is > 63 chars",
			map[string]string{"labelName": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}),
		Entry("workerselector key and value are both > 63 chars",
			map[string]string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}),
	)

	When("segments workerSelector key/value are both < 63 characters", func() {
		It("allows the request", func() {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.WorkerSelector = map[string]string{
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			}
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeTrue(), "did not match expected allowed value")
			Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			Expect(outputReview.Response.Result).To(BeNil())
		})
	})

	When("masterAndStandby cpu < 0", func() {
		It("rejects the request", func() {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.MasterAndStandby.CPU = resource.MustParse("-1")
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeFalse(), "did not match expected allowed value")
			expectedMessage := `invalid masterAndStandby cpu value: "-1": must be greater than or equal to 0`
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(expectedMessage))
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal(expectedMessage),
			})))
		})
	})

	DescribeTable("allows masterAndStandby cpu values >= 0",
		func(cpuValue resource.Quantity) {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.MasterAndStandby.CPU = cpuValue
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeTrue(), "did not match expected allowed value")

			Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			Expect(outputReview.Response.Result).To(BeNil())
		},
		Entry("cpu = 0", resource.MustParse("0")),
		Entry("cpu = 1", resource.MustParse("1")),
	)

	When("segments cpu < 0", func() {
		It("rejects the request", func() {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.CPU = resource.MustParse("-1")
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeFalse(), "did not match expected allowed value")
			expectedMessage := `invalid segments cpu value: "-1": must be greater than or equal to 0`
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(expectedMessage))
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal(expectedMessage),
			})))
		})
	})

	DescribeTable("allows segments cpu values >= 0",
		func(cpuValue resource.Quantity) {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.CPU = cpuValue
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeTrue(), "did not match expected allowed value")

			Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			Expect(outputReview.Response.Result).To(BeNil())
		},
		Entry("cpu = 0", resource.MustParse("0")),
		Entry("cpu = 1", resource.MustParse("1")),
	)

	When("masterAndStandby memory < 0", func() {
		It("rejects the request", func() {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.MasterAndStandby.Memory = resource.MustParse("-1")
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeFalse(), "did not match expected allowed value")
			expectedMessage := `invalid masterAndStandby memory value: "-1": must be greater than or equal to 0`
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(expectedMessage))
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal(expectedMessage),
			})))
		})
	})

	DescribeTable("allows masterAndStandby memory values >= 0",
		func(memoryValue resource.Quantity) {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.MasterAndStandby.Memory = memoryValue
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeTrue(), "did not match expected allowed value")

			Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			Expect(outputReview.Response.Result).To(BeNil())
		},
		Entry("memory = 0", resource.MustParse("0")),
		Entry("memory = 1", resource.MustParse("1")),
	)

	When("segments memory < 0", func() {
		It("rejects the request", func() {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.Memory = resource.MustParse("-1")
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeFalse(), "did not match expected allowed value")
			expectedMessage := `invalid segments memory value: "-1": must be greater than or equal to 0`
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(expectedMessage))
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal(expectedMessage),
			})))
		})
	})

	DescribeTable("allows segments memory values >= 0",
		func(memoryValue resource.Quantity) {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.Memory = memoryValue
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeTrue(), "did not match expected allowed value")

			Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			Expect(outputReview.Response.Result).To(BeNil())
		},
		Entry("memory = 0", resource.MustParse("0")),
		Entry("memory = 1", resource.MustParse("1")),
	)

	When("masterAndStandby storage < 0", func() {
		It("rejects the request", func() {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.MasterAndStandby.Storage = resource.MustParse("-1")
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeFalse(), "did not match expected allowed value")
			expectedMessage := `invalid masterAndStandby storage value: "-1": must be greater than or equal to 0`
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(expectedMessage))
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal(expectedMessage),
			})))
		})
	})

	DescribeTable("allows masterAndStandby storage values >= 0",
		func(storageValue resource.Quantity) {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.MasterAndStandby.Storage = storageValue
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeTrue(), "did not match expected allowed value")

			Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			Expect(outputReview.Response.Result).To(BeNil())
		},
		Entry("storage = 0", resource.MustParse("0")),
		Entry("storage = 1", resource.MustParse("1")),
	)

	When("segments storage < 0", func() {
		It("rejects the request", func() {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.Storage = resource.MustParse("-1")
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeFalse(), "did not match expected allowed value")
			expectedMessage := `invalid segments storage value: "-1": must be greater than or equal to 0`
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedGreenplumClusterEntry(expectedMessage))
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal(expectedMessage),
			})))
		})
	})

	DescribeTable("allows segments storage values >= 0",
		func(storageValue resource.Quantity) {
			newGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.Storage = storageValue
			outputReview := postValidateReview(subject.Handler(), newGreenplum, nil)
			Expect(outputReview.Response.Allowed).To(BeTrue(), "did not match expected allowed value")

			Expect(DecodeLogs(logBuf)).To(ContainAllowedGreenplumClusterEntry())
			Expect(outputReview.Response.Result).To(BeNil())
		},
		Entry("storage = 0", resource.MustParse("0")),
		Entry("storage = 1", resource.MustParse("1")),
	)
})

func generateGPDBLabels(additionalLabels map[string]string) map[string]string {
	labels := map[string]string{
		"app":               "greenplum",
		"greenplum-cluster": exampleGreenplum.Name,
	}
	for key, val := range additionalLabels {
		labels[key] = val
	}
	return labels
}

func createGPDBTestPVCs(kubeClient client.Client, masterCount, primaryCount, mirrorCount int, additionalLabels map[string]string) {
	labels := generateGPDBLabels(additionalLabels)
	pvcTemplate := generateTestPVCTemplate(labels)
	createTestMasterPVCs(kubeClient, pvcTemplate, masterCount)
	createTestSegmentPVCs(kubeClient, pvcTemplate, primaryCount, "segment-a")
	createTestSegmentPVCs(kubeClient, pvcTemplate, mirrorCount, "segment-b")
}

func generateTestPVCTemplate(labels map[string]string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: heapvalue.NewString("standard"),
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: storageRequirements("10G"),
		},
	}
}

func createTestMasterPVCs(kubeClient client.Client, pvcTemplate *corev1.PersistentVolumeClaim, masterCount int) {
	for i := 0; i < masterCount; i++ {
		masterPVC := pvcTemplate.DeepCopy()
		masterPVC.Name = fmt.Sprintf("master-%d", i)
		masterPVC.Labels["type"] = "master"
		Expect(kubeClient.Create(nil, masterPVC)).To(Succeed())
	}
}

func createTestSegmentPVCs(kubeClient client.Client, pvcTemplate *corev1.PersistentVolumeClaim, ssetSize int, ssetName string) {
	for i := 0; i < ssetSize; i++ {
		segmentPVC := pvcTemplate.DeepCopy()
		segmentPVC.Name = fmt.Sprintf("%s-%d", ssetName, i)
		segmentPVC.Labels["type"] = ssetName
		segmentPVC.Spec.Resources = storageRequirements("20G")
		Expect(kubeClient.Create(nil, segmentPVC)).To(Succeed())
	}
}

func storageRequirements(size string) corev1.ResourceRequirements {
	storageQuantity := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(size),
	}
	return corev1.ResourceRequirements{
		Limits:   storageQuantity,
		Requests: storageQuantity,
	}
}
