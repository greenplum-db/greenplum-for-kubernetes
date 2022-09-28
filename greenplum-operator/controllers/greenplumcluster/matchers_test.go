package greenplumcluster_test

import (
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

var beOwnedByGreenplum = gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
	"Name":       Equal("my-greenplum"),
	"Kind":       Equal("GreenplumCluster"),
	"Controller": gstruct.PointTo(BeTrue()),
})
