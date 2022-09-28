package kubewait

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubeexecpsql"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/dns"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/util/errors"
	apiwait "k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("kubewait")

var timeoutDurationOnce struct {
	sync.Once
	val time.Duration
}

func getPollTimeout() time.Duration {
	timeoutDurationOnce.Do(func() {
		const defaultTimeout = 120 * time.Second
		pollTimeoutEnv, exist := os.LookupEnv("POLL_TIMEOUT")
		if exist {
			pollTimeout, err := strconv.Atoi(pollTimeoutEnv)
			if err == nil {
				log.Info("Setting timeout from POLL_TIMEOUT", "timeoutSeconds", pollTimeout)
				timeoutDurationOnce.val = time.Duration(pollTimeout) * time.Second
				return
			}
		}

		log.Info("Setting timeout to default", "timeoutSeconds", defaultTimeout.Seconds())
		timeoutDurationOnce.val = defaultTimeout
	})
	return timeoutDurationOnce.val
}

func ForPodState(podName string, state string) error {
	return errors.Wrapf(apiwait.PollImmediate(1*time.Second, getPollTimeout(), func() (bool, error) {
		cmd := exec.Command("kubectl", "get", "pods")
		out, err := cmd.CombinedOutput()

		if err != nil {
			return false, nil
		}

		if podName == "greenplum-operator" {
			out, _ = exec.Command("bash", "-c", "kubectl get pods | grep greenplum-operator").CombinedOutput()
		} else {
			out, _ = exec.Command("kubectl", "get", "pods", podName).CombinedOutput()
		}
		if strings.Contains(string(out), podName) && strings.Contains(string(out), state) {
			// found something, finished
			return true, nil
		}

		// retry
		return false, nil
	}), "waiting for pod %s to become %s", podName, state)
}

func ForOperatorPod() error {
	return errors.Wrap(ForReplicasReady("deployment", "greenplum-operator"), "waiting for greenplum operator")
}

func getDesiredReplicas(kind string, objectName string) (desiredReplicas int, err error) {
	err = errors.Wrapf(apiwait.PollImmediate(1*time.Second, getPollTimeout(), func() (bool, error) {
		cmd := exec.Command("kubectl", "get", kind, objectName, "-o", "jsonpath={.spec.replicas}")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return false, nil
		}
		desiredReplicas, err = strconv.Atoi(string(out))
		if err != nil {
			return false, err
		}
		return true, nil
	}), "getting desired replicas from %s/%s", kind, objectName)
	return
}

func ForPodDestroyed(podName string) error {
	return errors.Wrapf(apiwait.PollImmediate(1*time.Second, getPollTimeout(), func() (bool, error) {
		cmd := exec.Command("kubectl", "get", "pods")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return false, nil
		}

		if !strings.Contains(string(out), podName) {
			// pod is gone
			return true, nil
		}

		// retry
		return false, nil
	}), "waiting for pod to be destroyed: %s", podName)
}

func ForDNSRefresh(targetPodName string, fromPodName string) error {
	return errors.Wrapf(apiwait.PollImmediate(1*time.Second, getPollTimeout(), func() (bool, error) {
		return performDNSLookup(targetPodName, fromPodName), nil
	}), "waiting for dns refresh %s -> %s", fromPodName, targetPodName)
}

// TODO: replace this with a call to a binary on master that check dns consistency
func ForConsistentDNSResolution(targetPodName string, fromPodName string) error {
	return dns.NewConsistentResolver().PollUntilConsistent(func() bool {
		return performDNSLookup(targetPodName, fromPodName)
	})
}

func performDNSLookup(targetPodName, fromPodName string) bool {
	cmd := exec.Command("kubectl", "exec", "-i", fromPodName, "--", "ping", "-c", "1", targetPodName)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func ForGreenplumInitialization() error {
	return errors.Wrap(apiwait.PollImmediate(1*time.Second, 500*time.Second, func() (bool, error) {
		queryOutput, err := kubeexecpsql.Query("master-0", "select * from gp_segment_configuration")
		if err != nil {
			return false, nil
		}
		if len(queryOutput) == 0 {
			return false, nil
		}
		return true, nil
	}), "waiting for gp_segment_configuration")
}

func ForGreenplumInitializationWithService() error {
	return errors.Wrap(apiwait.PollImmediate(1*time.Second, 500*time.Second, func() (bool, error) {
		queryOutput, err := kubeexecpsql.ExecutePsqlQueryThroughService("select * from gp_segment_configuration")
		if err != nil {
			return false, nil
		}
		if len(queryOutput) == 0 {
			return false, nil
		}
		return true, nil
	}), "waiting for gp_segment_configuration (via service)")
}

func ForResource(kind, name string) error {
	return errors.Wrapf(apiwait.PollImmediate(1*time.Second, getPollTimeout(), func() (bool, error) {
		cmd := exec.Command("kubectl", "get", kind, name, "-o", "jsonpath={.metadata.name}")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return false, nil
		}
		if string(out) == name {
			return true, nil
		}
		return false, nil
	}), "waiting for %s/%s", kind, name)
}

