package admission

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/hostpod"
	"github.com/pkg/errors"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reminder: This is not tested. It's mostly dependency injection,
// so testing is perhaps not useful, but tread carefully.
func NewWebhook(ctrlClient client.Client, cfg *rest.Config, podExec executor.PodExecInterface, instanceImage string) (*Webhook, error) {
	kubeClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "building kubernetes client set")
	}

	currentNS, err := hostpod.GetCurrentNamespace(vfs.OS())
	if err != nil {
		return nil, errors.Wrap(err, "getting current namespace")
	}

	operatorPod, err := hostpod.GetThisPod(context.Background(), ctrlClient, currentNS, os.Hostname)
	if err != nil {
		return nil, errors.Wrap(err, "getting operator pod")
	}
	operatorPodNameSuffix := strings.TrimPrefix(operatorPod.Name, "greenplum-operator")

	// TODO: is there a way we can test this? maybe not very interesting to test.
	var gpCRD apiextensionsv1.CustomResourceDefinition
	gpCRDKey := types.NamespacedName{Namespace: "", Name: "greenplumclusters.greenplum.pivotal.io"}
	if err := ctrlClient.Get(context.Background(), gpCRDKey, &gpCRD); err != nil {
		return nil, fmt.Errorf("missing crd %s: %w", gpCRDKey.String(), err)
	}

	handler := &Handler{
		KubeClient:     ctrlClient,
		InstanceImage:  instanceImage,
		PodCmdExecutor: podExec,
	}

	webhook := &Webhook{
		KubeClient:      ctrlClient,
		Namespace:       currentNS,
		ServiceOwner:    operatorPod,
		WebhookCfgOwner: &gpCRD,
		NameSuffix:      operatorPodNameSuffix,
		Handler:         handler.Handler(),
		Server:          NewTLSServer(),
		CertGenerator: &CertificateGenerator{
			CtrlClient:    ctrlClient,
			KubeClientSet: kubeClientset,
			Owner:         &gpCRD,
		},
	}

	return webhook, nil
}
