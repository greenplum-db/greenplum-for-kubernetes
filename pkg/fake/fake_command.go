package fake

import (
	"io"
)

type FakeCommand struct {
	Name  string
	Arg   []string
	Error error
}

func (f *FakeCommand) Run(command string, args ...string) error {
	f.Name = command
	f.Arg = args
	return f.Error
}

type FakeCommandWithOutput struct {
	StdOut io.Writer
	Name   string
	Arg    []string
	Error  error
}

func (f *FakeCommandWithOutput) Run(command string, args ...string) error {
	f.Name = command
	f.Arg = args
	f.StdOut.Write([]byte("command ran"))
	return f.Error
}
