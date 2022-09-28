package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/keyscanner"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/knownhosts"
	"github.com/pkg/errors"
)

const knownHostsFilename = "/home/gpadmin/.ssh/known_hosts"

type KeyScanApp struct {
	keyScanner       keyscanner.SSHKeyScannerInterface
	knownHostsReader knownhosts.ReaderInterface
	config           instanceconfig.Reader
	command          commandable.CommandFn
	stdoutBuffer     io.Writer
	stderrBuffer     io.Writer
	fileWriter       fileutil.FileWritable
}

func (k *KeyScanApp) scanSegmentHostKeys() error {
	dnsDomainCommand := k.command("dnsdomainname")
	dnsSuffixBytes, err := dnsDomainCommand.Output()
	if err != nil {
		return errors.New("scanning for segment host keys: dnsdomainname failed to determine this host's dns name")
	}
	dnsSuffix := "." + strings.TrimSuffix(string(dnsSuffixBytes), "\n")

	config, err := k.config.GetConfigValues()
	if err != nil {
		return err
	}

	hostList := net.GenerateHostList(config.SegmentCount, config.Mirrors, config.Standby, dnsSuffix)

	knownHosts, err := keyscanner.ScanHostKeys(k.keyScanner, k.knownHostsReader, hostList)
	if err != nil {
		return err
	}

	return errors.Wrapf(k.fileWriter.Append(knownHostsFilename, knownHosts),
		"failed to append known hosts to file: %v", knownHostsFilename)
}

func (k *KeyScanApp) Run() int {
	fmt.Fprintln(k.stdoutBuffer, "Key scanning started")
	err := k.scanSegmentHostKeys()
	if err != nil {
		fmt.Fprintln(k.stderrBuffer, err)
		return 1
	}
	return 0
}

func main() {
	fs := vfs.OS()
	scanner := KeyScanApp{
		keyScanner:       keyscanner.NewSSHKeyScanner(),
		knownHostsReader: knownhosts.NewReader(),
		config:           instanceconfig.NewReader(fs),
		command:          exec.Command,
		stdoutBuffer:     os.Stdout,
		stderrBuffer:     os.Stderr,
		fileWriter:       &fileutil.FileWriter{WritableFileSystem: fs},
	}
	os.Exit(scanner.Run())
}
