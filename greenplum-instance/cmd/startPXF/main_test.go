package main

import (
	"errors"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startPXF/s3downloader"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/testing/matcher"
)

var _ = Describe("Run", func() {
	var (
		pxfStarter     *PXFStarter
		fakeCommand    *commandable.CommandFake
		memoryfs       *memfs.MemFS
		stdout, stderr *gbytes.Buffer
		fakeDownloader *fakeS3Downloader
		env            map[string]string
	)
	BeforeEach(func() {
		fakeCommand = commandable.NewFakeCommand()
		env = map[string]string{
			"S3_BUCKET":             "test-bucket",
			"S3_FOLDER":             "test-folder",
			"S3_ENDPOINT":           "test-endpoint",
			"S3_ENDPOINT_IS_SECURE": "true",
			"S3_ACCESS_KEY_ID":      "test-access-key-id",
			"S3_SECRET_ACCESS_KEY":  "test-secret-access-key",
		}
		fakeDownloader = &fakeS3Downloader{}
		memoryfs = memfs.Create()
		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()
		pxfStarter = &PXFStarter{
			Getenv: func(key string) string {
				return env[key]
			},
			DownloaderFactory: func(s3 s3Location) (s3downloader.S3BucketDownloader, error) {
				fakeDownloader.secure = s3.isSecure
				return fakeDownloader, nil
			},
			Command:    fakeCommand.Command,
			Filesystem: memoryfs,
			Stdout:     stdout,
			Stderr:     stderr,
		}
	})
	Describe("Successful Run()", func() {
		var (
			pxfStartCalled int
			tailCalled     int
		)
		BeforeEach(func() {
			fakeCommand.ExpectCommand("/usr/local/pxf-gp6/bin/pxf", "start").
				CallCounter(&pxfStartCalled).
				SideEffect(func() {
					Expect(vfs.MkdirAll(memoryfs, "/usr/local/pxf-gp6/run", 0755))
					Expect(vfs.WriteFile(memoryfs, "/usr/local/pxf-gp6/run/catalina.pid", []byte("\t2345\n"), 0666)).To(Succeed())
				})
			fakeCommand.ExpectCommand("tail", "--pid=2345", "-f", "/etc/pxf/logs/pxf-service.log", "-n", "+0").
				CallCounter(&tailCalled).
				PrintsOutput("pxf webapp started\n").
				PrintsError("tail: lost my tail\n")
		})
		When("Downloader is successful", func() {
			It("succeeds", func() {
				Expect(pxfStarter.Run()).To(Succeed())
				Expect(fakeDownloader.localDirectory).To(Equal("/etc/pxf"))
				Expect(fakeDownloader.bucket).To(Equal("test-bucket"))
				Expect(fakeDownloader.s3Folder).To(Equal("test-folder"))
				Expect(fakeDownloader.secure).To(BeTrue())
				Expect("/usr/local/pxf-gp6/run/catalina.pid").To(matcher.ExistInFilesystem(memoryfs))
				Expect(pxfStartCalled).To(Equal(1))
				Expect(tailCalled).To(Equal(1), fakeCommand.CapturedArgs())
				Expect(stdout).To(gbytes.Say("pxf webapp started\n"))
				Expect(stderr).To(gbytes.Say("tail: lost my tail\n"))
			})
			When("S3_ENDPOINT_IS_SECURE is false", func() {
				BeforeEach(func() {
					env["S3_ENDPOINT_IS_SECURE"] = "false"
				})
				It("creates an insecure downloader", func() {
					Expect(pxfStarter.Run()).To(Succeed())
					Expect(fakeDownloader.secure).To(BeFalse())
				})
			})
		})
		When("Downloader can't be created", func() {
			BeforeEach(func() {
				pxfStarter.DownloaderFactory = func(_ s3Location) (s3downloader.S3BucketDownloader, error) {
					return nil, errors.New("custom downloader error")
				}
			})
			It("returns the error", func() {
				Expect(pxfStarter.Run()).To(MatchError("custom downloader error"))
			})
		})
		When("S3 download fails", func() {
			BeforeEach(func() {
				fakeDownloader.Err = errors.New("custom S3 Downloader error")
			})
			It("logs a message", func() {
				Expect(pxfStarter.Run()).To(MatchError("custom S3 Downloader error"))
			})
		})
		When("no s3 config is specified", func() {
			var downloaderCreated bool
			BeforeEach(func() {
				pxfStarter.Getenv = func(key string) string {
					return ""
				}
				pxfStarter.DownloaderFactory = func(_ s3Location) (s3downloader.S3BucketDownloader, error) {
					downloaderCreated = true
					return nil, nil
				}
			})
			It("does not run/create the downloader", func() {
				Expect(pxfStarter.Run()).To(Succeed())
				Expect(downloaderCreated).To(BeFalse())
			})
		})
	})
	When("PXF start errors", func() {
		BeforeEach(func() {
			fakeCommand.ExpectCommand("/usr/local/pxf-gp6/bin/pxf", "start").ReturnsStatus(1)
		})
		It("returns error", func() {
			err := pxfStarter.Run()
			Expect(err).To(MatchError("exit status 1"))
		})
	})

	When("unable to read catalina.pid file", func() {
		BeforeEach(func() {
			fakeCommand.ExpectCommand("/usr/local/pxf-gp6/bin/pxf", "start")
			Expect("/usr/local/pxf-gp6/run/catalina.pid").NotTo(matcher.ExistInFilesystem(memoryfs))
		})
		It("errors out", func() {
			err := pxfStarter.Run()
			Expect(err).To(MatchError("reading catalina.pid file: open /usr/local/pxf-gp6/run/catalina.pid: file does not exist"))
		})
	})

	When("tail on logs file fails", func() {
		BeforeEach(func() {
			fakeCommand.ExpectCommand("/usr/local/pxf-gp6/bin/pxf", "start").
				SideEffect(func() {
					Expect(vfs.MkdirAll(memoryfs, "/usr/local/pxf-gp6/run", 0755))
					Expect(vfs.WriteFile(memoryfs, "/usr/local/pxf-gp6/run/catalina.pid", []byte("\t2345\n"), 0666)).To(Succeed())
				})
			fakeCommand.ExpectCommand("tail", "--pid=2345", "-f", "/etc/pxf/logs/pxf-service.log", "-n", "+0").
				ReturnsStatus(1)
		})
		It("errors out", func() {
			err := pxfStarter.Run()
			Expect(err).To(MatchError("reading logs: exit status 1"))
		})
	})
})

type fakeS3Downloader struct {
	bucket         string
	s3Folder       string
	localDirectory string
	secure         bool
	Err            error
}

func (f *fakeS3Downloader) DownloadBucketPath(bucket string, s3Folder string, localDirectory string) error {
	f.bucket = bucket
	f.s3Folder = s3Folder
	f.localDirectory = localDirectory
	return f.Err
}
