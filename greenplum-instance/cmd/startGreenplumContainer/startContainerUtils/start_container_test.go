package startContainerUtils_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
)

type fakeStarter struct {
	called bool
	err    error
}

func (f *fakeStarter) Run() error {
	f.called = true
	return f.err
}

var _ = Describe("GreenplumContainerStarter", func() {
	var (
		args        []string
		uid         int
		errorBuffer *gbytes.Buffer
		outBuffer   *gbytes.Buffer
		fakeCmd     *commandable.CommandFake

		fakeRootStarter        fakeStarter
		fakeGpadminStarter     fakeStarter
		fakeLabelPVCStarter    fakeStarter
		fakeMultidaemonStarter fakeStarter
	)
	BeforeEach(func() {
		args = []string{"/ourselves"}
		uid = 1000
		fakeCmd = commandable.NewFakeCommand()
		errorBuffer = gbytes.NewBuffer()
		outBuffer = gbytes.NewBuffer()

		fakeRootStarter = fakeStarter{}
		fakeGpadminStarter = fakeStarter{}
		fakeLabelPVCStarter = fakeStarter{}
		fakeMultidaemonStarter = fakeStarter{}
	})

	var (
		app    *startContainerUtils.GreenplumContainerStarter
		status int
	)
	JustBeforeEach(func() {
		app = &startContainerUtils.GreenplumContainerStarter{
			App: &starter.App{
				Command:      fakeCmd.Command,
				StdoutBuffer: outBuffer,
				StderrBuffer: errorBuffer,
			},
			UID:                uid,
			Root:               &fakeRootStarter,
			Gpadmin:            &fakeGpadminStarter,
			LabelPVC:           &fakeLabelPVCStarter,
			MultidaemonStarter: &fakeMultidaemonStarter,
		}
		status = app.Run(args)
	})

	When("the gpadmin Starter fails", func() {
		BeforeEach(func() {
			fakeGpadminStarter.err = errors.New("An error")
		})
		It("Exits with a non-zero status code", func() {
			Expect(status).To(Equal(1))
		})
		It("Prints the returned error", func() {
			Expect(errorBuffer).To(gbytes.Say("An error"))
		})
		It("Doesn't label PVCs or run the multidaemon starter", func() {
			Expect(fakeGpadminStarter.called).To(BeTrue())
			Expect(fakeLabelPVCStarter.called).To(BeFalse())
			Expect(fakeMultidaemonStarter.called).To(BeFalse())
		})
	})

	When("the LabelPVCStarter fails", func() {
		BeforeEach(func() {
			fakeLabelPVCStarter.err = errors.New("An error")
		})
		It("Exits with a non-zero status code", func() {
			Expect(status).To(Equal(1))
		})
		It("Prints the returned error", func() {
			Expect(errorBuffer).To(gbytes.Say("An error"))
		})
		It("Doesn't run the multidaemon starter", func() {
			Expect(fakeGpadminStarter.called).To(BeTrue())
			Expect(fakeLabelPVCStarter.called).To(BeTrue())
			Expect(fakeMultidaemonStarter.called).To(BeFalse())
		})
	})

	When("the MultidaemonStarter fails", func() {
		BeforeEach(func() {
			fakeMultidaemonStarter.err = errors.New("An error")
		})
		It("Exits with a non-zero status code", func() {
			Expect(status).To(Equal(1))
		})
		It("Prints the returned error", func() {
			Expect(errorBuffer).To(gbytes.Say("An error"))
		})
		It("started all starters", func() {
			Expect(fakeGpadminStarter.called).To(BeTrue())
			Expect(fakeLabelPVCStarter.called).To(BeTrue())
			Expect(fakeMultidaemonStarter.called).To(BeTrue())
		})
	})

	When("--do-root-startup is passed", func() {
		BeforeEach(func() {
			args = append(args, "--do-root-startup")
		})
		When("UID is zero", func() {
			BeforeEach(func() {
				uid = 0
			})
			It("calls only the Root starter", func() {
				Expect(fakeRootStarter.called).To(BeTrue())
				Expect(fakeGpadminStarter.called).To(BeFalse())
				Expect(fakeLabelPVCStarter.called).To(BeFalse())
				Expect(fakeMultidaemonStarter.called).To(BeFalse())
			})
			When("RootStarter fails", func() {
				BeforeEach(func() {
					fakeRootStarter.err = errors.New("root startup failure")
				})
				It("Prints an error and exits non-zero", func() {
					Expect(errorBuffer).To(gbytes.Say("root startup failure\n"))
					Expect(status).To(Equal(1))
				})
			})
		})
		When("UID is not zero", func() {
			BeforeEach(func() {
				uid = 1001
			})
			It("Prints an error and exits non-zero", func() {
				Expect(errorBuffer).To(gbytes.Say("--do-root-startup was passed, but we are not root"))
				Expect(status).To(Equal(1))
			})
			It("does not call any Starters", func() {
				Expect(fakeRootStarter.called).To(BeFalse())
				Expect(fakeGpadminStarter.called).To(BeFalse())
				Expect(fakeLabelPVCStarter.called).To(BeFalse())
				Expect(fakeMultidaemonStarter.called).To(BeFalse())
			})
		})
	})
	When("no flags are passed", func() {
		When("root startup succeeds", func() {
			BeforeEach(func() {
				fakeCmd.ExpectCommand("/usr/bin/sudo", "/ourselves", "--do-root-startup").
					PrintsOutput("Running root startup")
			})
			It("does not call the Root starter", func() {
				Expect(fakeRootStarter.called).To(BeFalse())
				Expect(fakeGpadminStarter.called).To(BeTrue())
				Expect(fakeLabelPVCStarter.called).To(BeTrue())
				Expect(fakeMultidaemonStarter.called).To(BeTrue())
			})
			It("calls itself under sudo with --do-root-startup", func() {
				Expect(outBuffer).To(gbytes.Say("Running root startup"))
			})

		})
		When("root startup fails", func() {
			BeforeEach(func() {
				fakeCmd.ExpectCommand("/usr/bin/sudo", "/ourselves", "--do-root-startup").
					PrintsOutput("Running root startup").
					PrintsError("root startup failed!").
					ReturnsStatus(1)
			})
			It("does not call any starters", func() {
				Expect(fakeRootStarter.called).To(BeFalse())
				Expect(fakeGpadminStarter.called).To(BeFalse())
				Expect(fakeLabelPVCStarter.called).To(BeFalse())
				Expect(fakeMultidaemonStarter.called).To(BeFalse())
			})
			It("calls itself under sudo with --do-root-startup", func() {
				Expect(outBuffer).To(gbytes.Say("Running root startup"))
				Expect(errorBuffer).To(gbytes.Say("root startup failed!"))
			})

		})
	})
	When("bogus flags are passed", func() {
		BeforeEach(func() {
			args = append(args, "garbage", "junk")
		})
		It("Prints an error and exits non-zero", func() {
			Expect(errorBuffer).To(gbytes.Say(`Unexpected argument\(s\): \[garbage junk\]\n`))
			Expect(status).To(Equal(1))
		})
	})
})
