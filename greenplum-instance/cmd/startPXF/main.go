package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v6"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startPXF/s3downloader"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pkg/errors"
)

var log = ctrllog.Log.WithName("startPXF")

type S3DownloaderFactory func(s3 s3Location) (s3downloader.S3BucketDownloader, error)

type PXFStarter struct {
	Getenv            func(string) string
	Command           commandable.CommandFn
	Filesystem        vfs.Filesystem
	DownloaderFactory S3DownloaderFactory
	Stdout, Stderr    io.Writer
}

func main() {
	ctrllog.SetLogger(gplog.ForProd(false))
	pxfStarter := PXFStarter{
		Getenv:            os.Getenv,
		Command:           commandable.Command,
		Filesystem:        vfs.OS(),
		DownloaderFactory: DownloaderFactory,
		Stdout:            os.Stdout,
		Stderr:            os.Stderr,
	}
	if err := pxfStarter.Run(); err != nil {
		log.Error(err, "PXF failed to start")
		os.Exit(1)
	}
}

type EnvMissingError struct {
	env string
}

func (e EnvMissingError) Error() string {
	return fmt.Sprintf("%v is empty", e.env)
}

var _ error = EnvMissingError{}

type s3Location struct {
	accessKeyID     string
	secretAccessKey string
	endpoint        string
	isSecure        bool
	bucket          string
	folder          string
}

func (p *PXFStarter) readS3LocationFromEnv() *s3Location {
	s3 := &s3Location{}

	for _, env := range []struct {
		envName string
		strRef  *string
		boolRef *bool
	}{
		{envName: "S3_ACCESS_KEY_ID", strRef: &s3.accessKeyID},
		{envName: "S3_SECRET_ACCESS_KEY", strRef: &s3.secretAccessKey},
		{envName: "S3_ENDPOINT", strRef: &s3.endpoint},
		{envName: "S3_ENDPOINT_IS_SECURE", boolRef: &s3.isSecure},
		{envName: "S3_BUCKET", strRef: &s3.bucket},
		{envName: "S3_FOLDER", strRef: &s3.folder},
	} {
		value := p.Getenv(env.envName)
		if env.strRef != nil {
			*env.strRef = value
		} else if env.boolRef != nil {
			b, _ := strconv.ParseBool(value)
			*env.boolRef = b
		}
	}

	return s3
}

func DownloaderFactory(s3 s3Location) (s3downloader.S3BucketDownloader, error) {
	minioClient, err := minio.New(
		s3.endpoint,
		s3.accessKeyID,
		s3.secretAccessKey,
		s3.isSecure,
	)
	if err != nil {
		return nil, err
	}
	return &s3downloader.S3BucketDownloaderImpl{
		Filesystem: vfs.OS(),
		Client:     &s3downloader.MinioS3Client{Client: minioClient},
	}, nil
}

func (p *PXFStarter) Run() error {
	if s3 := p.readS3LocationFromEnv(); s3.endpoint != "" {
		log.Info("downloading PXF configuration", "bucket", s3.bucket, "folder", s3.folder)
		downloader, err := p.DownloaderFactory(*s3)
		if err != nil {
			return err
		}

		if err := downloader.DownloadBucketPath(s3.bucket, s3.folder, "/etc/pxf"); err != nil {
			return err
		}
	}
	log.Info("starting pxf")
	cmd := p.Command("/usr/local/pxf-gp6/bin/pxf", "start")
	err := cmd.Run()
	if err != nil {
		return err
	}

	const pidFile = "/usr/local/pxf-gp6/run/catalina.pid"
	pidBytes, err := vfs.ReadFile(p.Filesystem, pidFile)
	if err != nil {
		return errors.Wrap(err, "reading catalina.pid file")
	}
	cmd = p.Command("tail", "--pid="+strings.TrimSpace(string(pidBytes)), "-f", "/etc/pxf/logs/pxf-service.log", "-n", "+0")
	cmd.Stdout = p.Stdout
	cmd.Stderr = p.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "reading logs")
	}
	return nil
}
