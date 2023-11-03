package admission

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type Handler struct {
	KubeClient     client.Client
	InstanceImage  string
	RestClient     rest.Interface
	PodCmdExecutor executor.PodExecInterface
	KubeClientSet  *kubernetes.Clientset
}

func (h *Handler) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ready", h.HandleReady)
	mux.HandleFunc("/validate", h.HandleValidate)
	return mux
}

func (h *Handler) HandleReady(out http.ResponseWriter, req *http.Request) {
	_, err := out.Write([]byte("OK\n"))
	if err != nil {
		Log.Error(err, "unable to reply to readiness probe")
	}
}

func (h *Handler) HandleValidate(out http.ResponseWriter, req *http.Request) {
	log := Log
	defer func() { log.Info("/validate") }()

	ctx := req.Context()

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(out, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}
	reqBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(out, "couldn't read request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	var reviewResponse admissionv1.AdmissionReview
	reviewResponse.Response = func() (response *admissionv1.AdmissionResponse) {
		var reviewRequest admissionv1.AdmissionReview
		response = &admissionv1.AdmissionResponse{}
		if err := json.Unmarshal(reqBytes, &reviewRequest); err != nil {
			response.Result = &metav1.Status{Message: "parsing request: " + err.Error()}
			return
		}

		log = log.WithValues(
			"GVK", reviewRequest.Request.Kind.String(),
			"Name", reviewRequest.Request.Name,
			"Namespace", reviewRequest.Request.Namespace,
			"UID", reviewRequest.Request.UID,
			"Operation", reviewRequest.Request.Operation)

		response.UID = reviewRequest.Request.UID

		reqGVK := schema.GroupVersionKind{
			Group: reviewRequest.Request.Kind.Group,
			Version: reviewRequest.Request.Kind.Version,
			Kind: reviewRequest.Request.Kind.Kind,
		}

		switch reqGVK {
		case greenplumv1.GroupVersion.WithKind("GreenplumCluster"):
			var oldGreenplum, newGreenplum greenplumv1.GreenplumCluster
			if err := json.Unmarshal(reviewRequest.Request.Object.Raw, &newGreenplum); err != nil {
				response.Result = &metav1.Status{Message: "failed to unmarshal Request.Object into GreenplumCluster: " + err.Error()}
				return
			}
			op := reviewRequest.Request.Operation
			switch op {
			case admissionv1.Create:
				response.Allowed, response.Result = h.validateCreateGreenplumCluster(ctx, newGreenplum)
			case admissionv1.Update:
				if err := json.Unmarshal(reviewRequest.Request.OldObject.Raw, &oldGreenplum); err != nil {
					response.Result = &metav1.Status{Message: "failed to unmarshal Request.OldObject into GreenplumCluster: " + err.Error()}
					return
				}
				response.Allowed, response.Result = h.validateUpdateGreenplumCluster(ctx, oldGreenplum, newGreenplum)
			default:
				response.Allowed = false
				response.Result = &metav1.Status{Message: "unexpected operation for validation: " + string(op)}
			}
		case greenplumv1.GroupVersion.WithKind("GreenplumPXFService"):
			op := reviewRequest.Request.Operation
			var oldPXF, newPXF greenplumv1.GreenplumPXFService
			if err := json.Unmarshal(reviewRequest.Request.Object.Raw, &newPXF); err != nil {
				response.Result = &metav1.Status{Message: "failed to unmarshal Request.Object into GreenplumPXFService: " + err.Error()}
				return
			}
			switch op {
			case admissionv1.Create:
				response.Allowed, response.Result = h.validateGreenplumPXFService(ctx, nil, &newPXF)
			case admissionv1.Update:
				if err := json.Unmarshal(reviewRequest.Request.OldObject.Raw, &oldPXF); err != nil {
					response.Result = &metav1.Status{Message: "failed to unmarshal Request.OldObject into GreenplumPXFService: " + err.Error()}
					return
				}
				response.Allowed, response.Result = h.validateGreenplumPXFService(ctx, &oldPXF, &newPXF)
			default:
				response.Allowed = false
				response.Result = &metav1.Status{Message: "unexpected operation for validation: " + string(op)}
			}
		default:
			response.Allowed = false
			response.Result = &metav1.Status{Message: "unexpected validation request for object: " + reviewRequest.Request.Kind.String()}
		}

		return
	}()

	log = log.WithValues("Allowed", reviewResponse.Response.Allowed)
	if reviewResponse.Response.Result != nil && reviewResponse.Response.Result.Message != "" {
		log = log.WithValues("Message", reviewResponse.Response.Result.Message)
	}

	outBytes, _ := json.Marshal(reviewResponse)
	_, err = out.Write(outBytes)
	if err != nil {
		Log.Error(err, "responding to admission review")
	}
}
