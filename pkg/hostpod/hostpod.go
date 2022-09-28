package hostpod

import (
	"context"
	"fmt"

	"github.com/blang/vfs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetCurrentNamespace(filesystem vfs.Filesystem) (string, error) {
	const NamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	ns, err := vfs.ReadFile(filesystem, NamespaceFile)
	return string(ns), err
}

func GetThisPod(ctx context.Context, client client.Client, namespace string, hostname func() (string, error)) (*corev1.Pod, error) {
	host, err := hostname()
	if err != nil {
		return nil, fmt.Errorf("getting hostname: %w", err)
	}
	var thisPod corev1.Pod
	key := types.NamespacedName{Namespace: namespace, Name: host}
	err = client.Get(ctx, key, &thisPod)
	return &thisPod, err
}
