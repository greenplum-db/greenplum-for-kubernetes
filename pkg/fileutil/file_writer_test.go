package fileutil

import (
	"errors"
	"os"
	"path"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

/*********************
fakeFileSystem
**********************/
type fakeFileSystem struct {
	*memfs.MemFS
	isFileClosed bool
}

func (f *fakeFileSystem) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	f.SetFileClosed(false)
	file, err := f.MemFS.OpenFile(name, flag, perm)
	return fakeClosedFile{File: file, fs: f}, err
}
func (f *fakeFileSystem) GetFileClosed() bool   { return f.isFileClosed }
func (f *fakeFileSystem) SetFileClosed(in bool) { f.isFileClosed = in }

type fakeClosedFile struct {
	vfs.File
	fs *fakeFileSystem
}

func (u fakeClosedFile) Close() error {
	u.fs.SetFileClosed(true)
	return u.File.Close()
}

/*********************
fakeErrorOpenFileSystem
**********************/
type fakeErrorOpenFileSystem struct {
	fakeFileSystem
}

func (b fakeErrorOpenFileSystem) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	return vfs.DummyFile(errors.New("Cannot Open File")), nil
}

/*********************
fakeErrorWriteFileSystem
**********************/
type fakeErrorWriteFileSystem struct {
	fakeFileSystem
}

func (b *fakeErrorWriteFileSystem) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	b.SetFileClosed(false)
	file, err := b.MemFS.OpenFile(name, flag, perm)
	return unWritableFile{fakeClosedFile{File: file, fs: &b.fakeFileSystem}}, err
}

type unWritableFile struct {
	fakeClosedFile
}

func (unWritableFile) Write(p []byte) (n int, err error) {
	return 0, errors.New("Cannot Write to File")
}

/*********************
TESTS
**********************/

