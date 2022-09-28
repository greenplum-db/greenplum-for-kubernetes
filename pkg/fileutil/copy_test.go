package fileutil_test

import (
	"os"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/testing/matcher"
)

var _ = Describe("CopyFile", func() {
	var (
		fs vfs.Filesystem
	)
	BeforeEach(func() {
		fs = memfs.Create()
	})

	It("Copies a file", func() {
		Expect(vfs.MkdirAll(fs, "foo", 0755)).To(Succeed())
		Expect(vfs.MkdirAll(fs, "baz", 0755)).To(Succeed())
		Expect(vfs.WriteFile(fs, "foo/bar", []byte("Hello, world!"), 0644)).To(Succeed())

		Expect(fileutil.CopyFile(fs, "foo/bar", "baz/quux")).To(Succeed())
		Expect("baz/quux").To(EqualInFilesystem(fs, "Hello, world!"))
	})

	It("Duplicates the permission bits of the source file", func() {
		Expect(vfs.MkdirAll(fs, "srcdir", 0755)).To(Succeed())
		Expect(vfs.MkdirAll(fs, "dstdir", 0755)).To(Succeed())

		Expect(vfs.WriteFile(fs, "srcdir/file1", []byte("I am a secret file!\n"), 0600)).To(Succeed())
		Expect(fileutil.CopyFile(fs, "srcdir/file1", "dstdir/file1")).To(Succeed())
		stat, err := fs.Stat("dstdir/file1")
		Expect(err).NotTo(HaveOccurred())
		Expect(stat.Mode() & os.ModePerm).To(Equal(os.FileMode(0600)))

		Expect(vfs.WriteFile(fs, "srcdir/file2", []byte("I am an evil public file!\n"), 0666)).To(Succeed())
		Expect(fileutil.CopyFile(fs, "srcdir/file2", "dstdir/file2")).To(Succeed())
		stat, err = fs.Stat("dstdir/file2")
		Expect(err).NotTo(HaveOccurred())
		Expect(stat.Mode() & os.ModePerm).To(Equal(os.FileMode(0666)))
	})

	When("The source file does not exist", func() {
		It("returns an error", func() {
			err := fileutil.CopyFile(fs, "foo/bar", "baz/quux")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("foo/bar: file does not exist"))
		})
	})

	When("The destination can not be written", func() {
		BeforeEach(func() {
			Expect(vfs.MkdirAll(fs, "foo", 0755)).To(Succeed())
			Expect(vfs.WriteFile(fs, "foo/bar", []byte("Hello, world!"), 0644)).To(Succeed())
		})
		It("returns an error", func() {
			err := fileutil.CopyFile(fs, "foo/bar", "baz/quux")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("baz/quux: file does not exist"))
		})
	})

	When("There is an IO error", func() {
		BeforeEach(func() {
			Expect(vfs.WriteFile(fs, "srcfile", []byte("Hi"), 0644)).To(Succeed())

			hookfs := &fileutil.HookableFilesystem{Filesystem: fs}
			// If all files are read-only, we will get an error when io.Copy tries to write to the destination file
			hookfs.OpenFileHook = func(name string, flag int, perm os.FileMode) (file vfs.File, e error) {
				file, e = hookfs.Filesystem.OpenFile(name, flag, perm)
				return vfs.ReadOnlyFile(file), e
			}
			fs = hookfs
		})
		It("returns an error", func() {
			err := fileutil.CopyFile(fs, "srcfile", "dstfile")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Filesystem is read-only"))
		})
	})

})
