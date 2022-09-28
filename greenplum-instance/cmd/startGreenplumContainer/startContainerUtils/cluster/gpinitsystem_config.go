package cluster

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig"
	"github.com/pkg/errors"
)

const GpinitsystemConfigPath = "/home/gpadmin/gpinitsystem_config"
const gpinitsystemSuccessExitStatus = "exit status 1"

type GpInitSystem interface {
	GenerateConfig() error
	Run() error
}

type gpInitSystem struct {
	Filesystem       vfs.Filesystem
	Command          commandable.CommandFn
	Stdout           io.Writer
	Stderr           io.Writer
	greenplumCommand *GreenplumCommand
	configReader     instanceconfig.Reader
}

func NewGpInitSystem(fs vfs.Filesystem, command commandable.CommandFn, stdOut, stdErr io.Writer, configReader instanceconfig.Reader) GpInitSystem {
	return &gpInitSystem{
		Filesystem:       fs,
		Command:          command,
		Stdout:           stdOut,
		Stderr:           stdErr,
		greenplumCommand: NewGreenplumCommand(command),
		configReader:     configReader,
	}
}

func (g *gpInitSystem) GenerateConfig() error {
	PrintMessage(g.Stdout, "Generating gpinitsystem_config")
	segmentCount, err := g.configReader.GetSegmentCount()
	if err != nil {
		return err
	}
	useMirrors, err := g.configReader.GetMirrors()
	if err != nil {
		return err
	}

	cmd := g.Command("dnsdomainname")
	output, err := cmd.Output()
	if err != nil {
		return errors.Wrap(err, "dnsdomainname failed to determine this host's dns name")
	}
	subdomain := strings.TrimSuffix(string(output), "\n")
	fmt.Fprintln(g.Stdout, "Sub Domain for the cluster is:", subdomain)

	configFile, err := g.Filesystem.OpenFile(GpinitsystemConfigPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	dbID := 1
	fmt.Fprintf(configFile, "QD_PRIMARY_ARRAY=master-0.%v~5432~/greenplum/data-1~%d~-1~0\n", subdomain, dbID)
	dbID++
	fmt.Fprint(configFile, "declare -a PRIMARY_ARRAY=(\n")
	for segment := 0; segment < segmentCount; segment++ {
		fmt.Fprintf(configFile, "segment-a-%d.%v~40000~/greenplum/data~%d~%d\n", segment, subdomain, dbID, segment)
		dbID++
	}
	fmt.Fprint(configFile, ")\n")
	if useMirrors {
		fmt.Fprint(configFile, "declare -a MIRROR_ARRAY=(\n")
		for segment := 0; segment < segmentCount; segment++ {
			// We must use a different directory for mirrors, because gpinitsystem enforces this to make sure that on
			// bare metal systems that primaries and mirrors don't share storage.
			// https://github.com/greenplum-db/gpdb/blob/5X_STABLE/gpMgmt/bin/gpinitsystem#L460
			// TODO: enhance gpinitsystem to consider the hostname as well? i.e., sdw1:/data != sdw2:/data
			fmt.Fprintf(configFile, "segment-b-%d.%v~50000~/greenplum/mirror/data~%d~%d\n", segment, subdomain, dbID, segment)
			dbID++
		}
		fmt.Fprint(configFile, ")\n")
	}
	fmt.Fprint(configFile, "HBA_HOSTNAMES=1\n")
	return configFile.Close()
}

func (g *gpInitSystem) Run() error {
	PrintMessage(g.Stdout, "Running gpinitsystem")

	dnsDomainCommand := g.Command("dnsdomainname")
	dnsSuffixBytes, err := dnsDomainCommand.Output()
	if err != nil {
		return errors.Wrap(err, "dnsdomainname failed to determine this host's dns name")
	}
	dnsSuffix := strings.TrimSuffix(string(dnsSuffixBytes), "\n")

	args := []string{"-a", "-I", GpinitsystemConfigPath}

	if standby, err := g.configReader.GetStandby(); err != nil {
		return err
	} else if standby {
		args = append(args, []string{"-s", "master-1." + dnsSuffix}...)
	}

	_, err = g.Filesystem.Lstat("/etc/config/GUCs")
	if err == nil {
		args = append(args, "-p", "/etc/config/GUCs")
	}

	cmd := g.greenplumCommand.Command("/usr/local/greenplum-db/bin/gpinitsystem", args...)
	cmd.Stdout = g.Stdout
	cmd.Stderr = g.Stderr
	err = cmd.Run()
	if err != nil {
		if err.Error() != gpinitsystemSuccessExitStatus {
			return errors.Wrap(err, "gpinitsystem failed")
		}
	}

	return nil
}
