package admission_test

import (
	"crypto/tls"
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/admission"
)

var _ = Describe("Server", func() {
	const srvAddr = "localhost:11443"
	var (
		subject admission.Server
		ah      admission.Handler
	)

	BeforeEach(func() {
		subject = admission.NewTLSServer()
	})

	When("server is loaded with valid keys", func() {
		var (
			cert tls.Certificate
		)
		BeforeEach(func() {
			cmd := exec.Command("openssl", "req", "-new", "-nodes", "-x509",
				"-out", "server.pem",
				"-keyout", "server.key",
				"-days", "3650",
				"-subj", "/C=DE/ST=NRW/L=Earth/O=Test Company/OU=IT/CN=www.test.com/emailAddress=test@test.com")
			Expect(cmd.Run()).To(Succeed())

			var err error
			cert, err = tls.LoadX509KeyPair("server.pem", "server.key")
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			cmd := exec.Command("rm", "server.pem", "server.key")
			Expect(cmd.Run()).To(Succeed())
		})

		It("Starts/Stops a TLS server successfully", func() {
			doneCh := make(chan struct{})
			stopCh := make(chan struct{})

			go func() {
				err := subject.Start(stopCh, cert, srvAddr, ah.Handler())
				Expect(err).To(Equal(http.ErrServerClosed))
				close(doneCh)
			}()

			Eventually(TLSConnectionCheck(srvAddr)).Should(Succeed())

			close(stopCh)
			Eventually(doneCh).Should(BeClosed())
		})
	})

	When("server fails to start", func() {
		var doneCh, stopCh, startCh chan struct{}
		BeforeEach(func() {
			doneCh = make(chan struct{})
			stopCh = make(chan struct{})
			startCh = make(chan struct{})
			go func() {
				err := subject.Start(stopCh, tls.Certificate{}, srvAddr, ah.Handler())
				Expect(err).To(Equal(http.ErrServerClosed))
				close(doneCh)
			}()

		})

		It("does not close the startCh", func() {
			close(stopCh)
			Eventually(doneCh).Should(BeClosed())
			Eventually(startCh).ShouldNot(BeClosed())
		})
	})
})

func TLSConnectionCheck(addr string) func() error {
	return func() error {
		cfg := &tls.Config{
			InsecureSkipVerify: true,
		}
		conn, err := tls.Dial("tcp", addr, cfg)
		if conn != nil {
			return conn.Close()
		}
		return err
	}
}
