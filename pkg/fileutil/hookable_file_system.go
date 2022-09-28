package fileutil

import (
	"os"

	"github.com/blang/vfs"
)

type HookableFilesystem struct {
	vfs.Filesystem
	OpenFileHook func(name string, flag int, perm os.FileMode) (vfs.File, error)
	SymlinkHook  func(oldname, newname string) error
	MkdirHook    func(name string, perm os.FileMode) error
	StatHook     func(name string) (os.FileInfo, error)
}

func (fs *HookableFilesystem) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	if fs.OpenFileHook != nil {
		return fs.OpenFileHook(name, flag, perm)
	}
	return fs.Filesystem.OpenFile(name, flag, perm)
}

func (fs *HookableFilesystem) Symlink(oldname, newname string) error {
	if fs.SymlinkHook != nil {
		return fs.SymlinkHook(oldname, newname)
	}
	return fs.Filesystem.Symlink(oldname, newname)
}

func (fs *HookableFilesystem) Mkdir(name string, perm os.FileMode) error {
	if fs.MkdirHook != nil {
		return fs.MkdirHook(name, perm)
	}
	return fs.Filesystem.Mkdir(name, perm)
}

func (fs *HookableFilesystem) Stat(name string) (os.FileInfo, error) {
	if fs.StatHook != nil {
		return fs.StatHook(name)
	}
	return fs.Filesystem.Stat(name)
}
