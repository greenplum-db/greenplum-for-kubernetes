/*
.
*/

package v1

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GreenplumClusterSpec defines the desired state of GreenplumCluster
type GreenplumClusterSpec struct {
	MasterAndStandby GreenplumMasterAndStandbySpec `json:"masterAndStandby"`
	Segments         GreenplumSegmentsSpec         `json:"segments"`
	PXF              GreenplumPXFSpec              `json:"pxf,omitempty"`

	// YES or NO, specify whether or not to deploy a load-balancer
	// +kubebuilder:default=false
	LoadBalancer bool `json:"loadBalancer,omitempty"`

	// Optional Load Balancer Service Spec. if not provided, a default one will be used
	// +optional
	LoadBalancerServiceSpec *v1.ServiceSpec `json:"loadBalancerServiceSpec,omitempty"`
}

type GreenplumPodSpec struct {
	// Quantity expressed with an SI suffix, like 2Gi, 200m, 3.5, etc.
	Memory resource.Quantity `json:"memory,omitempty"`

	// Quantity expressed with an SI suffix, like 2Gi, 200m, 3.5, etc.
	CPU resource.Quantity `json:"cpu,omitempty"`

	// List of PersistentVolumeClaims
	PersistentVolumeClaims []PersistentVolumeClaim `json:"persistentVolumeClaims"`

	// A set of node labels for scheduling pods
	WorkerSelector map[string]string `json:"workerSelector,omitempty"`

	// YES or NO, specify whether or not to deploy with anti-affinity
	// +kubebuilder:default="no"
	// +kubebuilder:validation:Pattern=`^(?:yes|Yes|YES|no|No|NO|)$`
	AntiAffinity string `json:"antiAffinity,omitempty"`

	// Name of scheduler to be used for statefulset pods
	SchedulerName string `json:"schedulerName,omitempty"`

	// ImagePullSecret to be used for pods
	// +kubebuilder:default="regsecret"
	ImagePullSecret string `json:"imagePullSecret,omitempty"`

	// Optional PodSpec, for if you just want to get specific and manage all this yourself
	//+optional
	Spec v1.PodSpec `json:"spec,omitempty"`
}

type PersistentVolumeClaim struct {
	// Name of the PV (will be prepended with the Cluster Name)
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Name of storage class to use for statefulset PVs
	// +kubebuilder:validation:MinLength=1
	StorageClassName string `json:"storageClassName"`

	// Quantity expressed with an SI suffix, like 2Gi, 200m, 3.5, etc.
	Storage resource.Quantity `json:"storage"`

	// Path to mount on
	MountPath string `json:"mountPath"`
}

type GreenplumMasterAndStandbySpec struct {
	GreenplumPodSpec `json:",inline"`

	// Additional entries to add to pg_hba.conf
	HostBasedAuthentication string `json:"hostBasedAuthentication,omitempty"`

	// YES or NO, specify whether or not to deploy a standby master
	// +kubebuilder:default="no"
	// +kubebuilder:validation:Pattern=`^(?:yes|Yes|YES|no|No|NO|)$`
	Standby string `json:"standby,omitempty"`
}

type GreenplumSegmentsSpec struct {
	GreenplumPodSpec `json:",inline"`

	// Number of primary segments to create
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10000
	PrimarySegmentCount int32 `json:"primarySegmentCount"`

	// YES or NO, specify whether or not to deploy a PrimarySegmentCount number of mirror segments
	// +kubebuilder:default="no"
	// +kubebuilder:validation:Pattern=`^(?:yes|Yes|YES|no|No|NO|)$`
	Mirrors string `json:"mirrors,omitempty"`
}

type GreenplumPXFSpec struct {
	// Name of the PXF Service
	ServiceName string `json:"serviceName"`
}

type GreenplumClusterPhase string

const (
	GreenplumClusterPhasePending  GreenplumClusterPhase = "Pending"
	GreenplumClusterPhaseRunning  GreenplumClusterPhase = "Running"
	GreenplumClusterPhaseFailed   GreenplumClusterPhase = "Failed"
	GreenplumClusterPhaseDeleting GreenplumClusterPhase = "Deleting"
)

// GreenplumClusterStatus is the status for a GreenplumCluster resource
type GreenplumClusterStatus struct {
	InstanceImage   string                `json:"instanceImage,omitempty"`
	OperatorVersion string                `json:"operatorVersion,omitempty"`
	Phase           GreenplumClusterPhase `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`,description="The greenplum instance status"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="The greenplum instance age"
// +kubebuilder:resource:categories=all

// GreenplumCluster is the Schema for the greenplumclusters API
type GreenplumCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GreenplumClusterSpec   `json:"spec,omitempty"`
	Status GreenplumClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GreenplumClusterList contains a list of GreenplumCluster
type GreenplumClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GreenplumCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GreenplumCluster{}, &GreenplumClusterList{})
}
