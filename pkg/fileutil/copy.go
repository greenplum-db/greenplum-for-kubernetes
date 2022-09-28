package fileutil

import (
	"io"
	"os"

	"github.com/blang/vfs"
	"github.com/pkg/errors"
)

func CopyFile(fs vfs.Filesystem, src, dest string) error {
	srcStat, _ := fs.Stat(src)
	// Any error here (EACCES/ENOENT) would also be returned by Open, assuming no races.
	// If we were using os.File, then we would call srcFile.Stat() instead after open,
	// which removes the race. But blang/vfs has chosen to omit Stat() from its File interface.
	srcFile, err := vfs.Open(fs, src)
	if err != nil {
		return errors.Wrapf(err, "copy")
	}
	dstFile, err := fs.OpenFile(dest, os.O_WRONLY|os.O_CREATE, srcStat.Mode()&os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "copy")
	}
	_, err = io.Copy(dstFile, srcFile)
	return errors.Wrapf(err, "copy")
}
