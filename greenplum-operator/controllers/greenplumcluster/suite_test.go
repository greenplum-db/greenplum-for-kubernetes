/*
.
*/

package greenplumcluster_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	fakePodExec "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor/fake"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/reactive"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	reactiveClient *reactive.Client
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KubeBuilder Controller Suite (GreenplumCluster)")
}

var _ = BeforeEach(func() {
	reactiveClient = reactive.NewClient(fake.NewFakeClientWithScheme(scheme.Scheme))
})

const (
	namespaceName = "test-ns"
	clusterName   = "my-greenplum"
)

var exampleGreenplumCluster = &greenplumv1.GreenplumCluster{
	ObjectMeta: metav1.ObjectMeta{
		Name:      clusterName,
		Namespace: namespaceName,
	},
	Spec: greenplumv1.GreenplumClusterSpec{
		MasterAndStandby: greenplumv1.GreenplumMasterAndStandbySpec{
			GreenplumPodSpec: greenplumv1.GreenplumPodSpec{
				Memory:           resource.MustParse("1G"),
				CPU:              resource.MustParse("0.5"),
				StorageClassName: "standard",
				Storage:          resource.MustParse("1G"),
			},
		},
		Segments: greenplumv1.GreenplumSegmentsSpec{
			GreenplumPodSpec: greenplumv1.GreenplumPodSpec{
				Memory:           resource.MustParse("1G"),
				CPU:              resource.MustParse("0.5"),
				StorageClassName: "standard",
				Storage:          resource.MustParse("1G"),
			},
			PrimarySegmentCount: fakePodExec.DefaultSegmentCount,
		},
	},
}

var greenplumClusterRequest = reconcile.Request{
	NamespacedName: types.NamespacedName{
		Namespace: namespaceName,
		Name:      clusterName,
	},
}
