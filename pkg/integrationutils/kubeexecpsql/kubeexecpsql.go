package kubeexecpsql

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func QueryWithRetry(hostName, command string) ([]byte, error) {
	attempts := 3

	var out []byte
	var err error

	for i := 0; i < attempts; i++ {
		psqlQuery := fmt.Sprintf("source /usr/local/greenplum-db/greenplum_path.sh; psql -tAc '%v'", command)
		cmd := exec.Command("kubectl", "exec", "-i", hostName, "--", "bash", "-c", psqlQuery)
		out, err = cmd.CombinedOutput()

		if err == nil {
			return out, nil
		}
		log.Info("retrying kubectl exec", "output", string(out), "attempt", fmt.Sprintf("%d/%d", i+1, attempts), "command", command)
	}

	return out, err
}

func Query(hostName, command string) ([]byte, error) {
	psqlQuery := fmt.Sprintf(`source /usr/local/greenplum-db/greenplum_path.sh; psql -tAc "%v"`, command)
	cmd := exec.Command("kubectl", "exec", "-i", hostName, "--", "bash", "-c", psqlQuery)
	return cmd.CombinedOutput()
}

func ExecutePsqlQueryThroughService(query string) ([]byte, error) {
	greenplumServiceIP, greenplumPort, err := GreenplumService()
	if err != nil {
		return []byte{}, err
	}
	// on macos, no need to source
	psql := fmt.Sprintf(`psql -U gpadmin -h %s -p %s -tAc "%s"`, greenplumServiceIP, greenplumPort, query)
	return exec.Command("bash", "-c", psql).CombinedOutput()
}

func GreenplumService() (string, string, error) {
	var greenplumServiceIP string
	var ip []byte
	var greenplumPort = "5432"
	var err error
	// TODO: find a better check: Running locally or not
	if IsSingleNode() {
		port, err := exec.Command("kubectl", "get", "service", "greenplum", "-o", "jsonpath={.spec.ports[0].nodePort}").CombinedOutput()
		if err != nil {
			return "", "", err
		}
		if nodeIP, err := GetResolvableInternalNodeIP(port); err == nil {
			ip = []byte(nodeIP)
		}
		greenplumPort = strings.TrimSpace(string(port))
	} else {
		ip, err = exec.Command("kubectl", "get", "service", "greenplum", "-o", "jsonpath={.status.loadBalancer.ingress[0].ip}").CombinedOutput()
		if err != nil {
			return "", "", err
		}
	}
	greenplumServiceIP = strings.TrimSpace(string(ip))
	return greenplumServiceIP, greenplumPort, nil
}

func GetResolvableInternalNodeIP(nodePort []byte) (string, error) {
	out, err := exec.Command("kubectl", "get", "nodes", "-o", "json").CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "should be able to query for nodes", string(out))
	var nodeList corev1.NodeList
	Expect(json.Unmarshal(out, &nodeList)).To(Succeed())

	for _, node := range nodeList.Items {
		addresses := node.Status.Addresses
		for _, address := range addresses {
			if address.Type == corev1.NodeInternalIP {
				host := fmt.Sprintf("%s:%s", address.Address, nodePort)
				conn, err := net.DialTimeout("tcp", host, time.Duration(3)*time.Second)
				defer conn.Close()

				if err != nil {
					continue
				}
				return address.Address, nil
			}
		}
	}

	return "", errors.New("failed to resolve any node ports")
}
