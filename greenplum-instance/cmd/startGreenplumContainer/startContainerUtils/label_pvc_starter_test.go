package startContainerUtils

import (
	"errors"
	"fmt"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers/greenplumcluster"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/reactive"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/heapvalue"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const postgres = "/usr/local/greenplum-db/bin/postgres"

var _ = Describe("LabelPVCStarter", func() {

	const (
		pvcName     = "my-greenplum-pgdata-master-0"
		pvcLabelKey = "greenplum-major-version"
	)

	var (
		pvc            corev1.PersistentVolumeClaim
		reactiveClient *reactive.Client
		subject        *LabelPvcStarter

		cmdFake  *commandable.CommandFake
		memoryfs vfs.Filesystem
		pod      corev1.Pod
	)

	BeforeEach(func() {
		cmdFake = commandable.NewFakeCommand()
		gpdbSemVer := fmt.Sprintf("%s.1.2", greenplumcluster.SupportedGreenplumMajorVersion)
		cmdFake.ExpectCommand(postgres, "--gp-version").PrintsOutput(gpdbVer(gpdbSemVer))

		memoryfs = memfs.Create()
		Expect(vfs.MkdirAll(memoryfs, "/var/run/secrets/kubernetes.io/serviceaccount", 0777)).To(Succeed())
		Expect(vfs.WriteFile(memoryfs, "/var/run/secrets/kubernetes.io/serviceaccount/namespace", []byte("test-ns"), 0666))

		reactiveClient = reactive.NewClient(fakeClient.NewFakeClientWithScheme(clientgoscheme.Scheme))

		pod = corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-ns",
				Name:      "master-0",
			},
			Spec: corev1.PodSpec{
				Volumes: []corev1.Volume{
					{
						Name: "ssh-key-volume",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName:  "ssh-secrets",
								DefaultMode: heapvalue.NewInt32(0444),
							},
						},
					},
					{
						Name: "my-greenplum-pgdata",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: pvcName,
							},
						},
					},
				},
			},
		}
	})

	JustBeforeEach(func() {
		Expect(reactiveClient.Create(nil, &pod)).To(Succeed())

		subject = &LabelPvcStarter{
			App: &starter.App{
				Command: cmdFake.Command,
				Fs:      memoryfs,
			},
			Hostname:  func() (string, error) { return "master-0", nil },
			NewClient: func() (client.Client, error) { return reactiveClient, nil },
		}
	})

	When("the PVC is labeled", func() {
		When("the PVC label matches the Pod GPDB major version", func() {
			BeforeEach(func() {
				pvc = corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-ns",
						Name:      pvcName,
						Labels:    map[string]string{pvcLabelKey: greenplumcluster.SupportedGreenplumMajorVersion},
					},
				}

				Expect(reactiveClient.Create(nil, &pvc)).To(Succeed())
			})
			It("does not update the label", func() {
				Expect(subject.Run()).To(Succeed())

				Expect(reactiveClient.Get(nil, types.NamespacedName{Namespace: "test-ns", Name: pvcName}, &pvc)).To(Succeed())
				Expect(pvc.Labels).To(Equal(map[string]string{pvcLabelKey: greenplumcluster.SupportedGreenplumMajorVersion}))
			})
		})
		When("the PVC label does not match the Pod GPDB major version", func() {
			BeforeEach(func() {
				pvc = corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-ns",
						Name:      pvcName,
						Labels:    map[string]string{pvcLabelKey: "7000"},
					},
				}

				Expect(reactiveClient.Create(nil, &pvc)).To(Succeed())
			})
			It("fails", func() {
				Expect(subject.Run()).To(MatchError("GPDB version on PVC does not match pod version. PVC greenplum-major-version=7000; Pod version: " + greenplumcluster.SupportedGreenplumMajorVersion))
			})
		})
	})

	When("the PVC is not labeled", func() {
		BeforeEach(func() {
			pvc = corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-ns",
					Name:      pvcName,
				},
			}

			Expect(reactiveClient.Create(nil, &pvc)).To(Succeed())
		})
		It("labels the PVC", func() {
			Expect(subject.Run()).To(Succeed())

			Expect(reactiveClient.Get(nil, types.NamespacedName{Namespace: "test-ns", Name: pvcName}, &pvc)).To(Succeed())
			Expect(pvc.Labels).To(Equal(map[string]string{pvcLabelKey: greenplumcluster.SupportedGreenplumMajorVersion}))
		})

		When("patching pvc labels fails", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("patch", "persistentvolumeclaims", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("patch fail")
				})
			})
			It("fails", func() {
				Expect(subject.Run()).To(MatchError("patching pvc label: patch fail"))
			})
		})
	})

	When("there is more than one persistent volume in the pod", func() {
		BeforeEach(func() {
			extraPvc := corev1.Volume{
				Name: "extra-pvc",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "extra-pvc",
					},
				},
			}
			pod.Spec.Volumes = append(pod.Spec.Volumes, extraPvc)
		})
		It("fails", func() {
			Expect(subject.Run()).To(MatchError("found more pvc volumes than expected"))
		})
	})

	When("the PVC does not exist", func() {
		It("fails", func() {
			Expect(subject.Run()).To(MatchError(`persistentvolumeclaims "my-greenplum-pgdata-master-0" not found`))
		})
	})

	When("getting namespace fails", func() {
		JustBeforeEach(func() {
			subject.App.Fs = memfs.Create()
		})
		It("fails", func() {
			Expect(subject.Run()).To(MatchError("open /var/run/secrets/kubernetes.io/serviceaccount/namespace: file does not exist"))
		})
	})

	When("getting the pod fails", func() {
		BeforeEach(func() {
			reactiveClient.PrependReactor("get", "pods", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.New("can't find myself")
			})
		})
		It("fails", func() {
			Expect(subject.Run()).To(MatchError("can't find myself"))
		})
	})

	When("hostname lookup fails", func() {
		JustBeforeEach(func() {
			subject.Hostname = func() (string, error) {
				return "", errors.New("get hostname failure")
			}
		})
		It("returns an error", func() {
			Expect(subject.Run()).To(MatchError("getting hostname: get hostname failure"))
		})
	})

	When("creating a client fails", func() {
		JustBeforeEach(func() {
			subject.NewClient = func() (client.Client, error) {
				return nil, errors.New("no client for you")
			}
		})
		It("returns an error", func() {
			Expect(subject.Run()).To(MatchError("unable to create client: no client for you"))
		})

	})

	When("getting greenplum version fails", func() {
		JustBeforeEach(func() {
			cmdFake.ExpectCommand(postgres, "--gp-version").PrintsOutput("not a version")
		})
		It("fails", func() {
			Expect(subject.Run()).To(MatchError("couldn't parse greenplum version in: not a version"))
		})
	})
})

