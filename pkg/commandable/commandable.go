// Package commandable provides a fakeable implementation of exec.Command.
//
// commandable.Command is used to replace exec.Command with a test fake.
//
// See https://npf.io/2015/06/testing-exec-command/ for an explanation of how this
// works, and os/exec/exec_test.go for the original.
//
// To use, either inject a commandable.CommandFn in your production code, or
// substitute uses of `exec.Command()` in your production code with
//`commandable.Command()`. Then use in your tests as below:
//
// In your test module:
//
//   func TestHelperProcess(t *testing.T) {
//	   commandable.Command.HelperProcess()
//   }
//
// When using the global fake, then setup the fake using:
//
//   cmdFake := commandable.Command.InjectFake()
//
// After the test:
//
//   commandable.Command.UninjectFake()
//
// To check the expected program was run with the correct arguments, use:
//
//   cmdFake.CapturedArguments()
package commandable

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/onsi/gomega/gbytes"
)

type (
	CommandFn      func(path string, args ...string) *exec.Cmd
	CommandMatcher func(path string, args ...string) bool
)

var Command CommandFn = exec.Command

type CommandFake struct {
	defaultCommand ExpectedCommand
	expectations   []ExpectedCommand
}

type ExpectedCommand struct {
	matcher CommandMatcher

	stdout     string
	stderr     string
	status     int
	called     *int
	pids       *[]int
	sideEffect func()

	argsPipe   struct{ r, w *os.File }
	argsBuffer *gbytes.Buffer

	expectedEnvVars []string
	envChan         chan<- []string

	cmd *exec.Cmd
}

// NewFakeCommand creates a new exec.Command() substitute. Inject the
// fake.Command into production code.
func NewFakeCommand() *CommandFake {
	return &CommandFake{}
}

func (c *CommandFake) Command(name string, args ...string) *exec.Cmd {
	for i := len(c.expectations) - 1; i >= 0; i-- {
		e := c.expectations[i]
		if e.matcher(name, args...) {
			return e.fakeExecCommand(name, args...)
		}
	}
	return c.defaultCommand.fakeExecCommand(name, args...)
}

var _ CommandFn = (&CommandFake{}).Command

type closeAfterField string

const (
	closeAfterStart closeAfterField = "closeAfterStart"
	closeAfterWait  closeAfterField = "closeAfterWait"
)

// DIRTY HACK. Get cmd.Start() to close our pipe for us using reflection.
func closeAfterEvent(cmd *exec.Cmd, field closeAfterField, f *os.File) {
	rcmd := reflect.ValueOf(cmd).Elem()
	rcloseAfterField := rcmd.FieldByName(string(field))
	// Make the private field writeable
	rcloseAfterField = reflect.NewAt(rcloseAfterField.Type(), unsafe.Pointer(rcloseAfterField.UnsafeAddr())).Elem()
	rcloseAfterField.Set(reflect.Append(rcloseAfterField, reflect.ValueOf(f)))
}

// DIRTY HACK. Make cmd run our routine at Start time, and wait for it during Wait.
func injectGoroutine(cmd *exec.Cmd, goroutine func() error) {
	rcmd := reflect.ValueOf(cmd).Elem()
	rgoroutine := rcmd.FieldByName("goroutine")
	// Make the private field writeable
	rgoroutine = reflect.NewAt(rgoroutine.Type(), unsafe.Pointer(rgoroutine.UnsafeAddr())).Elem()
	rgoroutine.Set(reflect.Append(rgoroutine, reflect.ValueOf(goroutine)))
}

func (c *CommandFake) ExpectCommandMatching(matcher CommandMatcher) *ExpectedCommand {
	c.expectations = append(c.expectations, ExpectedCommand{
		matcher: matcher,
	})
	return &c.expectations[len(c.expectations)-1]
}

func (c *CommandFake) ExpectCommand(expectedPath string, expectedArgs ...string) *ExpectedCommand {
	return c.ExpectCommandMatching(func(path string, args ...string) bool {
		return path == expectedPath && reflect.DeepEqual(args, expectedArgs)
	})
}

func (e *ExpectedCommand) PrintsOutput(stdout string) *ExpectedCommand {
	e.stdout = stdout
	return e
}

