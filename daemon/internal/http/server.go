package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/daemon/internal/docker"
	"gh.tarampamp.am/indocker-app/daemon/internal/http/api"
	"gh.tarampamp.am/indocker-app/daemon/internal/http/middleware"
	"gh.tarampamp.am/indocker-app/daemon/internal/http/proxy"
	"gh.tarampamp.am/indocker-app/daemon/internal/version"
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
	var proxyHandler = proxy.NewProxy(docker)

	var apiOrigins = map[string]struct{}{
		"indocker.app":          {},
		"frontend.indocker.app": {}, // for local development
	}

	var (
		apiVer         = api.Version(version.Version())
		api404         = api.NotFound()
		apiDockerState = api.NewDockerState(docker)
	)

	for server, logger := range map[*http.Server]*zap.Logger{
		s.http:  s.log.Named("http"),
		s.https: s.log.Named("https"),
	} {
		server.Handler = middleware.HealthcheckMiddleware( // healthcheck requests will not be logged
			middleware.LogReq(logger, // named loggers for each server
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // API handlers wrapper
					const apiPrefix = "/api"

					// r.Host can be "localhost:8080" or "localhost"
					if hostPort := strings.Split(r.Host, ":"); len(hostPort) > 0 {
						var origin = hostPort[0]

						// check if the origin is allowed and the request is for the API
						if _, ok := apiOrigins[origin]; ok && strings.HasPrefix(r.URL.Path, apiPrefix+"/") {
							s.corsHeaders(w, origin)

							switch u, m := r.URL.Path, r.Method; { // the fastest multiplexer in the world :D
							case m == http.MethodGet && u == apiPrefix+"/version":
								apiVer.ServeHTTP(w, r)

							case m == http.MethodGet && u == apiPrefix+"/docker/state":
								apiDockerState.ServeHTTP(w, r)

							default:
								api404.ServeHTTP(w, r)
							}

							return
						}
					}

					proxyHandler.ServeHTTP(w, r)
				}),
			),
		)
	}

	return nil
}

func (s *Server) corsHeaders(w http.ResponseWriter, origin string) {
	w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("https://%s", origin))
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
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
