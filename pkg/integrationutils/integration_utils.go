package integrationutils

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/blang/vfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubeexecpsql"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubewait"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/yaml"
)

var (
	// to get the base directory path of this file
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)

	GreenplumImageRepository = flag.String(
		"greenplumImageRepository",
		"greenplum-for-kubernetes",
		"Sets the greenplum image repository")

	GreenplumImageTag = flag.String(
		"greenplumImageTag",
		"latest",
		"Sets the greenplum image tag")

	OperatorImageRepository = flag.String(
		"operatorImageRepository",
		"greenplum-operator",
		"Sets the operator image repo")

	OperatorImageTag = flag.String(
		"operatorImageTag",
		"latest",
		"Sets the operator image tag")
)

type SetupOperatorOptions struct {
	OperatorImageRepository  string
	OperatorImageTag         string
	GreenplumImageRepository string
	GreenplumImageTag        string
	OperatorPath             string
	Upgrade                  bool
}

func (o SetupOperatorOptions) WithUpgrade() SetupOperatorOptions {
	o.Upgrade = true
	return o
}

func CurrentOperatorOptions() (s SetupOperatorOptions) {
	s.OperatorImageRepository = *OperatorImageRepository
	s.OperatorImageTag = *OperatorImageTag
	s.GreenplumImageRepository = *GreenplumImageRepository
	s.GreenplumImageTag = *GreenplumImageTag
	s.OperatorPath = ReleasePath("operator", "greenplum-operator/operator")
	return
}

func OldOperatorOptions() (s SetupOperatorOptions) {
	s.OperatorImageRepository = *OperatorImageRepository
	s.OperatorImageTag = OldOperatorVersion
	s.GreenplumImageRepository = *GreenplumImageRepository
	s.GreenplumImageTag = OldOperatorVersion
	s.OperatorPath = *oldReleaseDir + "/operator"
	return
}

func EnsureRegSecretIsCreated() {
	cmd := exec.Command("kubectl", "get", "secret", "regsecret")
	if cmd.Run() != nil {
		var keyJSON string
		byteArray, err := vfs.ReadFile(vfs.OS(), basepath+"/../../greenplum-operator/operator/key.json")
		if err != nil {
			log.Info("Using GCP_SVC_ACCT_KEY to create regsecret")
			keyJSON = os.Getenv("GCP_SVC_ACCT_KEY")
		} else {
			log.Info("Using key.json to create regsecret")
			keyJSON = string(byteArray)
		}
		Expect(keyJSON).NotTo(BeEmpty(), "should be able to use key.json file or GCP_SVC_ACCT_KEY")
		cmd := exec.Command("kubectl", "create", "secret", "docker-registry", "regsecret",
			"--docker-server=https://gcr.io",
			"--docker-username=_json_key",
			"--docker-password="+keyJSON)
		cmd.Run()
	}
}

func SetupOperator(options SetupOperatorOptions) {
	var helmArgs []string
	if options.Upgrade {
		helmArgs = append(helmArgs, "upgrade")
	} else {
		helmArgs = append(helmArgs, "install")
	}
	helmArgs = append(helmArgs, "--wait", "greenplum-operator", "--timeout", "3m20s")
	helmArgs = append(helmArgs, options.OperatorPath)
	helmArgs = append(helmArgs, "--set", "operatorImageRepository="+options.OperatorImageRepository)
	helmArgs = append(helmArgs, "--set", "operatorImageTag="+options.OperatorImageTag)
	helmArgs = append(helmArgs, "--set", "greenplumImageRepository="+options.GreenplumImageRepository)
	helmArgs = append(helmArgs, "--set", "greenplumImageTag="+options.GreenplumImageTag)
	helmArgs = append(helmArgs, "--set", "logLevel=debug")
	operatorImage := options.OperatorImageRepository + ":" + options.OperatorImageTag
	log.Info("BEGIN setup Greenplum operator", "image", operatorImage)
	cmd := exec.Command("helm", helmArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Info("problem running helm", "cmd", cmd.Args[1], "output", string(output))
		log.Info("describe controller: ")
		cmd = exec.Command("kubectl", "describe", "pod", "greenplum-operator-")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		log.Info("logs from controller: ")
		cmd = exec.Command("kubectl", "logs", "-l", "name=greenplum-operator")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		log.Info("operator installation failed:")
		fmt.Println(string(output))
	}
	Expect(err).NotTo(HaveOccurred())
	log.Info("created Greenplum operator", "image", operatorImage)
	Expect(kubewait.ForPodState("greenplum-operator", string(greenplumv1.GreenplumClusterPhaseRunning))).To(Succeed())
	Expect(kubewait.ForResource("crd", "greenplumclusters.greenplum.pivotal.io")).To(Succeed())
	Expect(kubewait.ForAPIResource("greenplum.pivotal.io", "greenplumclusters")).To(Succeed())
	log.Info("END setup Greenplum operator", "image", operatorImage)
}

