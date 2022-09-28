package cluster_test

import (
	"errors"
	"io"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils/cluster"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	instanceconfigTesting "github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig/testing"
)

var _ = Describe("Gpinitsystem.GenerateConfig()", func() {
	var (
		g            cluster.GpInitSystem
		outBuffer    io.Writer
		errBuffer    io.Writer
		fs           vfs.Filesystem
		cmdFake      *commandable.CommandFake
		configReader *instanceconfigTesting.MockReader
	)

	BeforeEach(func() {
		outBuffer = gbytes.NewBuffer()
		errBuffer = gbytes.NewBuffer()
		fs = memfs.Create()
		cmdFake = commandable.NewFakeCommand()
		configReader = &instanceconfigTesting.MockReader{}
		g = cluster.NewGpInitSystem(fs, cmdFake.Command, outBuffer, errBuffer, configReader)
		// for hostname, make sure that the output reflects a changed "agent" name and a non-default namespace
	})

	When("all directories exist and a specific namespace and headless name are used", func() {
		BeforeEach(func() {
			Expect(vfs.MkdirAll(fs, "/home/gpadmin", 0755)).To(Succeed())
		})

		When("SEGMENT_COUNT is 1 and MIRRORS = true", func() {
			BeforeEach(func() {
				configReader.SegmentCount = 1
				configReader.Mirrors = true
			})

			It("checks dnsdomainname and prints the correct output", func() {
				cmdFake.FakeOutput("myheadlessservice.mynamespace.svc.cluster.local")
				Expect(g.GenerateConfig()).To(Succeed())
				Expect(cmdFake.CapturedArgs()).To(Equal("dnsdomainname"))
				Expect(outBuffer).To(gbytes.Say("Generating gpinitsystem_config"))
				Expect(outBuffer).To(gbytes.Say("Sub Domain for the cluster is: myheadlessservice.mynamespace.svc.cluster.local\n"))
			})

			It("generates gpinitsystem_config with QD_PRIMARY_ARRAY, 1 PRIMARY, and 1 MIRROR", func() {
				cmdFake.FakeOutput("myheadlessservice.mynamespace.svc.cluster.local")
				Expect(g.GenerateConfig()).To(Succeed())
				config, err := vfs.ReadFile(fs, "/home/gpadmin/gpinitsystem_config")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(config)).To(Equal(
					"QD_PRIMARY_ARRAY=master-0.myheadlessservice.mynamespace.svc.cluster.local~5432~/greenplum/data-1~1~-1~0\n" +
						"declare -a PRIMARY_ARRAY=(\n" +
						"segment-a-0.myheadlessservice.mynamespace.svc.cluster.local~40000~/greenplum/data~2~0\n" +
						")\n" +
						"declare -a MIRROR_ARRAY=(\n" +
						"segment-b-0.myheadlessservice.mynamespace.svc.cluster.local~50000~/greenplum/mirror/data~3~0\n" +
						")\n" +
						"HBA_HOSTNAMES=1\n"))
			})
		})

		When("SEGMENT_COUNT is 2 and MIRRORS = true", func() {
			BeforeEach(func() {
				configReader.SegmentCount = 2
				configReader.Mirrors = true
			})
			It("generates gpinitsystem_config with QD_PRIMARY_ARRAY, 2 PRIMARY, and 2 MIRROR", func() {
				cmdFake.FakeOutput("myheadlessservice.mynamespace.svc.cluster.local")
				Expect(g.GenerateConfig()).To(Succeed())
				config, err := vfs.ReadFile(fs, "/home/gpadmin/gpinitsystem_config")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(config)).To(Equal(
					"QD_PRIMARY_ARRAY=master-0.myheadlessservice.mynamespace.svc.cluster.local~5432~/greenplum/data-1~1~-1~0\n" +
						"declare -a PRIMARY_ARRAY=(\n" +
						"segment-a-0.myheadlessservice.mynamespace.svc.cluster.local~40000~/greenplum/data~2~0\n" +
						"segment-a-1.myheadlessservice.mynamespace.svc.cluster.local~40000~/greenplum/data~3~1\n" +
						")\n" +
						"declare -a MIRROR_ARRAY=(\n" +
						"segment-b-0.myheadlessservice.mynamespace.svc.cluster.local~50000~/greenplum/mirror/data~4~0\n" +
						"segment-b-1.myheadlessservice.mynamespace.svc.cluster.local~50000~/greenplum/mirror/data~5~1\n" +
						")\n" +
						"HBA_HOSTNAMES=1\n"))
			})
		})

		When("SEGMENT_COUNT fails to read", func() {
			BeforeEach(func() {
				configReader.SegmentCountErr = errors.New("foo bar")
			})
			It("returns an error", func() {
				Expect(g.GenerateConfig()).To(MatchError("foo bar"))
			})
		})

		When("SEGMENT_COUNT is 1 and MIRRORS = false", func() {
			BeforeEach(func() {
				configReader.SegmentCount = 1
				configReader.Mirrors = false
			})

			It("generates gpinitsystem_config with QD_PRIMARY_ARRAY, 1 PRIMARY, and 0 MIRRORS", func() {
				cmdFake.FakeOutput("myheadlessservice.mynamespace.svc.cluster.local")
				Expect(g.GenerateConfig()).To(Succeed())
				config, err := vfs.ReadFile(fs, "/home/gpadmin/gpinitsystem_config")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(config)).To(Equal(
					"QD_PRIMARY_ARRAY=master-0.myheadlessservice.mynamespace.svc.cluster.local~5432~/greenplum/data-1~1~-1~0\n" +
						"declare -a PRIMARY_ARRAY=(\n" +
						"segment-a-0.myheadlessservice.mynamespace.svc.cluster.local~40000~/greenplum/data~2~0\n" +
						")\n" +
						"HBA_HOSTNAMES=1\n"))
			})
		})
	})

	When("/home/gpadmin is missing, preventing gpinitsystem_config from being created", func() {
		It("returns an error", func() {
			_, err := fs.Stat("/home/gpadmin/gpinitsystem_config")
			Expect(err).To(HaveOccurred())
			Expect(g.GenerateConfig()).To(MatchError("open /home/gpadmin/gpinitsystem_config: file does not exist"))
		})
	})

	When("hostname fails", func() {
		It("returns error", func() {
			cmdFake.ExpectCommand("dnsdomainname").ReturnsStatus(1)
			Expect(vfs.MkdirAll(fs, "/home/gpadmin", 0755)).To(Succeed())
			Expect(g.GenerateConfig()).To(MatchError("dnsdomainname failed to determine this host's dns name: exit status 1"))
		})
	})
})

