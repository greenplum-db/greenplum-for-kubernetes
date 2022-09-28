package s3downloader_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestS3downloader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3downloader Suite")
}
