package ubuntuUtils

import (
	"os"
	"os/user"
	"path/filepath"
)

type SysFunctions struct {
	Chown      func(name string, uid, gid int) error
	LookupUser func(username string) (*user.User, error)
	Walk       func(root string, walkFn filepath.WalkFunc) error
}

func NewRealSysFunctions() *SysFunctions {
	return &SysFunctions{
		LookupUser: user.Lookup,
		Chown:      os.Chown,
		Walk:       filepath.Walk,
	}
}