var _ = Describe("FileWriter", func() {
	const knownHostsFilename = "/home/gpadmin/.ssh/known_hosts"
	var (
		fakeFs *memfs.MemFS
		target FileWriter
	)
	BeforeEach(func() {
		fakeFs = memfs.Create()
		target = FileWriter{fakeFs}

		// make sure the file doesn't exist
		_, err := fakeFs.Stat(knownHostsFilename)
		Expect(err).To(HaveOccurred())
	})

	When("file doesn't exist", func() {
		Describe("Append", func() {

			It("creates dir and file with proper permissions", func() {
				err := target.Append(knownHostsFilename, "bar")
				Expect(err).NotTo(HaveOccurred())

				// file permission should be 0600
				fileInfo, _ := fakeFs.Stat(knownHostsFilename)
				Expect(fileInfo.Mode()).To(Equal(os.FileMode(0600)))

				// path permission should be 0711
				dirInfo, _ := fakeFs.Stat(path.Dir(knownHostsFilename))
				Expect(dirInfo.Mode()).To(Equal(os.FileMode(0711)))
			})
			It("creates and writes to the file", func() {
				err := target.Append(knownHostsFilename, "bar")
				Expect(err).NotTo(HaveOccurred())
				knownHosts, err := vfs.ReadFile(fakeFs, knownHostsFilename)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(knownHosts)).To(Equal("bar"))
			})
			It("closes file after creates and writes to it", func() {
				target.WritableFileSystem = &fakeFileSystem{MemFS: fakeFs}
				err := target.Append(knownHostsFilename, "")
				Expect(err).NotTo(HaveOccurred())
				fakeFileSystem, _ := target.WritableFileSystem.(*fakeFileSystem)
				Expect(fakeFileSystem.GetFileClosed()).To(Equal(true))
			})
		})
		Describe("Insert", func() {
			It("creates dir and file with proper permissions", func() {
				err := target.Insert(knownHostsFilename, "bar")
				Expect(err).NotTo(HaveOccurred())

				// file permission should be 0600
				fileInfo, _ := fakeFs.Stat(knownHostsFilename)
				Expect(fileInfo.Mode()).To(Equal(os.FileMode(0600)))

				// path permission should be 0711
				dirInfo, _ := fakeFs.Stat(path.Dir(knownHostsFilename))
				Expect(dirInfo.Mode()).To(Equal(os.FileMode(0711)))
			})
			It("creates and writes to the file", func() {
				err := target.Insert(knownHostsFilename, "bar")
				Expect(err).NotTo(HaveOccurred())
				knownHosts, err := vfs.ReadFile(fakeFs, knownHostsFilename)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(knownHosts)).To(Equal("bar"))
			})
			It("closes file after creates and writes to it", func() {
				target.WritableFileSystem = &fakeFileSystem{MemFS: fakeFs}
				err := target.Insert(knownHostsFilename, "")
				Expect(err).NotTo(HaveOccurred())
				fakeFileSystem, _ := target.WritableFileSystem.(*fakeFileSystem)
				Expect(fakeFileSystem.GetFileClosed()).To(Equal(true))
			})
		})
	})
	When("file exists", func() {
		BeforeEach(func() {
			vfs.MkdirAll(fakeFs, path.Dir(knownHostsFilename), 0711)
			Expect(vfs.WriteFile(fakeFs, knownHostsFilename, []byte("previous key1\nprevious key2\n"), 0666)).To(Succeed())
		})
		Describe("Append", func() {
			It("appends to the existing file", func() {
				err := target.Append(knownHostsFilename, "bar")
				Expect(err).NotTo(HaveOccurred())
				knownHosts, err := vfs.ReadFile(fakeFs, knownHostsFilename)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(knownHosts)).To(Equal("previous key1\nprevious key2\nbar"))
			})
			It("returns error when OpenFile failed", func() {
				target.WritableFileSystem = &fakeErrorOpenFileSystem{fakeFileSystem{MemFS: fakeFs}}
				err := target.Append(knownHostsFilename, "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Cannot Open File"))
			})
			It("returns error when writing to file fails", func() {
				target.WritableFileSystem = &fakeErrorWriteFileSystem{fakeFileSystem{MemFS: fakeFs}}
				err := target.Append(knownHostsFilename, "")
				Expect(err).To(HaveOccurred())
				Expect(string(err.Error())).To(ContainSubstring("Cannot Write to File"))
			})
			It("closes file when fail to write to file", func() {
				target.WritableFileSystem = &fakeErrorWriteFileSystem{fakeFileSystem{MemFS: fakeFs}}
				err := target.Append(knownHostsFilename, "")
				Expect(err).To(HaveOccurred())
				badFileSystem, _ := target.WritableFileSystem.(*fakeErrorWriteFileSystem)
				Expect(badFileSystem.GetFileClosed()).To(Equal(true))
			})
		})
		Describe("Insert", func() {
			It("insert to beginning of the file", func() {
				err := target.Insert(knownHostsFilename, "bar\n")
				Expect(err).NotTo(HaveOccurred())
				knownHosts, err := vfs.ReadFile(fakeFs, knownHostsFilename)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(knownHosts)).To(Equal("bar\nprevious key1\nprevious key2\n"))
			})
			It("returns error when OpenFile failed", func() {
				target.WritableFileSystem = &fakeErrorOpenFileSystem{fakeFileSystem{MemFS: fakeFs}}
				err := target.Insert(knownHostsFilename, "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Cannot Open File"))
			})
			It("returns error when writing to file fails", func() {
				target.WritableFileSystem = &fakeErrorWriteFileSystem{fakeFileSystem{MemFS: fakeFs}}
				err := target.Insert(knownHostsFilename, "")
				Expect(err).To(HaveOccurred())
				Expect(string(err.Error())).To(ContainSubstring("Cannot Write to File"))
			})
			It("closes file when fail to write to file", func() {
				target.WritableFileSystem = &fakeErrorWriteFileSystem{fakeFileSystem{MemFS: fakeFs}}
				err := target.Insert(knownHostsFilename, "")
				Expect(err).To(HaveOccurred())
				badFileSystem, _ := target.WritableFileSystem.(*fakeErrorWriteFileSystem)
				Expect(badFileSystem.GetFileClosed()).To(Equal(true))
			})

		})
	})
})
