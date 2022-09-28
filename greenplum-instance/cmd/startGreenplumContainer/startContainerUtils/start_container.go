package startContainerUtils

import (
	"fmt"

	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
)

type GreenplumContainerStarter struct {
	*starter.App
	UID int

	Root, Gpadmin, LabelPVC, MultidaemonStarter starter.Starter
}

func (s *GreenplumContainerStarter) Run(args []string) (status int) {
	if len(args) == 2 && args[1] == "--do-root-startup" {
		if s.UID != 0 {
			fmt.Fprintf(s.StderrBuffer, "--do-root-startup was passed, but we are not root")
			return 1
		}

		if err := s.Root.Run(); err != nil {
			fmt.Fprintln(s.StderrBuffer, err.Error())
			return 1
		}
		return 0
	} else if len(args) != 1 {
		fmt.Fprintln(s.StderrBuffer, "Unexpected argument(s):", args[1:])
		return 1
	}

	// Get root privileges to perform root startup tasks
	cmd := s.Command("/usr/bin/sudo", args[0], "--do-root-startup")
	cmd.Stdout = s.StdoutBuffer
	cmd.Stderr = s.StderrBuffer
	if err := cmd.Run(); err != nil {
		return
	}

	starters := []starter.Starter{
		s.Gpadmin,
		s.LabelPVC,
		s.MultidaemonStarter,
	}
	for _, step := range starters {
		if err := step.Run(); err != nil {
			fmt.Fprintln(s.StderrBuffer, err)
			return 1
		}
	}
	return 0
}
