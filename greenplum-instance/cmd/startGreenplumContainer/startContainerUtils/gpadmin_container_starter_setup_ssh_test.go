package startContainerUtils_test

import (
	"errors"
	"os"
	"strings"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
)

var _ = Describe("GpadminContainerStarter.SetupSSHGpadmin()", func() {
	var (
		app      *startContainerUtils.GpadminContainerStarter
		memoryfs vfs.Filesystem
		fakeCmd  *commandable.CommandFake
	)
	BeforeEach(func() {
		fakeCmd = commandable.NewFakeCommand()
		memoryfs = memfs.Create()
		app = &startContainerUtils.GpadminContainerStarter{
			App: &starter.App{
				Command: fakeCmd.Command,
				Fs:      memoryfs,
			}}

		// simulate host key files that are generated at deployment time
		Expect(vfs.MkdirAll(memoryfs, "/etc/ssh-host-key", 755)).To(Succeed())
		Expect(vfs.WriteFile(memoryfs, "/etc/ssh-host-key/master.ssh_host_rsa_key",
			[]byte("i am host key for master"), 0400)).To(Succeed())
		Expect(vfs.WriteFile(memoryfs, "/etc/ssh-host-key/master.ssh_host_rsa_key.pub",
			[]byte("i am public host key for master"), 0400)).To(Succeed())

		// ensure destination directories that docker will have already created
		Expect(vfs.MkdirAll(memoryfs, LocalHostKeyPath, 755)).To(Succeed())
		Expect(vfs.MkdirAll(memoryfs, "/proc/self/", 0755)).To(Succeed())
		Expect(vfs.MkdirAll(memoryfs, "/sys/fs/cgroup/", 0755)).To(Succeed())
		Expect(vfs.WriteFile(memoryfs, "/proc/self/cgroup", []byte("4:cpu,"+
			"cpuacct:/docker/folder1\n3:devices:/docker/folder2\n2:net_cls:/docker/folder1"), 0644)).To(Succeed())
	})

	const LocalSSHDirPath = "/home/gpadmin/.ssh"
	When("/etc/ssh-key is fully populated", func() {
		BeforeEach(func() {
			Expect(vfs.MkdirAll(memoryfs, "/etc/ssh-key", 755)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/etc/ssh-key/id_rsa",
				[]byte("i am id_rsa"), 0400)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/etc/ssh-key/id_rsa.pub",
				[]byte("i am id_rsa.pub"), 0400)).To(Succeed())
		})
		It("fails when mkdirall fails", func() {
			app.Fs = vfs.Dummy(errors.New("could not make dir all"))

			err := app.SetupSSHForGpadmin()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("could not make dir all"))
		})
		for _, filename := range []string{
			"id_rsa",
			"id_rsa.pub",
			"authorized_keys",
			"config",
		} {
			When("opening ~/.ssh/"+filename+" will fail", func() {
				BeforeEach(func() {
					hookFs := &fileutil.HookableFilesystem{Filesystem: memoryfs}
					app.Fs = hookFs
					hookFs.OpenFileHook = func(name string, flag int, perm os.FileMode) (vfs.File, error) {
						if name == LocalSSHDirPath+"/"+filename {
							return nil, os.ErrInvalid
						}
						return memoryfs.OpenFile(name, flag, perm)
					}
				})
				It("fails", func() {
					err := app.SetupSSHForGpadmin()
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(os.ErrInvalid))
				})
			})
		}
	})
	When("mkdir ~/.ssh fails", func() {
		BeforeEach(func() {
			fakeFS := fileutil.HookableFilesystem{Filesystem: memoryfs}
			app.Fs = &fakeFS
			fakeFS.MkdirHook = func(name string, perm os.FileMode) error {
				if strings.Contains(name, ".ssh") {
					return errors.New("failed to mkdir .ssh")
				}
				return memoryfs.Mkdir(name, perm)
			}
		})
		It("returns the error from MkdirAll", func() {
			err := app.SetupSSHForGpadmin()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to mkdir .ssh"))
		})
	})
	When("/etc/ssh-key doesn't exist", func() {
		It("fails", func() {
			err := app.SetupSSHForGpadmin()
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))
		})
	})
	When("/etc/ssh-key/id_rsa doesn't exist", func() {
		BeforeEach(func() {
			Expect(vfs.MkdirAll(memoryfs, "/etc/ssh-key", 755)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/etc/ssh-key/id_rsa.pub", []byte("i am id_rsa.pub"), 0400)).To(Succeed())
		})
		It("fails", func() {
			err := app.SetupSSHForGpadmin()
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))
		})
	})
	When("/etc/ssh-key/id_rsa.pub doesn't exist", func() {
		BeforeEach(func() {
			Expect(vfs.MkdirAll(memoryfs, "/etc/ssh-key", 755)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/etc/ssh-key/id_rsa", []byte("i am id_rsa"), 0400)).To(Succeed())
		})
		It("fails", func() {
			err := app.SetupSSHForGpadmin()
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))
		})
	})
})
