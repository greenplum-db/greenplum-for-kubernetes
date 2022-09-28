package kubeexecpsql

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

var isSingleNodeOnce struct {
	sync.Once
	val bool
}

func IsSingleNode() bool {
	isSingleNodeOnce.Do(func() {
		var (
			err       error
			nodesInfo []byte
		)
		defer func() {
			if err != nil {
				log.Info("Output from node check:")
				fmt.Println(string(nodesInfo))
				panic(err)
			}
		}()
		cmd := exec.Command("kubectl", "--request-timeout=20s", "get", "nodes", "-o", "custom-columns=NAME:.metadata.name")
		nodesInfo, err = cmd.CombinedOutput()
		if err != nil {
			return
		}
		nodes := strings.Split(strings.TrimSuffix(string(nodesInfo), "\n"), "\n")
		if len(nodes) < 2 {
			err = errors.New("expected at least 2 lines of output")
			return
		}
		if nodes[0] != "NAME" {
			err = errors.New("unexpected output from command")
			return
		}
		isSingleNodeOnce.val = len(nodes) == 2
	})
	return isSingleNodeOnce.val
}

// TODO: This is perhaps not the best name. It is used in the former "IsMinikube" cases for which "IsSingleNode" doesn't make sense.
func ServicesAreOnLocalhostNodePort() bool {
	// TODO: find a better way to determine this.
	// History: we used to just check for a node named "minikube". Then minikube changed how it names nodes,
	// and in the meantime, "IsMinikube()" got used more frequently for situations that actually only cared
	// if there was one or many node(s). This function captures the remaining use cases.
	return IsSingleNode()
}
