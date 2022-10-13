package admission_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/admission"
	fakePodExec "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor/fake"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/reactive"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/gplog/testing"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("HandleValidate", func() {
	var (
		req     *http.Request
		resp    http.ResponseWriter
		subject admission.Handler
		logBuf  *gbytes.Buffer
	)
	JustBeforeEach(func() {
		reactiveClient := reactive.NewClient(fakeClient.NewFakeClientWithScheme(scheme.Scheme))
		subject = admission.Handler{
			KubeClient:     reactiveClient,
			PodCmdExecutor: &fakePodExec.PodExec{StdoutResult: "0\n"},
		}
		logBuf = gbytes.NewBuffer()
		admission.Log = gplog.ForTest(logBuf)
		req.Header.Add("Content-Type", "application/json")
		subject.HandleValidate(resp, req)
	})
	When("the request ContentType is not application/json", func() {
		var respRec *httptest.ResponseRecorder
		BeforeEach(func() {
			oldGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum := oldGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.PrimarySegmentCount++
			inputReview := SampleAdmissionReviewRequest().NewObj(newGreenplum).OldObj(oldGreenplum).Build()
			req = httptest.NewRequest("POST", "/validate", marshal(inputReview))
			respRec = httptest.NewRecorder()
			resp = respRec
			req.Header.Add("Content-Type", "foo/bar")
		})
		It("replies with an error message", func() {
			Expect(respRec.Code).To(Equal(http.StatusUnsupportedMediaType))
			buf, err := ioutil.ReadAll(respRec.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(buf)).To(Equal("invalid Content-Type, expect `application/json`\n"))
		})
	})
	When("the request can't be read", func() {
		var respRec *httptest.ResponseRecorder
		BeforeEach(func() {
			req = httptest.NewRequest("POST", "/validate", brokenReader{})
			respRec = httptest.NewRecorder()
			resp = respRec
		})
		It("replies with Bad Request", func() {
			Expect(respRec.Result().StatusCode).To(Equal(http.StatusBadRequest))
		})
		It("replies with an error message", func() {
			buf, err := ioutil.ReadAll(respRec.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(buf)).To(Equal("couldn't read request body: injected error from broken reader\n"))
		})
	})
	When("the request is corrupted", func() {
		var respRec *httptest.ResponseRecorder
		BeforeEach(func() {
			req = httptest.NewRequest("POST", "/validate", strings.NewReader("an invalid admission review"))
			respRec = httptest.NewRecorder()
			resp = respRec
		})
		It("replies with an error message", func() {
			Expect(respRec.Code).To(Equal(http.StatusOK))
			var admissionReview admissionv1beta1.AdmissionReview
			unmarshal(respRec.Body, &admissionReview)
			Expect(admissionReview.Response.Allowed).To(BeFalse(), "should not be allowed")
			Expect(admissionReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": HavePrefix("parsing request:"),
			})))
			Expect(DecodeLogs(logBuf)).To(ContainLogEntry(Keys{
				"msg":     Equal("/validate"),
				"Allowed": BeFalse(),
				"Message": HavePrefix("parsing request:"),
			}))
		})
	})
	When("the response can't be written", func() {
		BeforeEach(func() {
			oldGreenplum := exampleGreenplum.DeepCopy()
			newGreenplum := oldGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.PrimarySegmentCount++

			inputReview := SampleAdmissionReviewRequest().NewObj(newGreenplum).OldObj(oldGreenplum).Build()

			req = httptest.NewRequest("POST", "/validate", marshal(inputReview))
			resp = &brokenResponseWriter{}
		})
		It("logs an error", func() {
			Expect(logBuf).To(gbytes.Say(`"msg":"responding to admission review","error":"injected error from broken response writer"`))
		})
	})
	When("the request contains an object that exists in our scheme, but we don't have a validation handler for it", func() {
		var respRec *httptest.ResponseRecorder
		BeforeEach(func() {
			reader := marshal(corev1PodAdmissionReviewRequest())
			req = httptest.NewRequest("POST", "/validate", reader)
			respRec = httptest.NewRecorder()
			resp = respRec
		})
		It("rejects the request", func() {
			var admissionReview admissionv1beta1.AdmissionReview
			unmarshal(respRec.Body, &admissionReview)
			Expect(admissionReview.Response.Allowed).To(BeFalse(), "should not be allowed")
			Expect(admissionReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal("unexpected validation request for object: core/v1, Kind=Pod"),
			})))
			Expect(DecodeLogs(logBuf)).To(ContainLogEntry(Keys{
				"msg":       Equal("/validate"),
				"GVK":       Equal("core/v1, Kind=Pod"),
				"Name":      Equal("my-fake-instance"),
				"Namespace": BeEmpty(),
				"UID":       Equal("my-fake-uid"),
				"Operation": Equal("CREATE"),
				"Allowed":   BeFalse(),
				"Message":   Equal("unexpected validation request for object: core/v1, Kind=Pod"),
			}))
		})
	})
	When("the request contains a new GreenplumCluster that cannot be unmarshalled", func() {
		var respRec *httptest.ResponseRecorder
		BeforeEach(func() {
			reader := marshal(SampleAdmissionReviewRequest().NewObj(&exampleGreenplum).NewObjInvalid("greenplumcluster").Build())
			req = httptest.NewRequest("POST", "/validate", reader)
			respRec = httptest.NewRecorder()
			resp = respRec
		})
		It("rejects the request", func() {
			var admissionReview admissionv1beta1.AdmissionReview
			unmarshal(respRec.Body, &admissionReview)
			Expect(admissionReview.Response.Allowed).To(BeFalse(), "should not be allowed")
			Expect(admissionReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal("failed to unmarshal Request.Object into GreenplumCluster: json: cannot unmarshal array into Go value of type v1.GreenplumCluster"),
			})))
			Expect(DecodeLogs(logBuf)).To(ContainLogEntry(Keys{
				"msg":       Equal("/validate"),
				"GVK":       Equal("greenplum.pivotal.io/v1, Kind=GreenplumCluster"),
				"Name":      Equal("my-gp-instance"),
				"Namespace": Equal("test-ns"),
				"UID":       Equal("my-gp-instance-uid"),
				"Operation": Equal("CREATE"),
				"Allowed":   BeFalse(),
				"Message":   Equal("failed to unmarshal Request.Object into GreenplumCluster: json: cannot unmarshal array into Go value of type v1.GreenplumCluster"),
			}))
		})
	})
	When("the request contains an old GreenplumCluster that cannot be unmarshalled", func() {
		var respRec *httptest.ResponseRecorder
		BeforeEach(func() {
			reader := marshal(SampleAdmissionReviewRequest().NewObj(&exampleGreenplum).OldObjInvalid("greenplumcluster").Build())
			req = httptest.NewRequest("POST", "/validate", reader)
			respRec = httptest.NewRecorder()
			resp = respRec
		})
		It("rejects the request", func() {
			var admissionReview admissionv1beta1.AdmissionReview
			unmarshal(respRec.Body, &admissionReview)
			Expect(admissionReview.Response.Allowed).To(BeFalse(), "should not be allowed")
			Expect(admissionReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal("failed to unmarshal Request.OldObject into GreenplumCluster: json: cannot unmarshal array into Go value of type v1.GreenplumCluster"),
			})))
			Expect(DecodeLogs(logBuf)).To(ContainLogEntry(Keys{
				"msg":       Equal("/validate"),
				"GVK":       Equal("greenplum.pivotal.io/v1, Kind=GreenplumCluster"),
				"Name":      Equal("my-gp-instance"),
				"Namespace": Equal("test-ns"),
				"UID":       Equal("my-gp-instance-uid"),
				"Operation": Equal("UPDATE"),
				"Allowed":   BeFalse(),
				"Message":   Equal("failed to unmarshal Request.OldObject into GreenplumCluster: json: cannot unmarshal array into Go value of type v1.GreenplumCluster"),
			}))
		})
	})

	When("the request contains a new GreenplumPXFService that cannot be unmarshalled", func() {
		var respRec *httptest.ResponseRecorder
		BeforeEach(func() {
			reader := marshal(SampleAdmissionReviewRequest().OldObj(&examplePXF).NewObjInvalid("PXF").Build())
			req = httptest.NewRequest("POST", "/validate", reader)
			respRec = httptest.NewRecorder()
			resp = respRec
		})
		It("rejects the request", func() {
			var admissionReview admissionv1beta1.AdmissionReview
			unmarshal(respRec.Body, &admissionReview)
			Expect(admissionReview.Response.Allowed).To(BeFalse(), "should not be allowed")
			Expect(admissionReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal("failed to unmarshal Request.Object into GreenplumPXFService: json: cannot unmarshal array into Go value of type v1beta1.GreenplumPXFService"),
			})))
			Expect(DecodeLogs(logBuf)).To(ContainLogEntry(Keys{
				"msg":       Equal("/validate"),
				"GVK":       Equal("greenplum.pivotal.io/v1beta1, Kind=GreenplumPXFService"),
				"Name":      Equal("my-gp-pxf-instance"),
				"Namespace": Equal("test-ns"),
				"UID":       Equal("my-gp-pxf-instance-uid"),
				"Operation": Equal("UPDATE"),
				"Allowed":   BeFalse(),
				"Message":   Equal("failed to unmarshal Request.Object into GreenplumPXFService: json: cannot unmarshal array into Go value of type v1beta1.GreenplumPXFService"),
			}))
		})
	})
	When("the request contains an old GreenplumPXFService that cannot be unmarshalled", func() {
		var respRec *httptest.ResponseRecorder
		BeforeEach(func() {
			reader := marshal(SampleAdmissionReviewRequest().NewObj(&examplePXF).OldObjInvalid("PXF").Build())
			req = httptest.NewRequest("POST", "/validate", reader)
			respRec = httptest.NewRecorder()
			resp = respRec
		})
		It("rejects the request", func() {
			var admissionReview admissionv1beta1.AdmissionReview
			unmarshal(respRec.Body, &admissionReview)
			Expect(admissionReview.Response.Allowed).To(BeFalse(), "should not be allowed")
			Expect(admissionReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Message": Equal("failed to unmarshal Request.OldObject into GreenplumPXFService: json: cannot unmarshal array into Go value of type v1beta1.GreenplumPXFService"),
			})))
			Expect(DecodeLogs(logBuf)).To(ContainLogEntry(Keys{
				"msg":       Equal("/validate"),
				"GVK":       Equal("greenplum.pivotal.io/v1beta1, Kind=GreenplumPXFService"),
				"Name":      Equal("my-gp-pxf-instance"),
				"Namespace": Equal("test-ns"),
				"UID":       Equal("my-gp-pxf-instance-uid"),
				"Operation": Equal("UPDATE"),
				"Allowed":   BeFalse(),
				"Message":   Equal("failed to unmarshal Request.OldObject into GreenplumPXFService: json: cannot unmarshal array into Go value of type v1beta1.GreenplumPXFService"),
			}))
		})
	})

	for _, op := range []string{
		string(admissionv1beta1.Delete),
		string(admissionv1beta1.Connect),
		"BoGUs",
	} {
		operation := op
		When("we get an unsupported operation (like "+operation+") for a GreenplumCluster", func() {
			var respRec *httptest.ResponseRecorder
			BeforeEach(func() {
				reader := marshal(admissionv1beta1.AdmissionReview{
					TypeMeta: metav1.TypeMeta{},
					Request: &admissionv1beta1.AdmissionRequest{
						Kind:      metav1.GroupVersionKind{Group: "greenplum.pivotal.io", Version: "v1", Kind: "GreenplumCluster"},
						Name:      "my-gp-instance",
						UID:       "my-gp-instance-uid",
						Operation: admissionv1beta1.Operation(operation),
						Object:    runtime.RawExtension{Object: &exampleGreenplum},
					},
				})
				req = httptest.NewRequest("POST", "/validate", reader)
				respRec = httptest.NewRecorder()
				resp = respRec
			})
			It("rejects the request", func() {
				var admissionReview admissionv1beta1.AdmissionReview
				unmarshal(respRec.Body, &admissionReview)
				Expect(admissionReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(admissionReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal("unexpected operation for validation: " + operation),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainLogEntry(Keys{
					"msg":       Equal("/validate"),
					"GVK":       Equal("greenplum.pivotal.io/v1, Kind=GreenplumCluster"),
					"Name":      Equal("my-gp-instance"),
					"Namespace": BeEmpty(),
					"UID":       Equal("my-gp-instance-uid"),
					"Operation": Equal(operation),
					"Allowed":   BeFalse(),
					"Message":   Equal("unexpected operation for validation: " + operation),
				}))
			})
		})
	}

	for _, op := range []string{
		string(admissionv1beta1.Delete),
		string(admissionv1beta1.Connect),
		"BoGUs",
	} {
		operation := op
		When("we get an unsupported operation (like "+operation+") for a GreenplumPXFService", func() {
			var respRec *httptest.ResponseRecorder
			BeforeEach(func() {
				reader := marshal(admissionv1beta1.AdmissionReview{
					TypeMeta: metav1.TypeMeta{},
					Request: &admissionv1beta1.AdmissionRequest{
						Kind:      metav1.GroupVersionKind{Group: "greenplum.pivotal.io", Version: "v1beta1", Kind: "GreenplumPXFService"},
						Name:      "my-gp-pxf",
						UID:       "my-gp-pxf-uid",
						Operation: admissionv1beta1.Operation(operation),
						Object:    runtime.RawExtension{Object: &examplePXF},
					},
				})
				req = httptest.NewRequest("POST", "/validate", reader)
				respRec = httptest.NewRecorder()
				resp = respRec
			})
			It("rejects the request", func() {
				var admissionReview admissionv1beta1.AdmissionReview
				unmarshal(respRec.Body, &admissionReview)
				Expect(admissionReview.Response.Allowed).To(BeFalse(), "should not be allowed")
				Expect(admissionReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal("unexpected operation for validation: " + operation),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainLogEntry(Keys{
					"msg":       Equal("/validate"),
					"GVK":       Equal("greenplum.pivotal.io/v1beta1, Kind=GreenplumPXFService"),
					"Name":      Equal("my-gp-pxf"),
					"Namespace": BeEmpty(),
					"UID":       Equal("my-gp-pxf-uid"),
					"Operation": Equal(operation),
					"Allowed":   BeFalse(),
					"Message":   Equal("unexpected operation for validation: " + operation),
				}))
			})
		})
	}

})

