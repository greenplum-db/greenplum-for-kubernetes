package cluster

import (
	"fmt"
	"io"
	"os"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig"
	"github.com/pkg/errors"
)

const (
	masterDataDir = "/greenplum/data-1"
)

type ClusterInterface interface {
	Initialize() error
	GPStart() error
	RunPostInitialization() error
}

type Cluster struct {
	Filesystem vfs.Filesystem
	Command    commandable.CommandFn
	Stdout     io.Writer
	Stderr     io.Writer
	Config     instanceconfig.Reader

	gpInitSystem     GpInitSystem
	greenplumCommand *GreenplumCommand
}

var _ ClusterInterface = &Cluster{}

func New(
	fileSystem vfs.Filesystem,
	command commandable.CommandFn,
	stdout io.Writer,
	stderr io.Writer,
	config instanceconfig.Reader,
	gpInitSystem GpInitSystem,
) *Cluster {
	return &Cluster{
		Filesystem:       fileSystem,
		Command:          command,
		Stdout:           stdout,
		Stderr:           stderr,
		Config:           config,
		gpInitSystem:     gpInitSystem,
		greenplumCommand: NewGreenplumCommand(command),
	}
}

func (c *Cluster) Initialize() error {
	PrintMessage(c.Stdout, "Initializing Greenplum for Kubernetes Cluster")
	_, err := c.Filesystem.Stat(masterDataDir)
	if _, ok := err.(*os.PathError); !ok {
		return fmt.Errorf("master data directory already exists at %s", masterDataDir)
	}

	if err := c.gpInitSystem.GenerateConfig(); err != nil {
		return fmt.Errorf("gpinitsystem config failed: %w", err)
	}
	if err := c.gpInitSystem.Run(); err != nil {
		return fmt.Errorf("gpinitsystem failed: %w", err)
	}

	if err := c.createDB(); err != nil {
		return fmt.Errorf("createdb failed: %w", err)
	}

	// We reload the HBA config in RunPostInitialization
	return c.addMasterAndStandbyHostBasedAuthentication()
}

func (c *Cluster) createDB() error {
	PrintMessage(c.Stdout, "Running createdb")
	cmd := c.greenplumCommand.Command("/usr/local/greenplum-db/bin/createdb")
	cmd.Stderr = c.Stderr
	cmd.Stdout = c.Stdout
	return cmd.Run()
}

func (c *Cluster) addMasterAndStandbyHostBasedAuthentication() error {
	if err := c.addHostBasedAuthentication("master-0"); err != nil {
		return fmt.Errorf("adding host-based authentication failed: %w", err)
	}
	standby, err := c.Config.GetStandby()
	if err != nil {
		return fmt.Errorf("reading standby failed: %w", err)
	}

	if standby {
		if err = c.addHostBasedAuthentication("master-1"); err != nil {
			return fmt.Errorf("adding host-based authentication failed: %w", err)
		}
	}

	return nil
}

func (c *Cluster) addHostBasedAuthentication(host string) error {
	source := "/etc/config/hostBasedAuthentication"
	hasContent, err := fileutil.HasContent(c.Filesystem, source)
	if err != nil {
		return errors.Wrapf(err, "verifying if %v has any content failed", source)
	}
	if hasContent {
		PrintMessage(c.Stdout, "Adding host based authentication to "+host+" pg_hba.conf")
		destination := "/greenplum/data-1/pg_hba.conf"
		cmd := c.Command("/usr/bin/ssh", host, "cat", source, ">>", destination)
		cmd.Stderr = c.Stderr
		cmd.Stdout = c.Stdout

		err := cmd.Run()
		return errors.Wrap(err, "Attempting to append from '"+source+"' to end of "+destination)
	}
	return nil
}

func (c *Cluster) GPStart() error {
	cmd := c.greenplumCommand.Command("/usr/local/greenplum-db/bin/gpstart", "-am")
	cmd.Stderr = c.Stderr
	cmd.Stdout = c.Stdout

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("gpstart in maintenance mode failed: %w", err)
	}

	cmd = c.greenplumCommand.Command("/usr/local/greenplum-db/bin/gpstop", "-ar")
	cmd.Stderr = c.Stderr
	cmd.Stdout = c.Stdout

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("restart segments failed: %w", err)
	}
	return nil
}

func (c *Cluster) RunPostInitialization() error {
	cmd := c.greenplumCommand.Command("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-c", "select * from gp_segment_configuration")
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(c.Stdout, "the database is not running. skipping post-initialization.")
		return nil
	}

	if err := c.reloadGpConfig(); err != nil {
		return fmt.Errorf("gpstop failed: %w", err)
	}

	if err := c.createPXFExtension(); err != nil {
		return fmt.Errorf("createPXFExtension failed: %w", err)
	}

	return nil
}

func (c *Cluster) reloadGpConfig() error {
	PrintMessage(c.Stdout, "Reloading greenplum configs")
	cmd := c.greenplumCommand.Command("/usr/local/greenplum-db/bin/gpstop", "-u")
	cmd.Stderr = c.Stderr
	cmd.Stdout = c.Stdout
	return cmd.Run()
}

func (c *Cluster) createPXFExtension() error {
	pxfServiceName, err := c.Config.GetPXFServiceName()
	if err != nil {
		return err
	}
	if pxfServiceName != "" {
		return c.createExtension("pxf")
	}
	return nil
}

// TODO: turn this into its own starter.Starter.
//   Then make a []starter.Starter to run gpinitsystem and pxf
func (c *Cluster) createExtension(extensionName string) error {
	PrintMessage(c.Stdout, fmt.Sprintf("Creating %s Extension", extensionName))
	cmd := c.greenplumCommand.Command("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-d", "gpadmin", "-c", fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", extensionName))
	cmd.Stderr = c.Stderr
	cmd.Stdout = c.Stdout
	return cmd.Run()
}

func PrintMessage(writer io.Writer, message string) {
	fmt.Fprintln(writer, "*******************************")
	fmt.Fprintln(writer, message)
	fmt.Fprintln(writer, "*******************************")
}
