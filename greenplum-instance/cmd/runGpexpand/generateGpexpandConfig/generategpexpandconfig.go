package gpexpandconfig

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils/cluster"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pkg/errors"
)

type GenerateGpexpandConfigParams struct {
	maxDbID         int
	maxContentID    int
	namespace       string
	OldSegmentCount int
	NewSegmentCount int
	IsMirrored      bool
	Fs              vfs.Filesystem
	Command         commandable.CommandFn
}

func (p *GenerateGpexpandConfigParams) Run() error {
	var err error
	if p.NewSegmentCount <= 0 {
		return errors.New("new segment count cannot be <= 0")
	}

	if err = p.SetMaxDbID(); err != nil {
		return err
	}

	if err = p.SetMaxContentID(); err != nil {
		return err
	}

	if err = p.SetNamespace(); err != nil {
		return err
	}

	if err = p.GenerateConfig(); err != nil {
		return err
	}
	return nil
}

func (p *GenerateGpexpandConfigParams) GenerateConfig() error {
	if p.NewSegmentCount <= p.OldSegmentCount {
		return errors.New("newSegmentCount cannot be less than or equal to OldSegmentCount")
	}
	dbid := p.maxDbID
	contentID := p.maxContentID
	var configBuilder strings.Builder
	const gpexpandFmt = "%s-%d.agent.%s.svc.cluster.local|%s-%d|%d|%s|%d|%d|%s\n"
	for i := p.OldSegmentCount; i < p.NewSegmentCount; i++ {
		dbid++
		contentID++
		configBuilder.WriteString(fmt.Sprintf(gpexpandFmt, "segment-a", i, p.namespace, "segment-a",
			i, 40000, "/greenplum/data", dbid, contentID, "p"))
		if p.IsMirrored {
			dbid++
			configBuilder.WriteString(fmt.Sprintf(gpexpandFmt, "segment-b", i, p.namespace, "segment-b",
				i, 50000, "/greenplum/mirror/data", dbid, contentID, "m"))
		}
	}
	return vfs.WriteFile(p.Fs, "/tmp/gpexpand_config", []byte(configBuilder.String()), 0777)
}

func (p *GenerateGpexpandConfigParams) SetMaxDbID() (err error) {
	p.maxDbID, err = ExecPsqlQueryAndReturnInt(p.Command, "SELECT MAX(dbid) FROM gp_segment_configuration")
	return
}

func (p *GenerateGpexpandConfigParams) SetMaxContentID() (err error) {
	p.maxContentID, err = ExecPsqlQueryAndReturnInt(p.Command, "SELECT MAX(content) FROM gp_segment_configuration")
	return
}

func (p *GenerateGpexpandConfigParams) SetNamespace() error {
	ns, err := vfs.ReadFile(p.Fs, "/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return err
	}
	p.namespace = string(ns)
	return err
}

func ExecPsqlQueryAndReturnInt(command commandable.CommandFn, query string) (int, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	greenplumCommand := cluster.NewGreenplumCommand(command)
	cmd := greenplumCommand.Command("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-tAc", query)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return -1, errors.Wrap(err, stderr.String())
	}
	intResult, err := strconv.Atoi(strings.TrimSpace(stdout.String()))
	if err != nil {
		return -1, err
	}
	return intResult, nil
}
