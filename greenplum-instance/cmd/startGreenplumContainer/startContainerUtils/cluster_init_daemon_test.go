package startContainerUtils_test

import (
	"encoding/base64"
	"errors"
	"os"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils/cluster"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	instanceconfigTesting "github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig/testing"
	fakemultihost "github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost/testing"
	fakessh "github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/fake"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
	ubuntuTesting "github.com/pivotal/greenplum-for-kubernetes/pkg/ubuntuUtils/testing"
)

var _ = Describe("ClusterInitDaemon", func() {

	var (
		app              *startContainerUtils.ClusterInitDaemon
		errorBuffer      *gbytes.Buffer
		outBuffer        *gbytes.Buffer
		memoryfs         *fileutil.HookableFilesystem
		fakeCmd          *commandable.CommandFake
		mockConfig       *instanceconfigTesting.MockReader
		mockUbuntu       ubuntuTesting.MockUbuntu
		c                *fakeCluster
		fakeDNSResolver  *fakemultihost.FakeOperation
		keyScanner       *fakessh.KeyScanner
		knownHostsReader *fakessh.KnownHostsReader
	)

	BeforeEach(func() {
		fakeCmd = commandable.NewFakeCommand()
		errorBuffer = gbytes.NewBuffer()
		outBuffer = gbytes.NewBuffer()
		startContainerUtils.Log = gplog.ForTest(outBuffer)
		memoryfs = &fileutil.HookableFilesystem{Filesystem: memfs.Create()}
		mockConfig = &instanceconfigTesting.MockReader{
			SegmentCount: 1,
			Mirrors:      true,
			Standby:      true,
		}
		mockUbuntu = ubuntuTesting.MockUbuntu{}
		mockUbuntu.HostnameMock.Hostname = "master-0"
		c = &fakeCluster{}

		fakeDNSResolver = &fakemultihost.FakeOperation{}
		keyScanner = &fakessh.KeyScanner{}
		knownHostsReader = &fakessh.KnownHostsReader{}
		app = &startContainerUtils.ClusterInitDaemon{
			App: &starter.App{
				Command:      fakeCmd.Command,
				StdoutBuffer: outBuffer,
				StderrBuffer: errorBuffer,
				Fs:           memoryfs,
			},
			Config:           mockConfig,
			Ubuntu:           &mockUbuntu,
			DNSResolver:      fakeDNSResolver,
			KeyScanner:       keyScanner,
			KnownHostsReader: knownHostsReader,
			C:                c,
		}
	})

	Describe("InitializeCluster", func() {
		When("obtaining the container hostname fails", func() {
			BeforeEach(func() {
				mockUbuntu.HostnameMock.Hostname = "blah"
				mockUbuntu.HostnameMock.Err = errors.New("hostname error")
			})
			It("logs an error message", func() {
				app.InitializeCluster()
				Expect(outBuffer).To(gbytes.Say("hostname error"))
			})

		})

		ShouldRunKeyScanner := func() {
			app.InitializeCluster()
			Expect(keyScanner.WasCalled).To(BeTrue(), "expected keyScanner to be run")
		}

		ShouldNotGpinit := func() {
			app.InitializeCluster()
			Expect(c.initializeStub.wasCalled).To(BeFalse())
		}

		ShouldRunPostInitialization := func() {
			app.InitializeCluster()
			Expect(c.runPostInitializationStub.wasCalled).To(BeTrue(), "should run post-initialization")
		}

		ShouldNotRunPostInitialization := func() {
			app.InitializeCluster()
			Expect(c.runPostInitializationStub.wasCalled).To(BeFalse(), "should not run post-initialization")
		}

		ShouldRunPgCtl := func(pgctlArgs []string) func() {
			return func() {
				pgCtlCalled := 0
				fakeCmd.ExpectCommand("/usr/local/greenplum-db/bin/pg_ctl", pgctlArgs...).CallCounter(&pgCtlCalled)

				app.InitializeCluster()
				Expect(pgCtlCalled).To(Equal(1))
			}
		}

		ShouldNotRunPgCtl := func() {
			pgCtlCalled := 0
			fakeCmd.ExpectCommandMatching(func(path string, args ...string) bool {
				return path == "/usr/local/greenplum-db/bin/pg_ctl"
			}).CallCounter(&pgCtlCalled)

			app.InitializeCluster()

			Expect(pgCtlCalled).To(Equal(0))
		}

		ShouldFailOnPgCtlError := func() {
			Expect(app.InitializeCluster()).To(MatchError("pg_ctl failed to restart"))
			Expect(outBuffer).To(gbytes.Say("cluster has been initialized before; starting Postgres"))
			Expect(errorBuffer).To(gbytes.Say("custom pg_ctl error"))
		}

		masterPgctlArgs := []string{"-D", "/greenplum/data-1", "-l", "/greenplum/data-1/pg_log/startup.log", "restart"}
		segmentPgctlArgs := []string{"-D", "/greenplum/data", "-l", "/greenplum/data/pg_log/startup.log", "restart"}

		When("hostname is master-0", func() {
			BeforeEach(func() {
				mockUbuntu.HostnameMock.Hostname = "master-0"
			})
			When("/greenplum/data-1 directory exists", func() {
				BeforeEach(func() {
					Expect(vfs.MkdirAll(memoryfs, "/greenplum/data-1", 0755)).To(Succeed())
				})

				It("succeeds", func() {
					exitErr := app.InitializeCluster()
					Expect(exitErr).NotTo(HaveOccurred())
				})

				It("should not attempt to initialize the cluster", ShouldNotGpinit)

				It("should run keyscanner", ShouldRunKeyScanner)

				When("standby is off", func() {
					BeforeEach(func() {
						mockConfig.Standby = false
					})

					It("should run keyscanner", ShouldRunKeyScanner)

					It("GpStart gets invoked", func() {
						exitErr := app.InitializeCluster()
						Expect(outBuffer).To(gbytes.Say("cluster has been initialized before; starting Greenplum Cluster"))
						Expect(c.gpstartStub.wasCalled).To(BeTrue())
						Expect(exitErr).NotTo(HaveOccurred())
					})

					It("should run post initialization", ShouldRunPostInitialization)

					When("gpstart fails", func() {
						BeforeEach(func() {
							c.gpstartStub.err = errors.New("gpstart error")
						})

						It("returns an error", func() {
							exitErr := app.InitializeCluster()
							Expect(c.gpstartStub.wasCalled).To(BeTrue())
							Expect(exitErr).To(MatchError("gpstart error"))
						})

						It("should not run post initialization", ShouldNotRunPostInitialization)
					})
				})
				When("standby is ON", func() {
					BeforeEach(func() {
						mockConfig.Standby = true
					})

					It("does not invoke gpstart", func() {
						exitErr := app.InitializeCluster()
						Expect(c.gpstartStub.wasCalled).To(BeFalse())
						Expect(exitErr).NotTo(HaveOccurred())
						Expect(outBuffer).To(gbytes.Say("Automatic gpstart is not currently supported with standby masters. Skipping."))
					})

					It("should run post initialization", ShouldRunPostInitialization)
				})
			})

			When("/greenplum/data-1 directory does not exist", func() {
				It("succeeds", func() {
					exitErr := app.InitializeCluster()
					Expect(exitErr).NotTo(HaveOccurred())
				})

				It("should run keyscanner", ShouldRunKeyScanner)

				It("should attempt to initialize the cluster", func() {
					app.InitializeCluster()

					Expect(c.initializeStub.wasCalled).To(BeTrue(), "should call initialize")
				})

				It("should run post initialization", ShouldRunPostInitialization)

				It("does not run pg_ctl", ShouldNotRunPgCtl)

				When("an error occurs while initializing", func() {
					BeforeEach(func() {
						c.initializeStub.err = errors.New("gpinitsystem failed")
					})
					It("should return the error", func() {
						exitErr := app.InitializeCluster()
						Expect(c.initializeStub.wasCalled).To(BeTrue(), "should call initialize")
						Expect(exitErr).To(MatchError("gpinitsystem failed"))
					})

					It("should not run post initialization", ShouldNotRunPostInitialization)
				})
			})
		})

		When("hostname is master-1", func() {
			BeforeEach(func() {
				mockUbuntu.HostnameMock.Hostname = "master-1"
			})
			When("/greenplum/data-1 exists", func() {
				BeforeEach(func() {
					Expect(vfs.MkdirAll(memoryfs, "/greenplum/data-1", 0755)).To(Succeed())
				})

				It("succeeds", func() {
					exitErr := app.InitializeCluster()
					Expect(exitErr).NotTo(HaveOccurred())
				})

				It("should not attempt to initialize the cluster", ShouldNotGpinit)

				It("logs that cluster has been initialized", func() {
					app.InitializeCluster()
					Expect(outBuffer).To(gbytes.Say("cluster has been initialized before; starting Postgres"))
				})

				It("should not run post initialization", ShouldNotRunPostInitialization)

				It("runs pg_ctl to start the postgres process", ShouldRunPgCtl(masterPgctlArgs))

				When("pg_ctl command fails", func() {
					BeforeEach(func() {
						fakeCmd.ExpectCommand("/usr/local/greenplum-db/bin/pg_ctl", masterPgctlArgs...).ReturnsStatus(1).PrintsError("custom pg_ctl error")
					})
					It("logs error", ShouldFailOnPgCtlError)
				})
			})
			When("/greenplum/data-1 does not exist", func() {
				It("succeeds", func() {
					exitErr := app.InitializeCluster()
					Expect(exitErr).NotTo(HaveOccurred())
				})

				It("should not attempt to initialize the cluster", ShouldNotGpinit)

				It("should not run post initialization", ShouldNotRunPostInitialization)

				It("does not run pg_ctl", ShouldNotRunPgCtl)
			})
		})

		When("hostname is segment-b-42", func() {
			BeforeEach(func() {
				mockUbuntu.HostnameMock.Hostname = "segment-b-42"
			})
			When("/greenplum/data exists", func() {
				BeforeEach(func() {
					Expect(vfs.MkdirAll(memoryfs, "/greenplum/data", 0755)).To(Succeed())
				})

				It("succeeds", func() {
					exitErr := app.InitializeCluster()
					Expect(exitErr).NotTo(HaveOccurred())
				})

				It("should not attempt to initialize the cluster", ShouldNotGpinit)

				It("should not run post initialization", ShouldNotRunPostInitialization)

				It("logs that cluster has been initialized", func() {
					app.InitializeCluster()
					Expect(outBuffer).To(gbytes.Say("cluster has been initialized before; starting Postgres"))
				})

				It("runs pg_ctl to start the postgres process", ShouldRunPgCtl(segmentPgctlArgs))

				When("pg_ctl command fails", func() {
					BeforeEach(func() {
						fakeCmd.ExpectCommand("/usr/local/greenplum-db/bin/pg_ctl", segmentPgctlArgs...).ReturnsStatus(1).PrintsError("custom pg_ctl error")
					})
					It("logs error", ShouldFailOnPgCtlError)
				})
			})
			When("/greenplum/data does not exist", func() {
				It("succeeds", func() {
					exitErr := app.InitializeCluster()
					Expect(exitErr).NotTo(HaveOccurred())
				})
				It("should not attempt to initialize the cluster", ShouldNotGpinit)

				It("should not run post initialization", ShouldNotRunPostInitialization)

				It("does not run pg_ctl", ShouldNotRunPgCtl)
			})
		})

		When("sshKeyScan succeeds", func() {
			BeforeEach(func() {
				fakeCmd.ExpectCommand("dnsdomainname").PrintsOutput("myheadlessservice.mynamespace.svc.cluster.local")
			})
			It("calls ScanSegmentHostKeys()", func() {
				exitErr := app.InitializeCluster()
				Expect(exitErr).NotTo(HaveOccurred())
				Expect(outBuffer).To(gbytes.Say("started SSH KeyScan"))
				Expect(keyScanner.WasCalled).To(BeTrue(), "expected keyscanner to be called")
			})
			It("creates an entry in known Hosts file", func() {
				exitErr := app.InitializeCluster()
				Expect(exitErr).NotTo(HaveOccurred())
				_, err := app.Fs.Lstat("/home/gpadmin/.ssh/known_hosts")
				Expect(err).NotTo(HaveOccurred())
				knownHosts, err := vfs.ReadFile(app.Fs, "/home/gpadmin/.ssh/known_hosts")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(knownHosts)).To(ContainSubstring("master-0 FakeKey " + base64.StdEncoding.EncodeToString([]byte("master-0")) + "\n"))
				Expect(string(knownHosts)).To(ContainSubstring("master-0.myheadlessservice.mynamespace.svc.cluster.local FakeKey " + base64.StdEncoding.EncodeToString([]byte("master-0.myheadlessservice.mynamespace.svc.cluster.local")) + "\n"))
				Expect(string(knownHosts)).To(ContainSubstring("master-1 FakeKey " + base64.StdEncoding.EncodeToString([]byte("master-1")) + "\n"))
				Expect(string(knownHosts)).To(ContainSubstring("master-1.myheadlessservice.mynamespace.svc.cluster.local FakeKey " + base64.StdEncoding.EncodeToString([]byte("master-1.myheadlessservice.mynamespace.svc.cluster.local")) + "\n"))
				Expect(string(knownHosts)).To(ContainSubstring("segment-a-0 FakeKey " + base64.StdEncoding.EncodeToString([]byte("segment-a-0")) + "\n"))
				Expect(string(knownHosts)).To(ContainSubstring("segment-a-0.myheadlessservice.mynamespace.svc.cluster.local FakeKey " + base64.StdEncoding.EncodeToString([]byte("segment-a-0.myheadlessservice.mynamespace.svc.cluster.local")) + "\n"))
				Expect(string(knownHosts)).To(ContainSubstring("segment-b-0 FakeKey " + base64.StdEncoding.EncodeToString([]byte("segment-b-0")) + "\n"))
				Expect(string(knownHosts)).To(ContainSubstring("segment-b-0.myheadlessservice.mynamespace.svc.cluster.local FakeKey " + base64.StdEncoding.EncodeToString([]byte("segment-b-0.myheadlessservice.mynamespace.svc.cluster.local")) + "\n"))
			})
		})

		When("sshKeyScan fails", func() {
			BeforeEach(func() {
				keyScanner.Err = errors.New("mock failure")
			})
			It("exits on ScanHostKeys() failure", func() {
				exitErr := app.InitializeCluster()
				Expect(outBuffer).To(gbytes.Say("started SSH KeyScan"))
				Expect(keyScanner.WasCalled).To(BeTrue(), "expected key scanner to be called")
				Expect(outBuffer).To(gbytes.Say("failed to scan segment host keys: mock failure"))
				Expect(exitErr).To(MatchError("failed to scan segment host keys: mock failure"))
			})
		})

		When("Writing to known hosts fails", func() {
			BeforeEach(func() {
				memoryfs.OpenFileHook = func(name string, flag int, perm os.FileMode) (vfs.File, error) {
					if name == "/home/gpadmin/.ssh/known_hosts" {
						return nil, errors.New("OpenFile failed")
					}
					return memoryfs.Filesystem.OpenFile(name, flag, perm)
				}
			})
			It("exits", func() {
				exitErr := app.InitializeCluster()
				Expect(exitErr).To(MatchError("failed to write known_hosts file: OpenFile failed"))
			})
		})

		When("dns resolver fails", func() {
			BeforeEach(func() {
				fakeDNSResolver.FakeErrors = map[string]error{
					"master-0": errors.New("dns failure"),
				}
			})
			It("returns an error", func() {
				exitErr := app.InitializeCluster()
				Expect(exitErr).To(MatchError("failed to resolve dns entries for all masters and segments"))
			})
		})
	})

})

type fakeCluster struct {
	initializeStub struct {
		wasCalled bool
		err       error
	}
	gpstartStub struct {
		wasCalled bool
		err       error
	}
	runPostInitializationStub struct {
		wasCalled bool
		err       error
	}
}

var _ cluster.ClusterInterface = &fakeCluster{}

func (c *fakeCluster) Initialize() error {
	c.initializeStub.wasCalled = true
	return c.initializeStub.err
}

func (c *fakeCluster) GPStart() error {
	c.gpstartStub.wasCalled = true
	return c.gpstartStub.err
}

func (c *fakeCluster) RunPostInitialization() error {
	c.runPostInitializationStub.wasCalled = true
	return c.runPostInitializationStub.err
}
