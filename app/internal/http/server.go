package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/app/internal/docker"
	"gh.tarampamp.am/indocker-app/app/internal/http/api"
	"gh.tarampamp.am/indocker-app/app/internal/http/middleware"
	"gh.tarampamp.am/indocker-app/app/internal/http/proxy"
	"gh.tarampamp.am/indocker-app/app/internal/http/ws"
	"gh.tarampamp.am/indocker-app/app/internal/version"
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
		baseCtx = func(ln net.Listener) context.Context { return ctx }
		server  = Server{
			log:   log,
			http:  &http.Server{BaseContext: baseCtx},                //nolint:gosec
			https: &http.Server{BaseContext: baseCtx, TLSConfig: tc}, //nolint:gosec
		}
	)

	for _, option := range options {
		option(&server)
	}

	return &server
}

func (s *Server) Register(
	ctx context.Context,
	drw docker.ContainersRouter,
	dsw docker.ContainersStateWatcher,
	proxyClientTimeout time.Duration,
) error {
	var router = NewRouter("/indocker", proxy.NewProxy(drw, proxyClientTimeout))

	router.Register(http.MethodGet, "/api/version/current", api.VersionCurrent(version.Version()))
	router.Register(http.MethodGet, "/api/version/latest", api.VersionLatest(func() (*version.LatestVersion, error) {
		return version.GetLatestVersion(ctx, &http.Client{
			Timeout: time.Second * 30, //nolint:gomnd
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		})
	}, time.Minute*30)) //nolint:gomnd
	router.Register(http.MethodGet, "/ws/docker/state", ws.DockerState(dsw))

	for server, namedLogger := range map[*http.Server]*zap.Logger{
		s.http:  s.log.Named("http"),
		s.https: s.log.Named("https"),
	} {
		server.ErrorLog = zap.NewStdLog(namedLogger)       // replace the default logger with named
		server.Handler = middleware.HealthcheckMiddleware( // healthcheck requests will not be logged
			middleware.LogReq(namedLogger, // named loggers for each server
				router,
			),
		)
	}

	return nil
}

// Start the server.
func (s *Server) Start(http, https net.Listener) error {
	if s.https.TLSConfig == nil || s.https.TLSConfig.Certificates == nil {
		return errors.New("HTTPS server: TLS config was not set")
	}

	var errCh = make(chan error, 2) //nolint:gomnd

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