func (e *ExpectedCommand) PrintsError(stderr string) *ExpectedCommand {
	e.stderr = stderr
	return e
}

func (e *ExpectedCommand) ReturnsStatus(status int) *ExpectedCommand {
	e.status = status
	return e
}

func (e *ExpectedCommand) CallCounter(counter *int) *ExpectedCommand {
	e.called = counter
	return e
}

func (e *ExpectedCommand) PidList(list *[]int) *ExpectedCommand {
	e.pids = list
	return e
}

func (e *ExpectedCommand) SideEffect(sideEffect func()) *ExpectedCommand {
	e.sideEffect = sideEffect
	return e
}

func (e *ExpectedCommand) SendEnvironment(envs chan<- []string) *ExpectedCommand {
	e.envChan = envs
	return e
}

func (e *ExpectedCommand) fakeExecCommand(path string, args ...string) *exec.Cmd {
	var err error
	e.argsPipe.r, e.argsPipe.w, err = os.Pipe()
	if err != nil {
		panic(err)
	}

	cmdArgs := []string{"-test.run=TestHelperProcess", "--", path}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command(os.Args[0], cmdArgs...)
	cmd.Env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		"FAKE_STATUS=" + strconv.Itoa(e.status),
		"FAKE_RESULT=" + e.stdout,
		"FAKE_ERR_RESULT=" + e.stderr,
	}
	cmd.ExtraFiles = []*os.File{e.argsPipe.w}
	// set up Cmd to close our pipe for us at the right times (same as it does with stdout/err)
	closeAfterEvent(cmd, closeAfterStart, e.argsPipe.w)
	closeAfterEvent(cmd, closeAfterWait, e.argsPipe.r)

	// Similarly, set up Cmd to run a goroutine to read argsPipe, and execute side effects.
	e.argsBuffer = gbytes.NewBuffer()
	injectGoroutine(cmd, e.readArgsAndRunSideEffects)

	e.cmd = cmd

	return e.cmd
}

func (e *ExpectedCommand) readArgsAndRunSideEffects() error {
	io.Copy(e.argsBuffer, e.argsPipe.r)

	// argsPipe is exhausted (write end is closed),
	// so the command was run and we can perform our side effects

	if e.called != nil {
		*e.called++
	}

	if e.pids != nil {
		*e.pids = append(*e.pids, e.cmd.Process.Pid)
	}

	if e.envChan != nil {
		e.envChan <- e.cmd.Env
	}

	if e.sideEffect != nil {
		e.sideEffect()
	}
	return nil
}

func (c *CommandFake) FakeOutput(fakeOutput string) *CommandFake {
	c.defaultCommand.stdout = fakeOutput
	return c
}

func (c *CommandFake) FakeErrOutput(fakeErrOutput string) *CommandFake {
	c.defaultCommand.stderr = fakeErrOutput
	return c
}

func (c *CommandFake) FakeStatus(status int) *CommandFake {
	c.defaultCommand.status = status
	return c
}

// TODO: faithfully serialize and deserialize the args as []string.
// Currently you can't tell the difference between "hello world" and "hello" "world"
func (c *CommandFake) CapturedArgs() string {
	if c.defaultCommand.argsBuffer == nil {
		return ""
	}

	return string(c.defaultCommand.argsBuffer.Contents())
}

// InjectFake sets the global fakeable.Command to an exec.Command() substitute.
func (c *CommandFn) InjectFake() *CommandFake {
	fake := NewFakeCommand()
	*c = fake.Command
	return fake
}

// UninjectFake removes the fake from fakeable.Command and restores the actual exec.Command().
func (c *CommandFn) UninjectFake() {
	*c = exec.Command
}

func (c *CommandFn) HelperProcess() {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	pipe := os.NewFile(3, "args pipe")

	fmt.Fprint(pipe, strings.Join(os.Args[3:], " "))

	// Optionally output FAKE_RESULT
	fmt.Print(os.Getenv("FAKE_RESULT"))
	fmt.Fprint(os.Stderr, os.Getenv("FAKE_ERR_RESULT"))

	status, err := strconv.Atoi(os.Getenv("FAKE_STATUS"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing FAKE_STATUS:", err)
		status = -127
	}
	os.Exit(status)
}
