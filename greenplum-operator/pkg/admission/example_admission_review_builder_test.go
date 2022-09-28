package admission_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type AdmissionReviewRequestBuilder struct {
	r admissionv1beta1.AdmissionReview
}

func SampleAdmissionReviewRequest() AdmissionReviewRequestBuilder {
	return AdmissionReviewRequestBuilder{admissionv1beta1.AdmissionReview{
		Request: &admissionv1beta1.AdmissionRequest{},
	}}
}

func (b AdmissionReviewRequestBuilder) Kind(o runtime.Object) AdmissionReviewRequestBuilder {
	gvks, _, err := scheme.Scheme.ObjectKinds(o)
	Expect(err).NotTo(HaveOccurred())
	Expect(scheme.Scheme.Convert(&gvks[0], &b.r.Request.Kind, nil)).To(Succeed())
	return b
}

func (b AdmissionReviewRequestBuilder) namesFrom(o runtime.Object) AdmissionReviewRequestBuilder {
	if b.r.Request.Kind.Kind == "" {
		b.Kind(o)
	}
	mo, err := meta.Accessor(o)
	if err != nil {
		panic(err)
	}
	if b.r.Request.Name == "" {
		b.r.Request.Name = mo.GetName()
	}
	if b.r.Request.Namespace == "" {
		b.r.Request.Namespace = mo.GetNamespace()
	}
	if b.r.Request.UID == "" {
		b.r.Request.UID = types.UID(mo.GetName() + "-uid")
	}
	return b
}

func (b AdmissionReviewRequestBuilder) NewObj(o runtime.Object) AdmissionReviewRequestBuilder {
	b.namesFrom(o)
	b.r.Request.Object = runtime.RawExtension{Object: o}
	return b
}

func (b AdmissionReviewRequestBuilder) OldObj(o runtime.Object) AdmissionReviewRequestBuilder {
	b.namesFrom(o)
	b.r.Request.OldObject = runtime.RawExtension{Object: o}
	return b
}

func (b AdmissionReviewRequestBuilder) NewObjInvalid(note string) AdmissionReviewRequestBuilder {
	b.r.Request.Object = *invalidRaw("new " + note)
	return b
}

func (b AdmissionReviewRequestBuilder) OldObjInvalid(note string) AdmissionReviewRequestBuilder {
	b.r.Request.Operation = admissionv1beta1.Update
	b.r.Request.OldObject = *invalidRaw("old " + note)
	return b
}

func (b AdmissionReviewRequestBuilder) Build() admissionv1beta1.AdmissionReview {
	if b.r.Request.Operation == "" {
		oldObj := b.r.Request.OldObject
		if oldObj.Object == nil && oldObj.Raw == nil {
			b.r.Request.Operation = admissionv1beta1.Create
		} else {
			b.r.Request.Operation = admissionv1beta1.Update
		}
	}
	return b.r
}

func invalidRaw(note string) *runtime.RawExtension {
	return &runtime.RawExtension{
		Raw: []byte(fmt.Sprintf(`["invalid %s"]`, note)),
	}
}

var _ = Describe("SampleAdmissionReviewRequest", func() {
	var review admissionv1beta1.AdmissionReview
	When("given a PXF as a NewObj", func() {
		BeforeEach(func() {
			review = SampleAdmissionReviewRequest().NewObj(&examplePXF).Build()
		})
		It("sets the review object", func() {
			Expect(review.Request.Object.Object).To(BeIdenticalTo(&examplePXF))
		})
		It("infers the operation is CREATE", func() {
			Expect(review.Request.Operation).To(Equal(admissionv1beta1.Create))
		})
		It("sets the Kind", func() {
			Expect(review.Request.Kind).To(Equal(
				metav1.GroupVersionKind{Group: "greenplum.pivotal.io", Version: "v1beta1", Kind: "GreenplumPXFService"}))
		})
		It("sets the namespace from the NewObj", func() {
			Expect(review.Request.Namespace).To(Equal("test-ns"))
		})
		It("sets the name from the NewObj", func() {
			Expect(review.Request.Name).To(Equal("my-gp-pxf-instance"))
		})
		It("generates a UID based on the name from the NewObj", func() {
			Expect(review.Request.UID).To(Equal(types.UID("my-gp-pxf-instance-uid")))
		})
	})

	When("given a GreenplumCluster as an OldObj, and a PXF as NewObj", func() {
		var greenplum *v1.GreenplumCluster
		BeforeEach(func() {
			// It's not actually valid to use different Kinds in a review like this, but we can test
			// that we get the right kind/name
			greenplum = exampleGreenplum.DeepCopy()
			greenplum.Namespace = "another-ns"
			review = SampleAdmissionReviewRequest().OldObj(greenplum).NewObj(&examplePXF).Build()
		})
		It("sets the review OldObject", func() {
			Expect(review.Request.OldObject.Object).To(BeIdenticalTo(greenplum))
		})
		It("sets the review object", func() {
			Expect(review.Request.Object.Object).To(BeIdenticalTo(&examplePXF))
		})
		It("infers the operation is UPDATE", func() {
			Expect(review.Request.Operation).To(Equal(admissionv1beta1.Update))
		})
		It("sets the Kind", func() {
			Expect(review.Request.Kind).To(Equal(
				metav1.GroupVersionKind{Group: "greenplum.pivotal.io", Version: "v1", Kind: "GreenplumCluster"}))
		})
		It("sets the namespace from the NewObj", func() {
			Expect(review.Request.Namespace).To(Equal("another-ns"))
		})
		It("sets the name from the NewObj", func() {
			Expect(review.Request.Name).To(Equal("my-gp-instance"))
		})
		It("generates a UID based on the name from the NewObj", func() {
			Expect(review.Request.UID).To(Equal(types.UID("my-gp-instance-uid")))
		})
	})

	When("asked to assign an invalid NewObj", func() {
		BeforeEach(func() {
			review = SampleAdmissionReviewRequest().NewObjInvalid("thing").Build()
		})
		It("sets the review object", func() {
			Expect(review.Request.Object.Raw).To(Equal([]byte(`["invalid new thing"]`)))
		})
	})

	When("asked to assign an invalid OldObj", func() {
		BeforeEach(func() {
			review = SampleAdmissionReviewRequest().OldObjInvalid("thing").Build()
		})
		It("sets the review object", func() {
			Expect(review.Request.OldObject.Raw).To(Equal([]byte(`["invalid old thing"]`)))
		})
	})

})
