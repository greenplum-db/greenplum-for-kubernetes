package executor

import (
	"io"
	"net/url"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = ctrllog.Log.WithName("PodExec")

func NewPodExec(scheme *runtime.Scheme, config *rest.Config) PodExecInterface {
	codecFactory := serializer.NewCodecFactory(scheme)
	//parameterCodec := runtime.NewParameterCodec(scheme)
	podGVK, _ := apiutil.GVKForObject(&corev1.Pod{}, scheme)
	podRestInterface, _ := apiutil.RESTClientForGVK(podGVK, false, config, codecFactory)

	podExec := &PodExecRESTClient{
		RestCfg:    config,
		RestClient: podRestInterface,
		Upgrader:   &RealSPDYExecutorUpgrader{},
	}
	return podExec
}

type RemoteExecutorUpgrader interface {
	NewSPDYExecutor(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error)
}

type RealSPDYExecutorUpgrader struct{}

var _ RemoteExecutorUpgrader = &RealSPDYExecutorUpgrader{}

func (s *RealSPDYExecutorUpgrader) NewSPDYExecutor(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error) {
	return remotecommand.NewSPDYExecutor(config, method, url)
}

type PodExecInterface interface {
	Execute(command []string, namespace, podName string, stdout, stderr io.Writer) error
}

type PodExecRESTClient struct {
	RestCfg    *rest.Config
	RestClient rest.Interface
	Upgrader   RemoteExecutorUpgrader
}

var _ PodExecInterface = &PodExecRESTClient{}

func (p *PodExecRESTClient) Execute(command []string, namespace, podName string, stdout, stderr io.Writer) error {
	remoteCommandExecutor, err := p.Executor(namespace, podName, command)
	if err != nil {
		return err
	}
	return remoteCommandExecutor.Stream(remotecommand.StreamOptions{Stdout: stdout, Stderr: stderr})
}

func (p *PodExecRESTClient) Executor(namespace, podName string, command []string) (remotecommand.Executor, error) {
	url := p.RestClient.Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: command,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec).
		URL()
	return p.Upgrader.NewSPDYExecutor(p.RestCfg, "POST", url)
}