var _ = Describe("GpInitSystem.Run()", func() {
	var (
		fs                         vfs.Filesystem
		cmdFake                    *commandable.CommandFake
		outBuf                     io.Writer
		errBuf                     io.Writer
		configReader               *instanceconfigTesting.MockReader
		dnsDomainNameCalledCounter int
	)

	BeforeEach(func() {
		fs = memfs.Create()
		cmdFake = commandable.NewFakeCommand()
		outBuf = gbytes.NewBuffer()
		errBuf = gbytes.NewBuffer()
		configReader = &instanceconfigTesting.MockReader{Standby: true}
		dnsDomainNameCalledCounter = 0
		cmdFake.ExpectCommand("dnsdomainname").PrintsOutput("myHeadlessService.myNamespace.svc.cluster.local\n").CallCounter(&dnsDomainNameCalledCounter)
	})

	When("happy path", func() {
		var gpinitsystemCallCounter int
		var envs chan []string
		BeforeEach(func() {
			gpinitsystemCallCounter = 0
			envs = make(chan []string, 1)
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/gpinitsystem",
				"-a",
				"-I", "/home/gpadmin/gpinitsystem_config",
				"-s", "master-1.myHeadlessService.myNamespace.svc.cluster.local").
				CallCounter(&gpinitsystemCallCounter).
				PrintsOutput("gpinitsystem is called").SendEnvironment(envs)
		})
		It("prints gpinitsystem output to stdout and sets all required env", func() {
			g := cluster.NewGpInitSystem(fs, cmdFake.Command, outBuf, errBuf, configReader)
			Expect(g.Run()).To(Succeed())
			Expect(gpinitsystemCallCounter).To(Equal(1))
			Expect(outBuf).To(gbytes.Say("Running gpinitsystem"))
			Expect(outBuf).To(gbytes.Say("gpinitsystem is called"))

			var env []string
			Expect(envs).To(Receive(&env))
			Expect(env).To(And(
				ContainElement("HOME=/home/gpadmin"),
				ContainElement("USER=gpadmin"),
				ContainElement("LOGNAME=gpadmin"),
				ContainElement("GPHOME=/usr/local/greenplum-db"),
				ContainElement("PATH=/usr/local/greenplum-db/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"),
				ContainElement("LD_LIBRARY_PATH=/usr/local/greenplum-db/lib:/usr/local/greenplum-db/ext/python/lib"),
				ContainElement("MASTER_DATA_DIRECTORY=/greenplum/data-1"),
				ContainElement("PYTHONPATH=/usr/local/greenplum-db/lib/python"),
			))
		})
	})
	When("gpinitsystem exits with exitCode 1", func() {
		var gpinitsystemCallCount int
		BeforeEach(func() {
			// NOTE: when gpinitsystem succeed, it's exit code is 1.
			// It's legacy behavior and cannot be changed since customer depends on this behavior
			gpinitsystemCallCount = 0
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/gpinitsystem",
				"-a", "-I", "/home/gpadmin/gpinitsystem_config",
				"-s", "master-1.myHeadlessService.myNamespace.svc.cluster.local").
				PrintsError("succeed to call gpinitsystem").
				ReturnsStatus(1).
				CallCounter(&gpinitsystemCallCount)
		})
		It("successfully runs", func() {
			g := cluster.NewGpInitSystem(fs, cmdFake.Command, outBuf, errBuf, configReader)
			err := g.Run()
			Expect(err).ToNot(HaveOccurred())
			Expect(gpinitsystemCallCount).To(Equal(1))
		})
	})
	When("GUCs file exist (and Standby is True)", func() {
		var gpinitsystemCallCount int
		BeforeEach(func() {
			dnsDomainNameCalledCounter = 0
			gpinitsystemCallCount = 0
			Expect(vfs.MkdirAll(fs, "/etc/config", 0755)).To(Succeed())
			_, err := vfs.Create(fs, "/etc/config/GUCs")
			Expect(err).NotTo(HaveOccurred())
		})
		It("runs gpinitsystem with -p", func() {
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/gpinitsystem",
				"-a", "-I", "/home/gpadmin/gpinitsystem_config",
				"-s", "master-1.myHeadlessService.myNamespace.svc.cluster.local",
				"-p", "/etc/config/GUCs").CallCounter(&gpinitsystemCallCount)
			g := cluster.NewGpInitSystem(fs, cmdFake.Command, outBuf, errBuf, configReader)
			Expect(g.Run()).To(Succeed())
			Expect(dnsDomainNameCalledCounter).To(Equal(1))
			Expect(gpinitsystemCallCount).To(Equal(1))
		})
	})
	When("Standby is false", func() {
		var gpinitsystemCallCount int
		BeforeEach(func() {
			configReader.Standby = false
		})
		It("runs gpinitsystem without -s argument", func() {
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/gpinitsystem",
				"-a", "-I", "/home/gpadmin/gpinitsystem_config").
				CallCounter(&gpinitsystemCallCount)
			g := cluster.NewGpInitSystem(fs, cmdFake.Command, outBuf, errBuf, configReader)
			Expect(g.Run()).To(Succeed())
			Expect(dnsDomainNameCalledCounter).To(Equal(1))
			Expect(gpinitsystemCallCount).To(Equal(1))
		})
	})
	When("gpinitsystem fails", func() {
		BeforeEach(func() {
			cmdFake.FakeStatus(2).FakeErrOutput("gpinitsystem failed with some error")
		})
		It("exits with an error", func() {
			g := cluster.NewGpInitSystem(fs, cmdFake.Command, outBuf, errBuf, configReader)
			err := g.Run()
			Expect(err).To(HaveOccurred())
			Expect(errBuf).To(gbytes.Say("gpinitsystem failed with some error"))
			Expect(err.Error()).To(Equal("gpinitsystem failed: exit status 2"))
			Expect(dnsDomainNameCalledCounter).To(Equal(1))
		})
	})
	When("dnsdomainname fails", func() {
		BeforeEach(func() {
			cmdFake.ExpectCommand("dnsdomainname").ReturnsStatus(1)
		})
		It("exits with an error", func() {
			g := cluster.NewGpInitSystem(fs, cmdFake.Command, outBuf, errBuf, configReader)
			err := g.Run()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("dnsdomainname failed to determine this host's dns name: exit status 1"))
		})
	})
})
