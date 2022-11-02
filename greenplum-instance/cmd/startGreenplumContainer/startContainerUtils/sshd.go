package startContainerUtils

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
)

type SSHDaemon struct {
	*starter.App
}

func (s *SSHDaemon) Run(ctx context.Context) error {
	Log.Info("starting SSH Daemon")

	cmd := s.Command("/usr/bin/sudo", "/usr/sbin/sshd", "-D")
	cmd.Stdout = s.StdoutBuffer
	cmd.Stderr = s.StderrBuffer
	err := cmd.Start()
	if err != nil {
		Log.Error(err, "failed to start SSH in daemon mode")
		return err
	}

	sshErrorChan := make(chan error)
	go func(sshErrorChan chan<- error) {
		sshErrorChan <- cmd.Wait()
	}(sshErrorChan)

	select {
	// Got a SIGTERM: kill the sshd process
	case <-ctx.Done():
		Log.Info("killing sshd", "pid", cmd.Process.Pid)
		killCmd := s.Command("/usr/bin/sudo", "/bin/kill", "-SIGKILL", strconv.Itoa(cmd.Process.Pid))
		output, err := killCmd.CombinedOutput()
		if err != nil {
			Log.Error(err, "failed to kill sshd", "output", string(output))
			return nil
		}
		<-sshErrorChan
		return nil
	// The sshd process died: return an error
	case sshError := <-sshErrorChan:
		Log.Error(sshError, "sshd process terminated")
		return fmt.Errorf("sshd is not running: %w", sshError)
	}
}
