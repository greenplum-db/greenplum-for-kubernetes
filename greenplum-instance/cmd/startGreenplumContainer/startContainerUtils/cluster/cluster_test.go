package cluster_test

import (
	"errors"
	"os"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils/cluster"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	instanceconfigTesting "github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig/testing"
)

type fakeGpInitSystem struct {
	generateConfigCallCount int
	runCallCount            int
	runError                error
}

func (f *fakeGpInitSystem) GenerateConfig() error {
	f.generateConfigCallCount++
	return nil
}

func (f *fakeGpInitSystem) Run() error {
	f.runCallCount++
	return f.runError
}

type fakeStarter struct {
	called int
	err    error
}

func (f *fakeStarter) Run() error {
	f.called++
	return f.err
}

var _ = Describe("cluster", func() {
	var (
		c             *cluster.Cluster
		exitErr       error
		fs            *fileutil.HookableFilesystem
		cmdFake       *commandable.CommandFake
		outBuffer     *gbytes.Buffer
		errBuffer     *gbytes.Buffer
		fakeGpInit    *fakeGpInitSystem
		mockConfig    *instanceconfigTesting.MockReader
	)

	BeforeEach(func() {
		outBuffer = gbytes.NewBuffer()
		errBuffer = gbytes.NewBuffer()
		fs = &fileutil.HookableFilesystem{Filesystem: memfs.Create()}
		Expect(vfs.MkdirAll(fs, "/home/gpadmin", 0755)).To(Succeed())
		Expect(vfs.MkdirAll(fs, "/etc/config", 0755)).To(Succeed())
		Expect(vfs.WriteFile(fs, "/etc/config/segmentCount", []byte("1"), 0444)).To(Succeed())
		cmdFake = commandable.NewFakeCommand()
		fakeGpInit = &fakeGpInitSystem{}
		mockConfig = &instanceconfigTesting.MockReader{
			NamespaceName: "my-namespace",
			SegmentCount:  1,
			Mirrors:       true,
			Standby:       true,
		}
	})

	JustBeforeEach(func() {
		c = cluster.New(
			fs,
			cmdFake.Command,
			outBuffer,
			errBuffer,
			mockConfig,
			fakeGpInit)
	})

	When("MASTER_DATA_DIRECTORY doesn't exist", func() {
		BeforeEach(func() {
			Expect(vfs.MkdirAll(fs, "/greenplum", 0660)).To(Succeed())
			Expect(vfs.MkdirAll(fs, "/etc/config/", 0755)).To(Succeed())
			Expect(vfs.WriteFile(fs, "/etc/config/hostBasedAuthentication", []byte("hba line 1\nhba line 2"), 0444)).To(Succeed())
		})
		It("succeeds", func() {
			exitErr = c.Initialize()
			Expect(errBuffer.Contents()).To(BeEmpty())
			Expect(exitErr).NotTo(HaveOccurred())
			Expect(outBuffer).To(gbytes.Say("Initializing Greenplum for Kubernetes Cluster"))
			Expect(outBuffer).To(gbytes.Say("Running createdb"))
			Expect(outBuffer).To(gbytes.Say("Adding host based authentication to master-0 pg_hba.conf"))
			Expect(outBuffer).To(gbytes.Say("Adding host based authentication to master-1 pg_hba.conf"))
		})
	})
	When("MASTER_DATA_DIRECTORY already exists", func() {
		BeforeEach(func() {
			Expect(vfs.MkdirAll(fs, "/greenplum/data-1", 0660)).To(Succeed())
		})
		It("fails", func() {
			exitErr = c.Initialize()
			Expect(exitErr).To(MatchError("master data directory already exists at /greenplum/data-1"))
		})
	})

	It("runs GPStart", func() {
		gpstartCounter := 0
		restartSegmentsCounter := 0
		envs := make(chan []string, 1)
		cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/gpstart", "-am").CallCounter(&gpstartCounter).SendEnvironment(envs)
		cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/gpstop", "-ar").CallCounter(&restartSegmentsCounter)
		exitErr = c.GPStart()
		Expect(gpstartCounter).To(Equal(1))
		Expect(restartSegmentsCounter).To(Equal(1))
		Expect(exitErr).ToNot(HaveOccurred())
		Expect(envs).To(Receive(ContainElement(Equal("LD_LIBRARY_PATH=/usr/local/greenplum-db/lib:/usr/local/greenplum-db/ext/python/lib"))))
	})

	When("GPStart in maintenance mode fails", func() {
		It("returns an error", func() {
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/gpstart", "-am").ReturnsStatus(1)
			exitErr = c.GPStart()
			Expect(exitErr).To(MatchError("gpstart in maintenance mode failed: exit status 1"))
		})
	})

	When("restart segments fails", func() {
		It("returns an error", func() {
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/gpstop", "-ar").ReturnsStatus(1)
			exitErr = c.GPStart()
			Expect(exitErr).To(MatchError("restart segments failed: exit status 1"))
		})
	})

	It("called gpInitSystem.GenerateConfig()", func() {
		fakeGpInit.generateConfigCallCount = 0
		exitErr = c.Initialize()
		Expect(errBuffer.Contents()).To(BeEmpty())
		Expect(fakeGpInit.generateConfigCallCount).To(Equal(1))
	})

	It("runs gpinitsystem", func() {
		fakeGpInit.runCallCount = 0
		exitErr = c.Initialize()
		Expect(exitErr).ToNot(HaveOccurred())
		Expect(fakeGpInit.runCallCount).To(Equal(1))
	})

	It("returns an error in gpinitsystem", func() {
		fakeGpInit.runError = errors.New("gpinitsystem failed")
		exitErr = c.Initialize()
		Expect(exitErr).To(MatchError("gpinitsystem failed: gpinitsystem failed"))
	})

	When("/etc/config/hostBasedAuthentication exists", func() {
		var (
			masterCalled  int
			standbyCalled int
		)
		BeforeEach(func() {
			Expect(vfs.MkdirAll(fs, "/etc/config/", 0755)).To(Succeed())
			Expect(vfs.WriteFile(fs, "/etc/config/hostBasedAuthentication", []byte("hba line 1\nhba line 2"), 0444)).To(Succeed())

			masterCalled = 0
			cmdFake.ExpectCommand("/usr/bin/ssh", "master-0",
				"cat", "/etc/config/hostBasedAuthentication",
				">>", "/greenplum/data-1/pg_hba.conf").CallCounter(&masterCalled)
			standbyCalled = 0
			cmdFake.ExpectCommand("/usr/bin/ssh", "master-1",
				"cat", "/etc/config/hostBasedAuthentication",
				">>", "/greenplum/data-1/pg_hba.conf").CallCounter(&standbyCalled)
		})

		When("standby is yes", func() {
			It("adds hostBasedAuthentication to pg_hba.conf on master-0 and master-1", func() {
				exitErr = c.Initialize()
				Expect(exitErr).ToNot(HaveOccurred())
				Expect(errBuffer.Contents()).To(BeEmpty())
				Expect(masterCalled).To(Equal(1))
				Expect(standbyCalled).To(Equal(1))
			})
			It("returns an error on master-0", func() {
				hbaCalled := 0
				cmdFake.ExpectCommand("/usr/bin/ssh", "master-0",
					"cat", "/etc/config/hostBasedAuthentication",
					">>", "/greenplum/data-1/pg_hba.conf").
					CallCounter(&hbaCalled).
					ReturnsStatus(1).
					PrintsOutput("addHostBasedAuthentication is called").
					PrintsError("addHostBasedAuthentication failed with some error")

				exitErr = c.Initialize()
				Expect(hbaCalled).To(Equal(1))
				Expect(outBuffer).To(gbytes.Say("addHostBasedAuthentication is called"))
				Expect(errBuffer).To(gbytes.Say("addHostBasedAuthentication failed with some error"))
				Expect(exitErr).To(MatchError("adding host-based authentication failed: Attempting to append from '/etc/config/hostBasedAuthentication' to end of /greenplum/data-1/pg_hba.conf: exit status 1"))
			})
			It("returns an error on master-1", func() {
				hbaCalled := 0
				cmdFake.ExpectCommand("/usr/bin/ssh", "master-1",
					"cat", "/etc/config/hostBasedAuthentication",
					">>", "/greenplum/data-1/pg_hba.conf").
					CallCounter(&hbaCalled).
					ReturnsStatus(1).
					PrintsOutput("addHostBasedAuthentication is called").
					PrintsError("addHostBasedAuthentication failed with some error")

				exitErr = c.Initialize()
				Expect(hbaCalled).To(Equal(1))
				Expect(outBuffer).To(gbytes.Say("addHostBasedAuthentication is called"))
				Expect(errBuffer).To(gbytes.Say("addHostBasedAuthentication failed with some error"))
				Expect(exitErr).To(MatchError("adding host-based authentication failed: Attempting to append from '/etc/config/hostBasedAuthentication' to end of /greenplum/data-1/pg_hba.conf: exit status 1"))
			})
			It("returns an error from HasContent", func() {
				fakeLStatFS := &fakeLStatFileSystem{fakeFileSystem{MemFS: memfs.Create()}}
				c.Filesystem = fakeLStatFS
				exitErr = c.Initialize()
				Expect(exitErr).To(MatchError("adding host-based authentication failed: verifying if /etc/config/hostBasedAuthentication has any content failed: fake /etc/config/hostBasedAuthentication: invalid argument"))
			})
			When("configreader fails to read standby in configmap", func() {
				BeforeEach(func() {
					mockConfig.StandbyErr = errors.New("standby error")
				})
				It("returns an error", func() {
					exitErr = c.Initialize()
					Expect(exitErr).To(MatchError("reading standby failed: standby error"))
				})
			})
		})
		When("standby is no", func() {
			BeforeEach(func() {
				mockConfig.Standby = false
			})
			It("adds master-0, but does not add hostBasedAuthentication to master-1 pg_hba.conf", func() {
				exitErr = c.Initialize()
				Expect(exitErr).ToNot(HaveOccurred())
				Expect(errBuffer.Contents()).To(BeEmpty())
				Expect(masterCalled).To(Equal(1))
				Expect(standbyCalled).To(Equal(0))
			})
		})
	})
	var itDoesNotWriteToPgHba = func() {
		It("does not add hostBasedAuthentication to master-0/1 pg_hba.conf", func() {
			masterCalled := 0
			cmdFake.ExpectCommand("/usr/bin/ssh", "master-0",
				"cat", "/etc/config/hostBasedAuthentication",
				">>", "/greenplum/data-1/pg_hba.conf").CallCounter(&masterCalled)
			standbyCalled := 0
			cmdFake.ExpectCommand("/usr/bin/ssh", "master-1",
				"cat", "/etc/config/hostBasedAuthentication",
				">>", "/greenplum/data-1/pg_hba.conf").CallCounter(&standbyCalled)

			exitErr = c.Initialize()
			Expect(errBuffer.Contents()).To(BeEmpty())
			Expect(masterCalled).To(Equal(0))
			Expect(standbyCalled).To(Equal(0))
		})
	}
	When("/etc/config/hostBasedAuthentication does not exist", func() {
		BeforeEach(func() {
			Expect(vfs.MkdirAll(fs, "/etc/config/", 0755)).To(Succeed())
		})
		itDoesNotWriteToPgHba()
	})

	When("/etc/config/hostBasedAuthentication is empty", func() {
		BeforeEach(func() {
			Expect(vfs.MkdirAll(fs, "/etc/config/", 0755)).To(Succeed())
			Expect(vfs.WriteFile(fs, "/etc/config/hostBasedAuthentication", []byte{}, 0444)).To(Succeed())
		})
		itDoesNotWriteToPgHba()
	})

	ContainGreenplumEnvironment := And(
		ContainElement("HOME=/home/gpadmin"),
		ContainElement("USER=gpadmin"),
		ContainElement("LOGNAME=gpadmin"),
		ContainElement("GPHOME=/usr/local/greenplum-db"),
		ContainElement("PATH=/usr/local/greenplum-db/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"),
		ContainElement("LD_LIBRARY_PATH=/usr/local/greenplum-db/lib:/usr/local/greenplum-db/ext/python/lib"),
		ContainElement("MASTER_DATA_DIRECTORY=/greenplum/data-1"),
		ContainElement("PYTHONPATH=/usr/local/greenplum-db/lib/python"),
	)

	It("runs createdb", func() {
		createDbCallCount := 0
		envs := make(chan []string, 1)
		cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/createdb").
			CallCounter(&createDbCallCount).SendEnvironment(envs)
		exitErr = c.Initialize()
		Expect(errBuffer.Contents()).To(BeEmpty())
		Expect(createDbCallCount).To(Equal(1))

		var env []string
		Expect(envs).To(Receive(&env))
		Expect(env).To(ContainGreenplumEnvironment)
	})
	It("returns an error from createdb", func() {
		cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/createdb").
			ReturnsStatus(1).
			PrintsOutput("createdb is called").
			PrintsError("createdb failed with some error")

		exitErr = c.Initialize()
		Expect(outBuffer).To(gbytes.Say("createdb is called"))
		Expect(errBuffer).To(gbytes.Say("createdb failed with some error"))
		Expect(exitErr).To(MatchError("createdb failed: exit status 1"))
	})

	Describe("RunPostInitialization", func() {
		It("checks to see if the database is running", func() {
			psqlCalled := 0
			envs := make(chan []string, 1)
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-c", "select * from gp_segment_configuration").
				CallCounter(&psqlCalled).SendEnvironment(envs)

			Expect(c.RunPostInitialization()).To(Succeed())

			Expect(errBuffer.Contents()).To(BeEmpty())
			Expect(psqlCalled).To(Equal(1))

			var env []string
			Expect(envs).To(Receive(&env))
			Expect(env).To(ContainGreenplumEnvironment)
		})
		It("skips post-initialization when the database is not running", func() {
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-c", "select * from gp_segment_configuration").
				ReturnsStatus(1)

			Expect(c.RunPostInitialization()).To(Succeed())
			Expect(outBuffer).To(gbytes.Say("the database is not running. skipping post-initialization"))
		})
		It("runs gpstop -u", func() {
			gpStopCalled := 0
			envs := make(chan []string, 1)
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/gpstop", "-u").
				CallCounter(&gpStopCalled).SendEnvironment(envs)

			exitErr = c.RunPostInitialization()
			Expect(outBuffer).To(gbytes.Say("Reloading greenplum configs"))
			Expect(errBuffer.Contents()).To(BeEmpty())
			Expect(gpStopCalled).To(Equal(1))

			var env []string
			Expect(envs).To(Receive(&env))
			Expect(env).To(ContainGreenplumEnvironment)
		})
		It("returns an error from gpstop -u", func() {
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/gpstop", "-u").
				ReturnsStatus(1).
				PrintsOutput("gpstop is called").
				PrintsError("gpstop failed with some error")

			exitErr = c.RunPostInitialization()
			Expect(outBuffer).To(gbytes.Say("Reloading greenplum configs"))
			Expect(outBuffer).To(gbytes.Say("gpstop is called"))
			Expect(errBuffer).To(gbytes.Say("gpstop failed with some error"))
			Expect(exitErr).To(MatchError("gpstop failed: exit status 1"))
		})
		When("PXF_HOST is set", func() {
			BeforeEach(func() {
				mockConfig.PXFServiceName = "pxf-service"
			})
			It("creates PXF extension when PXF_HOST is set", func() {
				createExtensionCount := 0
				envs := make(chan []string, 1)
				cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-d", "gpadmin", "-c", "CREATE EXTENSION IF NOT EXISTS pxf").
					CallCounter(&createExtensionCount).SendEnvironment(envs)
				exitErr = c.RunPostInitialization()
				Expect(errBuffer.Contents()).To(BeEmpty())
				Expect(createExtensionCount).To(Equal(1))

				var env []string
				Expect(envs).To(Receive(&env))
				Expect(env).To(ContainGreenplumEnvironment)
			})
			It("returns an error when creating PXF extension", func() {
				cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-d", "gpadmin", "-c", "CREATE EXTENSION IF NOT EXISTS pxf").
					ReturnsStatus(1).
					PrintsOutput("createPXFExtension is called").
					PrintsError("createPXFExtension failed with some error")

				exitErr = c.RunPostInitialization()
				Expect(outBuffer).To(gbytes.Say("createPXFExtension is called"))
				Expect(errBuffer).To(gbytes.Say("createPXFExtension failed with some error"))
				Expect(exitErr).To(MatchError("createPXFExtension failed: exit status 1"))
			})
		})
		It("does not create PXF extension when PXF_HOST is not set", func() {
			createExtensionCount := 0
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-d", "gpadmin", "-c", "CREATE EXTENSION IF NOT EXISTS pxf").
				CallCounter(&createExtensionCount)
			exitErr = c.RunPostInitialization()
			Expect(errBuffer.Contents()).To(BeEmpty())
			Expect(createExtensionCount).To(Equal(0))
		})
	})
})

type fakeFileSystem struct {
	*memfs.MemFS
	isFileClosed bool
}

type fakeLStatFileSystem struct {
	fakeFileSystem
}

func (f *fakeLStatFileSystem) Lstat(name string) (os.FileInfo, error) {
	return nil, &os.PathError{Op: "fake", Path: name, Err: os.ErrInvalid}
}
