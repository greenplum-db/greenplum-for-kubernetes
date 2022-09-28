package integrationutils

import (
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubeexecpsql"
)

func EnsureOperatorIsDeployed(operatorOptions *SetupOperatorOptions) {
	cmd := exec.Command("kubectl", "get", "pods", "-l", "app=greenplum-operator")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		Expect(err).NotTo(HaveOccurred())
	}
	if !strings.Contains(string(output), "greenplum-operator") {
		SetupOperator(*operatorOptions)
	}
}

func EnsureNodesAreLabeled() {
	cmd := exec.Command("kubectl", "get", "nodes", "-l", "worker=my-gp-segments")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		Expect(err).NotTo(HaveOccurred())
	}
	if strings.Contains(string(output), "No resources found.") {
		LabelNodes()
	}
}

func EnsureGPDBIsDeployed(gpdbYamlFile string, greenplumImageTag string) {
	cmd := exec.Command("kubectl", "get", "greenplumclusters/my-greenplum")
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "my-greenplum") {
		SetupGreenplumCluster(gpdbYamlFile, greenplumImageTag)
	}
}

func EnsureDataIsLoaded() {
	out, err := Query("master-0", "SELECT * FROM foo")
	switch {
	case err == nil && strings.TrimSpace(string(out)) == "1":
		break
	case err != nil && strings.Contains(string(out), `relation "foo" does not exist`):
		LoadData()
	default:
		Fail(fmt.Sprintf("unexpected result checking for table `foo`\nerr: %e\noutput:\n%s\n", err, out))
	}
}
