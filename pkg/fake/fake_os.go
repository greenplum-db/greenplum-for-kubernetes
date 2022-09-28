package fake

import (
	"errors"
	"os"
)

type FakeOsSuccess struct {
	Path string
	Perm os.FileMode
}

func (f *FakeOsSuccess) FakeMkDirAll(path string, perm os.FileMode) error {
	f.Path = path
	f.Perm = perm
	return nil
}

type FakeOsFails struct{}

func (f *FakeOsFails) FakeMkDirAll(path string, perm os.FileMode) error {
	return errors.New("This is an error")
}
