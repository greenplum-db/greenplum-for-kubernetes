package main

import (
	"os"
	"os/exec"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils/cluster"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	ctrllog.SetLogger(gplog.ForProd(true))
	log := ctrllog.Log.WithName("initializeCluster")
	fs := vfs.OS()
	c := cluster.New(
		fs,
		exec.Command,
		os.Stdout,
		os.Stderr,
		instanceconfig.NewReader(fs),
		cluster.NewGpInitSystem(fs, exec.Command, os.Stdout, os.Stderr, instanceconfig.NewReader(fs)),
	)
	if err := c.Initialize(); err != nil {
		log.Error(err, "failed to initialize cluster")
		os.Exit(1)
	}
	os.Exit(0)
}
