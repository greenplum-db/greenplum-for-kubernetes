package scheme

import (
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	greenplumv1beta1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1beta1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/apis/apiserver"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var Scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(Scheme)
	_ = admissionregistrationv1.AddToScheme(Scheme)
	_ = apiextensions.AddToScheme(Scheme)
	_ = apiextensionsv1.AddToScheme(Scheme)
	_ = apiserver.AddToScheme(Scheme)
	_ = greenplumv1.AddToScheme(Scheme)
	_ = greenplumv1beta1.AddToScheme(Scheme)
}
