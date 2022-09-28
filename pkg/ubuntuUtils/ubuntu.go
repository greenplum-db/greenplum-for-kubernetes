package ubuntuUtils

import (
	"os"
	"strconv"
)

type UbuntuInterface interface {
	ChangeDirectoryOwner(dirName string, userName string) error
	Hostname() (name string, err error)
}

type Ubuntu struct {
	sysFunctions *SysFunctions
}

var _ UbuntuInterface = Ubuntu{}

func (u Ubuntu) ChangeDirectoryOwner(dirName string, userName string) error {
	user, err := u.sysFunctions.LookupUser(userName)
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		return err
	}
	return u.sysFunctions.Walk(dirName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return u.sysFunctions.Chown(path, uid, gid)
	})
}

// Hostname is wrapper for os.Hostname that can be replaced for testing.
func (u Ubuntu) Hostname() (name string, err error) {
	return os.Hostname()
}

func NewRealUbuntu() Ubuntu {
	return NewUbuntu(NewRealSysFunctions())
}

func NewUbuntu(sysFunctions *SysFunctions) Ubuntu {
	return Ubuntu{sysFunctions: sysFunctions}
}
