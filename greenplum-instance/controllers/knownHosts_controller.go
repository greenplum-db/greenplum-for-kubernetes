package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/blang/vfs"
	"github.com/go-logr/logr"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/dns"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/keyscanner"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/knownhosts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
)

const KnownHostsFilename = "/home/gpadmin/.ssh/known_hosts"

type Reconciler interface {
	Reconcile(eventObj interface{})
}

type KnownHostsController struct {
	ClientFn     func() (kubernetes.Interface, error)
	ConfigReader instanceconfig.Reader
	Reconciler   Reconciler
}

type KnownHostsReconciler struct {
	Log              logr.Logger
	Fs               vfs.Filesystem
	DNSResolver      multihost.Operation
	SSHKeyScanner    keyscanner.SSHKeyScannerInterface
	KnownHostsReader knownhosts.ReaderInterface
}

var _ Reconciler = &KnownHostsReconciler{}

func NewKnownHostsController() *KnownHostsController {
	reconciler := KnownHostsReconciler{
		Log:              ctrl.Log.WithName("knownHostsReconciler"),
		Fs:               vfs.OS(),
		DNSResolver:      dns.NewConsistentResolver(),
		SSHKeyScanner:    keyscanner.NewSSHKeyScanner(),
		KnownHostsReader: knownhosts.NewReader(),
	}

	return &KnownHostsController{
		ConfigReader: instanceconfig.NewReader(vfs.OS()),
		ClientFn: func() (kubernetes.Interface, error) {
			cfg := ctrl.GetConfigOrDie()
			return kubernetes.NewForConfig(cfg)
		},
		Reconciler: &reconciler,
	}
}

func (c *KnownHostsController) Run(ctx context.Context) error {
	client, err := c.ClientFn()
	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	greenplumClusterNamespace, err := c.ConfigReader.GetNamespace()
	if err != nil {
		return fmt.Errorf("failed to read namespace: %w", err)
	}
	greenplumClusterName, err := c.ConfigReader.GetGreenplumClusterName()
	if err != nil {
		return fmt.Errorf("failed to read greenplumcluster name: %w", err)
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		client,
		2*time.Hour,
		informers.WithNamespace(greenplumClusterNamespace),
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = "greenplum-cluster=" + greenplumClusterName
			options.FieldSelector = fields.OneTermEqualSelector("metadata.name", "agent").String()
		}),
	)
	endpointInformer := informerFactory.Core().V1().Endpoints().Informer()
	endpointInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.Reconciler.Reconcile(obj) },
		UpdateFunc: func(oldObj, newObj interface{}) { c.Reconciler.Reconcile(newObj) },
	})

	informerFactory.Start(ctx.Done())
	informerFactory.WaitForCacheSync(ctx.Done())

	return nil
}

func (r *KnownHostsReconciler) Reconcile(eventObj interface{}) {
	endpoints, ok := eventObj.(*corev1.Endpoints)
	if !ok {
		r.Log.Error(nil, "non *corev1.Endpoint event object received", "eventObj", eventObj)
		return
	}

	knownHostsMap, err := r.KnownHostsReader.GetKnownHosts()
	if err != nil {
		r.Log.Error(err, "unable to get known hosts")
		return
	}
	var knownHosts []string
	for knownHost := range knownHostsMap {
		knownHosts = append(knownHosts, knownHost)
	}

	newReadyHosts := getNewReadyHosts(knownHosts, endpoints)
	if len(newReadyHosts) == 0 {
		return
	}
	r.Log.Info("scanning ssh host key(s)", "hosts", newReadyHosts)

	if errs := multihost.ParallelForeach(r.DNSResolver, newReadyHosts); len(errs) != 0 {
		r.Log.Error(nil, "failed to resolve dns entries for all hosts", "hosts", newReadyHosts)
		return
	}

	newHostKeys, err := keyscanner.ScanHostKeys(r.SSHKeyScanner, r.KnownHostsReader, newReadyHosts)
	if err != nil {
		// error has already been logged by the keyscanner
		return
	}

	fileWriter := fileutil.FileWriter{WritableFileSystem: r.Fs}
	if err := fileWriter.Append(KnownHostsFilename, newHostKeys); err != nil {
		r.Log.Error(err, "failed to write known_hosts file")
		return
	}
}

func getNewReadyHosts(knownHosts []string, newEndpoints *corev1.Endpoints) (newReadyHosts []string) {
	readyAddresses := make(map[string]bool)

	for _, endpointsubset := range newEndpoints.Subsets {
		for _, address := range endpointsubset.Addresses {
			readyAddresses[address.Hostname] = true
		}
	}

	for _, hostname := range knownHosts {
		readyAddresses[hostname] = false
	}

	for hostname, isNew := range readyAddresses {
		if isNew {
			newReadyHosts = append(newReadyHosts, hostname)
		}
	}

	return
}
