package controllers_test

import (
	"bytes"
	"context"
	"errors"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gstruct"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/controllers"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/gplog/testing"
	instanceconfigTesting "github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig/testing"
	fakemultihost "github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost/testing"
	fakessh "github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/fake"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/keyscanner"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/testing"
)

var _ = Describe("KnownHostsController", func() {
	var subject *controllers.KnownHostsController

	Describe("Run happy path", func() {
		var (
			ctx           context.Context
			cancel        context.CancelFunc
			testClient    *fake.Clientset
			mockConfig    *instanceconfigTesting.MockReader
			reconciler    *reconcileSpy
			fakeEndpoints *corev1.Endpoints
		)
		BeforeEach(func() {
			ctx, cancel = context.WithCancel(context.Background())
			testClient = fake.NewSimpleClientset()
			mockConfig = &instanceconfigTesting.MockReader{}
			reconciler = &reconcileSpy{called: make(chan interface{})}
			subject = &controllers.KnownHostsController{
				ClientFn: func() (k kubernetes.Interface, err error) {
					return testClient, nil
				},
				ConfigReader: mockConfig,
				Reconciler:   reconciler,
			}

			fakeEndpoints = &corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-ns",
					Name:      "agent",
				},
			}
		})
		JustBeforeEach(func() {
			Expect(subject.Run(ctx)).To(Succeed())
		})
		AfterEach(func() {
			cancel()
		})
		When("endpoints are added", func() {
			JustBeforeEach(func() {
				_, err := testClient.CoreV1().Endpoints("test-ns").Create(ctx, fakeEndpoints, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})
			It("calls Reconcile", func() {
				Eventually(reconciler.called).Should(Receive(Equal(fakeEndpoints)))
			})
		})
		When("endpoints exist", func() {
			JustBeforeEach(func() {
				_, err := testClient.CoreV1().Endpoints("test-ns").Create(ctx, fakeEndpoints, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
				Eventually(reconciler.called).Should(Receive(Equal(fakeEndpoints)))
			})
			When("endpoints are updated", func() {
				JustBeforeEach(func() {
					fakeEndpoints.Subsets = []corev1.EndpointSubset{{
						Addresses: []corev1.EndpointAddress{
							{IP: "1.1.1.1", Hostname: "master-0"},
						},
					}}
					_, err := testClient.CoreV1().Endpoints("test-ns").Update(ctx, fakeEndpoints, metav1.UpdateOptions{})
					Expect(err).NotTo(HaveOccurred())
				})
				It("calls Reconcile", func() {
					Eventually(reconciler.called).Should(Receive(Equal(fakeEndpoints)))
				})
			})
			When("endpoints are deleted", func() {
				JustBeforeEach(func() {
					Expect(testClient.CoreV1().Endpoints("test-ns").Delete(ctx, "agent", metav1.DeleteOptions{})).To(Succeed())
				})
				It("does not call Reconcile", func() {
					Consistently(reconciler.called).ShouldNot(Receive())
				})
			})
		})
		Context("Informer filters", func() {
			var (
				actionCh    chan testing.WatchActionImpl
				watchAction testing.WatchActionImpl
			)
			BeforeEach(func() {
				actionCh = make(chan testing.WatchActionImpl)
				mockConfig.NamespaceName = "test-ns"
				mockConfig.GreenplumClusterName = "my-greenplum"
				testClient.PrependWatchReactor("endpoints", func(action testing.Action) (handled bool, ret watch.Interface, err error) {
					actionCh <- action.(testing.WatchActionImpl)
					return false, nil, nil
				})
			})
			It("filters on greenplumCluster namespace", func() {
				Eventually(actionCh).Should(Receive(&watchAction))
				Expect(watchAction.Namespace).To(Equal("test-ns"))
			})
			It("has greenplum-cluster label selector", func() {
				Eventually(actionCh).Should(Receive(&watchAction))
				labelSelector := watchAction.WatchRestrictions.Labels
				expectedLabels := labels.Set{"greenplum-cluster": "my-greenplum"}
				Expect(labelSelector.Matches(expectedLabels)).To(BeTrue(), "should match labels")

			})
			It("has greenplum-cluster field selector", func() {
				Eventually(actionCh).Should(Receive(&watchAction))
				fieldSelector := watchAction.WatchRestrictions.Fields
				expectedFields := fields.Set{"metadata.name": "agent"}
				Expect(fieldSelector.Matches(expectedFields)).To(BeTrue(), "should match fields")

			})
		})
	})
	When("Run error cases", func() {
		var (
			mockConfig *instanceconfigTesting.MockReader
		)
		BeforeEach(func() {
			mockConfig = &instanceconfigTesting.MockReader{}
			subject = &controllers.KnownHostsController{
				ClientFn: func() (k kubernetes.Interface, err error) {
					return nil, nil
				},
				ConfigReader: mockConfig,
			}
		})
		When("ClientFn fails", func() {
			BeforeEach(func() {
				subject.ClientFn = func() (k kubernetes.Interface, err error) {
					return nil, errors.New("epic fail")
				}
			})
			It("returns an error", func() {
				Expect(subject.Run(nil)).To(MatchError("failed to initialize client: epic fail"))
			})
		})
		When("getting namespace fails", func() {
			BeforeEach(func() {
				mockConfig.NamespaceNameErr = errors.New("no namespace here")
			})
			It("returns an error", func() {
				Expect(subject.Run(nil)).To(MatchError("failed to read namespace: no namespace here"))
			})
		})
		When("getting cluster name fails", func() {
			BeforeEach(func() {
				mockConfig.GreenplumClusterNameErr = errors.New("nope")
			})
			It("returns an error", func() {
				Expect(subject.Run(nil)).To(MatchError("failed to read greenplumcluster name: nope"))
			})
		})
	})
})

