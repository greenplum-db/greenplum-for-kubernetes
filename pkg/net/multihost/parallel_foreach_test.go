package multihost_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost"
	fakemultihost "github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost/testing"
)

var _ = Describe("ParallelForeach", func() {
	var (
		fakeOperation *fakemultihost.FakeOperation
	)

	BeforeEach(func() {
		fakeOperation = &fakemultihost.FakeOperation{}
	})
	DescribeTable("it executes operations for given hostname lists",
		func(useStandby, useMirrors bool, newSegmentCount int, expectedHostList []string) {
			Expect(multihost.ParallelForeach(fakeOperation, expectedHostList)).To(HaveLen(0))
			Expect(fakeOperation.HostRecords).To(HaveLen(len(expectedHostList)))
			Expect(fakeOperation.HostRecords).To(ContainElements(expectedHostList))
		},
		Entry("standby and mirrors", true, true, 2,
			[]string{"master-0", "master-1", "segment-a-0", "segment-a-1", "segment-b-0", "segment-b-1"}),
		Entry("standby and no mirrors", true, false, 2,
			[]string{"master-0", "master-1", "segment-a-0", "segment-a-1"}),
		Entry("no standby and mirrors", false, true, 2,
			[]string{"master-0", "segment-a-0", "segment-a-1", "segment-b-0", "segment-b-1"}),
		Entry("no standby and no mirrors", false, false, 2,
			[]string{"master-0", "segment-a-0", "segment-a-1"}),
	)
	When("an operation fails", func() {
		It("returns the error from the failed operation", func() {
			hostnameList := []string{
				"master-0",
				"segment-a-0",
				"segment-a-1",
			}
			fakeOperation.FakeErrors = map[string]error{
				"segment-a-1": errors.New("fake error"),
			}
			errs := multihost.ParallelForeach(fakeOperation, hostnameList)
			Expect(errs).To(HaveLen(1))
			Expect(errs[0]).To(MatchError("fake error"))
		})
	})
	When("multiple operations fail", func() {
		It("returns the errors from the failed operations", func() {
			hostnameList := []string{
				"master-0",
				"segment-a-0",
				"segment-a-1",
			}
			fakeOperation.FakeErrors = map[string]error{
				"segment-a-0": errors.New("fake error 1"),
				"segment-a-1": errors.New("fake error 2"),
			}
			errs := multihost.ParallelForeach(fakeOperation, hostnameList)
			Expect(errs).To(HaveLen(2))
			Expect(errs).To(ContainElements([]error{
				errors.New("fake error 1"),
				errors.New("fake error 2"),
			}))
		})
	})
})
