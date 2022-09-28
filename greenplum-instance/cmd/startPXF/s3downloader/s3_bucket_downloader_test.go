package s3downloader

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gstruct"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog/testing"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/testing/matcher"
)

var _ = Describe("S3BucketDownloader Operations", func() {
	var (
		s3ops             *S3BucketDownloaderImpl
		fakeClient        *fakeS3Client
		memoryFS          *fileutil.HookableFilesystem
		objectInfoDataMap map[string]string
		logbuf            *gbytes.Buffer
	)

	BeforeEach(func() {
		logbuf = gbytes.NewBuffer()
		log = gplog.ForTest(logbuf)
		memoryFS = &fileutil.HookableFilesystem{Filesystem: memfs.Create()}
		objectInfoDataMap = map[string]string{
			"servers/":              "",
			"servers/my-server.xml": "There are many like it, but this one is mine</s>",
		}
		fakeClient = newFakeS3Client(objectInfoDataMap)
		s3ops = &S3BucketDownloaderImpl{
			Filesystem: memoryFS,
			Client:     fakeClient,
		}
	})

	When("the user specifies a conf location that appears to contain valid conf directories", func() {
		DescribeTable("downloads and succeeds without logging a warning",
			func(goodDirName string) {
				objectInfoDataMap = map[string]string{
					"object1":            "I'm Mr. Bucket!",
					goodDirName + "/":    "",
					goodDirName + "/cat": "dog",
				}
				fakeClient = newFakeS3Client(objectInfoDataMap)
				s3ops = &S3BucketDownloaderImpl{
					Filesystem: memoryFS,
					Client:     fakeClient,
				}
				Expect(s3ops.DownloadBucketPath("mr-bucket", "", "localdir")).To(Succeed())
				Expect("localdir/object1").To(matcher.EqualInFilesystem(memoryFS, "I'm Mr. Bucket!"))
				Expect("localdir/" + goodDirName + "/cat").To(matcher.EqualInFilesystem(memoryFS, "dog"))
				logs, err := testing.DecodeLogs(logbuf)
				Expect(err).NotTo(HaveOccurred())
				Expect(logs).To(testing.ContainLogEntry(gstruct.Keys{"msg": Equal("downloading"), "object": Equal("object1")}))
				Expect(logs).To(testing.ContainLogEntry(gstruct.Keys{"msg": Equal("downloading"), "object": Equal(goodDirName + "/cat")}))
				Expect(logs).NotTo(testing.ContainLogEntry(gstruct.Keys{"msg": Equal(MalformedConfDirWarning)}))
			},
			Entry("contains a conf dir", "conf"),
			Entry("contains a keytabs dir", "keytabs"),
			Entry("contains a lib dir", "lib"),
			Entry("contains a logs dir", "logs"),
			Entry("contains a servers dir", "servers"),
			Entry("contains a templates dir", "templates"),
		)
	})

	When("the user specifies an s3 folder", func() {
		BeforeEach(func() {
			objectInfoDataMap = map[string]string{
				"object1":                     "I'm Mr. Bucket!",
				"s3dir/":                      "",
				"s3dir/object2":               "I'm inner bucket",
				"s3dir/servers/my-server.xml": "There are many like it, but this one is mine</s>",
			}
			fakeClient = newFakeS3Client(objectInfoDataMap)
			s3ops = &S3BucketDownloaderImpl{
				Filesystem: memoryFS,
				Client:     fakeClient,
			}
		})

		DescribeTable("downloads folders in s3",
			func(s3Folder string) {
				Expect(s3ops.DownloadBucketPath("mr-bucket", s3Folder, "localdir")).To(Succeed())
				Expect("localdir/object1").NotTo(matcher.ExistInFilesystem(memoryFS))
				Expect("localdir/s3dir/object2").NotTo(matcher.ExistInFilesystem(memoryFS))
				Expect("localdir/object2").To(matcher.EqualInFilesystem(memoryFS, "I'm inner bucket"))
				Expect("localdir/servers/my-server.xml").To(matcher.EqualInFilesystem(memoryFS, "There are many like it, but this one is mine</s>"))
				logs, err := testing.DecodeLogs(logbuf)
				Expect(err).NotTo(HaveOccurred())
				Expect(logs).To(testing.ContainLogEntry(gstruct.Keys{"msg": Equal("downloading"), "object": Equal("s3dir/object2")}))
				Expect(logs).To(testing.ContainLogEntry(gstruct.Keys{"msg": Equal("downloading"), "object": Equal("s3dir/servers/my-server.xml")}))
				Expect(logs).NotTo(testing.ContainLogEntry(gstruct.Keys{"msg": Equal(MalformedConfDirWarning)}))
			},
			Entry("no trailing /", "s3dir"),
			Entry("with trailing /", "s3dir/"),
		)
	})

	When("the user specifies a gcs bucket with subdirectories", func() {
		BeforeEach(func() {
			objectInfoDataMap = map[string]string{
				"object1":               "I'm Mr. Bucket!",
				"gcsdir/":               "test",
				"gcsdir/gcsobject":      "I'm inner bucket",
				"servers/my-server.xml": "There are many like it, but this one is mine</s>",
			}
			fakeClient = newFakeS3Client(objectInfoDataMap)
			s3ops = &S3BucketDownloaderImpl{
				Filesystem: memoryFS,
				Client:     fakeClient,
			}
		})
		It("succeeds", func() {
			Expect(s3ops.DownloadBucketPath("mr-bucket", "", "localdir")).To(Succeed())
			Expect("localdir/object1").To(matcher.EqualInFilesystem(memoryFS, "I'm Mr. Bucket!"))
			Expect("localdir/gcsdir/gcsobject").To(matcher.EqualInFilesystem(memoryFS, "I'm inner bucket"))
			Expect("localdir/servers/my-server.xml").To(matcher.EqualInFilesystem(memoryFS, "There are many like it, but this one is mine</s>"))
			logs, err := testing.DecodeLogs(logbuf)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs).To(testing.ContainLogEntry(gstruct.Keys{"msg": Equal("downloading"), "object": Equal("object1")}))
			Expect(logs).To(testing.ContainLogEntry(gstruct.Keys{"msg": Equal("downloading"), "object": Equal("gcsdir/gcsobject")}))
			Expect(logs).To(testing.ContainLogEntry(gstruct.Keys{"msg": Equal("downloading"), "object": Equal("servers/my-server.xml")}))
			Expect(logs).NotTo(testing.ContainLogEntry(gstruct.Keys{"msg": Equal(MalformedConfDirWarning)}))
		})
	})

	When("the user specifies a conf location that doesn't appear to contain valid conf directories", func() {
		BeforeEach(func() {
			objectInfoDataMap = map[string]string{
				"redherring/":     "",
				"redherring/yuck": "gross",
			}
			fakeClient = newFakeS3Client(objectInfoDataMap)
			s3ops = &S3BucketDownloaderImpl{
				Filesystem: memoryFS,
				Client:     fakeClient,
			}
		})
		It("downloads the data, but logs a warning", func() {
			Expect(s3ops.DownloadBucketPath("mr-bucket", "", "localdir")).To(Succeed())
			Expect("localdir/redherring/yuck").To(matcher.EqualInFilesystem(memoryFS, "gross"))
			logs, err := testing.DecodeLogs(logbuf)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs).To(testing.ContainLogEntry(gstruct.Keys{"msg": Equal("downloading"), "object": Equal("redherring/yuck")}))
			Expect(logs).To(testing.ContainLogEntry(gstruct.Keys{"msg": Equal(MalformedConfDirWarning)}))
		})
	})

	When("ListObjects encounters an error", func() {
		BeforeEach(func() {
			fakeClient.listObjectsStub.Err = errors.New("list objects error")
		})
		It("returns the error", func() {
			Expect(s3ops.DownloadBucketPath("mr-bucket", "", "localDir")).To(MatchError("s3 list objects: list objects error"))
		})
	})

	When("there are no objects found in ListObjects", func() {
		BeforeEach(func() {
			s3ops.Client = newFakeS3Client(nil)
		})
		It("returns an error including the bucket", func() {
			Expect(s3ops.DownloadBucketPath("mr-bucket", "", "localDir")).To(MatchError("no objects found in pxfConf.s3Source location: mr-bucket/"))
		})
		It("returns an error including the folder", func() {
			Expect(s3ops.DownloadBucketPath("mr-bucket", "not-a-set", "localDir")).To(MatchError("no objects found in pxfConf.s3Source location: mr-bucket/not-a-set"))
		})
	})

	When("GetObject encounters an error", func() {
		BeforeEach(func() {
			fakeClient.getObjectStub.Err = errors.New("get object error")
		})
		It("returns the error", func() {
			Expect(s3ops.DownloadBucketPath("mr-bucket", "", "localdir")).To(MatchError("s3 get object: get object error"))
		})
	})

	When("MkdirAll encounters an error", func() {
		BeforeEach(func() {
			memoryFS.MkdirHook = func(name string, perm os.FileMode) error {
				return errors.New("mkdirall error")
			}
		})
		It("returns the error", func() {
			Expect(s3ops.DownloadBucketPath("mr-bucket", "", "localdir")).To(MatchError("making download destination directory: mkdirall error"))
		})
	})

	When("OpenFile encounters an error", func() {
		BeforeEach(func() {
			memoryFS.OpenFileHook = func(name string, flag int, perm os.FileMode) (vfs.File, error) {
				return nil, errors.New("open file error")
			}
		})
		It("returns the error", func() {
			Expect(s3ops.DownloadBucketPath("mr-bucket", "", "localdir")).To(MatchError("open destination file: open file error"))
		})
	})

	When("Copy encounters an error", func() {
		BeforeEach(func() {
			fakeClient.getObjectStub.Reader = failingReader{}
		})
		It("returns the error", func() {
			Expect(s3ops.DownloadBucketPath("mr-bucket", "", "localdir")).To(MatchError("writing destination file: I am a failure"))
		})
	})

	When("Copy writes less than expected", func() {
		BeforeEach(func() {
			fakeClient.getObjectStub.Reader = strings.NewReader("short")
		})
		It("returns an error", func() {
			err := s3ops.DownloadBucketPath("mr-bucket", "", "localdir")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(`wrote 5 bytes to "localdir/.*" but expected to write \d* bytes`))
		})
	})
})

type failingReader struct{}

var _ io.Reader = failingReader{}

func (f failingReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("I am a failure")
}
