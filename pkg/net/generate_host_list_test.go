package net

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/gomega"
)

var _ = Describe("Generate host list", func() {
	table.DescribeTable("generate short host list",
		func(useStandby, useMirrors bool, segmentCount int, expectedHostList []string) {
			hostList := GenerateHostList(segmentCount, useMirrors, useStandby, "")
			gomega.Expect(hostList).To(gomega.ConsistOf(expectedHostList))
		},
		table.Entry("standby and mirrors", true, true, 2,
			[]string{"master-0", "master-1", "segment-a-0", "segment-a-1", "segment-b-0", "segment-b-1"}),
		table.Entry("standby and no mirrors", true, false, 2,
			[]string{"master-0", "master-1", "segment-a-0", "segment-a-1"}),
		table.Entry("no standby and mirrors", false, true, 2,
			[]string{"master-0", "segment-a-0", "segment-a-1", "segment-b-0", "segment-b-1"}),
		table.Entry("no standby and no mirrors", false, false, 2,
			[]string{"master-0", "segment-a-0", "segment-a-1"}),
	)

	table.DescribeTable("generate host list",
		func(useStandby, useMirrors bool, segmentCount int, expectedHostList []string) {
			hostList := GenerateHostList(segmentCount, useMirrors, useStandby, ".somedns")
			gomega.Expect(hostList).To(gomega.ConsistOf(expectedHostList))
		},
		table.Entry("standby and mirrors", true, true, 1,
			[]string{"master-0", "master-1", "segment-a-0", "segment-b-0", "master-0.somedns", "master-1.somedns", "segment-a-0.somedns", "segment-b-0.somedns"}),
		table.Entry("standby and no mirrors", true, false, 1,
			[]string{"master-0", "master-1", "segment-a-0", "master-0.somedns", "master-1.somedns", "segment-a-0.somedns"}),
		table.Entry("no standby and mirrors", false, true, 1,
			[]string{"master-0", "segment-a-0", "segment-b-0", "master-0.somedns", "segment-a-0.somedns", "segment-b-0.somedns"}),
		table.Entry("no standby and no mirrors", false, false, 1,
			[]string{"master-0", "segment-a-0", "master-0.somedns", "segment-a-0.somedns"}),
	)
})
