package http

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/app/internal/http/middleware/frontend"
	"gh.tarampamp.am/indocker-app/app/internal/http/middleware/logreq"
	"gh.tarampamp.am/indocker-app/app/internal/http/openapi"
	"gh.tarampamp.am/indocker-app/app/web"
)

type Server struct {
	http  *http.Server
	https *http.Server

	ShutdownTimeout time.Duration // Maximum amount of time to wait for the server to stop, default is 5 seconds
}

type ServerOption func(*Server)

func WithReadTimeout(d time.Duration) ServerOption {
	return func(s *Server) { s.http.ReadTimeout = d; s.https.ReadTimeout = d }
}

func WithWriteTimeout(d time.Duration) ServerOption {
	return func(s *Server) { s.http.WriteTimeout = d; s.https.WriteTimeout = d }
}

func WithIDLETimeout(d time.Duration) ServerOption {
	return func(s *Server) { s.http.IdleTimeout = d; s.https.IdleTimeout = d }
}

func NewServer(baseCtx context.Context, log *zap.Logger, opts ...ServerOption) Server {
	var (
		server = Server{
			http: &http.Server{ //nolint:gosec
				BaseContext: func(net.Listener) context.Context { return baseCtx },
				ErrorLog:    zap.NewStdLog(log.Named("http")),
			},
			https: &http.Server{ //nolint:gosec
				BaseContext: func(net.Listener) context.Context { return baseCtx },
				ErrorLog:    zap.NewStdLog(log.Named("https")),
			},
			ShutdownTimeout: 5 * time.Second, //nolint:mnd
		}
	)

	for _, opt := range opts {
		opt(&server)
	}

	return server
}

func (s *Server) Register(ctx context.Context, log *zap.Logger) {
	// since both servers uses the same logics, we can iterate over them, but with differently named loggers
	for namedLog, srv := range map[*zap.Logger]*http.Server{
		log.Named("http"):  s.http,
		log.Named("https"): s.https,
	} {
		var (
			// create openapi server implementation (it is used only for the monitor subdomain)
			openapiServer = NewOpenAPI(ctx, namedLog)

			// create the base router for the openapi server
			openapiMux = http.NewServeMux()

			// "convert" the openapi server to the [http.Handler]
			openapiHandler = openapi.HandlerWithOptions(openapiServer, openapi.StdHTTPServerOptions{
				ErrorHandlerFunc: openapiServer.HandleInternalError,
				BaseRouter:       openapiMux,
				Middlewares:      []openapi.MiddlewareFunc{openapi.CorsMiddleware()},
			})
		)

		// note that since a pattern ending in a slash names a rooted subtree, the pattern "/" matches all paths not
		// matched by other registered patterns, not just the URL with Path == "/". this allows us to use this pattern
		// as "catch-all" for all requests that are not handled by the openapi server (including static assets,
		// 404 errors, etc.).
		//
		// and one more important thing - to serve the frontend (located in the web directory) we use the frontend
		// middleware, which is a simple wrapper around the http.FileServer. it allows us to serve the frontend only
		// for the requests that are not intended for the API (i.e. requests that do not start with "/api").
		openapiMux.Handle("/", frontend.New(web.Dist(), func(r *http.Request) bool {
			return strings.HasPrefix(r.URL.Path, "/api") // skip the middleware, if the request is intended for the API
		})(http.HandlerFunc(openapiServer.HandleNotFoundError))) // <-- this is the general 404 handler

		// wrap the server handler with middleware
		srv.Handler = logreq.New(namedLog, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// extract the host from the request
			if h, _, err := net.SplitHostPort(r.Host); err == nil {
				// and check if it's a subdomain of the monitor
				if strings.ToLower(h) == "monitor.indocker.app" {
					// then serve the openapi handler
					openapiHandler.ServeHTTP(w, r)

					return
				}

				// otherwise, serve the proxy handler
				http.NotFound(w, r) // TODO: proxy handler
			}
		}))
	}
}

// StartHTTP starts the HTTP server. It listens on the provided listener and serves incoming requests.
// To stop the server, cancel the provided context.
//
// It blocks until the context is canceled or the server is stopped by some error.
func (s *Server) StartHTTP(ctx context.Context, ln net.Listener) error {
	var errCh = make(chan error)

	go func(ch chan<- error) { defer close(ch); ch <- s.http.Serve(ln) }(errCh)

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.ShutdownTimeout)
		defer cancel()

		if err := s.http.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case err, isOpened := <-errCh:
		switch {
		case !isOpened:
			return nil
		case err != nil:
			return err
		}
	}

	return nil
}

// StartHTTPs starts the HTTPS server. It listens on the provided listener and serves incoming requests.
// To stop the server, cancel the provided context.
//
// It blocks until the context is canceled or the server is stopped by some error.
func (s *Server) StartHTTPs(ctx context.Context, ln net.Listener, certFile, keyFile []byte) error {
	if s.https.TLSConfig == nil {
		s.https.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	if len(s.https.TLSConfig.Certificates) == 0 {
		cert, certErr := tls.X509KeyPair(certFile, keyFile)
		if certErr != nil {
			return fmt.Errorf("failed to load TLS certificate: %w", certErr)
		}

		s.https.TLSConfig.Certificates = []tls.Certificate{cert}
	}

	var errCh = make(chan error)

	go func(ch chan<- error) { defer close(ch); ch <- s.https.ServeTLS(ln, "", "") }(errCh)

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.ShutdownTimeout)
		defer cancel()

		if err := s.https.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case err, isOpened := <-errCh:
		switch {
		case !isOpened:
			return nil
		case err != nil:
			return err
		}
	}

	return nil
}
