package fileutil

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/blang/vfs"
)

type FileWritable interface {
	Append(filename string, toAppend string) error
	Insert(filename string, toInsert string) error
}

type FileWriter struct {
	WritableFileSystem vfs.Filesystem
}

func (f FileWriter) Insert(filename string, toInsert string) error {
	vfs.MkdirAll(f.WritableFileSystem, path.Dir(filename), 0711)

	file, err := f.WritableFileSystem.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	var contents []byte
	contents, err = ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// overwrite contents
	file.Seek(0, 0)

	for _, buf := range [][]byte{
		[]byte(toInsert),
		contents,
	} {
		_, err = file.Write(buf)
		if err != nil {
			file.Close()
			return err
		}
	}

	return file.Close()
}

func (f FileWriter) Append(filename string, toAppend string) error {
	vfs.MkdirAll(f.WritableFileSystem, path.Dir(filename), 0711)
	file, err := f.WritableFileSystem.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(toAppend))
	if err != nil {
		file.Close()
		return err
	}
	return file.Close()
}
