package executor_test

import (
	"bytes"
	"errors"
	"io"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/remotecommand"
)

var _ = Describe("podcommandexecutor", func() {

	var (
		podcommandexecutor   *executor.PodExecRESTClient
		restConfig           *rest.Config
		fakeRESTClient       *fake.RESTClient
		fakeExecutorUpgrader *FakeSPDYExecutorUpgrader
	)

	BeforeEach(func() {
		fakeRESTClient = &fake.RESTClient{
			NegotiatedSerializer: scheme.Codecs,
			GroupVersion:         schema.GroupVersion{Version: "v1"},
		}
		restConfig = &rest.Config{}
		fakeExecutorUpgrader = &FakeSPDYExecutorUpgrader{}
		podcommandexecutor = &executor.PodExecRESTClient{
			RestCfg:    restConfig,
			RestClient: fakeRESTClient,
			Upgrader:   fakeExecutorUpgrader,
		}
	})

	Context("NewPodCommandExecutor", func() {
		It("return podcommandexecutor", func() {
			Expect(podcommandexecutor.Upgrader).NotTo(BeNil())
		})
	})

	Context("Executor", func() {
		When("succeeds to Executor", func() {
			It("returns a valid executor", func() {
				executor, err := podcommandexecutor.Executor("testNamespace", "testPod", []string{})
				Expect(err).ToNot(HaveOccurred())
				Expect(executor).ToNot(BeNil())
			})
		})

		When("fails to Executor", func() {
			BeforeEach(func() {
				fakeExecutorUpgrader.SPDYError = errors.New("custom error")
			})
			It("returns an error and a nil executor", func() {
				executor, err := podcommandexecutor.Executor("testNamespace", "testPod", []string{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("custom error"))
				Expect(executor).To(BeNil())
			})
		})

		When("a command, podname, and namespace are specified", func() {
			It("returns a valid request issuing the command to the pod in the namespace specified", func() {
				_, err := podcommandexecutor.Executor("testNamespace", "testPod", []string{"fakeCmd"})
				Expect(err).NotTo(HaveOccurred())
				requestURL := fakeExecutorUpgrader.URL
				parameters, _ := url.ParseQuery(requestURL.RawQuery)
				Expect(parameters["command"]).To(Equal([]string{"fakeCmd"}))
				Expect(parameters["stdout"]).To(Equal([]string{"true"}))
				Expect(parameters["stderr"]).To(Equal([]string{"true"}))
				Expect(parameters["tty"]).To(BeNil())
				Expect(parameters["stdin"]).To(BeNil())

				Expect(requestURL.Path).To(ContainSubstring("pods/testPod"))
				Expect(requestURL.Path).To(ContainSubstring("namespaces/testNamespace"))
			})
		})

		It("passes the RESTConfig to the ExecutorUpgrader", func() {
			_, err := podcommandexecutor.Executor("testNamespace", "testPod", []string{"fakeCmd"})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeExecutorUpgrader.Config).To(BeIdenticalTo(restConfig))
		})
	})

	Context("Execute", func() {
		It("succeeds when remote command executor runs command successfully", func() {
			err := podcommandexecutor.Execute(nil, "testNamespace", "testPod", nil, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		When("Executor fails", func() {
			BeforeEach(func() {
				fakeExecutorUpgrader.SPDYError = errors.New("SPDY error")
			})
			It("returns an error", func() {
				err := podcommandexecutor.Execute(nil, "testNamespace", "testPod", nil, nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("SPDY error"))
			})
		})
		When("executor stream has an error", func() {
			BeforeEach(func() {
				fakeExecutorUpgrader.ExecError = errors.New("stream error")
			})
			It("returns an error", func() {
				var stderr bytes.Buffer
				err := podcommandexecutor.Execute(nil, "testNamespace", "testPod", nil, &stderr)
				Expect(err).To(MatchError("stream error"))
				Expect(stderr.String()).To(Equal("stream error"))
			})
		})
	})

})

// fake implementation of RemoteExecutorUpgrader interface
type FakeSPDYExecutorUpgrader struct {
	Config    *rest.Config
	URL       *url.URL
	SPDYError error
	ExecError error
}

var _ executor.RemoteExecutorUpgrader = &FakeSPDYExecutorUpgrader{}

func (s *FakeSPDYExecutorUpgrader) NewSPDYExecutor(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error) {
	s.Config = config
	s.URL = url
	if s.SPDYError != nil {
		return nil, s.SPDYError
	}
	return &FakeCommandExecutor{Err: s.ExecError}, nil
}

// fake implementation of Executor interface
type FakeCommandExecutor struct {
	Err error
}

var _ remotecommand.Executor = &FakeCommandExecutor{}

func (f *FakeCommandExecutor) Stream(options remotecommand.StreamOptions) error {
	if f.Err != nil {
		Expect(io.WriteString(options.Stderr, f.Err.Error())).To(Equal(len(f.Err.Error())))
		return f.Err
	}
	return nil
}