var _ = Describe("HandleReady", func() {
	var (
		subject admission.Handler
		logBuf  *gbytes.Buffer
		resp    *http.Response
		err     error
	)
	BeforeEach(func() {
		subject = admission.Handler{}
		logBuf = gbytes.NewBuffer()
		admission.Log = gplog.ForTest(logBuf)
	})
	When("happy path", func() {
		BeforeEach(func() {
			srv := httptest.NewServer(subject.Handler())
			defer srv.Close()
			resp, err = http.Get(srv.URL + "/ready")
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns OK", func() {
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			buf, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(buf)).To(Equal("OK\n"))
		})
	})
	When("writing response fails", func() {
		BeforeEach(func() {
			req := httptest.NewRequest("GET", "/ready", nil)
			req.Header.Add("Content-Type", "text/plain")
			subject.HandleReady(&brokenResponseWriter{}, req)
		})
		It("logs an error", func() {
			Expect(logBuf).To(gbytes.Say(`"msg":"unable to reply to readiness probe","error":"injected error from broken response writer"`))
		})
	})
})

func marshal(from interface{}) *bytes.Reader {
	data, err := json.Marshal(from)
	Expect(err).NotTo(HaveOccurred())
	return bytes.NewReader(data)
}

func unmarshal(r io.Reader, to interface{}) {
	data, err := ioutil.ReadAll(r)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(data, to)).To(Succeed())
}

func corev1PodAdmissionReviewRequest() admissionv1beta1.AdmissionReview {
	invalidObj := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "core/v1",
		},
	}
	return admissionv1beta1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{},
		Request: &admissionv1beta1.AdmissionRequest{
			Kind:      metav1.GroupVersionKind{Group: "core", Version: "v1", Kind: "Pod"},
			Name:      "my-fake-instance",
			UID:       "my-fake-uid",
			Operation: admissionv1beta1.Create,
			Object:    runtime.RawExtension{Object: &invalidObj},
		},
	}
}

type brokenReader struct{}

var _ io.Reader = brokenReader{}

func (brokenReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("injected error from broken reader")
}

type brokenResponseWriter struct {
	httptest.ResponseRecorder
}

var _ http.ResponseWriter = &brokenResponseWriter{}

func (*brokenResponseWriter) Write(buf []byte) (int, error) {
	return 0, errors.New("injected error from broken response writer")
}
