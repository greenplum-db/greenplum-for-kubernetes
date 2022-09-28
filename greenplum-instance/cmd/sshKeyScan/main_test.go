package main

import (
	"encoding/base64"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	instanceconfigTesting "github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig/testing"
	fakessh "github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/fake"
)

type fakeFailingFileWriter struct {
}

func (f *fakeFailingFileWriter) Append(filename string, knownHosts string) error {
	return errors.New("failed file writer")
}

func (f *fakeFailingFileWriter) Insert(filename string, knownHosts string) error {
	return errors.New("failed file writer")
}

type fakeSuccessFileWriter struct {
	fileutil.FileWriter
	filename     string
	fileContents string
}

func (f *fakeSuccessFileWriter) Append(filename string, knownHosts string) error {
	f.filename = filename
	f.fileContents = knownHosts
	return nil
}

var _ = Describe("KeyScanApp", func() {
	var (
		app              *KeyScanApp
		errBuffer        *gbytes.Buffer
		outBuffer        *gbytes.Buffer
		fileWriter       *fakeSuccessFileWriter
		keyScanner       *fakessh.KeyScanner
		knownHostsReader *fakessh.KnownHostsReader
		mockConfig       *instanceconfigTesting.MockReader
		fakeCmd          *commandable.CommandFake
	)
	BeforeEach(func() {
		errBuffer = gbytes.NewBuffer()
		outBuffer = gbytes.NewBuffer()

		fileWriter = &fakeSuccessFileWriter{}

		keyScanner = &fakessh.KeyScanner{}
		knownHostsReader = &fakessh.KnownHostsReader{}
		mockConfig = &instanceconfigTesting.MockReader{
			SegmentCount: 1,
			Mirrors:      true,
			Standby:      true,
		}
		fakeCmd = commandable.NewFakeCommand()

		app = &KeyScanApp{
			keyScanner:       keyScanner,
			knownHostsReader: knownHostsReader,
			config:           mockConfig,
			command:          fakeCmd.Command,
			fileWriter:       fileWriter,
			stdoutBuffer:     outBuffer,
			stderrBuffer:     errBuffer,
		}
	})
	When("sshKeyScan succeeds", func() {
		BeforeEach(func() {
			fakeCmd.ExpectCommand("dnsdomainname").PrintsOutput("myheadlessservice.mynamespace.svc.cluster.local")
		})
		It("creates an entry in known Hosts file", func() {
			status := app.Run()
			Expect(status).To(Equal(0))
			Expect(outBuffer).To(gbytes.Say("Key scanning started\n"))
			Expect(fileWriter.filename).To(Equal("/home/gpadmin/.ssh/known_hosts"))
			Expect(fileWriter.fileContents).To(ContainSubstring("master-0 FakeKey " + base64.StdEncoding.EncodeToString([]byte("master-0")) + "\n"))
			Expect(fileWriter.fileContents).To(ContainSubstring("master-0.myheadlessservice.mynamespace.svc.cluster.local FakeKey " + base64.StdEncoding.EncodeToString([]byte("master-0.myheadlessservice.mynamespace.svc.cluster.local")) + "\n"))
			Expect(fileWriter.fileContents).To(ContainSubstring("master-1 FakeKey " + base64.StdEncoding.EncodeToString([]byte("master-1")) + "\n"))
			Expect(fileWriter.fileContents).To(ContainSubstring("master-1.myheadlessservice.mynamespace.svc.cluster.local FakeKey " + base64.StdEncoding.EncodeToString([]byte("master-1.myheadlessservice.mynamespace.svc.cluster.local")) + "\n"))
			Expect(fileWriter.fileContents).To(ContainSubstring("segment-a-0 FakeKey " + base64.StdEncoding.EncodeToString([]byte("segment-a-0")) + "\n"))
			Expect(fileWriter.fileContents).To(ContainSubstring("segment-a-0.myheadlessservice.mynamespace.svc.cluster.local FakeKey " + base64.StdEncoding.EncodeToString([]byte("segment-a-0.myheadlessservice.mynamespace.svc.cluster.local")) + "\n"))
			Expect(fileWriter.fileContents).To(ContainSubstring("segment-b-0 FakeKey " + base64.StdEncoding.EncodeToString([]byte("segment-b-0")) + "\n"))
			Expect(fileWriter.fileContents).To(ContainSubstring("segment-b-0.myheadlessservice.mynamespace.svc.cluster.local FakeKey " + base64.StdEncoding.EncodeToString([]byte("segment-b-0.myheadlessservice.mynamespace.svc.cluster.local")) + "\n"))
		})
	})
	It("writes to stderr upon key scan error", func() {
		keyScanner.Err = errors.New("mock error")
		status := app.Run()

		Expect(status).To(Equal(1))
		Expect(errBuffer).To(gbytes.Say("mock error\n"))
	})

	It("writes to stderr upon fileWriter error", func() {
		app.fileWriter = &fakeFailingFileWriter{}
		status := app.Run()

		Expect(status).To(Equal(1))
		Expect(errBuffer).To(gbytes.Say("failed to append known hosts to file: /home/gpadmin/.ssh/known_hosts: failed file writer\n"))
	})
})
