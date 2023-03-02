package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/daemon/internal/docker"
	"gh.tarampamp.am/indocker-app/daemon/internal/http/handlers"
)

type Server struct {
	log   *zap.Logger
	http  *http.Server
	https *http.Server
}

type ServerOption func(*Server)

func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) { s.http.ReadTimeout = timeout; s.https.ReadTimeout = timeout }
}

func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) { s.http.WriteTimeout = timeout; s.https.WriteTimeout = timeout }
}

func WithIDLETimeout(timeout time.Duration) ServerOption {
	return func(s *Server) { s.http.IdleTimeout = timeout; s.https.IdleTimeout = timeout }
}

func NewServer(ctx context.Context, log *zap.Logger, tc *tls.Config, options ...ServerOption) *Server {
	var (
		stdLog  = zap.NewStdLog(log)
		baseCtx = func(ln net.Listener) context.Context { return ctx }
		server  = Server{
			log:   log,
			http:  &http.Server{ErrorLog: stdLog, BaseContext: baseCtx},
			https: &http.Server{ErrorLog: stdLog, BaseContext: baseCtx, TLSConfig: tc},
		}
	)

	for _, option := range options {
		option(&server)
	}

	return &server
}

func (s *Server) Register(docker *docker.Docker) error {
	var proxy = handlers.NewProxy(docker)

	// TODO redirect to https

	for _, server := range []*http.Server{s.http, s.https} {
		server.Handler = proxy
	}

	return nil
}

// Start the server.
func (s *Server) Start(http, https net.Listener) error {
	if s.https.TLSConfig == nil || s.https.TLSConfig.Certificates == nil {
		return errors.New("HTTPS server: TLS config was not set")
	}

	var errCh = make(chan error, 2)

	go func() { errCh <- s.http.Serve(http) }()
	go func() { errCh <- s.https.ServeTLS(https, "", "") }()

	return <-errCh
}

// Stop the server.
func (s *Server) Stop(ctx context.Context) error {
	if err := s.http.Shutdown(ctx); err != nil {
		return err
	}

	return s.https.Shutdown(ctx)
}
