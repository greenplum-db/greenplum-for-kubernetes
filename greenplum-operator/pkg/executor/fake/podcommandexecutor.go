package fake

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const DefaultSegmentCount = 1 // Used as the primarySegmentCount of exampleGreenplumCluster

type PodExec struct {
	ErrorMsgOnMaster0 string
	ErrorMsgOnMaster1 string

	SegmentCount    string
	SegmentCountErr error

	ErrorMsgOnCommand string
	CalledPodName     string

	RecordedCommands []string
	StdoutResult     string
}

// TODO: break import cycle so we can make this assertion
//var _ executor.PodExecInterface = &PodExec{}

func (f *PodExec) Execute(command []string, namespace, podName string, stdout, stderr io.Writer) error {
	cmdStr := strings.Join(command, " ")
	switch {
	case isActiveMasterQuery(cmdStr):
		return f.handleActiveMasterQuery(command, podName)
	case isSegmentCountQuery(cmdStr):
		if f.SegmentCountErr != nil {
			return f.SegmentCountErr
		}
		segCount := strconv.Itoa(DefaultSegmentCount) + "\n"
		if f.SegmentCount != "" {
			segCount = f.SegmentCount
		}
		_, err := io.WriteString(stdout, segCount)
		return err
	case f.ErrorMsgOnCommand != "":
		f.CalledPodName = podName
		fmt.Fprintf(stderr, f.ErrorMsgOnCommand)
		return errors.New(f.ErrorMsgOnCommand)
	default:
		f.CalledPodName = podName
		f.RecordedCommands = append(f.RecordedCommands, cmdStr)
		_, err := io.WriteString(stdout, f.StdoutResult)
		return err
	}
}

func isSegmentCountQuery(cmdStr string) bool {
	return strings.Contains(cmdStr, "SELECT COUNT(*) FROM gp_segment_configuration")
}

func isActiveMasterQuery(cmdStr string) bool {
	return strings.Contains(cmdStr, "psql -U gpadmin -c 'select * from gp_segment_configuration'")
}

func (f *PodExec) handleActiveMasterQuery(command []string, podName string) error {
	if f.ErrorMsgOnMaster0 != "" && podName == "master-0" {
		return errors.New(f.ErrorMsgOnMaster0)
	} else if f.ErrorMsgOnMaster1 != "" && podName == "master-1" {
		return errors.New(f.ErrorMsgOnMaster1)
	}

	return nil
}
