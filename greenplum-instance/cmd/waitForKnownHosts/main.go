package main

import (
	"flag"
	"os"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/knownhosts"
	apiwait "k8s.io/apimachinery/pkg/util/wait"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = ctrllog.Log.WithName("waitForKnownHosts")

func main() {
	ctrllog.SetLogger(gplog.ForProd(false))

	var newPrimarySegmentCount = flag.Int("newPrimarySegmentCount", 0, "new primary segment count")
	flag.Parse()

	config, err := instanceconfig.NewReader(vfs.OS()).GetConfigValues()
	if err != nil {
		log.Error(err, "error reading configmap")
		os.Exit(1)
	}
	gpdbClusterHostnames := net.GenerateHostList(*newPrimarySegmentCount, config.Mirrors, config.Standby, "")
	knownHostsWaiter := &ssh.KnownHostsWaiter{
		PollWait:         apiwait.PollImmediate,
		KnownHostsReader: knownhosts.NewReader(),
	}
	if errs := multihost.ParallelForeach(knownHostsWaiter, gpdbClusterHostnames); len(errs) != 0 {
		os.Exit(1)
	}
}
