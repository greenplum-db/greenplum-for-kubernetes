package multidaemon_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/multidaemon"
)

var _ = Describe("InitializeDaemons", func() {

	var (
		stopCh chan struct{}
	)

	BeforeEach(func() {
		stopCh = make(chan struct{})
	})

	When("all daemons are successful and there is a clean shutdown", func() {
		var (
			op1, op2 SuccessOperator
		)
		It("succeeds and allows all daemons to shutdown cleanly", func() {
			close(stopCh)
			errs := multidaemon.InitializeDaemons(stopCh, op1.Run, op2.Run)
			Expect(errs).To(HaveLen(0))
			Expect(op1.runCalled).To(BeTrue())
			Expect(op1.cleanShutdown).To(BeTrue())
			Expect(op2.runCalled).To(BeTrue())
			Expect(op2.cleanShutdown).To(BeTrue())
		})
	})

	When("all daemons are successful, but there is an error during shutdown", func() {
		var (
			op1 SuccessOperator
			op2 ShutdownFailureOperator
		)
		It("returns an error and allows all daemons to shutdown cleanly", func() {
			close(stopCh)
			errs := multidaemon.InitializeDaemons(stopCh, op1.Run, op2.Run)
			Expect(errs).To(HaveLen(1))
			Expect(errs).To(Equal([]error{errors.New("simulated failure")}))
			Expect(op1.runCalled).To(BeTrue())
			Expect(op1.cleanShutdown).To(BeTrue())
			Expect(op2.runCalled).To(BeTrue())
			Expect(op2.cleanShutdown).To(BeTrue())
		})
	})

	When("one operator fails with an error", func() {
		var (
			op1 SuccessOperator
			op2 FailureOperator
		)

		It("returns an error and allows all daemons to shutdown cleanly", func() {
			errs := multidaemon.InitializeDaemons(stopCh, op1.Run, op2.Run)
			Expect(errs).To(HaveLen(1))
			Expect(errs).To(Equal([]error{errors.New("simulated failure")}))
			Expect(op1.runCalled).To(BeTrue())
			Expect(op1.cleanShutdown).To(BeTrue())
			Expect(op2.runCalled).To(BeTrue())
		})
	})

	When("all daemons fail with errors", func() {
		var (
			op1 FailureOperator
			op2 FailureOperator
		)
		It("returns an error", func() {
			errs := multidaemon.InitializeDaemons(stopCh, op1.Run, op2.Run)
			Expect(errs).To(HaveLen(2))
			Expect(errs).To(Equal([]error{
				errors.New("simulated failure"),
				errors.New("simulated failure"),
			}))
			Expect(op1.runCalled).To(BeTrue())
			Expect(op2.runCalled).To(BeTrue())
		})
	})
})

type SuccessOperator struct {
	runCalled     bool
	cleanShutdown bool
}

func (o *SuccessOperator) Run(stopCh <-chan struct{}) error {
	o.runCalled = true
	<-stopCh
	o.cleanShutdown = true
	return nil
}

type FailureOperator struct {
	runCalled bool
}

func (o *FailureOperator) Run(stopCh <-chan struct{}) error {
	o.runCalled = true
	return errors.New("simulated failure")
}

type ShutdownFailureOperator struct {
	runCalled     bool
	cleanShutdown bool
}

func (o *ShutdownFailureOperator) Run(stopCh <-chan struct{}) error {
	o.runCalled = true
	<-stopCh
	o.cleanShutdown = true
	return errors.New("simulated failure")
}