func ForAPIResource(apigroup, name string) error {
	return errors.Wrapf(apiwait.PollImmediate(1*time.Second, getPollTimeout(), func() (bool, error) {
		cmd := exec.Command("kubectl", "api-resources", "--api-group", apigroup, "-o", "name")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return false, nil
		}
		for _, resource := range strings.Split(string(out), "\n") {
			if resource == name+"."+apigroup {
				return true, nil
			}
		}
		return false, nil
	}), "waiting for %s/%s", apigroup, name)
}

func ForGreenplumClusterStatus(status greenplumv1.GreenplumClusterPhase) error {
	return errors.Wrapf(apiwait.PollImmediate(1*time.Second, getPollTimeout(), func() (bool, error) {
		out, err := exec.Command("kubectl", "get", "greenplumCluster", "-o", "jsonpath={.items[0].status.phase}").CombinedOutput()
		if err != nil {
			return false, nil
		}
		if string(out) == string(status) {
			return true, nil
		}
		return false, nil
	}), "waiting for GreenplumCluster status %s", status)
}

func ForReplicasReady(kind string, objectName string) error {
	desiredReplicas, err := getDesiredReplicas(kind, objectName)
	if err != nil {
		return err
	}

	log.Info("Waiting for all replicas to be ready ...", "kind", kind, "name", objectName)

	err = errors.Wrapf(apiwait.PollImmediate(1*time.Second, 240*time.Second, func() (bool, error) {
		cmd := exec.Command("kubectl", "get", kind, objectName, "-o", "jsonpath={.status.readyReplicas}")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(out))
			return false, nil
		}

		if strings.TrimSpace(string(out)) == strconv.Itoa(desiredReplicas) {
			return true, nil
		}

		// retry
		return false, nil
	}), "waiting for %s/%s", kind, objectName)

	if err != nil {
		out, _ := exec.Command("kubectl", "get", "all").CombinedOutput()
		fmt.Println(string(out))
		return err
	}

	log.Info("all replicas are ready.", "kind", kind, "name", objectName)
	return nil
}

func ForNetworkReady(primarySegmentCount int, useMirrors bool) error {
	log.Info("waiting for all segments to be network reachable ...")
	totalNumSegs := primarySegmentCount * 2
	if !useMirrors {
		totalNumSegs = primarySegmentCount
	}
	results := make(chan error, totalNumSegs)
	for segmentID := 0; segmentID < primarySegmentCount; segmentID++ {
		go waitForDNSRefreshSegment("segment-a-", segmentID, results)
		if useMirrors {
			go waitForDNSRefreshSegment("segment-b-", segmentID, results)
		}
	}
	errs := make([]error, 0, cap(results))
	for i := 0; i < cap(results); i++ {
		errs = append(errs, <-results)
	}
	return errors.Wrap(k8serrors.NewAggregate(errs), "waiting ForNetworkReady")
}

func waitForDNSRefreshSegment(segmentSet string, segID int, results chan error) {
	segmentName := fmt.Sprint(segmentSet, segID)
	results <- func() error {
		err := ForDNSRefresh(segmentName, "master-0")
		if err != nil {
			log.Error(err, "failed to probe segment", "segment", segmentName)
			return err
		}
		log.Info("segment is ready.", "segment", segmentName)
		return nil
	}()
}

func ForGreenplumService(name string) error {
	var jsonPath string

	if kubeexecpsql.ServicesAreOnLocalhostNodePort() {
		log.Info("Checking if greenplum service has a valid internal port ...")
		jsonPath = "jsonpath={.spec.ports[0].nodePort}"
	} else {
		log.Info("Checking if greenplum service has a valid external ip ...")
		jsonPath = "jsonpath={.status.loadBalancer.ingress[0].ip}"
	}

	return errors.Wrapf(apiwait.PollImmediate(1*time.Second, getPollTimeout(), func() (bool, error) {
		cmd := exec.Command("kubectl", "get", "service", name, "-o", jsonPath)
		o, err := cmd.CombinedOutput()
		out := string(o)
		if err != nil {
			return false, nil
		}
		if len(out) > 0 {
			// look for port if minikube, look for IP if otherwise
			if kubeexecpsql.ServicesAreOnLocalhostNodePort() {
				port, err := strconv.Atoi(out)
				if err == nil && port > 0 {
					log.Info("got a valid port for service", "service", name, "port", out)
					return true, nil
				}
			} else {
				ip := net.ParseIP(out)
				if ip != nil {
					log.Info("got a valid ip for service", "service", name, "ip", out)
					return true, nil
				}
			}
		}
		return false, nil
	}), "waiting for service %s to have %s", name, jsonPath)
}

