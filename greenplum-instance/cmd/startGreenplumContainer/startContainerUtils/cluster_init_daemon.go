package startContainerUtils

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils/cluster"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/keyscanner"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/knownhosts"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/ubuntuUtils"
)

const knownHostsFilename = "/home/gpadmin/.ssh/known_hosts"

type ClusterInitDaemon struct {
	*starter.App
	Config           instanceconfig.Reader
	Ubuntu           ubuntuUtils.UbuntuInterface
	DNSResolver      multihost.Operation
	KeyScanner       keyscanner.SSHKeyScannerInterface
	KnownHostsReader knownhosts.ReaderInterface
	C                cluster.ClusterInterface
}

func (s *ClusterInitDaemon) Run(_ context.Context) error {
	go func() {
		if err := s.InitializeCluster(); err != nil {
			Log.Error(err, "failed to initialize cluster")
		}
	}()
	return nil
}

// TODO: Remove return type of integer after we removed initializeCluster executable
func (s *ClusterInitDaemon) InitializeCluster() error {
	hostname, err := s.Ubuntu.Hostname()
	if err != nil {
		Log.Error(err, "getting hostname")
		return err
	}

	return s.NewPostgresInitializer(hostname).InitializePostgres()
}

func (s *ClusterInitDaemon) SetupPasswordlessSSH(hostnameList []string) error {
	Log.Info("started SSH KeyScan")
	knownHosts, err := keyscanner.ScanHostKeys(s.KeyScanner, s.KnownHostsReader, hostnameList)
	if err != nil {
		return fmt.Errorf("failed to scan segment host keys: %s", err.Error())
	}
	fileWriter := &fileutil.FileWriter{WritableFileSystem: s.Fs}
	if err = fileWriter.Append(knownHostsFilename, knownHosts); err != nil {
		return fmt.Errorf("failed to write known_hosts file: %s", err.Error())
	}
	return nil
}

type PostgresInitializer interface {
	InitializePostgres() error
}

func (s *ClusterInitDaemon) NewPostgresInitializer(hostname string) PostgresInitializer {
	switch hostname {
	case "master-0":
		return &masterPostgresInitializer{postgresInitializer{
			clusterStarter: s,
			hostname:       hostname,
			dataDir:        "/greenplum/data-1",
		}}
	case "master-1":
		return &segmentPostgresInitializer{postgresInitializer{
			clusterStarter: s,
			hostname:       hostname,
			dataDir:        "/greenplum/data-1",
		}}
	default:
		return &segmentPostgresInitializer{postgresInitializer{
			clusterStarter: s,
			hostname:       hostname,
			dataDir:        "/greenplum/data",
		}}
	}
}

type postgresInitializer struct {
	clusterStarter *ClusterInitDaemon
	hostname       string
	dataDir        string
}

type masterPostgresInitializer struct {
	postgresInitializer
}

type segmentPostgresInitializer struct {
	postgresInitializer
}

var _ PostgresInitializer = &masterPostgresInitializer{}

func (i *masterPostgresInitializer) InitializePostgres() error {
	dnsDomainCommand := i.clusterStarter.Command("dnsdomainname")
	dnsSuffixBytes, err := dnsDomainCommand.Output()
	if err != nil {
		Log.Error(err, "scanning for segment host keys: dnsdomainname failed to determine this host's dns name")
		return err
	}
	dnsSuffix := "." + strings.TrimSuffix(string(dnsSuffixBytes), "\n")

	config, err := i.clusterStarter.Config.GetConfigValues()
	if err != nil {
		Log.Error(err, "error reading configmap")
		return err
	}

	hostnameList := net.GenerateHostList(config.SegmentCount, config.Mirrors, config.Standby, dnsSuffix)

	// Block until all hosts are Ready
	Log.Info("resolving DNS entries for all masters and segments")
	if errs := multihost.ParallelForeach(i.clusterStarter.DNSResolver, hostnameList); len(errs) != 0 {
		return errors.New("failed to resolve dns entries for all masters and segments")
	}

	// TODO: Wait for passwordless SSH to be setup by the instance-controller
	//       instead of manually scanning host keys
	if err := i.clusterStarter.SetupPasswordlessSSH(hostnameList); err != nil {
		Log.Error(err, "error setting up passwordless SSH")
		return err
	}

	if i.CheckPreinitalizedCluster() {
		Log.Info("cluster has been initialized before; starting Greenplum Cluster")
		if err := i.checkAndRunGPStart(config.Standby); err != nil {
			return err
		}
	} else {
		Log.Info("initializing Greenplum Cluster")
		if err := i.clusterStarter.C.Initialize(); err != nil {
			return err
		}
	}

	return i.clusterStarter.C.RunPostInitialization()
}

func (i *masterPostgresInitializer) checkAndRunGPStart(hasStandby bool) error {
	if hasStandby {
		Log.Info("Automatic gpstart is not currently supported with standby masters. Skipping.")
		return nil
	}
	return i.clusterStarter.C.GPStart()
}

var _ PostgresInitializer = &segmentPostgresInitializer{}

func (i *segmentPostgresInitializer) InitializePostgres() error {
	if i.CheckPreinitalizedCluster() {
		Log.Info("cluster has been initialized before; starting Postgres")
		return i.pgCtlRestart()
	}
	return nil
}

func (i *postgresInitializer) CheckPreinitalizedCluster() bool {
	_, err := i.clusterStarter.Fs.Stat(i.dataDir)
	return err == nil
}

func (i *postgresInitializer) pgCtlRestart() error {
	startupLog := filepath.Join(i.dataDir, "pg_log", "startup.log")
	cmd := cluster.NewGreenplumCommand(i.clusterStarter.Command).Command("/usr/local/greenplum-db/bin/pg_ctl", "-D", i.dataDir, "-l", startupLog, "restart")
	cmd.Stderr = i.clusterStarter.StderrBuffer
	cmd.Stdout = i.clusterStarter.StdoutBuffer

	if err := cmd.Run(); err != nil {
		return errors.New("pg_ctl failed to restart")
	}
	return nil
}
