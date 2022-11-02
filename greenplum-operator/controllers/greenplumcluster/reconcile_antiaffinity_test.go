package greenplumcluster_test

import (
	"context"
	"errors"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers/greenplumcluster"
	fakeexec "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor/fake"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/reactive"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("antiAffinity tests", func() {
	var (
		fakeGreenplumClusterSpec *greenplumv1.GreenplumCluster
		testNodes                *corev1.NodeList
		exampleValidNodeList     *corev1.NodeList

		logBuf              *gbytes.Buffer
		greenplumReconciler *greenplumcluster.GreenplumClusterReconciler
		ctx                 context.Context
	)

	BeforeEach(func() {
		fakeGreenplumClusterSpec = exampleGreenplumCluster.DeepCopy()
		fakeGreenplumClusterSpec.Spec.MasterAndStandby.AntiAffinity = "yes"
		fakeGreenplumClusterSpec.Spec.MasterAndStandby.Standby = "yes"
		fakeGreenplumClusterSpec.Spec.Segments.AntiAffinity = "yes"
		fakeGreenplumClusterSpec.Spec.Segments.Mirrors = "yes"

		exampleValidNodeList = &corev1.NodeList{
			Items: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node1",
						Labels: map[string]string{"worker": "my-gp-masters"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node2",
						Labels: map[string]string{"worker": "my-gp-masters"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node3",
						Labels: map[string]string{"worker": "my-gp-segments"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node4",
						Labels: map[string]string{"worker": "my-gp-segments"},
					},
				},
				{ //node does not have a workerselector label
					ObjectMeta: metav1.ObjectMeta{
						Name: "node5",
					},
				},
			},
		}

		logBuf = gbytes.NewBuffer()
		ctx = context.WithValue(context.Background(), struct{ key string }{"test"}, CurrentGinkgoTestDescription().TestText)
	})

	JustBeforeEach(func() {
		reactiveClient = reactive.NewClient(fake.NewFakeClientWithScheme(scheme.Scheme, testNodes))
		Expect(reactiveClient.Create(nil, fakeGreenplumClusterSpec)).To(Succeed())
		greenplumReconciler = &greenplumcluster.GreenplumClusterReconciler{
			Client:        reactiveClient,
			Log:           gplog.ForTest(logBuf),
			SSHCreator:    fakeSecretCreator{},
			PodExec:       &fakeexec.PodExec{},
			InstanceImage: "greenplum-for-kubernetes:tag",
			OperatorImage: "greenplum-operator:tag",
		}
	})

	Describe("HandleAntiAffinity", func() {
		When("both masterAndStandby.workerSelector and segments.workerSelector are set", func() {
			BeforeEach(func() {
				fakeGreenplumClusterSpec.Spec.MasterAndStandby.WorkerSelector = map[string]string{"worker": "my-gp-masters"}
				fakeGreenplumClusterSpec.Spec.Segments.WorkerSelector = map[string]string{"worker": "my-gp-segments"}
				testNodes = exampleValidNodeList
			})
			When("at least 2 worker nodes for master and segment are available", func() {
				It("succeeds and labels nodes", func() {
					Expect(greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)).To(Equal(ctrl.Result{}))
					checkNodeLabels(fakeGreenplumClusterSpec)
				})
			})
			When("gpdb cluster resources already exist", func() {
				var nodePatched bool
				JustBeforeEach(func() {
					Expect(greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)).To(Equal(ctrl.Result{}))
					reactiveClient.PrependReactor("patch", "nodes", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						nodePatched = true
						return false, nil, nil
					})
				})
				It("does not label the nodes with antiaffinity labels", func() {
					Expect(greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)).To(Equal(ctrl.Result{}))
					Expect(nodePatched).To(BeFalse())
				})
			})
			When("there is an error checking if gpdb cluster resources exist", func() {
				var errMsg string
				JustBeforeEach(func() {
					errMsg = "failed to list statefulsets"
					reactiveClient.PrependReactor("list", "statefulsets", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						listAction := action.(testing.ListAction)
						listLabels := listAction.GetListRestrictions().Labels
						if listLabels.Matches(labels.Set{"greenplum-cluster": greenplumClusterRequest.Name}) {
							return true, nil, errors.New(errMsg)
						}
						return false, nil, nil
					})
				})
				It("returns an error", func() {
					reconcileResult, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
					Expect(reconcileResult).To(Equal(ctrl.Result{}))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("unable to check if GreenplumCluster resources exist: " + errMsg))
				})
			})
			When("there are no nodes available with matching master and/or segment worker selectors", func() {
				BeforeEach(func() {
					testNodes = &corev1.NodeList{
						Items: []corev1.Node{
							{ObjectMeta: metav1.ObjectMeta{Name: "node1"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "node2"}},
						},
					}
				})
				It("returns an error", func() {
					reconcileResult, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
					Expect(reconcileResult).To(Equal(ctrl.Result{}))
					Expect(err).To(HaveOccurred())
					expectedErrMsg := fmt.Sprintf(greenplumcluster.NodeCountErrorFmt, 0, 0)
					Expect(err.Error()).To(Equal("antiAffinity: instance my-greenplum does not meet requirements: " + expectedErrMsg))
				})
			})
			When("there are not enough nodes labeled with master worker selector", func() {
				BeforeEach(func() {
					testNodes = &corev1.NodeList{
						Items: []corev1.Node{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:   "node1",
									Labels: map[string]string{"worker": "my-gp-masters"},
								},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:   "node2",
									Labels: map[string]string{"worker": "my-gp-segments"},
								},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:   "node3",
									Labels: map[string]string{"worker": "my-gp-segments"},
								},
							},
						},
					}
				})
				It("returns an error", func() {
					reconcileResult, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
					Expect(reconcileResult).To(Equal(ctrl.Result{}))
					Expect(err).To(HaveOccurred())
					expectedErrMsg := fmt.Sprintf(greenplumcluster.NodeCountErrorFmt, 1, 2)
					Expect(err.Error()).To(Equal("antiAffinity: instance my-greenplum does not meet requirements: " + expectedErrMsg))
				})
			})
			When("there are not enough nodes labeled with segment worker selector", func() {
				BeforeEach(func() {
					testNodes = &corev1.NodeList{
						Items: []corev1.Node{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:   "node1",
									Labels: map[string]string{"worker": "my-gp-masters"},
								},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:   "node2",
									Labels: map[string]string{"worker": "my-gp-masters"},
								},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:   "node3",
									Labels: map[string]string{"worker": "my-gp-segments"},
								},
							},
						},
					}
				})
				It("returns an error", func() {
					reconcileResult, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
					Expect(reconcileResult).To(Equal(ctrl.Result{}))
					Expect(err).To(HaveOccurred())
					expectedErrMsg := fmt.Sprintf(greenplumcluster.NodeCountErrorFmt, 2, 1)
					Expect(err.Error()).To(Equal("antiAffinity: instance my-greenplum does not meet requirements: " + expectedErrMsg))
				})
			})
			When("listing master nodes returns an error", func() {
				var errMsg string
				JustBeforeEach(func() {
					errMsg = "failed to list nodes"
					reactiveClient.PrependReactor("list", "nodes", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						listAction := action.(testing.ListAction)
						listLabels := listAction.GetListRestrictions().Labels
						if listLabels.Matches(labels.Set(fakeGreenplumClusterSpec.Spec.MasterAndStandby.WorkerSelector)) {
							return true, nil, errors.New(errMsg)
						}
						return false, nil, nil
					})
				})
				It("returns an error", func() {
					reconcileResult, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
					Expect(reconcileResult).To(Equal(ctrl.Result{}))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("antiAffinity: master node worker selector list: " + errMsg))
				})
			})
			When("listing segments nodes returns an error", func() {
				var errMsg string
				JustBeforeEach(func() {
					errMsg = "failed to list nodes"
					reactiveClient.PrependReactor("list", "nodes", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						listAction := action.(testing.ListAction)
						listLabels := listAction.GetListRestrictions().Labels
						if listLabels.Matches(labels.Set(fakeGreenplumClusterSpec.Spec.Segments.WorkerSelector)) {
							return true, nil, errors.New(errMsg)
						}
						return false, nil, nil
					})
				})
				It("returns an error", func() {
					reconcileResult, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
					Expect(reconcileResult).To(Equal(ctrl.Result{}))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("antiAffinity: segment node worker selector list: " + errMsg))
				})
			})
			When("patching nodes with master label returns an error", func() {
				var errMsg string
				JustBeforeEach(func() {
					errMsg = "failed to patch node"
					reactiveClient.PrependReactor("patch", "nodes", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						patchAction := action.(testing.PatchAction)
						patchString := string(patchAction.GetPatch())
						if strings.Contains(patchString, "greenplum-affinity-test-ns-master") {
							return true, nil, errors.New(errMsg)
						}
						return false, nil, nil
					})
				})
				It("returns an error", func() {
					reconcileResult, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
					Expect(reconcileResult).To(Equal(ctrl.Result{}))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(MatchRegexp("antiAffinity: failed to add label '.+=.+' to node '.+': " + errMsg))
				})
			})
			When("patching nodes with segment label returns an error", func() {
				var errMsg string
				JustBeforeEach(func() {
					errMsg = "failed to patch node"
					reactiveClient.PrependReactor("patch", "nodes", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						patchAction := action.(testing.PatchAction)
						patchString := string(patchAction.GetPatch())
						if strings.Contains(patchString, "greenplum-affinity-test-ns-segment") {
							return true, nil, errors.New(errMsg)
						}
						return false, nil, nil
					})
				})
				It("returns an error", func() {
					reconcileResult, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
					Expect(reconcileResult).To(Equal(ctrl.Result{}))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(MatchRegexp("antiAffinity: failed to add label '.+=.+' to node '.+': " + errMsg))
				})
			})
		})

		When("masterAndStandby.workerSelector is set but segments.workerSelector is not set", func() {
			BeforeEach(func() {
				fakeGreenplumClusterSpec.Spec.MasterAndStandby.WorkerSelector = map[string]string{"worker": "my-gp-masters"}
				testNodes = exampleValidNodeList
			})
			When("at least 2 worker nodes for master and segment are available", func() {
				It("succeeds and labels nodes", func() {
					Expect(greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)).To(Equal(ctrl.Result{}))
					checkNodeLabels(fakeGreenplumClusterSpec)
				})
			})
			When("there are not enough nodes labeled with master worker selector", func() {
				BeforeEach(func() {
					testNodes = &corev1.NodeList{
						Items: []corev1.Node{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:   "node1",
									Labels: map[string]string{"worker": "my-gp-masters"},
								},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "node2",
								},
							},
						},
					}
				})
				It("returns an error", func() {
					reconcileResult, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
					Expect(reconcileResult).To(Equal(ctrl.Result{}))
					Expect(err).To(HaveOccurred())
					expectedErrMsg := fmt.Sprintf(greenplumcluster.NodeCountErrorFmt, 1, 2)
					Expect(err.Error()).To(Equal("antiAffinity: instance my-greenplum does not meet requirements: " + expectedErrMsg))
				})
			})
		})

		When("segments.workerSelector is set but masterAndStandby.workerSelector is not set", func() {
			BeforeEach(func() {
				fakeGreenplumClusterSpec.Spec.Segments.WorkerSelector = map[string]string{"worker": "my-gp-segments"}
				testNodes = exampleValidNodeList
			})
			When("at least 2 worker nodes for master and segment are available", func() {
				It("succeeds and labels nodes", func() {
					Expect(greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)).To(Equal(ctrl.Result{}))
					checkNodeLabels(fakeGreenplumClusterSpec)
				})
			})
			When("there are not enough nodes labeled with segment worker selector", func() {
				BeforeEach(func() {
					testNodes = &corev1.NodeList{
						Items: []corev1.Node{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:   "node1",
									Labels: map[string]string{"worker": "my-gp-segments"},
								},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "node2",
								},
							},
						},
					}
				})
				It("returns an error", func() {
					reconcileResult, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
					Expect(reconcileResult).To(Equal(ctrl.Result{}))
					Expect(err).To(HaveOccurred())
					expectedErrMsg := fmt.Sprintf(greenplumcluster.NodeCountErrorFmt, 2, 1)
					Expect(err.Error()).To(Equal("antiAffinity: instance my-greenplum does not meet requirements: " + expectedErrMsg))
				})
			})
		})

		When("neither masterAndStandby.workerSelector nor segment.workerSelector is set", func() {
			When("there are at least two nodes available", func() {
				BeforeEach(func() {
					testNodes = &corev1.NodeList{
						Items: []corev1.Node{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "node1",
								},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "node2",
								},
							},
						},
					}
				})
				It("succeeds and labels nodes", func() {
					Expect(greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)).To(Equal(ctrl.Result{}))
					checkNodeLabels(fakeGreenplumClusterSpec)
				})
			})
			When("there is only one node available (i.e. minikube)", func() {
				BeforeEach(func() {
					testNodes = &corev1.NodeList{
						Items: []corev1.Node{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "node1",
								},
							},
						},
					}
				})
				It("returns an error", func() {
					reconcileResult, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
					Expect(reconcileResult).To(Equal(ctrl.Result{}))
					Expect(err).To(HaveOccurred())
					expectedErrMsg := fmt.Sprintf(greenplumcluster.NodeCountErrorFmt, 1, 1)
					Expect(err.Error()).To(Equal("antiAffinity: instance my-greenplum does not meet requirements: " + expectedErrMsg))
				})
			})
		})

		When("master.antiaffinity and segment.antiaffinity are set to different values", func() {
			BeforeEach(func() {
				fakeGreenplumClusterSpec.Spec.MasterAndStandby.AntiAffinity = "yes"
				fakeGreenplumClusterSpec.Spec.Segments.AntiAffinity = "no"
				testNodes = exampleValidNodeList
			})
			It("returns an error", func() {
				reconcileResult, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
				Expect(reconcileResult).To(Equal(ctrl.Result{}))
				Expect(err).To(HaveOccurred())
				expectedErrMsg := fmt.Sprintf(greenplumcluster.AntiAffinityMismatchErrorFmt, fakeGreenplumClusterSpec.Spec.Segments.AntiAffinity, fakeGreenplumClusterSpec.Spec.MasterAndStandby.AntiAffinity)
				Expect(err.Error()).To(Equal("antiAffinity: instance my-greenplum does not meet requirements: " + expectedErrMsg))
			})
		})
	})
})

func checkNodeLabels(greenplumCluster *greenplumv1.GreenplumCluster) {
	var nodeList corev1.NodeList
	Expect(reactiveClient.List(nil, &nodeList)).To(Succeed())
	masterNodeLabelKey := fmt.Sprintf("greenplum-affinity-%s-master", greenplumCluster.Namespace)
	segNodeLabelKey := fmt.Sprintf("greenplum-affinity-%s-segment", greenplumCluster.Namespace)
	for _, node := range nodeList.Items {
		nodeLabels := node.GetLabels()
		if greenplumCluster.Spec.MasterAndStandby.WorkerSelector == nil ||
			nodeLabels["worker"] == greenplumCluster.Spec.MasterAndStandby.WorkerSelector["worker"] {
			Expect(nodeLabels[masterNodeLabelKey] == "true").To(BeTrue())
		}
		if greenplumCluster.Spec.Segments.WorkerSelector == nil ||
			nodeLabels["worker"] == greenplumCluster.Spec.Segments.WorkerSelector["worker"] {
			Expect((nodeLabels[segNodeLabelKey] == "a") || (nodeLabels[segNodeLabelKey] == "b")).To(BeTrue())
		}
	}
}
