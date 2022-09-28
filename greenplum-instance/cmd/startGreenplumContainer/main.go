package main

import (
	"os"
	"os/exec"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils"
	clusterinit "github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils/cluster"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/controllers"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/multidaemon"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/dns"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/keyscanner"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/knownhosts"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/ubuntuUtils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	ctrllog.SetLogger(gplog.ForProd(true))
	fs := vfs.OS()
	s := &starter.App{
		Command:      exec.Command,
		StdoutBuffer: os.Stdout,
		StderrBuffer: os.Stderr,
		Fs:           fs,
	}
	u := ubuntuUtils.NewRealUbuntu()
	cluster := clusterinit.New(
		fs,
		exec.Command,
		os.Stdout,
		os.Stderr,
		instanceconfig.NewReader(fs),
		clusterinit.NewGpInitSystem(vfs.OS(), exec.Command, os.Stdout, os.Stderr, instanceconfig.NewReader(fs)),
	)

	knownHostsController := controllers.NewKnownHostsController()
	clusterInitDaemon := &startContainerUtils.ClusterInitDaemon{
		App:              s,
		Config:           instanceconfig.NewReader(fs),
		Ubuntu:           u,
		DNSResolver:      dns.NewConsistentResolver(),
		KeyScanner:       keyscanner.NewSSHKeyScanner(),
		KnownHostsReader: knownhosts.NewReader(),
		C:                cluster,
	}
	sshDaemon := &startContainerUtils.SSHDaemon{App: s}
	containerStarter := startContainerUtils.GreenplumContainerStarter{
		App:     s,
		UID:     os.Getuid(),
		Root:    &startContainerUtils.RootContainerStarter{App: s, Ubuntu: u},
		Gpadmin: &startContainerUtils.GpadminContainerStarter{App: s},
		LabelPVC: &startContainerUtils.LabelPvcStarter{
			App:      s,
			Hostname: os.Hostname,
			NewClient: func() (client.Client, error) {
				return client.New(ctrl.GetConfigOrDie(), client.Options{})
			},
		},
		MultidaemonStarter: &startContainerUtils.MultidaemonStarter{
			Daemons: []multidaemon.DaemonFunc{
				knownHostsController.Run,
				clusterInitDaemon.Run,
				sshDaemon.Run,
			},
		},
	}

	os.Exit(containerStarter.Run(os.Args))
}
