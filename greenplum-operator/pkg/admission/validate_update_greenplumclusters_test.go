package admission_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/admission"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor/fake"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/gpexpandjob"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/reactive"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/gplog/testing"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("validateUpdateGreenplumCluster", func() {

	var (
		subject admission.Handler
		logBuf  *gbytes.Buffer
	)
	BeforeEach(func() {
		reactiveClient := reactive.NewClient(fakeClient.NewFakeClientWithScheme(scheme.Scheme))
		subject = admission.Handler{
			KubeClient:     reactiveClient,
			PodCmdExecutor: &fake.PodExec{StdoutResult: "0\n"},
		}
		logBuf = gbytes.NewBuffer()
		admission.Log = gplog.ForTest(logBuf)
	})

	ContainDisallowedEntry := func(expectedMessage string) types.GomegaMatcher {
		return ContainLogEntry(Keys{
			"msg":       Equal("/validate"),
			"GVK":       Equal("greenplum.pivotal.io/v1, Kind=GreenplumCluster"),
			"Name":      Equal("my-gp-instance"),
			"Namespace": Equal("test-ns"),
			"UID":       Equal("my-gp-instance-uid"),
			"Operation": Equal("UPDATE"),
			"Allowed":   BeFalse(),
			"Message":   Equal(expectedMessage),
		})
	}
	ContainAllowedEntry := func() types.GomegaMatcher {
		return ContainLogEntry(Keys{
			"msg":       Equal("/validate"),
			"GVK":       Equal("greenplum.pivotal.io/v1, Kind=GreenplumCluster"),
			"Name":      Equal("my-gp-instance"),
			"Namespace": Equal("test-ns"),
			"UID":       Equal("my-gp-instance-uid"),
			"Operation": Equal("UPDATE"),
			"Allowed":   BeTrue(),
		})
	}

	When("increasing primarySegmentCount", func() {
		var (
			job            batchv1.Job
			reactiveClient *reactive.Client
			oldGreenplum   *greenplumv1.GreenplumCluster
			newGreenplum   *greenplumv1.GreenplumCluster
		)
		BeforeEach(func() {
			reactiveClient = reactive.NewClient(fakeClient.NewFakeClientWithScheme(scheme.Scheme))
			subject.KubeClient = reactiveClient

			oldGreenplum = exampleGreenplum.DeepCopy()
			newGreenplum = oldGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.PrimarySegmentCount++
		})

		Allowed := func() {
			outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

			Expect(outputReview.Response.Allowed).To(BeTrue())
			Expect(DecodeLogs(logBuf)).To(ContainAllowedEntry())
		}

		Disallowed := func(expectedMessage string) func() {
			return func() {
				outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

				Expect(outputReview.Response.Allowed).To(BeFalse())
				Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Message": Equal(expectedMessage),
				})))
				Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry(expectedMessage))
			}
		}

		When("the cluster has not been expanded before", func() {
			It("allows requests that change only PrimarySegmentCount", Allowed)
		})
		When("the cluster is not Running", func() {
			BeforeEach(func() {
				oldGreenplum.Status.Phase = greenplumv1.GreenplumClusterPhasePending
			})
			It("does not allow requests to increase primarySegmentCount",
				Disallowed("updates only supported when cluster is Running"))
		})
		When("there is a gpexpand job with status Completed", func() {
			BeforeEach(func() {
				job = gpexpandjob.GenerateJob("blah", "some-hostname", 2)
				job.Status = batchv1.JobStatus{
					Active:    0,
					Succeeded: 1,
					Failed:    0,
				}
				job.Namespace = "test-ns"
				job.Name = "my-gp-instance-gpexpand-job"
				Expect(reactiveClient.Create(nil, &job)).To(Succeed())
			})
			It("allows requests to increase primarySegmentCount", Allowed)
		})
		When("there is a gpexpand job with status Failed", func() {
			BeforeEach(func() {
				job = gpexpandjob.GenerateJob("blah", "some-hostname", 2)
				job.Status = batchv1.JobStatus{
					Active:    0,
					Succeeded: 0,
					Failed:    1,
				}
				job.Namespace = "test-ns"
				job.Name = "my-gp-instance-gpexpand-job"
				Expect(reactiveClient.Create(nil, &job)).To(Succeed())
			})
			It("does not allow requests to increase primarySegmentCount",
				Disallowed("cannot expand cluster because previous gpexpand job failed"))
		})
		When("there is a gpexpand job that is still running", func() {
			BeforeEach(func() {
				job = gpexpandjob.GenerateJob("blah", "some-hostname", 2)
				job.Status = batchv1.JobStatus{
					Active:    1,
					Succeeded: 0,
					Failed:    0,
				}
				job.Namespace = "test-ns"
				job.Name = "my-gp-instance-gpexpand-job"
				Expect(reactiveClient.Create(nil, &job)).To(Succeed())
			})
			It("does not allow requests to increase primarySegmentCount",
				Disallowed("cannot expand cluster because a gpexpand job is currently running"))
		})
		When("there's a job with uninitialized status", func() {
			BeforeEach(func() {
				job = gpexpandjob.GenerateJob("blah", "some-hostname", 2)
				job.Status = batchv1.JobStatus{
					Active:    0,
					Succeeded: 0,
					Failed:    0,
				}
				job.Namespace = "test-ns"
				job.Name = "my-gp-instance-gpexpand-job"
				Expect(reactiveClient.Create(nil, &job)).To(Succeed())
			})
			It("does not allow requests to increase primarySegmentCount",
				Disallowed("cannot expand cluster because a gpexpand job is currently running"))
		})
		When("webhook fails to get the gpexpand job", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("get", "jobs", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("error getting my-gp-instance-gpexpand-job")
				})
			})
			It("does not allow requests to increase primarySegmentCount",
				Disallowed("failed to check for previous expand job: error getting my-gp-instance-gpexpand-job"))
		})
		When("webhook fails to contact an active gpdb master", func() {
			BeforeEach(func() {
				subject.PodCmdExecutor = &fake.PodExec{
					ErrorMsgOnMaster0: "fake error",
					ErrorMsgOnMaster1: "fake error",
				}
			})
			It("does not allow increasing primarySegmentCount",
				Disallowed("failed to contact an active gpdb master"))
		})
		When("webhook fails to run query on gpdb master", func() {
			BeforeEach(func() {
				subject.PodCmdExecutor = &fake.PodExec{
					ErrorMsgOnCommand: "fake error",
				}
			})
			It("does not allow increasing primarySegmentCount",
				Disallowed("failed to check for expansion schema: fake error"))
		})
		When("previous expansion schema exists", func() {
			BeforeEach(func() {
				subject.PodCmdExecutor = &fake.PodExec{
					StdoutResult: "1\n",
				}
			})
			It("does not allow increasing primarySegmentCount",
				Disallowed("previous expansion schema exists. you must redistribute data and clean up expansion schema prior to performing another expansion"))
		})
	})

	It("disallows requests that decrease PrimarySegmentCount", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.Segments.PrimarySegmentCount--

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse())
		const expectedMessage = "primarySegmentCount cannot be decreased after the cluster has been created"
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal(expectedMessage),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry(expectedMessage))
	})

	It("disallows update requests that update an old greenplum instance", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.Segments.PrimarySegmentCount = 1
		oldGreenplum.Status.InstanceImage = "v1.0.0"
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.Segments.PrimarySegmentCount = 2
		subject.InstanceImage = "v1.0.1"

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		const expectedMessage = admission.UpgradeClusterHelpMsg + `; GreenplumCluster has image: v1.0.0; Operator supports image: v1.0.1`
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal(expectedMessage),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry(expectedMessage))
	})

	It("disallows requests that change cpu for MasterAndStandby", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.MasterAndStandby.CPU = resource.MustParse("1000000.0")

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("CPU reservation cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("CPU reservation cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change cpu for Segments", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.Segments.CPU = resource.MustParse("1000000.0")

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("CPU reservation cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("CPU reservation cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change memory for MasterAndStandby", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.MasterAndStandby.Memory = resource.MustParse("1.21G")

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("Memory reservation cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("Memory reservation cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change memory for Segments", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.Segments.Memory = resource.MustParse("1.21G")

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("Memory reservation cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("Memory reservation cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change standby", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.MasterAndStandby.Standby = "no"
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.MasterAndStandby.Standby = "yes"

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("standby value cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("standby value cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change hostBasedAuthentication", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.MasterAndStandby.HostBasedAuthentication = "initial value"
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.MasterAndStandby.HostBasedAuthentication = "changed value"

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("hostBasedAuthentication cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("hostBasedAuthentication cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change MasterAndStandby workerSelector", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.MasterAndStandby.WorkerSelector = map[string]string{
			"my-gp-master": "true",
		}
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.MasterAndStandby.WorkerSelector = map[string]string{
			"my-gp-master": "false",
		}

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("workerSelector cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("workerSelector cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change Segments workerSelector", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.Segments.WorkerSelector = map[string]string{
			"my-gp-master": "true",
		}
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.Segments.WorkerSelector = map[string]string{
			"my-gp-master": "false",
		}

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("workerSelector cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("workerSelector cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change masterAndStandby antiAffinity", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.MasterAndStandby.AntiAffinity = "no"
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.MasterAndStandby.AntiAffinity = "yes"

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("antiAffinity cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("antiAffinity cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change segments antiAffinity", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.Segments.AntiAffinity = "no"
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.Segments.AntiAffinity = "yes"

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("antiAffinity cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("antiAffinity cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change segments mirrors", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.Segments.Mirrors = "no"
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.Segments.Mirrors = "yes"

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("mirrors cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("mirrors cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change segments mirrors", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.Segments.Mirrors = "yes"
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.Segments.Mirrors = "no"

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("mirrors cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("mirrors cannot be changed after the cluster has been created"))
	})

	DescribeTable("allows requests that only change the case of segments mirrors",
		func(oldMirrors, newMirrors string) {
			oldGreenplum := exampleGreenplum.DeepCopy()
			oldGreenplum.Spec.Segments.Mirrors = oldMirrors
			newGreenplum := oldGreenplum.DeepCopy()
			newGreenplum.Spec.Segments.Mirrors = newMirrors

			outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

			Expect(outputReview.Response.Allowed).To(BeTrue(), "should be allowed")
			Expect(DecodeLogs(logBuf)).To(ContainAllowedEntry())
		},
		Entry("no -> NO", "no", "NO"),
		Entry("NO -> no", "NO", "no"),
	)

	It("disallows requests that change masterAndStandby storage", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.MasterAndStandby.Storage = resource.MustParse("10G")
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.MasterAndStandby.Storage = resource.MustParse("20G")

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("storage cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("storage cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change segments storage", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.Segments.Storage = resource.MustParse("10G")
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.Segments.Storage = resource.MustParse("20G")

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("storage cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("storage cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change masterAndStandby storageClassName", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.MasterAndStandby.StorageClassName = "foo"
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.MasterAndStandby.StorageClassName = "bar"

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("storageClassName cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("storageClassName cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change segments storageClassName", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.Segments.StorageClassName = "foo"
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.Segments.StorageClassName = "bar"

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("storageClassName cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("storageClassName cannot be changed after the cluster has been created"))
	})

	It("disallows requests that change pxf serviceName", func() {
		oldGreenplum := exampleGreenplum.DeepCopy()
		oldGreenplum.Spec.PXF.ServiceName = "foo"
		newGreenplum := oldGreenplum.DeepCopy()
		newGreenplum.Spec.PXF.ServiceName = "bar"

		outputReview := postValidateReview(subject.Handler(), newGreenplum, oldGreenplum)

		Expect(outputReview.Response.Allowed).To(BeFalse(), "should not be allowed")
		Expect(outputReview.Response.Result).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Message": Equal("PXF serviceName cannot be changed after the cluster has been created"),
		})))
		Expect(DecodeLogs(logBuf)).To(ContainDisallowedEntry("PXF serviceName cannot be changed after the cluster has been created"))
	})
})
