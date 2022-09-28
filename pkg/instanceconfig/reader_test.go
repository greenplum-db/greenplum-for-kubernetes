package instanceconfig_test

import (
	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig"
)

var _ = Describe("instanceconfig.Reader", func() {

	var (
		memoryfs *memfs.MemFS
		subject  instanceconfig.Reader
	)

	BeforeEach(func() {
		memoryfs = memfs.Create()
		Expect(vfs.MkdirAll(memoryfs, instanceconfig.ConfigMapPathPrefix, 0755)).To(Succeed())
		Expect(vfs.MkdirAll(memoryfs, instanceconfig.PodInfoPathPrefix, 0755)).To(Succeed())
		subject = instanceconfig.NewReader(memoryfs)
	})

	Describe("GetNamespace", func() {
		When("namespace file exists", func() {
			It("reads a string successfully", func() {
				Expect(vfs.WriteFile(memoryfs, "/etc/podinfo/namespace", []byte("testName"), 0777)).To(Succeed())
				name, err := subject.GetNamespace()
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("testName"))
			})
		})
		When("namespace file does not exist", func() {
			It("returns an error", func() {
				_, err := subject.GetNamespace()
				Expect(err).To(MatchError("open /etc/podinfo/namespace: file does not exist"))
			})
		})
	})

	Describe("GetGreenplumClusterName", func() {
		When("greenplum-cluster file exists", func() {
			It("reads a string successfully", func() {
				Expect(vfs.WriteFile(memoryfs, "/etc/podinfo/greenplumClusterName", []byte("testName"), 0777)).To(Succeed())
				name, err := subject.GetGreenplumClusterName()
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("testName"))
			})
		})
		When("greenplum-cluster file does not exist", func() {
			It("returns an error", func() {
				_, err := subject.GetGreenplumClusterName()
				Expect(err).To(MatchError("open /etc/podinfo/greenplumClusterName: file does not exist"))
			})
		})
	})

	Describe("GetSegmentCount", func() {
		When("/etc/config/segmentCount is positive", func() {
			BeforeEach(func() {
				Expect(vfs.WriteFile(memoryfs, "/etc/config/segmentCount", []byte("1"), 0777)).To(Succeed())
			})
			It("successfully reads segmentCount", func() {
				count, err := subject.GetSegmentCount()
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(1))
			})
		})

		When("/etc/config/segmentCount is zero", func() {
			BeforeEach(func() {
				Expect(vfs.WriteFile(memoryfs, "/etc/config/segmentCount", []byte("0"), 0777)).To(Succeed())
			})
			It("successfully reads segmentcount", func() {
				count, err := subject.GetSegmentCount()
				Expect(err).To(MatchError("segmentCount must be > 0"))
				Expect(count).To(Equal(0))
			})
		})

		When("/etc/config/segmentCount is negative", func() {
			BeforeEach(func() {
				Expect(vfs.WriteFile(memoryfs, "/etc/config/segmentCount", []byte("-1"), 0777)).To(Succeed())
			})
			It("returns an error", func() {
				count, err := subject.GetSegmentCount()
				Expect(err).To(MatchError("segmentCount must be > 0"))
				Expect(count).To(Equal(0))
			})
		})

		When("/etc/config/segmentCount is absent", func() {
			It("returns an error", func() {
				count, err := subject.GetSegmentCount()
				Expect(err).To(MatchError("open /etc/config/segmentCount: file does not exist"))
				Expect(count).To(Equal(0))
			})
		})

		When("/etc/config contains non-integer value", func() {
			BeforeEach(func() {
				Expect(vfs.WriteFile(memoryfs, "/etc/config/segmentCount", []byte("blah"), 0777)).To(Succeed())
			})
			It("returns an error", func() {
				count, err := subject.GetSegmentCount()
				Expect(err).To(MatchError("strconv.Atoi: parsing \"blah\": invalid syntax"))
				Expect(count).To(Equal(0))
			})
		})
	})

	Describe("GetMirrors", func() {
		It("does not fail when mirrors is a valid boolean", func() {
			Expect(vfs.WriteFile(memoryfs, "/etc/config/mirrors", []byte("true"), 0777)).To(Succeed())
			_, err := subject.GetMirrors()
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails when mirrors not a boolean", func() {
			Expect(vfs.WriteFile(memoryfs, "/etc/config/mirrors", []byte("not-a-bool"), 0777)).To(Succeed())
			_, err := subject.GetMirrors()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error parsing mirrors, must be a boolean, got: not-a-bool"))
		})

		It("fails when mirrors config file is not present", func() {
			_, err := subject.GetMirrors()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error reading mirrors: open /etc/config/mirrors: file does not exist"))
		})
	})

	Describe("GetStandby", func() {
		It("does not fail when standby is a valid boolean", func() {
			Expect(vfs.WriteFile(memoryfs, "/etc/config/standby", []byte("true"), 0777)).To(Succeed())
			_, err := subject.GetStandby()
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails when standby not a boolean", func() {
			Expect(vfs.WriteFile(memoryfs, "/etc/config/standby", []byte("not-a-bool"), 0777)).To(Succeed())
			_, err := subject.GetStandby()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error parsing standby, must be a boolean, got: not-a-bool"))
		})

		It("fails when standby config file is not present", func() {
			_, err := subject.GetStandby()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error reading standby: open /etc/config/standby: file does not exist"))
		})
	})

	Describe("GetPXFServiceName", func() {
		When("pxfServiceName is defined", func() {
			It("reads a string successfully", func() {
				Expect(vfs.WriteFile(memoryfs, "/etc/config/pxfServiceName", []byte("testName"), 0777)).To(Succeed())
				name, err := subject.GetPXFServiceName()
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("testName"))
			})
		})
		When("pxfServiceName is empty", func() {
			It("returns empty string without error", func() {
				name, err := subject.GetPXFServiceName()
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal(""))
			})
		})
	})

	Describe("GetConfigValues", func() {
		BeforeEach(func() {
			Expect(vfs.WriteFile(memoryfs, "/etc/podinfo/namespace", []byte("testns"), 0777)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/etc/podinfo/greenplumClusterName", []byte("my-greenplum"), 0777)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/etc/config/segmentCount", []byte("1"), 0777)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/etc/config/mirrors", []byte("true"), 0777)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/etc/config/standby", []byte("true"), 0777)).To(Succeed())
			Expect(vfs.WriteFile(memoryfs, "/etc/config/pxfServiceName", []byte("testPXFName"), 0777)).To(Succeed())
		})
		When("all values are populated", func() {
			It("populates the struct", func() {
				config, err := subject.GetConfigValues()
				Expect(err).NotTo(HaveOccurred())
				Expect(config).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Namespace":            Equal("testns"),
					"GreenplumClusterName": Equal("my-greenplum"),
					"SegmentCount":         Equal(1),
					"Mirrors":              BeTrue(),
					"Standby":              BeTrue(),
					"PXFServiceName":       Equal("testPXFName"),
				}))
			})
		})

		When("getting one of the values fail", func() {
			BeforeEach(func() {
				Expect(vfs.RemoveAll(memoryfs, "/etc/config/standby")).To(Succeed())
			})
			It("returns error", func() {
				_, err := subject.GetConfigValues()
				Expect(err).To(MatchError("error reading standby: open /etc/config/standby: file does not exist"))
			})
		})
	})
})
