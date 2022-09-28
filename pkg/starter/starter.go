package starter

import (
	"io"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
)

type App struct {
	Command      commandable.CommandFn
	StdoutBuffer io.Writer
	StderrBuffer io.Writer
	Fs           vfs.Filesystem
}

type Starter interface {
	Run() error
}