func CleanUpK8s(operatorImageTag string) {
	log.Info("BEGIN cleaning up Kubernetes cluster ...")
	// Ignore errors from these commands.
	KubeDelete("--all", "greenplumclusters")
	log.Info("     deleting all charts...")
	DeleteAllCharts()

	// clean up the GPDB deployment
	log.Info("     deleting all resources...")
	KubeDelete(
		"configmaps/greenplum-config",
		"statefulsets/master", "statefulset/segment-a", "statefulset/segment-b",
		"secrets/ssh-secrets", "secrets/regsecret",
		"service/agent", "service/greenplum")
	for _, pod := range []string{
		"greenplum-operator", "master", "segment-a", "segment-b",
	} {
		Expect(kubewait.ForPodDestroyed(pod)).To(Succeed())
	}

	// clean up all the PVCs, PVs and events
	log.Info("     deleting all PVCs...")
	KubeDelete("--all", "pvc")
	log.Info("     deleting all events...")
	KubeDelete("--all", "events")

	// Clean up operator resources
	log.Info("     deleting all operator resources...")
	KubeDelete(
		"crd/greenplumclusters.greenplum.pivotal.io",
		"deployment.apps/greenplum-operator",
		"service/greenplum-validating-webhook",
		"validatingwebhookconfiguration/greenplum-validating-webhook",
		"certificatesigningrequest/greenplum-validating-webhook",
	)
	log.Info("END cleaning up Kubernetes cluster ...")
}

func VerifyPVCsAreDeleted() {
	cmd := exec.Command("kubectl", "get", "pvc", "-l", "greenplum-cluster=my-greenplum")
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred())
	Expect(string(out)).To(Equal("No resources found in default namespace.\n"))
}

type GpYamlOptions struct {
	APIVersion     string
	MasterStorage  string
	SegmentStorage string
	StorageClass   string
	Standby        string
	Mirrors        string

	PrimarySegmentCount int
	DefaultStorageClass string
	IsSingleNode        bool

	PXFService    string
}

var greenplumClusterManifestTemplate = template.Must(
	template.New("GreenplumCluster manifest template").
		Funcs(sprig.TxtFuncMap()).
		Parse(`---
apiVersion: "greenplum.pivotal.io/{{.APIVersion}}"
kind: "GreenplumCluster"
metadata:
  name: my-greenplum
spec:
  masterAndStandby:
    hostBasedAuthentication: ""
{{- if not .IsSingleNode }}
    memory: "800Mi"
    cpu: "0.5"
{{- end }}
    storageClassName: {{ .StorageClass | default .DefaultStorageClass | quote }}
    storage: {{ .MasterStorage | default "1G" | quote }}
{{- if and (not .IsSingleNode) (eq .Mirrors "yes") (eq .Standby "yes") }}
    antiAffinity: "yes"
{{- end}}
{{- if .Standby }}
    standby: {{ .Standby | quote }}
{{- end }}
  segments:
    primarySegmentCount: {{ .PrimarySegmentCount }}
{{- if not .IsSingleNode }}
    memory: "800Mi"
    cpu: "0.5"
{{- end }}
    storageClassName: {{ .StorageClass | default .DefaultStorageClass | quote }}
    storage: {{ .SegmentStorage | default "2G" | quote }}
{{- if and (not .IsSingleNode) (eq .Mirrors "yes") (eq .Standby "yes") }}
    antiAffinity: "yes"
{{- end }}
{{- if .Mirrors }}
    mirrors: {{ .Mirrors | quote }}
{{- end }}
{{- if .PXFService }}
  pxf:
    serviceName: {{ .PXFService | quote }}
{{- end }}
`))

