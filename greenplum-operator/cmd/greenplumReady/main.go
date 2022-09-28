package main

import (
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubewait"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("greenplumReady")

func main() {
	ctrl.SetLogger(gplog.ForProd(true))

	log.Info("Waiting for cluster to be ready ...")
	err := kubewait.ForClusterReady(true)

	if err != nil {
		log.Error(err, "cluster failed to become ready")
	} else {
		log.Info("cluster is ready")
	}
}
