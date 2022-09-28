module github.com/pivotal/greenplum-for-kubernetes

go 1.14

require (
	code.cloudfoundry.org/clock v1.0.0
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/blang/vfs v0.0.0-00010101000000-000000000000
	github.com/cppforlife/go-semi-semantic v0.0.0-20160921010311-576b6af77ae4
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/loads v0.19.5 // indirect
	github.com/go-openapi/runtime v0.19.11 // indirect
	github.com/go-openapi/validate v0.19.8
	github.com/gocarina/gocsv v0.0.0-20200302151839-87c60d755c58
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/greenplum-db/gp-common-go-libs v1.0.4
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/jessevdk/go-flags v1.4.0
	github.com/knqyf263/go-deb-version v0.0.0-20190517075300-09fca494f03d
	github.com/lib/pq v1.3.0
	github.com/minio/minio-go/v6 v6.0.49
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.7.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
	github.com/tedsuo/ifrit v0.0.0-20191009134036-9a97d0632f00 // indirect
	go.mongodb.org/mongo-driver v1.3.1 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	go.uber.org/zap v1.14.0
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/oauth2 v0.0.0-20191122200657-5d9234df094c // indirect
	golang.org/x/tools v0.0.0-20200304193943-95d2e580d8eb // indirect
	gopkg.in/ini.v1 v1.52.0 // indirect
	honnef.co/go/tools v0.0.1-2020.1.3 // indirect
	k8s.io/api v0.17.8
	k8s.io/apiextensions-apiserver v0.17.8
	k8s.io/apimachinery v0.17.8
	k8s.io/apiserver v0.17.8
	k8s.io/client-go v0.17.8
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200410163147-594e756bea31 // indirect
	k8s.io/kubectl v0.17.8
	sigs.k8s.io/controller-runtime v0.5.7
	sigs.k8s.io/kustomize/api v0.3.2
	sigs.k8s.io/yaml v1.2.0
)

// Fork of blang/vfs:

replace github.com/blang/vfs => github.com/dsharp-pivotal/vfs v0.0.0-20180917171731-4c8d59de28be

replace vbom.ml/util => github.com/fvbommel/util v0.0.0-20160121211510-db5cfe13f5cc