func ForClusterReady(waitForAutoInit bool) error {
	out, _ := exec.Command("kubectl", "get", "greenplumcluster", "my-greenplum", "-o", "jsonpath={.status.phase}").CombinedOutput()
	if string(out) == string(greenplumv1.GreenplumClusterPhaseRunning) {
		return nil
	}

	useMirrors, standby := getMirrorsAndStandby()
	err := ForReplicasReady("statefulset", "master")
	if err != nil {
		return err
	}

	err = ForReplicasReady("statefulset", "segment-a")
	if err != nil {
		return err
	}

	if useMirrors {
		err = ForReplicasReady("statefulset", "segment-b")
		if err != nil {
			return err
		}
	}

	if standby {
		// TODO: Do we need 2 way check?
		log.Info("Waiting for master-* to be network reachable ...")
		err = ForDNSRefresh("master-1", "master-0")
		if err != nil {
			return err
		}
		err = ForDNSRefresh("master-0", "master-1")
		if err != nil {
			return err
		}
	}

	primarySegmentCount, err := getDesiredReplicas("statefulset", "segment-a")
	if err != nil {
		return err
	}

	err = ForResource("configmap", "greenplum-config")
	if err != nil {
		return err
	}

	err = ForResource("secret", "ssh-secrets")
	if err != nil {
		return err
	}

	err = ForNetworkReady(primarySegmentCount, useMirrors)
	if err != nil {
		return err
	}

	err = ForGreenplumService("greenplum")
	if err != nil {
		return err
	}
	if waitForAutoInit {
		log.Info("Waiting for cluster to be initialized ... This could take a few minutes")
		err = ForGreenplumInitialization()
		if err != nil {
			out, _ := exec.Command("kubectl", "logs", "master-0", "--tail", "50").CombinedOutput()
			fmt.Println(string(out))
			log.Error(err, "initialization failed")
			return err
		}
	}

	return nil
}

func ForPsqlQuery(query, expectedResult string) error {
	var resultBytes []byte
	var err error
	return errors.Wrapf(apiwait.PollImmediate(1*time.Second, getPollTimeout(), func() (bool, error) {
		if resultBytes, err = kubeexecpsql.Query("master-0", query); err != nil {
			return false, nil
		}
		if strings.Contains(string(resultBytes), expectedResult) {
			return true, nil
		}
		return false, errors.New("mismatched output: " + string(resultBytes))
	}), "waiting for psql query to finish")
}

func ForPsqlQueryThroughService(query, expectedResult string) error {
	var resultBytes []byte
	var err error
	return errors.Wrapf(apiwait.PollImmediate(1*time.Second, getPollTimeout(), func() (bool, error) {
		if resultBytes, err = kubeexecpsql.ExecutePsqlQueryThroughService(query); err != nil {
			return false, nil
		}
		if strings.Contains(string(resultBytes), expectedResult) {
			return true, nil
		}
		return false, errors.New("mismatched output: " + string(resultBytes))
	}), "waiting for psql query to finish")
}

func getMirrorsAndStandby() (useMirrors, standby bool) {
	nameCmd := exec.Command("kubectl", "get", "greenplumCluster", "-o", "jsonpath={.items[0].metadata.name}")
	clusterName, err := nameCmd.CombinedOutput()
	if err != nil {
		log.Error(err, "getting name of GreenplumCluster")
	}
	cmd := exec.Command("kubectl", "get", "greenplumCluster", string(clusterName), "-o", "jsonpath={.spec.segments.mirrors},{.spec.masterAndStandby.standby}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err, "getting mirrors/standby of GreenplumCluster", "name", clusterName)
	}
	mirrorStandbyArr := strings.Split(string(out), ",")
	if mirrorStandbyArr[0] == "yes" {
		useMirrors = true
	}
	if mirrorStandbyArr[1] == "yes" {
		standby = true
	}
	return
}

func ForJobSuccess(jobName string) error {
	return errors.Wrapf(apiwait.PollImmediate(1*time.Second, 5*time.Minute, func() (bool, error) {
		cmd := exec.Command("kubectl", "get", "job", jobName, "-o", "jsonpath={.status.succeeded}")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return false, nil
		}
		if strings.TrimSpace(string(out)) != "1" {
			return false, err
		}
		return true, nil
	}), "waiting for job to succeed")
}

type Existence string

const (
	Exist    Existence = "exist"
	NotExist Existence = "not exist"
)

func ForServiceTo(svcName string, existence Existence) error {
	return errors.Wrapf(apiwait.PollImmediate(1*time.Second, 5*time.Minute, func() (bool, error) {
		cmd := exec.Command("kubectl", "get", "svc", svcName)
		err := cmd.Run()
		if err != nil {
			// it doesn't exist
			return existence == NotExist, nil
		}
		// it does exist
		return existence == Exist, nil
	}), "waiting for service to %s", existence)
}

func ForAPIService() error {
	return errors.Wrapf(apiwait.PollImmediate(1*time.Second, getPollTimeout(), func() (bool, error) {
		cmd := exec.Command("kubectl", "wait", "--for", "condition=available", "--timeout=0", "--all", "apiservice")
		err := cmd.Run()
		if err != nil {
			return false, nil
		}
		return true, nil
	}), "waiting for apiservice to be ready")
}
