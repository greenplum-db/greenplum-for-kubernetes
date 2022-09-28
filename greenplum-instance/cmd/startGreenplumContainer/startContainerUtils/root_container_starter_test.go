package startContainerUtils_test

import (
	"errors"
	"os"
	"strings"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/testing/matcher"
	ubuntutest "github.com/pivotal/greenplum-for-kubernetes/pkg/ubuntuUtils/testing"
)

var _ = Describe("RootContainerStarter", func() {
	var (
		app         *startContainerUtils.RootContainerStarter
		errorBuffer *gbytes.Buffer
		outBuffer   *gbytes.Buffer
		memoryfs    vfs.Filesystem
		fakeCmd     *commandable.CommandFake
		mockUbuntu  ubuntutest.MockUbuntu
	)
	BeforeEach(func() {
		fakeCmd = commandable.NewFakeCommand()
		mockUbuntu = ubuntutest.MockUbuntu{}
		errorBuffer = gbytes.NewBuffer()
		outBuffer = gbytes.NewBuffer()
		startContainerUtils.Log = gplog.ForTest(outBuffer)
		memoryfs = memfs.Create()
		app = &startContainerUtils.RootContainerStarter{
			App: &starter.App{
				Command:      fakeCmd.Command,
				StdoutBuffer: outBuffer,
				StderrBuffer: errorBuffer,
				Fs:           memoryfs,
			},
			Ubuntu: &mockUbuntu,
		}

		fakeCmd.ExpectCommand("/usr/bin/ssh-keygen", "-t", "rsa", "-f", "/greenplum/hostKeyDir/ssh_host_rsa_key", "-N", "").SideEffect(func() {
			Expect(vfs.WriteFile(memoryfs, "/greenplum/hostKeyDir/ssh_host_rsa_key", []byte("generated SSH host key"), 0600)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/greenplum/hostKeyDir/ssh_host_rsa_key.pub", []byte("generated SSH host public key"), 0600)).To(Succeed())
		})

		// simulate that mount has already created /greenplum dir. Unfortunately, we cannot simulate it is owned by root
		Expect(vfs.MkdirAll(memoryfs, "/greenplum", 0755)).To(Succeed())
	})

	// ensure destination directories that docker will have already created
	BeforeEach(func() {
		Expect(vfs.MkdirAll(memoryfs, "/etc/ssh", 755)).To(Succeed())
		Expect(vfs.MkdirAll(memoryfs, "/home/gpadmin", 755)).To(Succeed())
		Expect(vfs.WriteFile(memoryfs, "/home/gpadmin/.bashrc",
			[]byte("this is a .bashrc file \n"), 0600)).To(Succeed())
		Expect(vfs.MkdirAll(memoryfs, "/proc/self/", 0755)).To(Succeed())
		Expect(vfs.MkdirAll(memoryfs, "/sys/fs/cgroup/", 0755)).To(Succeed())
		Expect(vfs.WriteFile(memoryfs, "/proc/mounts", []byte(`cgroup /sys/fs/cgroup/cpu,cpuacct cgroup rw,noatime,cpu,cpuacct 0 0
cgroup /sys/fs/cgroup/cpuset cgroup rw,noatime,cpuset 0 0
cgroup /sys/fs/cgroup/devices cgroup rw,noatime,devices 0 0
`), 0644)).To(Succeed())
		Expect(vfs.WriteFile(memoryfs, "/proc/self/cgroup",
			[]byte(`4:cpu,cpuacct:/docker/folder4
3:devices:/docker/folder3
2:cpuset:/docker/folder2`), 0644)).To(Succeed())
		Expect(vfs.MkdirAll(memoryfs, "/etc/", 0755)).To(Succeed())
		Expect(vfs.WriteFile(memoryfs, "/etc/resolv.conf",
			[]byte("nameserver 10.96.0.10\n"+
				"search default.svc.cluster.local svc.cluster.local cluster.local\n"+
				"options ndots:5\n"), 0644)).To(Succeed())
	})

	It("succeeds", func() {
		Expect(app.Run()).To(Succeed())
	})

	It("gpadmin owns /greenplum dir", func() {
		Expect(app.ChownGreenplumDir()).To(Succeed())
		Expect(mockUbuntu.ChangeDirectoryOwnerMock.DirName).To(Equal("/greenplum"))
		Expect(mockUbuntu.ChangeDirectoryOwnerMock.UserName).To(Equal("gpadmin"))
	})
	It("Dir /greenplum/mirror exits on error", func() {
		mockUbuntu.ChangeDirectoryOwnerMock.Err = errors.New("chown -R error")
		err := app.ChownGreenplumDir()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("changing ownership of /greenplum dir to gpadmin failed: chown -R error"))
	})
	It("exits on ssh-keygen failure in SetupSSHHostKeys", func() {
		app.Command = commandable.NewFakeCommand().
			FakeStatus(1).
			FakeErrOutput("error").
			Command
		err := app.SetupSSHHostKeys()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("failed to generate SSH host key at /greenplum/hostKeyDir/ssh_host_rsa_key : exit status 1"))
	})
	It("exits on cp failure in SetupSSHHostKeys", func() {
		Expect(memoryfs.Remove("/etc/ssh")).To(Succeed())

		err := app.SetupSSHHostKeys()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to copy SSH host key to /etc/ssh/ssh_host_rsa_key"))
	})
	It("chowns /greenplum before attempting to add ssh host key there", func() {
		var keygenCalled int
		fakeCmd.ExpectCommandMatching(func(path string, args ...string) bool {
			if path == "/usr/bin/ssh-keygen" {
				Expect(mockUbuntu.ChangeDirectoryOwnerMock.DirName).To(Equal("/greenplum"), "chown was not done before setup ssh hostkeys")
				keygenCalled++
				// Don't return true: let the previous expectation for ssh-keygen match so that its side effects happen.
			}
			return false
		})

		Expect(app.Run()).To(Succeed())
		Expect(keygenCalled).To(Equal(1))
	})
	Describe("Editing /etc/resolv.conf", func() {
		It("insert a new entry for subdomain as the first item in search in /etc/resolv.conf", func() {
			Expect(app.Run()).To(Succeed())
			Expect("/etc/resolv.conf").To(EqualInFilesystem(memoryfs,
				"nameserver 10.96.0.10\n"+
					"search agent.default.svc.cluster.local default.svc.cluster.local svc.cluster.local cluster.local\n"+
					"options ndots:5\n"))
		})
		When("the namespace is 'myns'", func() {
			BeforeEach(func() {
				Expect(vfs.WriteFile(memoryfs, "/etc/resolv.conf",
					[]byte("nameserver 10.96.0.12\n"+
						"search myns.svc.cluster.local svc.cluster.local cluster.local\n"+
						"options ndots:2\n"), 0644)).To(Succeed())
			})
			It("uses that namespace", func() {
				Expect(app.Run()).To(Succeed())
				Expect("/etc/resolv.conf").To(EqualInFilesystem(memoryfs,
					"nameserver 10.96.0.12\n"+
						"search agent.myns.svc.cluster.local myns.svc.cluster.local svc.cluster.local cluster.local\n"+
						"options ndots:2\n"))
			})

		})
		When("the subdomain is already present", func() {
			BeforeEach(func() {
				Expect(vfs.WriteFile(memoryfs, "/etc/resolv.conf",
					[]byte("nameserver 10.96.0.12\n"+
						"search agent.default.svc.cluster.local default.svc.cluster.local svc.cluster.local cluster.local\n"+
						"options ndots:2\n"), 0644)).To(Succeed())
			})
			It("does not edit the file", func() {
				Expect(app.Run()).To(Succeed())
				Expect("/etc/resolv.conf").To(EqualInFilesystem(memoryfs,
					"nameserver 10.96.0.12\n"+
						"search agent.default.svc.cluster.local default.svc.cluster.local svc.cluster.local cluster.local\n"+
						"options ndots:2\n"))
			})
		})
		When("the subdomain is present in a different place on the line", func() {
			BeforeEach(func() {
				Expect(vfs.WriteFile(memoryfs, "/etc/resolv.conf",
					[]byte("nameserver 10.96.0.246\n"+
						"search default.svc.cluster.local svc.cluster.local cluster.local agent.default.svc.cluster.local\n"+
						"options ndots:7\n"), 0644)).To(Succeed())
			})
			It("edits the file", func() {
				Expect(app.Run()).To(Succeed())
				Expect("/etc/resolv.conf").To(EqualInFilesystem(memoryfs,
					"nameserver 10.96.0.246\n"+
						"search agent.default.svc.cluster.local default.svc.cluster.local svc.cluster.local cluster.local\n"+
						"options ndots:7\n"))

			})
		})
		It("errors out when /etc/resolv.conf is missing", func() {
			Expect(memoryfs.Remove("/etc/resolv.conf")).To(Succeed())
			err := app.Run()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("open /etc/resolv.conf: file does not exist"))
		})
		It("errors out when /etc/resolv.conf can't be opened for write", func() {
			hookFs := &fileutil.HookableFilesystem{Filesystem: memoryfs}
			app.Fs = hookFs
			hookFs.OpenFileHook = func(name string, flag int, perm os.FileMode) (vfs.File, error) {
				if name == "/etc/resolv.conf" && flag&os.O_WRONLY != 0 {
					return nil, errors.New("cannot open file for write")
				}
				return hookFs.Filesystem.OpenFile(name, flag, perm)
			}
			err := app.Run()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("cannot open file for write"))
		})
	})
	When("multiple entries are present in /proc/self/cgroup", func() {
		BeforeEach(func() {
			Expect(vfs.MkdirAll(memoryfs, "/proc/self/", 0755)).To(Succeed())
			Expect(vfs.MkdirAll(memoryfs, "/sys/fs/cgroup/", 0755)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/proc/self/cgroup",
				[]byte(`6:rdma:/
5:memory:/docker/folder5
4:devices:/docker/folder4
3:cpuacct:/docker/folder3
2:cpu:/docker/folder2
1:cpuset:/docker/folder1`), 0644)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/proc/mounts",
				[]byte(`cgroup /sys/fs/cgroup/rdma cgroup x,x,rdma 0 0
cgroup /sys/fs/cgroup/memory cgroup x,x,memory 0 0
cgroup /sys/fs/cgroup/devices cgroup x,x,devices 0 0
cgroup /sys/fs/cgroup/cpuacct cgroup x,x,cpuacct 0 0
cgroup /sys/fs/cgroup/cpu cgroup x,x,cpu 0 0
cgroup /sys/fs/cgroup/cpuset cgroup x,x,cpuset 0 0`), 0644)).To(Succeed())
		})

		Context("checking that we create desired cgroup directories", func() {
			var createCPUCalled, createCPUAcctCalled, createCPUSetCalled, createMemoryCalled int
			BeforeEach(func() {
				fakeCmd.ExpectCommand("install", "-d", "-o", "gpadmin", "-g", "gpadmin",
					"/sys/fs/cgroup/cpu/docker/folder2/gpdb").
					CallCounter(&createCPUCalled)
				fakeCmd.ExpectCommand("install", "-d", "-o", "gpadmin", "-g", "gpadmin",
					"/sys/fs/cgroup/cpuacct/docker/folder3/gpdb").
					CallCounter(&createCPUAcctCalled)
				fakeCmd.ExpectCommand("install", "-d", "-o", "gpadmin", "-g", "gpadmin",
					"/sys/fs/cgroup/cpuset/docker/folder1/gpdb").
					CallCounter(&createCPUSetCalled)
				fakeCmd.ExpectCommand("install", "-d", "-o", "gpadmin", "-g", "gpadmin",
					"/sys/fs/cgroup/memory/docker/folder5/gpdb").
					CallCounter(&createMemoryCalled)

				Expect(app.Run()).To(Succeed())
			})
			It("creates gpdb dir in cpu, cpuacct, cpuset and memory", func() {
				Expect(createCPUCalled).To(Equal(1))
				Expect(createCPUAcctCalled).To(Equal(1))
				Expect(createCPUSetCalled).To(Equal(1))
				Expect(createMemoryCalled).To(Equal(1))
			})
		})
		Context("checking that we chown the directories", func() {
			var chownCPUCalled, chownCPUAcctCalled, chownCPUSetCalled, chownMemoryCalled int
			BeforeEach(func() {
				fakeCmd.ExpectCommand("chown", "-R", "gpadmin:gpadmin",
					"/sys/fs/cgroup/cpu/docker/folder2/gpdb").
					CallCounter(&chownCPUCalled)
				fakeCmd.ExpectCommand("chown", "-R", "gpadmin:gpadmin",
					"/sys/fs/cgroup/cpuacct/docker/folder3/gpdb").
					CallCounter(&chownCPUAcctCalled)
				fakeCmd.ExpectCommand("chown", "-R", "gpadmin:gpadmin",
					"/sys/fs/cgroup/cpuset/docker/folder1/gpdb").
					CallCounter(&chownCPUSetCalled)
				fakeCmd.ExpectCommand("chown", "-R", "gpadmin:gpadmin",
					"/sys/fs/cgroup/memory/docker/folder5/gpdb").
					CallCounter(&chownMemoryCalled)

				Expect(app.Run()).To(Succeed())
			})
			It("chowns gpdb dir in cpu, cpuacct, cpuset and memory", func() {
				Expect(chownCPUCalled).To(Equal(1))
				Expect(chownCPUAcctCalled).To(Equal(1))
				Expect(chownCPUSetCalled).To(Equal(1))
				Expect(chownMemoryCalled).To(Equal(1))
			})
		})
		Context("checking that we do not create undesired cgroup dirs", func() {
			var createDevicesCalled, createRDMACalled int
			BeforeEach(func() {
				fakeCmd.ExpectCommand("install", "-d", "-o", "gpadmin", "-g", "gpadmin",
					"/sys/fs/cgroup/devices/docker/folder4/gpdb").
					CallCounter(&createDevicesCalled)
				fakeCmd.ExpectCommand("install", "-d", "-o", "gpadmin", "-g", "gpadmin",
					"/sys/fs/cgroup/rdma/gpdb").
					CallCounter(&createRDMACalled)
			})
			It("does not create gpdb dir in devices or rdma", func() {
				Expect(createDevicesCalled).To(Equal(0))
				Expect(createRDMACalled).To(Equal(0))
			})
		})
	})
	It("returns an error if /proc/mounts doesn't exist", func() {
		Expect(vfs.RemoveAll(memoryfs, "/proc/mounts")).To(Succeed())
		err := app.Run()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("open /proc/mounts: file does not exist"))
	})
	It("returns an error if /proc/self/cgroup doesn't exist", func() {
		Expect(vfs.RemoveAll(memoryfs, "/proc/self/cgroup")).To(Succeed())
		createCgroupCalled := 0
		fakeCmd.ExpectCommand("install", "-d", "-o", "gpadmin", "-g", "gpadmin",
			"/sys/fs/cgroup").CallCounter(&createCgroupCalled)
		err := app.Run()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("open /proc/self/cgroup: file does not exist"))
		Expect(createCgroupCalled).To(Equal(0))
	})
	It("returns an error if running `install -d` fails", func() {
		fakeCmd.ExpectCommandMatching(func(path string, args ...string) bool {
			return path == "install" && len(args) > 5 &&
				args[0] == "-d" && strings.Contains(args[5], "/sys/fs/cgroup")
		}).ReturnsStatus(1).PrintsError("install -d failed")

		err := app.Run()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("failed to create cgroup dir /sys/fs/cgroup/cpu,cpuacct/docker/folder4/gpdb with gpadmin as owner: exit status 1"))
		Expect(errorBuffer).To(gbytes.Say("install -d failed"))
	})
	It("returns an error if running `chown` fails", func() {
		fakeCmd.ExpectCommandMatching(func(path string, args ...string) bool {
			return path == "chown" && len(args) > 2 && strings.Contains(args[2], "/sys/fs/cgroup")
		}).ReturnsStatus(1).PrintsError("chown failed")

		err := app.Run()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("failed to change ownership of /sys/fs/cgroup/cpu,cpuacct/docker/folder4/gpdb to gpadmin: exit status 1"))
		Expect(errorBuffer).To(gbytes.Say("chown failed"))
	})
	When("the mountpoint name differs from the subsystems in /proc/self/cgroup", func() {
		BeforeEach(func() {
			const (
				procMounts = `cgroup /sys/fs/cgroup/the_cpu_one cgroup rw,noatime,cpu,cpuacct 0 0
cgroup /sys/fs/cgroup/the_memory_one cgroup rw,noatime,memory 0 0`
				procCgroup = `1:cpu,cpuacct:/docker/folder1
2:memory:/docker/folder2`
			)
			Expect(vfs.WriteFile(memoryfs, "/proc/mounts", []byte(procMounts), 0644)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/proc/self/cgroup", []byte(procCgroup), 0644)).To(Succeed())
		})
		It("creates the /gpdb sub-group inside the cpu,cpuacct mountpoint", func() {
			createCgroupCalled := 0
			fakeCmd.ExpectCommand("install", "-d", "-o", "gpadmin", "-g", "gpadmin",
				"/sys/fs/cgroup/the_cpu_one/docker/folder1/gpdb").
				CallCounter(&createCgroupCalled)
			Expect(app.Run()).To(Succeed())
			Expect(errorBuffer.Contents()).To(BeEmpty())
			// Expect to be called 2x, once for each subsystem.
			// That's okay because the install and chown are idempotent.
			Expect(createCgroupCalled).To(Equal(2), fakeCmd.CapturedArgs())
		})
		It("creates the /gpdb sub-group inside the memory mountpoint", func() {
			createCgroupCalled := 0
			fakeCmd.ExpectCommand("install", "-d", "-o", "gpadmin", "-g", "gpadmin",
				"/sys/fs/cgroup/the_memory_one/docker/folder2/gpdb").
				CallCounter(&createCgroupCalled)
			Expect(app.Run()).To(Succeed())
			Expect(errorBuffer.Contents()).To(BeEmpty())
			Expect(createCgroupCalled).To(Equal(1), fakeCmd.CapturedArgs())
		})
	})

	When("there are non-cgroup filesystems in /proc/mounts", func() {
		BeforeEach(func() {
			const (
				procMounts = `trick-cgroup /sys/fs/cgroup/the_cpu_one notcgroup rw,noatime,cpu,cpuacct 0 0
cgroup /sys/fs/cgroup/the_memory_one cgroup rw,noatime,memory 0 0
trick-cgroup /sys/fs/cgroup/the_cpu_one notcgroup rw,noatime,cpu,cpuacct 0 0
`
				procCgroup = `1:cpu,cpuacct:/docker/folder1
2:memory:/docker/folder2`
			)
			Expect(vfs.WriteFile(memoryfs, "/proc/mounts", []byte(procMounts), 0644)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/proc/self/cgroup", []byte(procCgroup), 0644)).To(Succeed())
		})
		It("creates the /gpdb sub-group inside the mountpoint", func() {
			createCgroupCalled := 0
			fakeCmd.ExpectCommand("install", "-d", "-o", "gpadmin", "-g", "gpadmin",
				"/sys/fs/cgroup/the_memory_one/docker/folder2/gpdb").
				CallCounter(&createCgroupCalled)
			Expect(app.Run()).To(Succeed())
			Expect(errorBuffer.Contents()).To(BeEmpty())
			Expect(createCgroupCalled).To(Equal(1), fakeCmd.CapturedArgs())
		})
		It("doesn't create any other cgroups", func() {
			createCgroupCalled := 0
			fakeCmd.ExpectCommandMatching(func(path string, args ...string) bool {
				return path == "install"
			}).CallCounter(&createCgroupCalled)
			Expect(app.Run()).To(Succeed())
			Expect(errorBuffer.Contents()).To(BeEmpty())
			Expect(createCgroupCalled).To(Equal(1), fakeCmd.CapturedArgs())
		})
	})

	When("a combined cgroup hierarchy is present with only one GPDB-relevant subsystem", func() {
		BeforeEach(func() {
			const (
				procMounts = `combocgroup /sys/fs/cgroup/the_cpu_and_devices_one cgroup rw,noatime,cpu,devices 0 0`
				procCgroup = `1:devices,cpu:/docker/folder1`
			)
			Expect(vfs.WriteFile(memoryfs, "/proc/mounts", []byte(procMounts), 0644)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/proc/self/cgroup", []byte(procCgroup), 0644)).To(Succeed())
		})
		It("creates the /gpdb sub-group inside the mountpoint", func() {
			createCgroupCalled := 0
			fakeCmd.ExpectCommand("install", "-d", "-o", "gpadmin", "-g", "gpadmin",
				"/sys/fs/cgroup/the_cpu_and_devices_one/docker/folder1/gpdb").
				CallCounter(&createCgroupCalled)
			Expect(app.Run()).To(Succeed())
			Expect(errorBuffer.Contents()).To(BeEmpty())
			Expect(createCgroupCalled).To(Equal(1), fakeCmd.CapturedArgs())
		})
	})

	When("host key is present persistent disk", func() {
		const (
			sshHostRSAKeyPath      = startContainerUtils.HostKeyDir + "/ssh_host_rsa_key"
			sshHostRSAKeyLocalPath = LocalHostKeyPath + "/ssh_host_rsa_key"
		)
		BeforeEach(func() {
			Expect(vfs.MkdirAll(memoryfs, startContainerUtils.HostKeyDir, 0755)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, sshHostRSAKeyPath, []byte("i am private host key"), 0600)).To(Succeed())
		})
		It("copies existing host key to /etc/ssh/ dir", func() {
			Expect(app.Run()).To(Succeed())
			Expect(sshHostRSAKeyLocalPath).To(EqualInFilesystem(memoryfs, "i am private host key"))
			stat, err := memoryfs.Stat(sshHostRSAKeyLocalPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(stat.Mode() & os.ModePerm).To(Equal(os.FileMode(0600)))
		})
	})
	When("host key is not present", func() {
		It("creates a new host key and stores it on persistent disk", func() {
			const sshHostRSAKeyPath = startContainerUtils.HostKeyDir + "/ssh_host_rsa_key"
			fakeCmd.ExpectCommandMatching(func(path string, args ...string) bool {
				return path == "/usr/bin/ssh-keygen" && len(args) > 5 &&
					args[1] == "rsa" && args[3] == sshHostRSAKeyPath && args[5] == ""
			}).SideEffect(func() {
				Expect(vfs.WriteFile(memoryfs, sshHostRSAKeyPath, []byte("i am private host key"), 0600)).To(Succeed())
				Expect(vfs.WriteFile(memoryfs, sshHostRSAKeyPath+".pub", []byte("i am private host key"), 0600)).To(Succeed())
			}).PrintsOutput("ssh-keygen is called")

			Expect(app.Run()).To(Succeed())
			_, err := memoryfs.Stat(startContainerUtils.HostKeyDir + "/ssh_host_rsa_key")
			Expect(err).NotTo(HaveOccurred())
			_, err = memoryfs.Stat(startContainerUtils.HostKeyDir + "/ssh_host_rsa_key.pub")
			Expect(err).NotTo(HaveOccurred())
			// Make sure not to print sshkeygen randomart output
			Expect(outBuffer).NotTo(gbytes.Say("ssh-keygen is called"))
		})
	})

	When("Run method is called", func() {
		It("print log message", func() {
			Expect(app.Run()).To(Succeed())
			Expect(outBuffer).To(gbytes.Say(`"Creating cgroup dirs"`))
			Expect(outBuffer).To(gbytes.Say(`"changing ownership of /greenplum to gpadmin"`))
		})
	})

})
