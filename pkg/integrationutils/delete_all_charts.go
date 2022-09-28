package integrationutils

import (
	"os/exec"
	"strings"

	. "github.com/onsi/gomega"
)

func DeleteAllCharts() {
	// delete every helm chart that is found
	out, err := exec.Command("helm", "list", "-aq").CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "error is %s", string(out))
	if len(out) == 0 {
		return
	}
	array := strings.Split(string(out), "\n")
	for _, chart := range array {
		if len(chart) == 0 {
			continue
		}

		out, err = exec.Command("helm", "uninstall", chart).CombinedOutput()
		if err != nil {
			log.Info("Problem with helm delete", "output", string(out))
		}
		Expect(err).ToNot(HaveOccurred())
	}
}
