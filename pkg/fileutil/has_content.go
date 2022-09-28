package fileutil

import (
	"os"

	"github.com/blang/vfs"
)

// returns false if source is directory, or file does not exist, or is empty
func HasContent(filesystem vfs.Filesystem, source string) (bool, error) {
	stat, err := filesystem.Lstat(source)
	if err != nil {
		if err.(*os.PathError).Err == os.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	if stat.IsDir() {
		return false, os.ErrInvalid
	}
	return stat.Size() > 0, nil
}