func GetGreenplumManifestYaml(primarySegmentCount int, options GpYamlOptions) string {
	var m strings.Builder
	// apiVersion
	if options.APIVersion == "" {
		options.APIVersion = "v1"
	}
	options.PrimarySegmentCount = primarySegmentCount

	options.DefaultStorageClass = "standard"
	if sc := GetDefaultStorageClass(); sc != nil {
		options.DefaultStorageClass = sc.Name
	}

	options.IsSingleNode = IsSingleNode()

	Expect(greenplumClusterManifestTemplate.Execute(&m, options)).To(Succeed())
	return m.String()
}

func CreateTempFile(contents string) (string, string) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "gpdb-integration")
	Expect(err).NotTo(HaveOccurred())
	tempFile := path.Join(tempDir, "gpdb-integration-temp.yaml")
	Expect(ioutil.WriteFile(tempFile, []byte(contents), 0644)).To(Succeed())
	return tempDir, tempFile
}

func SetupGreenplumCluster(gpdbYamlFile, version string) {
	log.Info("BEGIN setup Greenplum cluster", "version", version)
	out, err := ApplyManifest(gpdbYamlFile, "greenplumcluster")
	if err != nil {
		log.Info("Output from kubectl apply -f")
		fmt.Println(out)
		log.Info("problem creating greenplum cluster instance", "output", out)
		log.Info("describe greenplum cluster: ")
		cmd := exec.Command("kubectl", "describe", "greenplumcluster", "my-greenplum")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		log.Info("logs from master: ")
		cmd = exec.Command("kubectl", "logs", "master-0")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(out).To(ContainSubstring("my-greenplum"))
	log.Info("END setup Greenplum cluster", "version", version)
}

func CleanupGreenplumCluster(gpdbYamlFile string) {
	log.Info("BEGIN clean up Greenplum cluster ...")
	DeleteGreenplumCluster()
	Expect(KubeDelete("--all", "pvc")).To(Succeed())
	VerifyPVCsAreDeleted()
	log.Info("END clean up Greenplum cluster")
}

func DeleteGreenplumCluster() {
	cmd := exec.Command("kubectl", "delete", "greenplumcluster/my-greenplum")
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred())
	Expect(string(out)).To(ContainSubstring("my-greenplum"))
	Expect(string(out)).To(ContainSubstring("deleted"))

	Expect(kubewait.ForPodDestroyed("master-0")).To(Succeed())
	Expect(kubewait.ForPodDestroyed("segment-a-0")).To(Succeed())
	Expect(kubewait.ForPodDestroyed("segment-b-0")).To(Succeed())
}

func CleanupComponentService(tempdir string, yamlFileName string, serviceName string) {
	log.Info("Deleting Component", "serviceName", serviceName)
	cmd := exec.Command("kubectl", "delete", "-f", yamlFileName)
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred())
	Expect(string(out)).To(ContainSubstring(fmt.Sprintf(`%q deleted`, serviceName)))
	Expect(os.RemoveAll(tempdir)).To(Succeed())

	Expect(kubewait.ForPodDestroyed(serviceName)).To(Succeed())
}

func DeleteGreenplumClusterIfExists() {
	out, err := exec.Command("kubectl", "get", "greenplumclusters", "my-greenplum").CombinedOutput()
	if err == nil || !strings.Contains(string(out), "not found") {
		DeleteGreenplumCluster()
	}
}

func KubeDelete(resources ...string) error {
	cmd := exec.Command("kubectl", "delete", "--wait", "--grace-period=0", "--force")
	cmd.Args = append(cmd.Args, resources...)
	return cmd.Run()
}

func KubeExec(hostName, command string) ([]byte, error) {
	return exec.Command("kubectl", "exec", "-i", hostName, "--", "bash", "-c", command).CombinedOutput()
}

