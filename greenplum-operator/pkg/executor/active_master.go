package executor

import (
	"io/ioutil"
)

func GetCurrentActiveMaster(p PodExecInterface, namespace string) string {
	testIfPrimaryMasterCommand := []string{
		"/bin/bash",
		"-c",
		"--",
		"source /usr/local/greenplum-db/greenplum_path.sh && psql -U gpadmin -c 'select * from gp_segment_configuration'",
	}

	stdout, stderr := ioutil.Discard, ioutil.Discard
	err := p.Execute(testIfPrimaryMasterCommand, namespace, "master-0", stdout, stderr)
	if err == nil {
		return "master-0"
	}
	log.V(1).Info("master-0 is not active master", "namespace", namespace, "error", err)

	err = p.Execute(testIfPrimaryMasterCommand, namespace, "master-1", stdout, stderr)
	if err == nil {
		return "master-1"
	}
	log.V(1).Info("master-1 is not active master", "namespace", namespace, "error", err)

	return ""
}
