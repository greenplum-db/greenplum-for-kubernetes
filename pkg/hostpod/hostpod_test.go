package hostpod_test

import (
	"errors"
	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/greenplum-db/gp-common-go-libs/structmatcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/reactive"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/hostpod"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("GetCurrentNamespace", func() {
	var (
		memoryFS vfs.Filesystem
	)
	BeforeEach(func() {
		memoryFS = memfs.Create()
	})
	It("returns the current namespace of the pod from the file 'namespace' in local FS", func() {
		const nsFilename = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
		Expect(vfs.MkdirAll(memoryFS, filepath.Dir(nsFilename), 0755)).To(Succeed())
		Expect(vfs.WriteFile(memoryFS, nsFilename, []byte("test"), 0644)).To(Succeed())
		currentNS, err := hostpod.GetCurrentNamespace(memoryFS)
		Expect(err).NotTo(HaveOccurred())
		Expect(currentNS).To(Equal("test"))
	})

	It("returns error if the file 'namespace' is not present", func() {
		currentNS, err := hostpod.GetCurrentNamespace(memoryFS)
		Expect(currentNS).To(BeEmpty())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("open /var/run/secrets/kubernetes.io/serviceaccount/namespace: file does not exist"))
	})
})

var _ = Describe("GetThisPod", func() {

	var (
		reactiveClient *reactive.Client
		operatorPod    *corev1.Pod
		testNS         string
	)
	fakehostname := func() (string, error) {
		return operatorPod.Name, nil
	}

	BeforeEach(func() {
		testNS = "testNamespace"

		operatorPod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "greenplum-operator-hash123-hash456",
				Namespace: testNS,
				Labels: map[string]string{
					"app": "greenplum-operator",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "greenplum-operator",
						Image: "greenplum-operator:latest",
					},
				},
			},
		}

		// 2nd pod with a different name
		operatorPod2 := operatorPod.DeepCopy()
		operatorPod2.Name = "greenplum-operator2"
		reactiveClient = reactive.NewClient(fake.NewFakeClientWithScheme(scheme.Scheme, operatorPod2))
	})

	When("there is a pod named with our hostname", func() {
		BeforeEach(func() {
			Expect(reactiveClient.Create(nil, operatorPod.DeepCopy())).To(Succeed())
		})
		It("returns the operator pod", func() {
			resultPod, err := hostpod.GetThisPod(nil, reactiveClient, testNS, fakehostname)
			Expect(err).NotTo(HaveOccurred())
			Expect(resultPod).To(structmatcher.MatchStruct(operatorPod).ExcludingFields("ObjectMeta.ResourceVersion"))
		})
	})

	When("there are no matching pods", func() {
		BeforeEach(func() {
			reactiveClient = reactive.NewClient(fake.NewFakeClientWithScheme(scheme.Scheme))
		})
		It("returns an error", func() {
			_, err := hostpod.GetThisPod(nil, reactiveClient, testNS, fakehostname)
			Expect(err).To(MatchError(`pods "greenplum-operator-hash123-hash456" not found`))
		})
	})

	When("there is an error fetching the pod", func() {
		BeforeEach(func() {
			reactiveClient.PrependReactor("get", "pods", func(testing.Action) (bool, runtime.Object, error) {
				return true, nil, errors.New("intentional failure to get pods")
			})
		})

		It("returns an error", func() {
			_, err := hostpod.GetThisPod(nil, reactiveClient, testNS, fakehostname)
			Expect(err).To(MatchError("intentional failure to get pods"))
		})
	})

	When("hostname returns an error", func() {
		It("fails", func() {
			_, err := hostpod.GetThisPod(nil, reactiveClient, testNS, func() (string, error) {
				return "", errors.New("failed to get hostname")
			})
			Expect(err).To(MatchError("getting hostname: failed to get hostname"))
		})
	})
})
