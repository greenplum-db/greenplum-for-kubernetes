package startContainerUtils_test

import (
	"context"
	"strconv"
	"time"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gstruct"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog/testing"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
)

const (
	LocalSSHDirPath  = "/home/gpadmin/.ssh"
	LocalHostKeyPath = "/etc/ssh"
)

var _ = Describe("SSHDaemon", func() {
	var (
		app         *startContainerUtils.SSHDaemon
		errorBuffer *gbytes.Buffer
		outBuffer   *gbytes.Buffer
		memoryfs    vfs.Filesystem
		fakeCmd     *commandable.CommandFake
	)
	BeforeEach(func() {
		fakeCmd = commandable.NewFakeCommand()
		errorBuffer = gbytes.NewBuffer()
		outBuffer = gbytes.NewBuffer()
		startContainerUtils.Log = gplog.ForTest(outBuffer)
		memoryfs = memfs.Create()
		app = &startContainerUtils.SSHDaemon{
			App: &starter.App{
				Command:      fakeCmd.Command,
				StdoutBuffer: outBuffer,
				StderrBuffer: errorBuffer,
				Fs:           memoryfs,
			}}

		// simulate that mount has already created /greenplum dir. Unfortunately, we cannot simulate it is owned by root
		Expect(vfs.MkdirAll(memoryfs, "/greenplum", 0755)).To(Succeed())
	})
	Describe("on Run()", func() {
		It("starts SSH Daemon in foreground", func() {
			Expect(app.Run(context.TODO()).Error()).To(ContainSubstring("sshd is not running"))
			Expect(outBuffer).To(gbytes.Say(`"starting SSH Daemon"`))
			Expect(fakeCmd.CapturedArgs()).To(Equal("/usr/bin/sudo /usr/sbin/sshd -D"))
		})
		When("sshd terminates uncleanly", func() {
			var err error
			BeforeEach(func() {
				fakeCmd.ExpectCommand("/usr/bin/sudo", "/usr/sbin/sshd", "-D").ReturnsStatus(1)
				err = app.Run(context.TODO())
			})
			It("logs the return code", func() {
				Expect(testing.DecodeLogs(outBuffer)).To(testing.ContainLogEntry(gstruct.Keys{
					"level": Equal("ERROR"),
					"msg":   Equal("sshd process terminated"),
					"error": Equal("exit status 1"),
				}))
			})
			It("returns an error", func() {
				Expect(err).To(MatchError("sshd is not running: exit status 1"))
			})
		})
		When("stopChan is closed", func() {
			var (
				sshdCallCount int
				pidList       []int
				sshWaitCh     chan struct{}
			)
			BeforeEach(func() {
				sshWaitCh = make(chan struct{})
				fakeCmd.ExpectCommand("/usr/bin/sudo", "/usr/sbin/sshd", "-D").CallCounter(&sshdCallCount).PidList(&pidList).SideEffect(func() {
					<-sshWaitCh
				})
				fakeCmd.ExpectCommandMatching(func(path string, args ...string) bool {
					if path == "/usr/bin/sudo" {
						if args[0] == "/bin/kill" && args[1] == "-SIGKILL" {
							close(sshWaitCh)
							return true
						}
					}
					return false
				})
			})
			It("kills sshd", func() {
				ctx, cancel := context.WithCancel(context.Background())
				resultChan := make(chan error)
				go func() {
					resultChan <- app.Run(ctx)
				}()
				cancel()
				Eventually(resultChan, 3*time.Second).Should(Receive(nil))
				Expect(sshdCallCount).To(Equal(1))
				Expect(testing.DecodeLogs(outBuffer)).To(testing.ContainLogEntry(gstruct.Keys{
					"level": Equal("INFO"),
					"msg":   Equal("killing sshd"),
					"pid":   Equal(float64(pidList[0])),
				}))
			})
		})
		When("sudo kill sshd fails", func() {
			var (
				sshdCallCount int
				sshWaitCh     chan struct{}
				pidCh         chan int
			)
			BeforeEach(func() {
				sshWaitCh = make(chan struct{})
				pidCh = make(chan int)
				fakeCmd.ExpectCommand("/usr/bin/sudo", "/usr/sbin/sshd", "-D").CallCounter(&sshdCallCount).SideEffect(func() {
					<-sshWaitCh
				})
				fakeCmd.ExpectCommandMatching(func(path string, args ...string) bool {
					if path == "/usr/bin/sudo" {
						if args[0] == "/bin/kill" && args[1] == "-SIGKILL" {
							close(sshWaitCh)
							pidInt, err := strconv.Atoi(args[2])
							Expect(err).NotTo(HaveOccurred())
							pidCh <- pidInt
							return true
						}
					}
					return false
				}).PrintsOutput("error message").ReturnsStatus(1)
			})
			It("logs an error", func() {
				ctx, cancel := context.WithCancel(context.Background())
				resultChan := make(chan error)
				go func() {
					resultChan <- app.Run(ctx)
				}()
				cancel()
				expectedPid := <-pidCh
				Eventually(resultChan, 3*time.Second).Should(Receive(nil))
				decodedLogs, err := testing.DecodeLogs(outBuffer)
				Expect(err).NotTo(HaveOccurred())
				Expect(decodedLogs).To(testing.ContainLogEntry(gstruct.Keys{
					"level": Equal("INFO"),
					"msg":   Equal("killing sshd"),
					"pid":   Equal(float64(expectedPid)),
				}))
				Expect(decodedLogs).To(testing.ContainLogEntry(gstruct.Keys{
					"level":  Equal("ERROR"),
					"error":  Equal("exit status 1"),
					"msg":    Equal("failed to kill sshd"),
					"output": Equal("error message"),
				}))
			})
		})
	})
})