var _ = Describe("GetGreenplumMajorVersion", func() {
	var (
		cmdFake *commandable.CommandFake
		subject *LabelPvcStarter
	)

	BeforeEach(func() {
		cmdFake = commandable.NewFakeCommand()

		subject = &LabelPvcStarter{
			App: &starter.App{
				Command: cmdFake.Command,
			},
		}
	})

	table.DescribeTable("parses the major version",
		func(majVersion, versionOutput string) {
			cmdFake.ExpectCommand(postgres, "--gp-version").PrintsOutput(versionOutput)

			version, err := subject.GetGreenplumMajorVersion()
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal(majVersion))

		},
		table.Entry("GPDB 6 release build", "6", gpdbVer("6.7.0")),
		table.Entry("GPDB 6 dev build", "6", gpdbVer("6.5.0+dev.37.g935509d994")),
		table.Entry("GPDB 6 alpha build", "6", gpdbVer("6.0.0-alpha.1")),
		table.Entry("GPDB 7 release build", "7", gpdbVer("7.7.0")),
		table.Entry("Unusual output", "six", gpdbVer("six.oh.oh")),
	)

	table.DescribeTable("major version can't be parsed",
		func(errString, versionOutput string) {
			cmdFake.ExpectCommand(postgres, "--gp-version").PrintsOutput(versionOutput)

			version, err := subject.GetGreenplumMajorVersion()
			Expect(err).To(MatchError(errString), version)

		},
		table.Entry("empty version",
			"couldn't parse greenplum version in: postgres (Greenplum Database)  build commit:deadc0de",
			gpdbVer("")),
		table.Entry("non-matching output",
			"couldn't parse greenplum version in: hi, I am not Postgres",
			"hi, I am not Postgres"),
	)

	When("postgres fails to run", func() {
		BeforeEach(func() {
			cmdFake.ExpectCommand(postgres, "--gp-version").ReturnsStatus(42).PrintsError("postgres: explode\n")
		})

		It("outputs the error", func() {
			_, err := subject.GetGreenplumMajorVersion()
			Expect(err).To(MatchError("trying to get greenplum version: exit status 42; stderr: postgres: explode"), err.Error())
		})
	})
})

func gpdbVer(ver string) string {
	return fmt.Sprintf("postgres (Greenplum Database) %s build commit:deadc0de\n", ver)
}
