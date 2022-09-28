package ubuntuUtils_test

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/ubuntuUtils"
)

var _ = Describe("ChangeDirectoryOwner", func() {
	var userName, dirPath string
	var fileInfo os.FileInfo
	var userLookupSuccess func(username string) (*user.User, error)
	var fakeSystemFunctions *ubuntuUtils.SysFunctions
	var ubuntu ubuntuUtils.Ubuntu
	var chown struct {
		FilePath string
		UID      int
		GID      int
	}
	BeforeEach(func() {
		userLookupSuccess = func(username string) (*user.User, error) {
			userName = username
			return &user.User{Name: userName, Gid: "10", Uid: "20"}, nil
		}
		dirWalkerSuccess := func(root string, walkFunction filepath.WalkFunc) error {
			dirPath = root
			_, fileName, _, _ := runtime.Caller(0)
			fileInfo, _ = os.Lstat(fileName)
			return walkFunction(root+"/bar", fileInfo, nil)
		}
		fakeChown := func(path string, uid int, gid int) error {
			chown.FilePath = path
			chown.UID = uid
			chown.GID = gid
			return nil
		}
		fakeSystemFunctions = &ubuntuUtils.SysFunctions{
			LookupUser: userLookupSuccess,
			Walk:       dirWalkerSuccess,
			Chown:      fakeChown,
		}
		ubuntu = ubuntuUtils.NewUbuntu(fakeSystemFunctions)
	})
	It("walks the directory", func() {
		Expect(ubuntu.ChangeDirectoryOwner("foo", "bob")).To(Succeed())
		Expect(userName).To(Equal("bob"))
		Expect(dirPath).To(Equal("foo"))
		Expect(chown.FilePath).To(Equal("foo/bar"))
		Expect(fileInfo.Name()).To(Equal("ubuntu_test.go"))
	})
	It("changes the directory permission", func() {
		Expect(ubuntu.ChangeDirectoryOwner("foo", "bob")).To(Succeed())
		Expect(chown.FilePath).To(Equal("foo/bar"))
		Expect(chown.UID).To(Equal(20))
		Expect(chown.GID).To(Equal(10))
	})
	It("returns an error if LookupUser fails", func() {
		userLookupFail := func(username string) (*user.User, error) {
			return nil, errors.New("LookupUser failed")
		}
		fakeSystemFunctions.LookupUser = userLookupFail
		errReturned := ubuntu.ChangeDirectoryOwner("", "")
		Expect(errReturned).To(HaveOccurred())
		Expect(errReturned.Error()).To(Equal("LookupUser failed"))
	})
	It("returns an error if Atoi for UserID fails", func() {
		userLookupFail := func(username string) (*user.User, error) {
			userName = username
			return &user.User{Name: userName, Gid: "10", Uid: "test"}, nil
		}
		fakeSystemFunctions.LookupUser = userLookupFail
		errReturned := ubuntu.ChangeDirectoryOwner("", "")
		Expect(errReturned).To(HaveOccurred())
		Expect(errReturned.Error()).To(Equal("strconv.Atoi: parsing \"test\": invalid syntax"))
	})
	It("returns an error if Atoi for GroupID fails", func() {
		userLookupFail := func(username string) (*user.User, error) {
			userName = username
			return &user.User{Name: userName, Gid: "test", Uid: "20"}, nil
		}
		fakeSystemFunctions.LookupUser = userLookupFail
		errReturned := ubuntu.ChangeDirectoryOwner("", "")
		Expect(errReturned).To(HaveOccurred())
		Expect(errReturned.Error()).To(Equal("strconv.Atoi: parsing \"test\": invalid syntax"))
	})
	It("returns an error if ChangeFilePermission gets an error as input", func() {
		dirWalkerFail := func(root string, walkFunction filepath.WalkFunc) error {
			err := errors.New("failed to walk dir")
			walkError := walkFunction("", nil, err)
			return walkError
		}
		fakeSystemFunctions.Walk = dirWalkerFail
		errReturned := ubuntu.ChangeDirectoryOwner("", "")
		Expect(errReturned).To(HaveOccurred())
		Expect(errReturned.Error()).To(Equal("failed to walk dir"))
	})
	It("returns an error if ChangeFilePermission has an error", func() {
		fakeChownFail := func(path string, uid int, gid int) error {
			return errors.New("failed to change file permission")
		}
		fakeSystemFunctions.Chown = fakeChownFail
		errReturned := ubuntu.ChangeDirectoryOwner("", "")
		Expect(errReturned).To(HaveOccurred())
		Expect(errReturned.Error()).To(Equal("failed to change file permission"))
	})
})
