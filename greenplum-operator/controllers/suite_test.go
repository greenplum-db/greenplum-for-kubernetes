/*
.
*/

package controllers

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/reactive"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	reactiveClient *reactive.Client
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KubeBuilder Controller Suite")
}

var _ = BeforeEach(func() {
	reactiveClient = reactive.NewClient(fake.NewFakeClientWithScheme(scheme.Scheme))
})
