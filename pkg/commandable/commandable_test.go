package commandable_test

import (
	"io"
	"os/exec"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
)

type sampleObject struct {
	Command        commandable.CommandFn
	Stdout, Stderr io.Writer
}

func (t *sampleObject) RunCommands() (err error) {
	var cmd *exec.Cmd
	cmd = t.Command("command1", "with", "arguments")
	cmd.Stdout = t.Stdout
	cmd.Stderr = t.Stderr
	if err = cmd.Run(); err != nil {
		return
	}
	cmd = t.Command("command2", "with", "more", "arguments")
	cmd.Stdout = t.Stdout
	cmd.Stderr = t.Stderr
	if err = cmd.Run(); err != nil {
		return
	}
	return nil
}

var _ = Describe("commandable.NewFakeCommand", func() {
	var (
		sample         sampleObject
		cmdFake        *commandable.CommandFake
		stdout, stderr *gbytes.Buffer
	)
	BeforeEach(func() {
		cmdFake = commandable.NewFakeCommand()
		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()
		sample = sampleObject{Command: cmdFake.Command, Stdout: stdout, Stderr: stderr}
	})
	It("Runs commands", func() {
		Expect(sample.RunCommands()).To(Succeed())
	})
	Describe("CapturedArgs()", func() {
		It("Records the last command", func() {
			Expect(sample.RunCommands()).To(Succeed())
			Expect(cmdFake.CapturedArgs()).To(Equal("command2 with more arguments"))
		})
		It("Returns empty string when nothing was run", func() {
			Expect(cmdFake.CapturedArgs()).To(Equal(""))
		})
	})
	It("Produces fake output", func() {
		cmdFake.FakeOutput("data\n")
		Expect(sample.RunCommands()).To(Succeed())
		Expect(stdout).To(gbytes.Say("data\n"))
		Expect(stdout).To(gbytes.Say("data\n"))
	})
	It("Can set a result status", func() {
		cmdFake.FakeStatus(123)
		err := sample.RunCommands()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("exit status 123"))
		Expect(stdout.Contents()).To(BeEmpty())
		Expect(stderr.Contents()).To(BeEmpty())
	})
	It("Produces fake error output", func() {
		cmdFake.FakeErrOutput("error: You forgot to say please!\n")
		Expect(sample.RunCommands()).To(Succeed())
		Expect(stderr).To(gbytes.Say("error: You forgot to say please!\n"))
		Expect(stderr).To(gbytes.Say("error: You forgot to say please!\n"))
	})

	Describe("ExpectCommand()", func() {
		It("Modifies how a specific command behaves", func() {
			cmdFake.FakeOutput("initial output\n")
			cmdFake.ExpectCommand("command2", "with", "more", "arguments").PrintsOutput("I am command2\n").PrintsError("I had an error\n")
			Expect(sample.RunCommands()).To(Succeed())
			Expect(stdout).To(gbytes.Say("initial output\n"), "output should come from running command1")
			Expect(stdout).To(gbytes.Say("I am command2\n"), "output should come from running command2")
			Expect(stderr).To(gbytes.Say("I had an error\n"))
		})
		It("Retains the FakeOutput for other commands", func() {
			cmdFake.FakeOutput("initial output\n")
			cmdFake.ExpectCommand("command1", "with", "arguments").PrintsOutput("I am command1\n")
			Expect(sample.RunCommands()).To(Succeed())
			Expect(stdout).To(gbytes.Say("I am command1\n"), "output should come from running command1")
			Expect(stdout).To(gbytes.Say("initial output\n"), "output should come from running command2")
		})
		It("Can change the return status", func() {
			cmdFake.FakeOutput("initial output\n")
			cmdFake.ExpectCommand("command2", "with", "more", "arguments").ReturnsStatus(2)
			err := sample.RunCommands()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("exit status 2"))
			Expect(stdout).To(gbytes.Say("initial output\n"), "output should come from running command1")
		})

		It("Requires matching all arguments exactly", func() {
			cmdFake.ExpectCommand("command1", "with", "WRONG arguments").PrintsOutput("I am command1\n")
			Expect(sample.RunCommands()).To(Succeed())
			Expect(stdout.Contents()).NotTo(ContainSubstring("I am command1\n"), "command1 should NOT have matched")
		})
		It("Records how many times a command was called", func() {
			var cmd2Called, nonCmdCalled int
			cmdFake.ExpectCommand("command2", "with", "more", "arguments").CallCounter(&cmd2Called)
			cmdFake.ExpectCommand("nonCmd").CallCounter(&nonCmdCalled).PrintsOutput("Unexpected output from nonCmd\n")
			Expect(sample.RunCommands()).To(Succeed())
			Expect(sample.RunCommands()).To(Succeed())
			Expect(cmd2Called).To(Equal(2))
			Expect(nonCmdCalled).To(Equal(0))
			Expect(stdout.Contents()).To(BeEmpty(), "Should not see output from nonCmd")
		})
		It("Can produce a side effect", func() {
			sideEffectHappened := false
			cmdFake.ExpectCommand("command1", "with", "arguments").SideEffect(func() { sideEffectHappened = true })
			Expect(sample.RunCommands()).To(Succeed())
			Expect(sideEffectHappened).To(BeTrue())
		})
		It("Prefers the most recently added matching expectation", func() {
			var matchedExpect [2]bool
			cmdFake.ExpectCommand("command1", "with", "arguments").SideEffect(func() { matchedExpect[0] = true })
			cmdFake.ExpectCommand("command1", "with", "arguments").SideEffect(func() { matchedExpect[1] = true })
			Expect(sample.RunCommands()).To(Succeed())
			Expect(matchedExpect[0]).To(BeFalse())
			Expect(matchedExpect[1]).To(BeTrue())
		})
		It("Can send the environment for each call over a channel", func() {
			envs := make(chan []string, 2)
			cmdFake.ExpectCommand("command3").SendEnvironment(envs)

			for _, e := range []string{
				"FOO=bar",
				"BAZ=quux",
			} {
				cmd := cmdFake.Command("command3")
				cmd.Env = append(cmd.Env, e)
				cmd.Stdout = stdout
				cmd.Stderr = stderr
				Expect(cmd.Run()).To(Succeed())
			}

			var env []string
			Expect(envs).To(Receive(&env))
			Expect(env).To(ContainElement("FOO=bar"))
			Expect(envs).To(Receive(&env))
			Expect(env).To(ContainElement("BAZ=quux"))
		})
		It("Uses a builder pattern", func() {
			envs := make(chan []string, 1)
			sideEffectHappened := false
			cmdFake.ExpectCommand("command1", "with", "arguments").
				SendEnvironment(envs).
				SideEffect(func() { sideEffectHappened = true }).
				ReturnsStatus(3).
				PrintsOutput("hello\n").
				PrintsError("WRONG\n")
			err := sample.RunCommands()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("exit status 3"))
			Expect(stdout).To(gbytes.Say("hello\n"))
			Expect(stderr).To(gbytes.Say("WRONG\n"))
			Expect(sideEffectHappened).To(BeTrue())
			Expect(envs).To(Receive())
		})
	})

	Describe("ExpectCommandMatching()", func() {
		It("Takes a custom matcher function", func() {
			cmdsCalled := 0
			cmdFake.ExpectCommandMatching(func(path string, _ ...string) bool {
				return strings.Contains(path, "command")
			}).CallCounter(&cmdsCalled)
			Expect(sample.RunCommands()).To(Succeed())
			Expect(cmdsCalled).To(Equal(2))
		})
	})
})

var _ = Describe("commandable.NewFakeCommand", func() {
	It("Uses a builder pattern", func() {
		cmdFake := commandable.NewFakeCommand().FakeStatus(3).FakeOutput("hello\n").FakeErrOutput("WRONG\n")
		stdout := gbytes.NewBuffer()
		stderr := gbytes.NewBuffer()
		sample := sampleObject{Command: cmdFake.Command, Stdout: stdout, Stderr: stderr}
		err := sample.RunCommands()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("exit status 3"))
		Expect(stdout).To(gbytes.Say("hello\n"))
		Expect(stderr).To(gbytes.Say("WRONG\n"))
	})
})

func TestHelperProcess(t *testing.T) {
	commandable.Command.HelperProcess()
}
