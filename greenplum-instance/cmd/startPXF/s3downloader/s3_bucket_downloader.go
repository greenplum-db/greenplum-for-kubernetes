package s3downloader

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/blang/vfs"
	"github.com/minio/minio-go/v6"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = ctrllog.Log.WithName("s3downloader")

const MalformedConfDirWarning = "warning: PXF S3 conf data doesn't look like a PXF conf directory. Please check pxfConf.s3Source setting"

type S3BucketDownloader interface {
	DownloadBucketPath(bucket string, s3Folder string, localDirectory string) error
}

type S3BucketDownloaderImpl struct {
	Filesystem vfs.Filesystem
	Client     S3Client
}

var KnownPxfConfSubdirs = map[string]bool{
	"conf":      true,
	"keytabs":   true,
	"lib":       true,
	"logs":      true,
	"servers":   true,
	"templates": true,
}

func (s *S3BucketDownloaderImpl) DownloadBucketPath(bucket string, s3Folder string, localDirectory string) error {
	doneCh := make(chan struct{})
	defer close(doneCh)

	numObjects := 0
	sawKnownPXFConfSubdir := false
	objectCh := s.Client.ListObjects(bucket, s3Folder, true, doneCh)
	for info := range objectCh {
		if info.Err != nil {
			return fmt.Errorf("s3 list objects: %w", info.Err)
		}
		if info.Size > 0 && !strings.HasSuffix(info.Key, "/") {
			numObjects++
			log.Info("downloading", "object", info.Key)
			obj, err := s.Client.GetObject(bucket, info.Key, minio.GetObjectOptions{})
			if err != nil {
				return fmt.Errorf("s3 get object: %w", err)
			}
			fileKey := info.Key
			if s3Folder != "" {
				fileKey = strings.TrimPrefix(fileKey, s3Folder)
				fileKey = strings.TrimPrefix(fileKey, "/")
			}
			if KnownPxfConfSubdirs[firstPathComponent(fileKey)] {
				sawKnownPXFConfSubdir = true
			}
			filename := path.Join(localDirectory, fileKey)
			err = vfs.MkdirAll(s.Filesystem, path.Dir(filename), os.ModePerm)
			if err != nil {
				return fmt.Errorf("making download destination directory: %w", err)
			}
			file, err := s.Filesystem.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("open destination file: %w", err)
			}
			numBytes, err := io.Copy(file, obj)
			if err != nil {
				return fmt.Errorf("writing destination file: %w", err)
			}
			if numBytes < info.Size {
				return fmt.Errorf("wrote %d bytes to %#v but expected to write %d bytes", numBytes, filename, info.Size)
			}
		}
	}
	if numObjects == 0 {
		return fmt.Errorf("no objects found in pxfConf.s3Source location: %s/%s", bucket, s3Folder)
	}
	if !sawKnownPXFConfSubdir {
		log.Info("warning: PXF S3 conf data doesn't look like a PXF conf directory. Please check pxfConf.s3Source setting")
	}
	return nil
}

// like path.Split, but from the other end
func firstPathComponent(path string) string {
	i := strings.Index(path, "/")
	if i < 0 {
		return path
	}
	return path[0:i]
}
