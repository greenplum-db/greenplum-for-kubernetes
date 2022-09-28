package executor

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	fakeExecutor "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor/fake"
)

var _ = Context("GetCurrentActiveMaster", func() {
	It("returns master-0 when master-0 is active master", func() {
		fakePodCommandExecutor := &fakeExecutor.PodExec{}
		activeMaster := GetCurrentActiveMaster(fakePodCommandExecutor, "testNamespace")
		Expect(activeMaster).To(Equal("master-0"))
	})
	It("returns master-1 when master-1 is active master", func() {
		fakePodCommandExecutor := &fakeExecutor.PodExec{
			ErrorMsgOnMaster0: "postgres not running on port 5432",
		}
		activeMaster := GetCurrentActiveMaster(fakePodCommandExecutor, "testNamespace")
		Expect(activeMaster).To(Equal("master-1"))
	})
	It("returns '' when neither master-0 nor master-1 is active", func() {
		fakePodCommandExecutor := &fakeExecutor.PodExec{
			ErrorMsgOnMaster0: "postgres not running on port 5432",
			ErrorMsgOnMaster1: "postgres not running on port 5432",
		}
		activeMaster := GetCurrentActiveMaster(fakePodCommandExecutor, "testNamespace")
		Expect(activeMaster).To(Equal(""))
	})
})