func ApplyManifest(manifestYamlFileName string, manifestType string) (string, error) {
	log.Info("Applying Manifest", "type", manifestType)
	cmd := exec.Command("kubectl", "apply", "-f", manifestYamlFileName)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func CheckCleanClusterStartup() {
	// ensure cluster was shut down properly before startup
	out, err := exec.Command("kubectl", "logs", "master-0").CombinedOutput()
	Expect(err).NotTo(HaveOccurred())
	Expect(string(out)).NotTo(ContainSubstring("[WARNING]:-postmaster.pid file exists on Master, checking if recovery startup required"))
}

func VerifyDataThroughService() {
	log.Info("verify if data is intact")
	Expect(kubewait.ForPsqlQueryThroughService("select * from foo", "1")).To(Succeed())
}

func VerifyData() {
	log.Info("verify if data is intact")
	Expect(kubewait.ForPsqlQuery("select * from foo", "1")).To(Succeed())
}
func LoadData() {
	log.Info("load some data ...")
	out, err := QueryWithRetry("master-0", "CREATE TABLE foo(a int);")
	if err != nil {
		fmt.Println(string(out))
		Fail("create table failed")
	}
	out, err = QueryWithRetry("master-0", "INSERT INTO foo VALUES (1)")
	if err != nil {
		fmt.Println(string(out))
		Fail("insert row failed")
	}
	out, err = QueryWithRetry("master-0", "SELECT * FROM foo")
	if err != nil {
		fmt.Println(string(out))
		Fail("select query failed")
	}
	Expect(strings.TrimSpace(string(out))).To(Equal("1"))
}

func VerifyGreenplumForKubernetesVersion(pod string, version string) {
	out, err := exec.Command("kubectl", "get", "pod", pod, "-o", "jsonpath={.spec.containers[0].image}").CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "should be able to get container image for master-0: %s", string(out))
	Expect(string(out)).To(ContainSubstring(version))
}

func VerifyStatusInstanceImage(version string) {
	out, err := exec.Command("kubectl", "get", "greenplumCluster", "-o", "jsonpath={.items[0].status.instanceImage}").CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "should be able to get container image for gpInstance: %s", string(out))
	imageName := fmt.Sprintf("%s:%s", *GreenplumImageRepository, version)
	Expect(string(out)).To(Equal(imageName))
}

func VerifyStatusOperatorVersion(version string) {
	out, err := exec.Command("kubectl", "get", "greenplumCluster", "-o", "jsonpath={.items[0].status.operatorVersion}").CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "should be able to get status.operatorversion for gpInstance: %s", string(out))
	imageName := fmt.Sprintf("%s:%s", *OperatorImageRepository, version)
	Expect(string(out)).To(Equal(imageName))
}

func AddHbaToAllowAccessToGreenplumThroughService(masterPodName string) {
	additionalHBAEntry, err := GetAdditionalHBAEntryForService()
	Expect(err).NotTo(HaveOccurred())
	if additionalHBAEntry != "" {
		addHbaEntryToMasterCommand := fmt.Sprintf(`echo "%s" >> /greenplum/data-1/pg_hba.conf`, additionalHBAEntry)
		out, err := KubeExec(masterPodName, addHbaEntryToMasterCommand)
		if err != nil {
			log.Info("writing entry to pg_hba.conf failed")
			fmt.Println(string(out))
		}
		Expect(err).NotTo(HaveOccurred())
		pgHbaFile, err := KubeExec(masterPodName, "cat /greenplum/data-1/pg_hba.conf")
		Expect(pgHbaFile).To(ContainSubstring(additionalHBAEntry))
		Expect(err).NotTo(HaveOccurred())
		out, err = KubeExec(masterPodName, "source /usr/local/greenplum-db/greenplum_path.sh && gpstop -u")
		if err != nil {
			log.Info("restarting gpdb failed")
			fmt.Println(string(out))
		}
		Expect(err).NotTo(HaveOccurred())
	}
}

