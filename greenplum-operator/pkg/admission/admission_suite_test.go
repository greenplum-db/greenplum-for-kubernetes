package admission_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestValidationwebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validationwebhook Suite")
}

var exampleGreenplum = greenplumv1.GreenplumCluster{
	TypeMeta: metav1.TypeMeta{
		Kind:       "GreenplumCluster",
		APIVersion: "greenplum.pivotal.io/v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "my-gp-instance",
		Namespace: "test-ns",
	},
	Spec: greenplumv1.GreenplumClusterSpec{
		MasterAndStandby: greenplumv1.GreenplumMasterAndStandbySpec{
			GreenplumPodSpec: greenplumv1.GreenplumPodSpec{
				Memory:           resource.MustParse("1G"),
				CPU:              resource.MustParse("1.0"),
				Storage:          resource.MustParse("10G"),
				StorageClassName: "standard",
				AntiAffinity:     "yes",
			},
			Standby: "yes",
		},
		Segments: greenplumv1.GreenplumSegmentsSpec{
			GreenplumPodSpec: greenplumv1.GreenplumPodSpec{
				Memory:           resource.MustParse("1G"),
				CPU:              resource.MustParse("1.0"),
				Storage:          resource.MustParse("20G"),
				StorageClassName: "standard",
				AntiAffinity:     "yes",
			},
			PrimarySegmentCount: 5,
			Mirrors:             "yes",
		},
	},
	Status: greenplumv1.GreenplumClusterStatus{
		Phase: greenplumv1.GreenplumClusterPhaseRunning,
	},
}

var examplePXFObjectMeta = metav1.ObjectMeta{
	Name:      "my-gp-pxf-instance",
	Namespace: "test-ns",
}
var examplePXF = greenplumv1.GreenplumPXFService{
	TypeMeta: metav1.TypeMeta{
		Kind:       "GreenplumPXFService",
		APIVersion: "greenplum.pivotal.io/v1",
	},
	ObjectMeta: examplePXFObjectMeta,
	Spec: greenplumv1.GreenplumPXFServiceSpec{
		Replicas: 2,
		CPU:      resource.MustParse("2"),
		Memory:   resource.MustParse("2G"),
	},
}

func postValidateReview(handler http.Handler, newObj, oldObj runtime.Object) (outputReview admissionv1.AdmissionReview) {
	srv := httptest.NewServer(handler)
	defer srv.Close()

	var inputReview admissionv1.AdmissionReview
	rb := SampleAdmissionReviewRequest()
	if newObj != nil {
		rb.NewObj(newObj)
	}
	if oldObj != nil {
		rb.OldObj(oldObj)
	}
	inputReview = rb.Build()
	resp, err := http.Post(srv.URL+"/validate", "application/json", marshal(inputReview))
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	unmarshal(resp.Body, &outputReview)
	return
}
