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
)

var _ = Describe("GpadminContainerStarter", func() {
	var (
		app       *startContainerUtils.GpadminContainerStarter
		outBuffer *gbytes.Buffer
		memoryfs  vfs.Filesystem
		fakeCmd   *commandable.CommandFake
	)

	BeforeEach(func() {
		fakeCmd = commandable.NewFakeCommand()
		outBuffer = gbytes.NewBuffer()
		startContainerUtils.Log = gplog.ForTest(outBuffer)
		memoryfs = memfs.Create()
		app = &startContainerUtils.GpadminContainerStarter{
			App: &starter.App{
				Command: fakeCmd.Command,
				Fs:      memoryfs,
			}}
		Expect(vfs.MkdirAll(memoryfs, "/etc/config/", 755)).To(Succeed())
		Expect(vfs.WriteFile(memoryfs, "/etc/config/pxfServiceName",
			[]byte("pxf-service"), 0400)).To(Succeed())
	})

	Describe("CreatePsqlHistory()", func() {
		It("touch psql_history file throws error when failed to create the file", func() {
			message := "open /home/gpadmin/.psql_history: file does not exist"
			err := app.CreatePsqlHistory()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(message))
		})

	})
	Describe("on Run()", func() {
		BeforeEach(func() {
			// simulate ssh key files that are generated at deployment time (shared by all containers)
			Expect(vfs.MkdirAll(memoryfs, "/etc/ssh-key", 755)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/etc/ssh-key/id_rsa",
				[]byte("i am id_rsa"), 0400)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/etc/ssh-key/id_rsa.pub",
				[]byte("i am id_rsa.pub"), 0400)).To(Succeed())
		})
		It("prints the expected log messages", func() {
			Expect(app.Run()).To(Succeed())
			Expect(outBuffer).To(gbytes.Say(`"starting Greenplum Container"`))
			Expect(outBuffer).To(gbytes.Say(`"setting up ssh for gpadmin"`))
			Expect(outBuffer).To(gbytes.Say(`"creating symlink for gpAdminLogs"`))
			Expect(outBuffer).To(gbytes.Say(`"creating /home/gpadmin/.psql_history file"`))
			Expect(outBuffer).To(gbytes.Say(`"creating mirror dir /greenplum/mirror"`))

		})
		It("create psql_history file throws an error and exits on Run()", func() {
			fakeFS := fileutil.HookableFilesystem{Filesystem: memoryfs}
			app.Fs = &fakeFS
			fakeFS.OpenFileHook = func(name string, flag int, perm os.FileMode) (vfs.File, error) {
				if strings.Contains(name, ".psql_history") {
					return nil, errors.New("failed to create .psql_history")
				}
				return memoryfs.OpenFile(name, flag, perm)
			}
			err := app.Run()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to create .psql_history"))
		})

		It("sources greenplum path", func() {
			Expect(app.Run()).To(Succeed())
			b, _ := vfs.ReadFile(memoryfs, "/home/gpadmin/.bashrc")
			Expect(string(b)).To(HavePrefix("source /usr/local/greenplum-db/greenplum_path.sh\n"))
		})
		It("exports PXF_HOST when /etc/config/pxfServiceName is not empty", func() {
			Expect(vfs.MkdirAll(memoryfs, "/etc/config", 0755)).To(Succeed())
			_, err := vfs.Create(memoryfs, "/etc/config/pxfServiceName")
			Expect(err).NotTo(HaveOccurred())
			err = vfs.WriteFile(memoryfs, "/etc/config/pxfServiceName", []byte("my-pxf-host"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			Expect(app.Run()).To(Succeed())
			b, _ := vfs.ReadFile(memoryfs, "/home/gpadmin/.bashrc")
			Expect(string(b)).To(ContainSubstring("export PXF_HOST=my-pxf-host\n"))
		})
		It("does not export PXF_HOST when /etc/config/pxfServiceName is empty", func() {
			Expect(vfs.MkdirAll(memoryfs, "/etc/config", 0755)).To(Succeed())
			_, err := vfs.Create(memoryfs, "/etc/config/pxfServiceName")
			Expect(err).NotTo(HaveOccurred())

			Expect(app.Run()).To(Succeed())
			b, _ := vfs.ReadFile(memoryfs, "/home/gpadmin/.bashrc")
			Expect(string(b)).NotTo(ContainSubstring("export PXF_HOST="))
		})
		It("does not export PXF_HOST when /etc/config/pxfServiceName does not exist", func() {
			fake := fileutil.HookableFilesystem{Filesystem: memoryfs}
			fake.StatHook = func(name string) (info os.FileInfo, e error) {
				return nil, errors.New("file does not exist")
			}
			app.Fs = &fake
			Expect(app.Run()).To(Succeed())
			b, _ := vfs.ReadFile(memoryfs, "/home/gpadmin/.bashrc")
			Expect(string(b)).NotTo(ContainSubstring("export PXF_HOST="))
		})
		It("returns error when /etc/config/pxfServiceName exists but fails to read", func() {
			fake := fileutil.HookableFilesystem{Filesystem: memoryfs}
			fake.OpenFileHook = func(name string, flag int, perm os.FileMode) (vfs.File, error) {
				if strings.Contains(name, "pxfServiceName") {
					return nil, errors.New("read failed on /etc/config/pxfServiceName")
				}
				return memoryfs.OpenFile(name, flag, perm)
			}
			app.Fs = &fake

			err := app.Run()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to read /etc/config/pxfServiceName, was configMap mounted properly?"))
		})
		It("fails on file writer error", func() {
			fake := fileutil.HookableFilesystem{Filesystem: memoryfs}
			app.Fs = &fake
			fake.OpenFileHook = func(name string, flag int, perm os.FileMode) (vfs.File, error) {
				if strings.Contains(name, ".bashrc") {
					return nil, errors.New("failed to open .bashrc")
				}
				return memoryfs.OpenFile(name, flag, perm)
			}
			err := app.Run()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to open .bashrc"))
		})

		It("create symlink gpAdminLogs successfully create gpAdminLogs dir in /greenplum", func() {
			Expect(app.Run()).To(Succeed())

			fileInfo, err := memoryfs.Stat("/greenplum/gpAdminLogs")
			Expect(err).To(BeNil())
			Expect(fileInfo.IsDir()).To(BeTrue())
		})

		It("create symlink gpAdminLogs successfully and read files from the target dir", func() {
			Expect(vfs.MkdirAll(memoryfs, "/greenplum/gpAdminLogs", 0755)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/greenplum/gpAdminLogs/realfile",
				[]byte("Hey! I am real file"), 0644)).To(Succeed())
			_, err := vfs.ReadFile(memoryfs, "/home/gpadmin/gpAdminLogs/realfile")
			Expect(err).To(HaveOccurred())

			Expect(app.Run()).To(Succeed())

			gpAdminLogsBytes, err := vfs.ReadFile(memoryfs, "/home/gpadmin/gpAdminLogs/realfile")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(gpAdminLogsBytes)).To(ContainSubstring("Hey! I am real file"))
		})
		It("fails when creating gpAdminLogs/ fails", func() {
			fake := fileutil.HookableFilesystem{Filesystem: memoryfs}
			app.Fs = &fake
			fake.MkdirHook = func(name string, perm os.FileMode) error {
				if strings.Contains(name, "gpAdminLogs") {
					return errors.New("failed to mkdir gpAdminLogs")
				}
				return memoryfs.Mkdir(name, perm)
			}

			err := app.Run()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to mkdir gpAdminLogs"))
		})
		It("throws error when creating symlink gpAdminLogs/ fails", func() {
			fake := fileutil.HookableFilesystem{Filesystem: memoryfs}
			app.Fs = &fake
			fake.SymlinkHook = func(oldname, newname string) error {
				return errors.New("failed to create symlink")
			}

			err := app.Run()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to create symlink"))
		})

		It("Dir /greenplum/mirror exists", func() {
			Expect(app.Run()).To(Succeed())

			_, err := memoryfs.Stat("/greenplum/mirror")
			Expect(err).To(BeNil())
		})
		It("Dir /greenplum/mirror exits on error on Run()", func() {

			fakeFS := fileutil.HookableFilesystem{Filesystem: memoryfs}
			app.Fs = &fakeFS
			fakeFS.MkdirHook = func(name string, perm os.FileMode) error {
				if strings.Contains(name, "mirror") {
					return errors.New("failed to create dir /greenplum/mirror")
				}
				return memoryfs.Mkdir(name, perm)
			}

			err := app.Run()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to create dir /greenplum/mirror"))
		})

		It("writes all ~/.ssh files", func() {
			Expect(app.Run()).To(Succeed())
			dir, err := memoryfs.Stat(LocalSSHDirPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(dir.IsDir()).To(BeTrue())

			Expect(LocalSSHDirPath + "/id_rsa").To(EqualInFilesystem(memoryfs, "i am id_rsa"))
			Expect(LocalSSHDirPath + "/id_rsa.pub").To(EqualInFilesystem(memoryfs, "i am id_rsa.pub"))
			Expect(LocalSSHDirPath + "/authorized_keys").To(EqualInFilesystem(memoryfs, "i am id_rsa.pub"))
			Expect(LocalSSHDirPath + "/known_hosts").To(EqualInFilesystem(memoryfs, ""))
			Expect(LocalSSHDirPath + "/config").To(EqualInFilesystem(memoryfs, "Host *\n    ConnectionAttempts 5"))
			fileInfo, _ := memoryfs.Stat(LocalSSHDirPath + "/known_hosts")
			Expect(int(fileInfo.Mode().Perm())).To(Equal(0600))
		})
		It("touch psql_history file successfully", func() {
			Expect(vfs.MkdirAll(memoryfs, "/home/gpadmin", 0755)).To(Succeed())
			fileName := ".psql_history"
			fileNameWithPath := "/home/gpadmin/.psql_history"

			app.Run()

			info, err := memoryfs.Stat(fileNameWithPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Name()).To(Equal(fileName))
			Expect(info.Mode()).To(Equal(os.FileMode(0600)))
		})
	})

})
