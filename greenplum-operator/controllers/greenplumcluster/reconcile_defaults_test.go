package greenplumcluster_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers/greenplumcluster"
)

var _ = Describe("greenplumcluster.SetDefaultGreenplumClusterValues", func() {
	var fakeGreenplumCluster *greenplumv1.GreenplumCluster
	BeforeEach(func() {
		fakeGreenplumCluster = exampleGreenplumCluster.DeepCopy()
	})
	When("given a greenplumCluster with antiAffinity possibly containing uppercase characters", func() {
		yesAndNoes := []string{"Yes", "YES", "yes", "No", "NO", "no"}
		It("sets segments.antiAffinity to lowercase when given", func() {
			for _, value := range yesAndNoes {
				fakeGreenplumCluster.Spec.Segments.AntiAffinity = value
				greenplumcluster.SetDefaultGreenplumClusterValues(fakeGreenplumCluster)
				Expect(fakeGreenplumCluster.Spec.Segments.AntiAffinity).To(Equal(strings.ToLower(value)))
			}
		})
		It("sets masterAndStandby.antiAffinity to lowercase when given", func() {
			for _, value := range yesAndNoes {
				fakeGreenplumCluster.Spec.MasterAndStandby.AntiAffinity = value
				greenplumcluster.SetDefaultGreenplumClusterValues(fakeGreenplumCluster)
				Expect(fakeGreenplumCluster.Spec.MasterAndStandby.AntiAffinity).To(Equal(strings.ToLower(value)))
			}
		})
	})

	yesAndNoes := []string{"Yes", "YES", "yes", "No", "NO", "no"}
	When("given a greenplumCluster with mirrors possibly containing uppercase characters", func() {
		It("sets segments.mirrors to lowercase when given", func() {
			for _, value := range yesAndNoes {
				fakeGreenplumCluster.Spec.Segments.Mirrors = value
				greenplumcluster.SetDefaultGreenplumClusterValues(fakeGreenplumCluster)
				Expect(fakeGreenplumCluster.Spec.Segments.Mirrors).To(Equal(strings.ToLower(value)))
			}
		})
	})
	When("given a greenplumCluster with standby possibly containing uppercase characters", func() {
		It("sets masterAndStandby.standby to lowercase when given", func() {
			for _, value := range yesAndNoes {
				fakeGreenplumCluster.Spec.MasterAndStandby.Standby = value
				greenplumcluster.SetDefaultGreenplumClusterValues(fakeGreenplumCluster)
				Expect(fakeGreenplumCluster.Spec.MasterAndStandby.Standby).To(Equal(strings.ToLower(value)))
			}
		})
	})
})
