package integrationutils

import (
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/minio/minio-go/v6"
	. "github.com/onsi/gomega"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubeexecpsql"
)

const (
	pxfTestBucketName = "pxf-test-data"
	pxfConfFolder     = "integration-test-conf"
	gsSitePath        = "integration-test-conf/servers/gs/gs-site.xml"
	pxfDataPath       = "1/lineitem.tbl"
)

func generateGSSiteContents(accessKeyID, secretAccessKey string) string {
	return fmt.Sprintf(`
<configuration>
	<property>
		<name>fs.s3a.endpoint</name>
		<value>https://storage.googleapis.com</value>
	</property>
	<property>
		<name>fs.s3a.access.key</name>
		<value>%s</value>
	</property>
	<property>
		<name>fs.s3a.secret.key</name>
		<value>%s</value>
	</property>
</configuration>
`, accessKeyID, secretAccessKey)
}

func EnsurePXFTestFilesArePresent(minioClient *minio.Client, accessKeyID, secretAccessKey string) {
	found, err := minioClient.BucketExists(pxfTestBucketName)
	Expect(err).NotTo(HaveOccurred())
	if !found {
		Expect(minioClient.MakeBucket(pxfTestBucketName, "us-central1")).To(Succeed())
	}

	// data file
	_, err = minioClient.StatObject(pxfTestBucketName, pxfDataPath, minio.StatObjectOptions{})
	Expect(err).NotTo(HaveOccurred())

	// gs-site.xml file
	gsSiteFile := generateGSSiteContents(accessKeyID, secretAccessKey)
	_, err = minioClient.PutObject(pxfTestBucketName, gsSitePath, strings.NewReader(gsSiteFile), -1, minio.PutObjectOptions{})
	Expect(err).NotTo(HaveOccurred())
	log.Info("S3 upload successful", "bucket", pxfTestBucketName, "path", gsSitePath)
}

type PXFYamlOptions struct {
	UseConfBucket   bool
	AccessKeyID     string
	SecretAccessKey string

	// Note: The following options are not intended to be set manually
	NumReplicas           int
	ConfBucketName        string
	ConfBucketFolder      string
	Base64AccessKeyID     string
	Base64SecretAccessKey string
}

var pxfManifestTemplate = template.Must(
	template.New("GreenplumCluster manifest template").
		Funcs(sprig.TxtFuncMap()).
		Parse(`---
apiVersion: "greenplum.pivotal.io/v1beta1"
kind: GreenplumPXFService
metadata:
  name: my-greenplum-pxf
spec:
  replicas: {{ .NumReplicas }}
  cpu: "0.5"
  memory: "1Gi"
  workerSelector: {}
{{- if .UseConfBucket }}
  pxfConf:
    s3Source:
      secret: "my-greenplum-pxf-configs"
      endpoint: "storage.googleapis.com"
      bucket: {{ .ConfBucketName }}
      folder: {{ .ConfBucketFolder }}
---
apiVersion: v1
kind: Secret
metadata:
  name: "my-greenplum-pxf-configs"
type: Opaque
data:
  access_key_id: {{ .Base64AccessKeyID }}
  secret_access_key: {{ .Base64SecretAccessKey }}
{{- end }}
`))

func GetPXFManifestYaml(numReplicas int, options PXFYamlOptions) string {
	var m strings.Builder
	options.NumReplicas = numReplicas
	options.ConfBucketName = pxfTestBucketName
	options.ConfBucketFolder = pxfConfFolder
	options.Base64AccessKeyID = base64.StdEncoding.EncodeToString([]byte(options.AccessKeyID))
	options.Base64SecretAccessKey = base64.StdEncoding.EncodeToString([]byte(options.SecretAccessKey))
	Expect(pxfManifestTemplate.Execute(&m, options)).To(Succeed())
	return m.String()
}

func CreatePXFExternalTable() {
	DropPXFExternalTable("lineitem_s3_1")
	createTableQuery := `CREATE EXTERNAL TABLE lineitem_s3_1 (
   l_orderkey    BIGINT,
   l_partkey     BIGINT,
   l_suppkey     BIGINT,
   l_linenumber  BIGINT,
   l_quantity    DECIMAL(15,2),
   l_extendedprice  DECIMAL(15,2),
   l_discount    DECIMAL(15,2),
   l_tax         DECIMAL(15,2),
   l_returnflag  CHAR(1),
   l_linestatus  CHAR(1),
   l_shipdate    DATE,
   l_commitdate  DATE,
   l_receiptdate DATE,
   l_shipinstruct CHAR(25),
   l_shipmode     CHAR(10),
   l_comment VARCHAR(44)
)
LOCATION ('pxf://pxf-test-data/1/?PROFILE=s3:text&SERVER=gs')
FORMAT 'CSV' (DELIMITER '|');`
	out, err := Query("master-0", createTableQuery)
	Expect(err).NotTo(HaveOccurred(), string(out))
}

func DropPXFExternalTable(table string) {
	dropTableQuery := fmt.Sprintf(`DROP EXTERNAL TABLE IF EXISTS %s;`, table)
	out, err := Query("master-0", dropTableQuery)
	Expect(err).NotTo(HaveOccurred(), string(out))
}

func CreatePXFExternalTableNoS3() {
	var pxfCreateTableCmd strings.Builder
	DropPXFExternalTable("pxf_read_test")
	pxfCreateTableCmd.WriteString(`CREATE EXTERNAL TABLE pxf_read_test (a TEXT, b TEXT, c TEXT)`)
	pxfCreateTableCmd.WriteString(` LOCATION ('\''pxf://tmp/dummy1?`)
	pxfCreateTableCmd.WriteString(`FRAGMENTER=org.greenplum.pxf.api.examples.DemoFragmenter`)
	pxfCreateTableCmd.WriteString(`&ACCESSOR=org.greenplum.pxf.api.examples.DemoAccessor`)
	pxfCreateTableCmd.WriteString(`&RESOLVER=org.greenplum.pxf.api.examples.DemoTextResolver'\'')`)
	pxfCreateTableCmd.WriteString(` FORMAT '\''TEXT'\'' (DELIMITER '\'','\'');`)
	pxfCreateTableResult, err := QueryWithRetry("master-0", pxfCreateTableCmd.String())
	Expect(err).NotTo(HaveOccurred(), string(pxfCreateTableResult))
}
