package s3downloader

import (
	"io"
	"strings"

	"github.com/minio/minio-go/v6"
)

type fakeS3Client struct {
	objectInfos     map[string]minio.ObjectInfo
	objectDatas     map[string]string
	listObjectsStub struct {
		Err error
	}
	getObjectStub struct {
		Reader io.Reader
		Err    error
	}
}

func newFakeS3Client(objects map[string]string) *fakeS3Client {
	infos := map[string]minio.ObjectInfo{}
	for name, data := range objects {
		infos[name] = minio.ObjectInfo{
			Key:  name,
			Size: int64(len(data)),
		}
	}

	return &fakeS3Client{
		objectInfos: infos,
		objectDatas: objects,
	}
}

func (f *fakeS3Client) ListObjects(bucketName, objectPrefix string, recursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo {
	objectCh := make(chan minio.ObjectInfo)
	go func() {
		if f.listObjectsStub.Err != nil {
			objectCh <- minio.ObjectInfo{Err: f.listObjectsStub.Err}
		} else {
			for name, objectInfo := range f.objectInfos {
				if objectPrefix != "" {
					if strings.HasPrefix(name, objectPrefix) {
						objectCh <- objectInfo
					}
					continue
				}
				objectCh <- objectInfo
			}
		}
		close(objectCh)
	}()
	return objectCh
}

func (f *fakeS3Client) GetObject(bucketName, objectName string, opts minio.GetObjectOptions) (io.Reader, error) {
	if f.getObjectStub.Reader != nil {
		return f.getObjectStub.Reader, f.getObjectStub.Err
	}
	return strings.NewReader(f.objectDatas[objectName]), f.getObjectStub.Err
}
