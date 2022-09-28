/*
.
*/

package v1beta1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const PXFAppName = "greenplum-pxf"

// GreenplumPXFServiceSpec defines the desired state of GreenplumPXFService
type GreenplumPXFServiceSpec struct {
	// Number of pods to create
	// +kubebuilder:default=2
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000
	Replicas int32 `json:"replicas,omitempty"`

	// Quantity expressed with an SI suffix, like 2Gi, 200m, 3.5, etc.
	CPU resource.Quantity `json:"cpu,omitempty"`

	// Quantity expressed with an SI suffix, like 2Gi, 200m, 3.5, etc.
	Memory resource.Quantity `json:"memory,omitempty"` // TODO: limit to 31Gi

	// A set of node labels for scheduling pods
	WorkerSelector map[string]string `json:"workerSelector,omitempty"`

	// S3 Bucket and Secret for downloading PXF configs
	PXFConf *GreenplumPXFConf `json:"pxfConf,omitempty"`
}

type GreenplumPXFServicePhase string

const (
	GreenplumPXFServicePhasePending  GreenplumPXFServicePhase = "Pending"
	GreenplumPXFServicePhaseDegraded GreenplumPXFServicePhase = "Degraded"
	GreenplumPXFServicePhaseRunning  GreenplumPXFServicePhase = "Running"
)

// GreenplumPXFServiceStatus defines the observed state of GreenplumPXFService
type GreenplumPXFServiceStatus struct {
	Phase GreenplumPXFServicePhase `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`,description="The greenplum pxf service status"
// +kubebuilder:resource:categories=all

// GreenplumPXFService is the Schema for the greenplumpxfservices API
type GreenplumPXFService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GreenplumPXFServiceSpec   `json:"spec,omitempty"`
	Status GreenplumPXFServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GreenplumPXFServiceList contains a list of GreenplumPXFService
type GreenplumPXFServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GreenplumPXFService `json:"items"`
}

type GreenplumPXFConf struct {
	// +kubebuilder:validation:Required
	S3Source S3Source `json:"s3Source"`
}

type S3Source struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Secret string `json:"secret"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Bucket string `json:"bucket"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	EndPoint string `json:"endpoint"`

	// +kubebuilder:validation:Enum=http;https
	Protocol string `json:"protocol,omitempty"`

	// +kubebuilder:validation:MinLength=1
	Folder string `json:"folder,omitempty"`
}

func init() {
	SchemeBuilder.Register(&GreenplumPXFService{}, &GreenplumPXFServiceList{})
}
