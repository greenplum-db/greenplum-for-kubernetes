package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/blang/vfs"
	gpexpandconfig "github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/runGpexpand/generateGpexpandConfig"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/runGpexpand/gpexpand"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/dns"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = ctrllog.Log.WithName("runGpexpand")

func main() {
	ctrllog.SetLogger(gplog.ForProd(false))

	var newPrimarySegmentCount = flag.Int("newPrimarySegmentCount", 0, "new primary segment count")
	flag.Parse()

	oldSegmentCount, err := GetOldSegmentCount(exec.Command)
	if err != nil {
		log.Error(err, "error getting existing segment count")
		os.Exit(1)
	}
	config, err := instanceconfig.NewReader(vfs.OS()).GetConfigValues()
	if err != nil {
		log.Error(err, "error reading configmap")
		os.Exit(1)
	}

	generateGpexpandConfig := &gpexpandconfig.GenerateGpexpandConfigParams{
		OldSegmentCount: oldSegmentCount,
		NewSegmentCount: *newPrimarySegmentCount,
		IsMirrored:      config.Mirrors,
		Fs:              vfs.OS(),
		Command:         exec.Command,
	}
	if err := generateGpexpandConfig.Run(); err != nil {
		log.Error(err, "error generating gpexpand configuration")
		os.Exit(1)
	}

	gpexpandRunner := &gpexpand.RunGpexpandConfig{
		Log:              log,
		NewSegmentCount:  *newPrimarySegmentCount,
		IsMirrored:       config.Mirrors,
		Standby:          config.Standby,
		Stdout:           os.Stdout,
		Stderr:           os.Stderr,
		DNSResolver:      dns.NewConsistentResolver(),
		KnownHostsWaiter: ssh.NewKnownHostsWaiter(),
		SSHExecutor:      ssh.NewMultiHostExec(fmt.Sprintf("/tools/waitForKnownHosts --newPrimarySegmentCount %d", *newPrimarySegmentCount)),
		Command:          exec.Command,
	}
	if err := gpexpandRunner.Run(); err != nil {
		log.Error(err, "error running gpexpand")
		os.Exit(1)
	}
}

func GetOldSegmentCount(command commandable.CommandFn) (int, error) {
	oldSegmentCount, err := gpexpandconfig.ExecPsqlQueryAndReturnInt(command, "SELECT COUNT(*) FROM gp_segment_configuration WHERE hostname LIKE 'segment-a%'")
	if err != nil {
		return 0, err
	}

	return oldSegmentCount, nil
}
