package main

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGreenplumClusterOperator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "greenplumOperator Test Suite")
}
