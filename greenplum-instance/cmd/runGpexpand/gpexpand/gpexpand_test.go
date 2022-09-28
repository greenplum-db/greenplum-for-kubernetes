package gpexpand

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gstruct"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/gplog/testing"
	fakemultihost "github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost/testing"
)

var _ = Describe("RunGpexpand", func() {
	var (
		subject              *RunGpexpandConfig
		logBuf               *gbytes.Buffer
		expectedHosts        []string
		stdout               *gbytes.Buffer
		stderr               *gbytes.Buffer
		fakeDNSResolver      *fakemultihost.FakeOperation
		fakeKnownHostsWaiter *fakemultihost.FakeOperation
		fakeSSHExecutor      *fakemultihost.FakeOperation
		cmdFake              *commandable.CommandFake
		gpexpandCalled       int
	)
	BeforeEach(func() {
		expectedHosts = []string{
			"master-0",
			"master-1",
			"segment-a-0",
			"segment-a-1",
			"segment-b-0",
			"segment-b-1",
		}
		cmdFake = commandable.NewFakeCommand()
		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()
		fakeDNSResolver = &fakemultihost.FakeOperation{}
		fakeKnownHostsWaiter = &fakemultihost.FakeOperation{}
		fakeSSHExecutor = &fakemultihost.FakeOperation{}
		logBuf = gbytes.NewBuffer()
		subject = &RunGpexpandConfig{
			Log:              gplog.ForTest(logBuf),
			NewSegmentCount:  2,
			IsMirrored:       true,
			Standby:          true,
			Stdout:           stdout,
			Stderr:           stderr,
			DNSResolver:      fakeDNSResolver,
			KnownHostsWaiter: fakeKnownHostsWaiter,
			SSHExecutor:      fakeSSHExecutor,
			Command:          cmdFake.Command,
		}
		gpexpandCalled = 0
		cmdFake.ExpectCommand("bash", "-c",
			"source /usr/local/greenplum-db/greenplum_path.sh && MASTER_DATA_DIRECTORY=/greenplum/data-1 gpexpand -i /tmp/gpexpand_config").
			CallCounter(&gpexpandCalled).
			PrintsOutput("gpexpand successful")
	})

	It("resolves DNS entries for every pod in cluster", func() {
		Expect(subject.Run()).To(Succeed())
		Expect(DecodeLogs(logBuf)).To(ContainLogEntry(gstruct.Keys{
			"msg": Equal("resolving DNS entries for all masters and segments"),
		}))
		Expect(fakeDNSResolver.HostRecords).To(ConsistOf(expectedHosts))
	})

	It("waits for known_hosts on master", func() {
		Expect(subject.Run()).To(Succeed())
		Expect(DecodeLogs(logBuf)).To(ContainLogEntry(gstruct.Keys{
			"msg": Equal("waiting for known_hosts file to be populated on master"),
		}))
		Expect(fakeKnownHostsWaiter.HostRecords).To(ConsistOf(expectedHosts))
	})

	It("waits for known_hosts on every pod in cluster", func() {
		Expect(subject.Run()).To(Succeed())
		Expect(DecodeLogs(logBuf)).To(ContainLogEntry(gstruct.Keys{
			"msg": Equal("waiting for known_hosts file to be populated on all masters and segments"),
		}))
		Expect(fakeSSHExecutor.HostRecords).To(ConsistOf(expectedHosts))
	})

	It("runs gpexpand", func() {
		Expect(subject.Run()).To(Succeed())
		Expect(DecodeLogs(logBuf)).To(ContainLogEntry(gstruct.Keys{
			"msg": Equal("running gpexpand"),
		}))
		Expect(gpexpandCalled).To(Equal(1))
		Expect(stdout).To(gbytes.Say("gpexpand successful"))
	})

	When("dns resolver fails", func() {
		BeforeEach(func() {
			fakeDNSResolver.FakeErrors = map[string]error{
				"segment-b-0": errors.New("injected error"),
			}
		})
		It("returns an error", func() {
			Expect(subject.Run()).To(MatchError("failed to resolve DNS entries for all masters and segments"))
			Expect(fakeDNSResolver.HostRecords).To(ConsistOf(expectedHosts))
			Expect(gpexpandCalled).To(Equal(0))
		})
	})

	When("known_hosts waiter fails", func() {
		BeforeEach(func() {
			fakeKnownHostsWaiter.FakeErrors = map[string]error{
				"segment-b-1": errors.New("injected error"),
			}
		})
		It("returns an error", func() {
			Expect(subject.Run()).To(MatchError("timed out waiting for known_hosts on master"))
			Expect(fakeKnownHostsWaiter.HostRecords).To(ConsistOf(expectedHosts))
			Expect(gpexpandCalled).To(Equal(0))
		})
	})

	When("ssh multihost exec waitForKnownHosts fails", func() {
		BeforeEach(func() {
			fakeSSHExecutor.FakeErrors = map[string]error{
				"segment-a-1": errors.New("injected error"),
			}
		})
		It("returns an error", func() {
			Expect(subject.Run()).To(MatchError("timed out waiting for known_hosts on all masters and segments"))
			Expect(fakeSSHExecutor.HostRecords).To(ConsistOf(expectedHosts))
			Expect(gpexpandCalled).To(Equal(0))
		})
	})

	When("gpexpand command fails", func() {
		BeforeEach(func() {
			cmdFake.ExpectCommand("bash", "-c",
				"source /usr/local/greenplum-db/greenplum_path.sh && MASTER_DATA_DIRECTORY=/greenplum/data-1 gpexpand -i /tmp/gpexpand_config").
				ReturnsStatus(1).
				PrintsError("gpexpand failed")
		})
		It("returns error", func() {
			Expect(subject.Run()).NotTo(Succeed())
			Expect(stderr).To(gbytes.Say("gpexpand failed"))
		})
	})
})
