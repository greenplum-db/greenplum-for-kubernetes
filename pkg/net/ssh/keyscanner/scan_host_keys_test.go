package keyscanner_test

import (
	"encoding/base64"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gstruct"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/gplog/testing"
	fakessh "github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/fake"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/keyscanner"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

var _ = Describe("scanKeysUtils", func() {
	var (
		fakeKeyScanner       *fakessh.KeyScanner
		fakeKnownHostsReader *fakessh.KnownHostsReader
		fakeHostList         []string
		outBuffer            *gbytes.Buffer
	)

	BeforeEach(func() {
		outBuffer = gbytes.NewBuffer()
		fakeKeyScanner = &fakessh.KeyScanner{}
		fakeKnownHostsReader = &fakessh.KnownHostsReader{}
		fakeHostList = []string{
			"master-0",
			"master-1",
			"segment-a-0",
			"segment-b-0",
		}

		keyscanner.Log = gplog.ForTest(outBuffer)
	})

	It("scans host keys", func() {
		hostKeys, err := keyscanner.ScanHostKeys(fakeKeyScanner, fakeKnownHostsReader, fakeHostList)
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeKeyScanner.WasCalled).To(BeTrue())
		Expect(fakeKeyScanner.Hostnames).To(ConsistOf([]string{"master-0", "master-1", "segment-a-0", "segment-b-0"}))
		Expect(hostKeys).To(ContainSubstring("master-0 FakeKey " + base64.StdEncoding.EncodeToString([]byte("master-0")) + "\n"))
		Expect(hostKeys).To(ContainSubstring("master-1 FakeKey " + base64.StdEncoding.EncodeToString([]byte("master-1")) + "\n"))
		Expect(hostKeys).To(ContainSubstring("segment-a-0 FakeKey " + base64.StdEncoding.EncodeToString([]byte("segment-a-0")) + "\n"))
		Expect(hostKeys).To(ContainSubstring("segment-b-0 FakeKey " + base64.StdEncoding.EncodeToString([]byte("segment-b-0")) + "\n"))
	})

	It("calls GetKnownHosts", func() {
		_, err := keyscanner.ScanHostKeys(fakeKeyScanner, fakeKnownHostsReader, fakeHostList)
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeKnownHostsReader.WasCalled).To(BeTrue())
	})

	It("writes log messages", func() {
		_, err := keyscanner.ScanHostKeys(fakeKeyScanner, fakeKnownHostsReader, fakeHostList)
		Expect(err).NotTo(HaveOccurred())

		logs, err := DecodeLogs(outBuffer)
		Expect(err).NotTo(HaveOccurred())
		Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("starting keyscan"), "host": Equal("master-0")}))
		Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("starting keyscan"), "host": Equal("master-1")}))
		Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("starting keyscan"), "host": Equal("segment-a-0")}))
		Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("starting keyscan"), "host": Equal("segment-b-0")}))
		Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("keyscan successful"), "host": Equal("master-0")}))
		Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("keyscan successful"), "host": Equal("master-1")}))
		Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("keyscan successful"), "host": Equal("segment-a-0")}))
		Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("keyscan successful"), "host": Equal("segment-b-0")}))
	})

	When("the key in known_hosts matches the callback key", func() {
		BeforeEach(func() {
			fakeKnownHostsReader.KnownHosts = map[string]ssh.PublicKey{
				"master-0": fakessh.KeyForHost("master-0"),
			}
		})
		It("does not return a key for that host", func() {
			hostKeys, err := keyscanner.ScanHostKeys(fakeKeyScanner, fakeKnownHostsReader, fakeHostList)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeKeyScanner.WasCalled).To(BeTrue())
			Expect(fakeKeyScanner.Hostnames).To(ConsistOf([]string{"master-0", "master-1", "segment-a-0", "segment-b-0"}))
			Expect(hostKeys).NotTo(ContainSubstring("master-0 FakeKey " + base64.StdEncoding.EncodeToString([]byte("master-0")) + "\n"))
			Expect(hostKeys).To(ContainSubstring("master-1 FakeKey " + base64.StdEncoding.EncodeToString([]byte("master-1")) + "\n"))
			Expect(hostKeys).To(ContainSubstring("segment-a-0 FakeKey " + base64.StdEncoding.EncodeToString([]byte("segment-a-0")) + "\n"))
			Expect(hostKeys).To(ContainSubstring("segment-b-0 FakeKey " + base64.StdEncoding.EncodeToString([]byte("segment-b-0")) + "\n"))
		})
	})
	When("the key in known hosts does not match the callback key", func() {
		BeforeEach(func() {
			fakeKnownHostsReader.KnownHosts = map[string]ssh.PublicKey{
				"master-0": fakessh.KeyForHost("imposter-master-0"),
			}
		})
		It("returns an error", func() {
			_, err := keyscanner.ScanHostKeys(fakeKeyScanner, fakeKnownHostsReader, fakeHostList)
			Expect(err).To(MatchError("scanned key does not match known key"))
		})
		It("logs an error message", func() {
			_, err := keyscanner.ScanHostKeys(fakeKeyScanner, fakeKnownHostsReader, fakeHostList)
			logs, err := DecodeLogs(outBuffer)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs).To(ContainLogEntry(gstruct.Keys{
				"level": Equal("ERROR"),
				"msg":   Equal("keyscan failed"),
				"host":  Equal("master-0"),
				"error": Equal("scanned key does not match known key"),
			}))
		})
	})

	It("throws err if ssh-keyscan fails", func() {
		fakeKeyScanner.Err = errors.New("timed out waiting for keyscan on host: master-0")

		hostKeys, err := keyscanner.ScanHostKeys(fakeKeyScanner, fakeKnownHostsReader, fakeHostList)
		Expect(fakeKeyScanner.WasCalled).To(BeTrue())
		Expect(hostKeys).To(BeEmpty())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("timed out waiting for keyscan on host: master-0"))
	})

	When("calling GetKnownHosts fails", func() {
		var (
			hostKeysErr error
		)
		BeforeEach(func() {
			fakeKnownHostsReader.Err = errors.New("could not get known hosts")
			_, hostKeysErr = keyscanner.ScanHostKeys(fakeKeyScanner, fakeKnownHostsReader, fakeHostList)
		})
		It("throws an error", func() {
			Expect(hostKeysErr).To(MatchError("could not get known hosts"))
		})
	})
})
