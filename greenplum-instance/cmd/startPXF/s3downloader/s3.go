package s3downloader

import (
	"io"

	"github.com/minio/minio-go/v6"
)

type S3Client interface {
	ListObjects(bucketName, objectPrefix string, recursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo
	GetObject(bucketName, objectName string, opts minio.GetObjectOptions) (io.Reader, error)
}

type MinioS3Client struct {
	*minio.Client
}

var _ S3Client = &MinioS3Client{}

// GetObject implements S3Client by changing the return type of the embedded minio.Client.GetObject to io.Reader
// which makes it easier to mock for testing.
func (m *MinioS3Client) GetObject(bucketName, objectName string, opts minio.GetObjectOptions) (io.Reader, error) {
	return m.Client.GetObject(bucketName, objectName, opts)
}
