package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
)

var _ = Describe("GetOldSegmentCount", func() {
	var (
		cmdFake *commandable.CommandFake
	)
	BeforeEach(func() {
		cmdFake = commandable.NewFakeCommand()
	})
	When("there are no errors", func() {
		BeforeEach(func() {
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-tAc",
				"SELECT COUNT(*) FROM gp_segment_configuration WHERE hostname LIKE 'segment-a%'",
			).PrintsOutput("1\n")
		})
		It("succeeds", func() {
			oldSegmentCount, err := GetOldSegmentCount(cmdFake.Command)
			Expect(err).NotTo(HaveOccurred())
			Expect(oldSegmentCount).To(Equal(1))
		})
	})

	When("querying segment count fails", func() {
		BeforeEach(func() {
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-tAc",
				"SELECT COUNT(*) FROM gp_segment_configuration WHERE hostname LIKE 'segment-a%'",
			).ReturnsStatus(1).PrintsError("custom get segment count error")
		})
		It("returns error", func() {
			_, err := GetOldSegmentCount(cmdFake.Command)
			Expect(err).To(MatchError("custom get segment count error: exit status 1"))
		})
	})
})
