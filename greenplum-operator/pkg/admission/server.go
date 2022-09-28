package admission

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
)

type Server interface {
	Start(stopCh <-chan struct{}, cert tls.Certificate, addr string, handler http.Handler) error
	Shutdown() error
}

type tlsServer struct {
	http.Server
}

var _ Server = &tlsServer{}

func NewTLSServer() Server {
	return &tlsServer{}
}

func (srv *tlsServer) Start(stopCh <-chan struct{}, cert tls.Certificate, addr string, handler http.Handler) error {
	srv.Addr = addr
	srv.Handler = handler
	srv.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	go func() {
		<-stopCh
		if err := srv.Shutdown(); err != nil {
			log.Fatalf("error shutting down admission webhook server: %s", err.Error())
		}
	}()

	return srv.Server.ListenAndServeTLS("", "")
}

func (srv *tlsServer) Shutdown() error {
	ctx := context.Background()
	return srv.Server.Shutdown(ctx)
}