func GetAdditionalHBAEntryForService() (result string, err error) {
	err = wait.PollImmediate(1*time.Second, 20*time.Second, func() (done bool, err error) {
		out, err := ExecutePsqlQueryThroughService(`select 'connected'`)
		if err != nil {
			if _, ok := err.(*exec.ExitError); ok {
				reg := regexp.MustCompile(`psql: FATAL:  no pg_hba.conf entry for host "(\d+.\d+.\d+.\d+)"`)
				matches := reg.FindSubmatch(out)
				if len(matches) == 2 {
					ip := matches[1]

					result = fmt.Sprintf("host  all  gpadmin  %s/32  trust", ip)
					return true, nil
				}
				if strings.HasPrefix(string(out), `psql: could not connect to server: Connection timed out`) {
					return false, nil
				}
			}
		}
		if strings.Contains(string(out), "connected") {
			result = ""
			return true, nil
		}
		return true, fmt.Errorf("psql query had unexpected output: %s, %w", out, err)
	})

	return
}

func GetDefaultStorageClass() *storagev1.StorageClass {
	out, err := exec.Command("kubectl", "get", "storageclasses", "-o", "json").CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "should be able to query for storageclasses", string(out))
	var scList storagev1.StorageClassList
	Expect(json.Unmarshal(out, &scList)).To(Succeed())

	for _, sc := range scList.Items {
		if val, ok := sc.ObjectMeta.Annotations["storageclass.kubernetes.io/is-default-class"]; ok {
			if val == "true" {
				return &sc
			}
		}
	}
	return nil
}

func LabelNodes() {
	// get all nodes
	nodeNames := GetNodeNames()

	// label nodes
	for i := 0; i < 2; i++ {
		cmd := exec.Command("kubectl", "label", "nodes", nodeNames[i], "worker=my-gp-masters")
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "should be able to label nodes with my-gp-masters: %s", string(out))
	}

	for i := 2; i < len(nodeNames); i++ {
		cmd := exec.Command("kubectl", "label", "nodes", nodeNames[i], "worker=my-gp-segments")
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "should be able to label nodes with my-gp-segments: %s", string(out))
	}
}

func AnnotateObject(objectIdentifier, annotation string) {
	cmd := exec.Command("kubectl", "annotate", objectIdentifier, annotation)
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "should be able to annotate %s with %s: %s", objectIdentifier, annotation, string(out))
}

func GetNodeNames() []string {
	cmd := exec.Command("kubectl", "get", "nodes", "-o", "yaml")
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "should be able to list nodes: %s", string(out))
	var nodes map[string]interface{}
	err = yaml.Unmarshal(out, &nodes)
	if err != nil {
		log.Error(err, "failed to parse node list yaml")
	}
	nodeList := nodes["items"].([]interface{})
	Expect(len(nodeList) >= 4).To(BeTrue())

	// get node names
	nodeNames := make([]string, 0)
	for i := 0; i < len(nodeList); i++ {
		node := nodeList[i].(map[string]interface{})
		nodeNames = append(nodeNames, node["metadata"].(map[string]interface{})["name"].(string))
	}

	return nodeNames
}

func GetPodsOnNodes() map[string][]string {
	podsOnNodes := make(map[string][]string)

	nodeNames := GetNodeNames()
	for _, node := range nodeNames {
		podsOnNodes[node] = make([]string, 0)
	}

	cmd := exec.Command("kubectl", "get", "pods", "-o", "yaml")
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "should be able to list pods: %s", string(out))
	var nodes map[string]interface{}
	err = yaml.Unmarshal(out, &nodes)
	if err != nil {
		log.Error(err, "failed to parse pod list yaml")
	}
	podList := nodes["items"].([]interface{})

	for i := 0; i < len(podList); i++ {
		pod := podList[i].(map[string]interface{})
		podName := pod["metadata"].(map[string]interface{})["name"].(string)
		nodeName := pod["spec"].(map[string]interface{})["nodeName"].(string)
		podsOnNodes[nodeName] = append(podsOnNodes[nodeName], podName)
	}

	return podsOnNodes
}

func GetVWHSvcName() string {
	cmd := exec.Command("kubectl", "get", "svc", "-o", "jsonpath={.items[*].metadata.name}", "-l", "app=greenplum-operator")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Info("cannot fetch validating webhook svc name")
	}
	return string(out)
}
