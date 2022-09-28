package knownhosts_test

import (
	"errors"
	"os"
	"strings"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	fakessh "github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/fake"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/knownhosts"
	"golang.org/x/crypto/ssh"
)

var _ = Describe("Reader", func() {
	var (
		memoryfs vfs.Filesystem
	)

	BeforeEach(func() {
		memoryfs = memfs.Create()
		Expect(vfs.MkdirAll(memoryfs, "/home/gpadmin/.ssh/", 755)).To(Succeed())
	})

	When("GetKnownHosts() is called", func() {
		var (
			app              *knownhosts.Reader
			knownHosts       map[string]ssh.PublicKey
			getKnownHostsErr error
		)
		BeforeEach(func() {
			var knownHostsFileContents strings.Builder

			knownHostsFileContents.WriteString("# Example comment line")
			knownHostsFileContents.WriteRune('\n')

			knownHostsFileContents.WriteString("127.0.0.1 ")
			knownHostsFileContents.WriteString(fakessh.ExamplePublicKey)
			knownHostsFileContents.WriteRune('\n')

			knownHostsFileContents.WriteString("examplehost.com ")
			knownHostsFileContents.WriteString(fakessh.ExamplePublicKey)

			Expect(vfs.WriteFile(memoryfs, "/home/gpadmin/.ssh/known_hosts", []byte(knownHostsFileContents.String()), 644)).To(Succeed())
			app = &knownhosts.Reader{Fs: memoryfs}
		})
		JustBeforeEach(func() {
			knownHosts, getKnownHostsErr = app.GetKnownHosts()
		})
		It("returns the known hosts", func() {
			expectedPublicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(fakessh.ExamplePublicKey))
			Expect(err).NotTo(HaveOccurred())

			Expect(getKnownHostsErr).NotTo(HaveOccurred())
			Expect(knownHosts).To(HaveKey("127.0.0.1"))
			Expect(knownHosts["127.0.0.1"].Type()).To(Equal("ssh-rsa"))
			Expect(knownHosts["127.0.0.1"].Marshal()).To(Equal(expectedPublicKey.Marshal()))

			Expect(knownHosts).To(HaveKey("examplehost.com"))
			Expect(knownHosts["examplehost.com"].Type()).To(Equal("ssh-rsa"))
			Expect(knownHosts["examplehost.com"].Type()).To(Equal("ssh-rsa"))
			Expect(knownHosts["examplehost.com"].Marshal()).To(Equal(expectedPublicKey.Marshal()))
		})
		When("the known_hosts file does not exist", func() {
			BeforeEach(func() {
				memoryfs = memfs.Create()
				app.Fs = memoryfs
			})
			It("returns an empty map", func() {
				Expect(getKnownHostsErr).NotTo(HaveOccurred())
				Expect(knownHosts).To(HaveLen(0))
			})
		})
		When("the known_hosts file is empty", func() {
			BeforeEach(func() {
				Expect(vfs.WriteFile(memoryfs, "/home/gpadmin/.ssh/known_hosts", []byte(""), 644)).To(Succeed())
			})
			It("returns an empty map", func() {
				Expect(getKnownHostsErr).NotTo(HaveOccurred())
				Expect(knownHosts).To(HaveLen(0))
			})
		})
		When("a bad public key exists in the known_hosts file", func() {
			BeforeEach(func() {
				Expect(vfs.WriteFile(memoryfs, "/home/gpadmin/.ssh/known_hosts", []byte("bad public key"), 644)).To(Succeed())
			})
			It("returns an error", func() {
				Expect(getKnownHostsErr).To(HaveOccurred())
				Expect(knownHosts).To(BeNil())
				Expect(getKnownHostsErr).To(MatchError("could not parse /home/gpadmin/.ssh/known_hosts: illegal base64 data at input byte 0"))
			})
		})
		When("a public key includes a marker", func() {
			BeforeEach(func() {
				Expect(vfs.WriteFile(memoryfs, "/home/gpadmin/.ssh/known_hosts", []byte("@revoked revoked.example.com "+fakessh.ExamplePublicKey), 644)).To(Succeed())
			})
			It("returns an error", func() {
				Expect(getKnownHostsErr).To(HaveOccurred())
				Expect(knownHosts).To(BeNil())
				Expect(getKnownHostsErr).To(MatchError("known_hosts markers are not currently supported"))
			})
		})

		When("the OS can't read known_hosts", func() {
			BeforeEach(func() {
				fakeFS := fileutil.HookableFilesystem{Filesystem: memoryfs}
				app.Fs = &fakeFS
				fakeFS.OpenFileHook = func(name string, flag int, perm os.FileMode) (vfs.File, error) {
					return nil, errors.New("fake open file error")
				}
			})
			It("returns an error", func() {
				Expect(getKnownHostsErr).To(HaveOccurred())
				Expect(knownHosts).To(BeNil())
				Expect(getKnownHostsErr).To(MatchError("could not read /home/gpadmin/.ssh/known_hosts: fake open file error"))
			})
		})
	})

	When("GetHostPublicKey() is called", func() {
		var (
			fakeKnownHostsReader *fakessh.KnownHostsReader
		)
		BeforeEach(func() {
			fakeKnownHostsReader = &fakessh.KnownHostsReader{}
			fakeKnownHostsReader.KnownHosts = map[string]ssh.PublicKey{
				"test-host-1": fakessh.KeyForHost("test-host-1"),
				"test-host-2": fakessh.KeyForHost("test-host-2"),
			}
		})
		It("returns the public key associated with the host", func() {
			key1, err := knownhosts.GetHostPublicKey(fakeKnownHostsReader, "test-host-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(key1).To(Equal(fakessh.KeyForHost("test-host-1")))
			key2, err := knownhosts.GetHostPublicKey(fakeKnownHostsReader, "test-host-2")
			Expect(err).NotTo(HaveOccurred())
			Expect(key2).To(Equal(fakessh.KeyForHost("test-host-2")))
		})
		When("there is an error getting the known hosts", func() {
			BeforeEach(func() {
				fakeKnownHostsReader.Err = errors.New("failed to read hosts")
			})
			It("returns an error", func() {
				_, err := knownhosts.GetHostPublicKey(fakeKnownHostsReader, "test-host-1")
				Expect(err).To(MatchError("failed to read hosts"))
			})
		})
		When("the host is unknown", func() {
			It("returns an error", func() {
				_, err := knownhosts.GetHostPublicKey(fakeKnownHostsReader, "test-host-3")
				Expect(err).To(MatchError("host test-host-3 was not found in the known_hosts file"))
			})
		})
	})
})
