package cluster

import (
	"os/exec"

	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
)

type GreenplumCommand struct {
	command commandable.CommandFn
}

func NewGreenplumCommand(cmd commandable.CommandFn) *GreenplumCommand {
	return &GreenplumCommand{command: cmd}
}

func (g *GreenplumCommand) Command(command string, args ...string) *exec.Cmd {
	cmd := g.command(command, args...)
	cmd.Env = append(cmd.Env,
		"HOME=/home/gpadmin",
		"USER=gpadmin",
		"LOGNAME=gpadmin",
		"GPHOME=/usr/local/greenplum-db",
		"PATH=/usr/local/greenplum-db/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"LD_LIBRARY_PATH=/usr/local/greenplum-db/lib:/usr/local/greenplum-db/ext/python/lib",
		"MASTER_DATA_DIRECTORY=/greenplum/data-1",
		"PYTHONHOME=/usr/local/greenplum-db/ext/python",
		"PYTHONPATH=/usr/local/greenplum-db/lib/python")
	return cmd
}
