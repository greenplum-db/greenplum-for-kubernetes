package gpexpand

import (
	"io"

	"github.com/go-logr/logr"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost"
	"github.com/pkg/errors"
)

type RunGpexpandConfig struct {
	Log              logr.Logger
	NewSegmentCount  int
	IsMirrored       bool
	Standby          bool
	Stdout           io.Writer
	Stderr           io.Writer
	DNSResolver      multihost.Operation
	KnownHostsWaiter multihost.Operation
	SSHExecutor      multihost.Operation
	Command          commandable.CommandFn
}

func (r *RunGpexpandConfig) Run() error {
	hostnameList := net.GenerateHostList(r.NewSegmentCount, r.IsMirrored, r.Standby, "")

	r.Log.Info("resolving DNS entries for all masters and segments")
	if errs := multihost.ParallelForeach(r.DNSResolver, hostnameList); len(errs) != 0 {
		return errors.New("failed to resolve DNS entries for all masters and segments")
	}

	// Before running waitForKnownHosts on all pods via SSH, ensure that (on the master pod)
	// we have known_host entries for every other pod. Otherwise, the SSH commands will fail.
	r.Log.Info("waiting for known_hosts file to be populated on master")
	if errs := multihost.ParallelForeach(r.KnownHostsWaiter, hostnameList); len(errs) != 0 {
		return errors.New("timed out waiting for known_hosts on master")
	}

	// Run waitForKnownHosts on every pod in the cluster to ensure that every pod has known_host
	// file entries for every other pod.
	r.Log.Info("waiting for known_hosts file to be populated on all masters and segments")
	if errs := multihost.ParallelForeach(r.SSHExecutor, hostnameList); len(errs) != 0 {
		// log errors
		return errors.New("timed out waiting for known_hosts on all masters and segments")
	}

	r.Log.Info("running gpexpand")
	cmd := r.Command("bash", "-c",
		"source /usr/local/greenplum-db/greenplum_path.sh && MASTER_DATA_DIRECTORY=/greenplum/data-1 gpexpand -i /tmp/gpexpand_config")
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	return cmd.Run()
}
