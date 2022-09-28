package testing

import "github.com/pivotal/greenplum-for-kubernetes/pkg/ubuntuUtils"

type MockUbuntu struct {
	ChangeDirectoryOwnerMock struct {
		DirName  string
		UserName string
		Err      error
	}
	HostnameMock struct {
		Hostname string
		Err      error
	}
}

var _ = ubuntuUtils.UbuntuInterface(&MockUbuntu{}) // MockUbuntu is an UbuntuInterface

func (f *MockUbuntu) ChangeDirectoryOwner(dirName string, userName string) error {
	f.ChangeDirectoryOwnerMock.DirName = dirName
	f.ChangeDirectoryOwnerMock.UserName = userName
	return f.ChangeDirectoryOwnerMock.Err
}

func (f *MockUbuntu) Hostname() (name string, err error) {
	return f.HostnameMock.Hostname, f.HostnameMock.Err
}
