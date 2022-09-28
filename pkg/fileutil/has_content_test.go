package fileutil

import (
	"os"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HasContent test", func() {
	var memFS vfs.Filesystem
	BeforeEach(func() {
		memFS = memfs.Create()
	})
	When("there is no file", func() {
		It("returns false", func() {
			result, err := HasContent(memFS, "/foo")
			Expect(result).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("file exists and has some content", func() {
		It("return true", func() {
			vfs.WriteFile(memFS, "/foo", []byte{'0'}, 0644)
			result, err := HasContent(memFS, "/foo")
			Expect(result).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("file exists and has no content", func() {
		It("return false", func() {
			vfs.WriteFile(memFS, "/foo", []byte{}, 0644)
			result, err := HasContent(memFS, "/foo")
			Expect(result).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("path is a directory", func() {
		It("return error", func() {
			vfs.MkdirAll(memFS, "/a", 0000)
			//Expect(vfs.WriteFile(memFS, "/a/foo", []byte{'0'}, 0000)).NotTo(Succeed())
			_, err := HasContent(memFS, "/a")
			Expect(err).To(HaveOccurred())
		})
	})
	When("path is a symlink that is not dir but has no parent", func() {
		It("return error", func() {
			fakeLStatFS := &fakeLStatFileSystem{fakeFileSystem{MemFS: memfs.Create()}}
			_, err := HasContent(fakeLStatFS, "/arbitrarysymlink")
			Expect(err).To(HaveOccurred())
		})
	})
})

type fakeLStatFileSystem struct {
	fakeFileSystem
}

func (f *fakeLStatFileSystem) Lstat(name string) (os.FileInfo, error) {
	return nil, &os.PathError{Op: "fake", Path: name, Err: os.ErrInvalid}
}