var _ = Describe("KnownHostsReconciler", func() {
	Describe("Reconcile", func() {
		var (
			subject              *controllers.KnownHostsReconciler
			logBuf               *gbytes.Buffer
			fakeFs               *memfs.MemFS
			fakeDNSResolver      *fakemultihost.FakeOperation
			fakeSSHKeyScanner    *fakessh.KeyScanner
			fakeKnownHostsReader *fakessh.KnownHostsReader
			fakeEndpoints        *corev1.Endpoints

			previousKnownHosts []byte
		)
		BeforeEach(func() {
			logBuf = gbytes.NewBuffer()

			fakeFs = memfs.Create()
			Expect(vfs.MkdirAll(fakeFs, "/home/gpadmin/.ssh/", 0711)).To(Succeed())
			var knownHostsFile bytes.Buffer
			knownHostsFile.WriteString("master-0 ")
			knownHostsFile.Write(ssh.MarshalAuthorizedKey(fakessh.KeyForHost("master-0")))
			knownHostsFile.WriteString("segment-a-0 ")
			knownHostsFile.Write(ssh.MarshalAuthorizedKey(fakessh.KeyForHost("segment-a-0")))
			previousKnownHosts = knownHostsFile.Bytes()
			Expect(vfs.WriteFile(fakeFs, controllers.KnownHostsFilename, previousKnownHosts, 0600)).To(Succeed())

			fakeSSHKeyScanner = &fakessh.KeyScanner{}
			fakeKnownHostsReader = &fakessh.KnownHostsReader{}
			fakeKnownHostsReader.KnownHosts = map[string]ssh.PublicKey{
				"master-0":    fakessh.KeyForHost("master-0"),
				"segment-a-0": fakessh.KeyForHost("segment-a-0"),
			}

			fakeDNSResolver = &fakemultihost.FakeOperation{}

			subject = &controllers.KnownHostsReconciler{
				Log:              gplog.ForTest(logBuf),
				Fs:               fakeFs,
				DNSResolver:      fakeDNSResolver,
				SSHKeyScanner:    fakeSSHKeyScanner,
				KnownHostsReader: fakeKnownHostsReader,
			}

			fakeEndpoints = &corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-ns",
					Name:      "agent",
				},
				Subsets: []corev1.EndpointSubset{{
					Addresses: []corev1.EndpointAddress{
						{IP: "1.1.1.1", Hostname: "master-0"},
						{IP: "2.2.2.2", Hostname: "segment-a-0"},
						{IP: "3.3.3.3", Hostname: "segment-a-1"},
					},
				}},
			}
		})
		JustBeforeEach(func() {
			subject.Reconcile(fakeEndpoints)
		})
		It("gets current known hosts", func() {
			Expect(fakeKnownHostsReader.WasCalled).To(BeTrue())
		})
		It("dns resolves unknown hosts", func() {
			Expect(fakeDNSResolver.HostRecords).To(Equal([]string{"segment-a-1"}))
		})
		It("ssh scans unknown hosts", func() {
			Expect(fakeSSHKeyScanner.Hostnames).To(Equal([]string{"segment-a-1"}))
		})
		It("appends new ssh host keys to /home/gpadmin/.ssh/known_hosts", func() {
			actualKnownHosts, err := vfs.ReadFile(fakeFs, controllers.KnownHostsFilename)
			Expect(err).NotTo(HaveOccurred())
			var expectedKnownHosts bytes.Buffer
			expectedKnownHosts.Write(previousKnownHosts)
			expectedKnownHosts.WriteString("segment-a-1 ")
			expectedKnownHosts.Write(ssh.MarshalAuthorizedKey(fakessh.KeyForHost("segment-a-1")))
			Expect(actualKnownHosts).To(Equal(expectedKnownHosts.Bytes()))
		})
		It("logs the hosts that are scanned", func() {
			logs, err := DecodeLogs(logBuf)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("scanning ssh host key(s)"), "hosts": ConsistOf("segment-a-1")}))
		})
		When("a host is removed", func() {
			BeforeEach(func() {
				fakeKnownHostsReader.KnownHosts = map[string]ssh.PublicKey{
					"master-0":    fakessh.KeyForHost("master-0"),
					"segment-a-0": fakessh.KeyForHost("segment-a-0"),
					"segment-a-1": fakessh.KeyForHost("segment-a-1"),
				}
				Expect(vfs.WriteFile(fakeFs, controllers.KnownHostsFilename, []byte("unmodified"), 0600)).To(Succeed())
				fakeEndpoints.Subsets[0].Addresses = []corev1.EndpointAddress{
					{IP: "1.1.1.1", Hostname: "master-0"},
					{IP: "2.2.2.2", Hostname: "segment-a-0"},
				}
			})
			It("does not scan any hosts", func() {
				Expect(fakeDNSResolver.HostRecords).To(BeEmpty())
				Expect(fakeSSHKeyScanner.WasCalled).To(BeFalse())
			})
			It("does not log", func() {
				Expect(logBuf.Contents()).To(BeEmpty(), string(logBuf.Contents()))
			})

			It("does not modify .ssh/known_hosts", func() {
				Expect(vfs.ReadFile(fakeFs, controllers.KnownHostsFilename)).To(BeEquivalentTo("unmodified"))
			})
		})
		When("a host key changes", func() {
			BeforeEach(func() {
				fakeKnownHostsReader.KnownHosts = map[string]ssh.PublicKey{
					"master-0": fakessh.KeyForHost("It's an older code, sir, but it checks out."),
				}
				Expect(vfs.WriteFile(fakeFs, controllers.KnownHostsFilename, []byte("unmodified"), 0600)).To(Succeed())
				fakeEndpoints.Subsets[0].Addresses = []corev1.EndpointAddress{
					{IP: "1.1.1.1", Hostname: "master-0"},
				}
			})
			It("does not modify .ssh/known_hosts", func() {
				Expect(vfs.ReadFile(fakeFs, controllers.KnownHostsFilename)).To(BeEquivalentTo("unmodified"))
			})
		})
		When("Reading known_hosts fails", func() {
			BeforeEach(func() {
				fakeKnownHostsReader.Err = errors.New("GetKnownHosts failed")
				fakeKnownHostsReader.KnownHosts = nil
			})
			It("logs an error", func() {
				logs, err := DecodeLogs(logBuf)
				Expect(err).NotTo(HaveOccurred())
				Expect(logs).To(ContainLogEntry(gstruct.Keys{
					"msg":   Equal("unable to get known hosts"),
					"error": Equal("GetKnownHosts failed"),
				}))
			})
			It("does not scan any hosts", func() {
				Expect(fakeDNSResolver.HostRecords).To(BeEmpty())
				Expect(fakeSSHKeyScanner.WasCalled).To(BeFalse())
			})
		})
		When("dns resolution fails", func() {
			BeforeEach(func() {
				subject.DNSResolver = &fakemultihost.FakeOperation{
					FakeErrors: map[string]error{
						"segment-a-1": errors.New("dns failure"),
					},
				}
			})
			It("logs the failure", func() {
				logs, err := DecodeLogs(logBuf)
				Expect(err).NotTo(HaveOccurred())
				Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("failed to resolve dns entries for all hosts"), "hosts": ConsistOf("segment-a-1")}))
			})
			It("does not scan any hosts", func() {
				Expect(fakeSSHKeyScanner.WasCalled).To(BeFalse())
			})
		})
		When("key scanning fails", func() {
			BeforeEach(func() {
				fakeSSHKeyScanner.Err = errors.New("SSHKeyScan error")
				Expect(vfs.WriteFile(fakeFs, controllers.KnownHostsFilename, []byte("unmodified"), 0600)).To(Succeed())
				keyscanner.Log = subject.Log
			})
			It("does not modify .ssh/known_hosts", func() {
				Expect(vfs.ReadFile(fakeFs, controllers.KnownHostsFilename)).To(BeEquivalentTo("unmodified"))
			})
			It("logs the keyscan failure", func() {
				// NB: This logging is actually done by ScanHostKeys. This is a contract test:
				// If ScanHostKeys stops logging, then we should change something here.
				logs, err := DecodeLogs(logBuf)
				Expect(err).NotTo(HaveOccurred())
				Expect(logs).To(ContainLogEntry(gstruct.Keys{
					"msg":   Equal("keyscan failed"),
					"host":  Equal("segment-a-1"),
					"error": Equal("SSHKeyScan error"),
				}))
			})
		})
		When("appending to .ssh/known_hosts fails", func() {
			BeforeEach(func() {
				subject.Fs = vfs.Dummy(errors.New("filesystem is cheese"))
			})
			It("logs an error", func() {
				logs, err := DecodeLogs(logBuf)
				Expect(err).NotTo(HaveOccurred())
				Expect(logs).To(ContainLogEntry(gstruct.Keys{
					"msg":   Equal("failed to write known_hosts file"),
					"error": Equal("filesystem is cheese"),
				}))
			})
		})
	})
	Describe("Reconcile non-Endpoints type", func() {
		// This test doesn't use the "JustBeforeEach()" in the other Describe("Reconcile")
		// b/c it needs to pass a different type to Reconcile().
		It("logs an error when non-Endpoint is passed", func() {
			logBuf := gbytes.NewBuffer()
			subject := &controllers.KnownHostsReconciler{
				Log: gplog.ForTest(logBuf),
			}

			subject.Reconcile(nil)

			logs, err := DecodeLogs(logBuf)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs).To(ContainLogEntry(gstruct.Keys{
				"msg": Equal("non *corev1.Endpoint event object received"),
			}))
		})
	})
})

type reconcileSpy struct {
	called chan interface{}
}

func (r *reconcileSpy) Reconcile(eventObj interface{}) {
	r.called <- eventObj
}
