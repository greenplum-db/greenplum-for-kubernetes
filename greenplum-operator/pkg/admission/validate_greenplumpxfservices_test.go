package admission_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1beta1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/admission"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/reactive"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/gplog/testing"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("validateGreenplumPXFService", func() {

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
	ContainDisallowedPXFEntry := func(expectedMessage string, operation string) types.GomegaMatcher {
		return ContainLogEntry(Keys{
			"msg":       Equal("/validate"),
			"GVK":       Equal("greenplum.pivotal.io/v1beta1, Kind=GreenplumPXFService"),
			"Name":      Equal("my-gp-pxf-instance"),
			"Namespace": Equal("test-ns"),
			"UID":       Equal("my-gp-pxf-instance-uid"),
			"Operation": Equal(operation),
			"Allowed":   BeFalse(),
			"Message":   Equal(expectedMessage),
		})
	}
	ContainAllowedPXFEntry := func(operation string) types.GomegaMatcher {
		return ContainLogEntry(Keys{
			"msg":       Equal("/validate"),
			"GVK":       Equal("greenplum.pivotal.io/v1beta1, Kind=GreenplumPXFService"),
			"Name":      Equal("my-gp-pxf-instance"),
			"Namespace": Equal("test-ns"),
			"UID":       Equal("my-gp-pxf-instance-uid"),
			"Operation": Equal(operation),
			"Allowed":   BeTrue(),
		})
	}

	DescribeTable("rejects invalid pxf workerSelector key/value",
		func(workerSelectorMap map[string]string) {
			newPXF := examplePXF.DeepCopy()
			newPXF.Spec.WorkerSelector = workerSelectorMap
			outputReview := postValidateReview(subject.Handler(), newPXF, nil)
			Expect(outputReview.Response.Allowed).To(BeFalse(), "did not match expected allowed value")

			expectedMessage := "pxf workerSelector key/value is longer than 63 characters"
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedPXFEntry(expectedMessage, "CREATE"))
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

	When("pxf workerSelector key/value are both < 63 characters", func() {
		It("allows the request", func() {
			newPXF := examplePXF.DeepCopy()
			newPXF.Spec.WorkerSelector = map[string]string{
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			}
			outputReview := postValidateReview(subject.Handler(), newPXF, nil)
			Expect(outputReview.Response.Allowed).To(BeTrue(), "did not match expected allowed value")
			Expect(DecodeLogs(logBuf)).To(ContainAllowedPXFEntry("CREATE"))
			Expect(outputReview.Response.Result).To(BeNil())
		})
	})

	When("pxf cpu < 0", func() {
		It("rejects the request", func() {
			newPXF := examplePXF.DeepCopy()
			newPXF.Spec.CPU = resource.MustParse("-1")
			outputReview := postValidateReview(subject.Handler(), newPXF, nil)
			Expect(outputReview.Response.Allowed).To(BeFalse(), "did not match expected allowed value")
			expectedMessage := `invalid pxf cpu value: "-1": must be greater than or equal to 0`
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedPXFEntry(expectedMessage, "CREATE"))
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal(expectedMessage),
			})))
		})
	})

	DescribeTable("allows pxf cpu values >= 0",
		func(cpuValue resource.Quantity) {
			newPXF := examplePXF.DeepCopy()
			newPXF.Spec.CPU = cpuValue
			outputReview := postValidateReview(subject.Handler(), newPXF, nil)
			Expect(outputReview.Response.Allowed).To(BeTrue(), "did not match expected allowed value")

			Expect(DecodeLogs(logBuf)).To(ContainAllowedPXFEntry("CREATE"))
			Expect(outputReview.Response.Result).To(BeNil())
		},
		Entry("cpu = 0", resource.MustParse("0")),
		Entry("cpu = 1", resource.MustParse("1")),
	)

	When("pxf memory < 0", func() {
		It("rejects the request", func() {
			newPXF := examplePXF.DeepCopy()
			newPXF.Spec.Memory = resource.MustParse("-1")
			outputReview := postValidateReview(subject.Handler(), newPXF, nil)
			Expect(outputReview.Response.Allowed).To(BeFalse(), "did not match expected allowed value")
			expectedMessage := `invalid pxf memory value: "-1": must be greater than or equal to 0`
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedPXFEntry(expectedMessage, "CREATE"))
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal(expectedMessage),
			})))
		})
	})

	DescribeTable("allows pxf memory values >= 0",
		func(memoryValue resource.Quantity) {
			newPXF := examplePXF.DeepCopy()
			newPXF.Spec.Memory = memoryValue
			outputReview := postValidateReview(subject.Handler(), newPXF, nil)
			Expect(outputReview.Response.Allowed).To(BeTrue(), "did not match expected allowed value")

			Expect(DecodeLogs(logBuf)).To(ContainAllowedPXFEntry("CREATE"))
			Expect(outputReview.Response.Result).To(BeNil())
		},
		Entry("memory = 0", resource.MustParse("0")),
		Entry("memory = 1", resource.MustParse("1")),
	)

	When("a PXF exists from the current controller", func() {
		var oldPXF, newPXF *v1beta1.GreenplumPXFService
		BeforeEach(func() {
			oldPXF = examplePXF.DeepCopy()
			oldPXF.Spec.Replicas = 1
			newPXF = oldPXF.DeepCopy()
			newPXF.Spec.Replicas = 1
			subject.InstanceImage = "v1.0.1"

			pxfDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      oldPXF.Name,
					Namespace: oldPXF.Namespace,
				},
			}
			pxfDeployment.Spec.Template.Spec.Containers = []corev1.Container{{
				Image: "v1.0.1",
			}}
			err := subject.KubeClient.Create(nil, pxfDeployment)
			Expect(err).NotTo(HaveOccurred())
		})
		It("allows updates that make no changes", func() {
			outputReview := postValidateReview(subject.Handler(), newPXF, oldPXF)

			Expect(outputReview.Response.Allowed).To(BeTrue(), "should be allowed")
			Expect(DecodeLogs(logBuf)).To(ContainAllowedPXFEntry("UPDATE"))

		})
	})

	When("a PXF exists from an old controller", func() {
		var oldPXF, newPXF *v1beta1.GreenplumPXFService
		BeforeEach(func() {
			oldPXF = examplePXF.DeepCopy()
			oldPXF.Spec.Replicas = 1
			newPXF = oldPXF.DeepCopy()
			newPXF.Spec.Replicas = 2
			subject.InstanceImage = "v1.0.1"

			pxfDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      oldPXF.Name,
					Namespace: oldPXF.Namespace,
				},
			}
			pxfDeployment.Spec.Template.Spec.Containers = []corev1.Container{{
				Image: "v1.0.0",
			}}
			err := subject.KubeClient.Create(nil, pxfDeployment)
			Expect(err).NotTo(HaveOccurred())
		})

		It("disallows update requests that update an old PXF instance", func() {
			outputReview := postValidateReview(subject.Handler(), newPXF, oldPXF)

			Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
			const expectedMessage = admission.UpgradePXFHelpMsg + `; GreenplumPXFService has image: v1.0.0; Operator supports image: v1.0.1`
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal(expectedMessage),
			})))
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedPXFEntry(expectedMessage, "UPDATE"))
		})
	})

	When("No pxf deployment exists", func() {
		var oldPXF, newPXF *v1beta1.GreenplumPXFService
		BeforeEach(func() {
			oldPXF = examplePXF.DeepCopy()
			oldPXF.Spec.Replicas = 1
			newPXF = oldPXF.DeepCopy()
			newPXF.Spec.Replicas = 2

			oldDeployment := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: oldPXF.Namespace,
					Name:      oldPXF.Name,
				},
			}
			err := subject.KubeClient.Delete(nil, &oldDeployment)
			Expect(err).To(MatchError(`deployments.apps "my-gp-pxf-instance" not found`))
		})
		It("rejects update and logs an error", func() {
			outputReview := postValidateReview(subject.Handler(), newPXF, oldPXF)
			Expect(outputReview.Response.Allowed).To(BeFalse())
			expectedMessage := `failed to get PXF Deployment. Try again later: deployments.apps "my-gp-pxf-instance" not found`
			Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal(expectedMessage),
			})))
			Expect(DecodeLogs(logBuf)).To(ContainDisallowedPXFEntry(expectedMessage, "UPDATE"))
		})
	})

})
