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
	"gh.tarampamp.am/indocker-app/app/internal/http/fileserver"
	"gh.tarampamp.am/indocker-app/app/internal/http/middleware"
	"gh.tarampamp.am/indocker-app/app/internal/http/proxy"
	"gh.tarampamp.am/indocker-app/app/internal/httptools"
	ver "gh.tarampamp.am/indocker-app/app/internal/version"
	"gh.tarampamp.am/indocker-app/app/web"
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
		baseCtx   = func(ln net.Listener) context.Context { return ctx }
		logBridge = zap.NewStdLog(log)
		server    = Server{
			log:   log,
			http:  &http.Server{BaseContext: baseCtx, ErrorLog: logBridge},                //nolint:gosec
			https: &http.Server{BaseContext: baseCtx, ErrorLog: logBridge, TLSConfig: tc}, //nolint:gosec
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
	dashboardDomain string,
	dontUseEmbeddedFront bool,
) error {
	var (
		proxyHandler = proxy.NewProxy(s.log, drw)
		router       *api.Router
	)

	if dashboardDomain != "" {
		var fallback http.Handler

		if dontUseEmbeddedFront {
			fallback = proxyHandler // if the embedded front is disabled, proxy the request
		} else {
			fallback = fileserver.NewHandler(http.FS(web.Content())) // otherwise, serve the embedded front
		}

		router = api.NewRouter("/api", fallback)

		const latestVerCacheTTL = time.Minute * 30

		router.
			// Register(http.MethodGet, "/ws/docker/state", ws.DockerState(dsw)) // TODO: under construction
			Register(http.MethodGet, "/version/current", api.VersionCurrent(ver.Version())).
			Register(http.MethodGet, "/version/latest", api.VersionLatest(ver.NewLatest(ver.WithContext(ctx)), latestVerCacheTTL)) //nolint:lll
	}

	for server, namedLogger := range map[*http.Server]*zap.Logger{
		s.http:  s.log.Named("http"),
		s.https: s.log.Named("https"),
	} {
		server.ErrorLog = zap.NewStdLog(namedLogger)       // replace the default logger with named
		server.Handler = middleware.HealthcheckMiddleware( // healthcheck requests will not be logged
			middleware.DiscoverMiddleware(dashboardDomain,
				middleware.LogReq(namedLogger, // named loggers for each server
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if dashboardDomain != "" && httptools.TrimHostPortSuffix(r.Host) == dashboardDomain && router != nil {
							router.ServeHTTP(w, r)

							return
						}

						proxyHandler.ServeHTTP(w, r) // otherwise, proxy the request
					}),
				),
			),
		)
	}

	return nil
}

// Start the server(s).
func (s *Server) Start(http, https net.Listener) error {
	if s.https.TLSConfig == nil || s.https.TLSConfig.Certificates == nil {
		return errors.New("HTTPS server: TLS config was not set")
	}

	var errCh = make(chan error)

	go func() { errCh <- s.http.Serve(http) }()
	go func() { errCh <- s.https.ServeTLS(https, "", "") }()

	if err := <-errCh; err != nil {
		defer func() { <-errCh; close(errCh) }()

		return err
	}

	defer close(errCh)

	return <-errCh
}

// Stop the server(s).
func (s *Server) Stop(ctx context.Context) error {
	if err := s.http.Shutdown(ctx); err != nil {
		defer func() { _ = s.https.Shutdown(ctx) }()

		return err
	}

	return s.https.Shutdown(ctx)
}
