package matcher

import (
	"fmt"

	"github.com/blang/vfs"
	"github.com/onsi/gomega/types"
)

type vfsFileMatcher struct {
	fs       vfs.Filesystem
	expected string
	contents string
}

func EqualInFilesystem(fs vfs.Filesystem, contents string) types.GomegaMatcher {
	return &vfsFileMatcher{fs: fs, expected: contents}
}

func (m *vfsFileMatcher) Match(actual interface{}) (success bool, err error) {
	filename := actual.(string)
	contents, err := vfs.ReadFile(m.fs, filename)
	if err != nil {
		return false, err
	}
	m.contents = string(contents)
	return m.contents == m.expected, nil
}

func (m *vfsFileMatcher) FailureMessage(file interface{}) string {
	return fmt.Sprintf("Expected the file\n\t%#v\nto contain\n\t%#v\nbut found\n\t%#v", file, m.expected, m.contents)
}

func (m *vfsFileMatcher) NegatedFailureMessage(file interface{}) string {
	return fmt.Sprintf("Expected the file\n\t%#v\nnot to contain\n\t%#v\nbut found \n\t%#v", file, m.expected, m.contents)
}

type vfsExistMatcher struct {
	fs       vfs.Filesystem
	expected string
}

func ExistInFilesystem(fs vfs.Filesystem) types.GomegaMatcher {
	return &vfsExistMatcher{fs: fs}
}

func (m *vfsExistMatcher) Match(actual interface{}) (success bool, err error) {
	filename := actual.(string)
	if _, err := m.fs.Stat(filename); err != nil {
		return false, nil
	}
	m.expected = filename
	return true, nil
}

func (m *vfsExistMatcher) FailureMessage(file interface{}) string {
	return fmt.Sprintf("Expected file\n\t%#v\nto exist", file)
}

func (m *vfsExistMatcher) NegatedFailureMessage(file interface{}) string {
	return fmt.Sprintf("Expected file\n\t%#v\nnot to exist", file)
}
