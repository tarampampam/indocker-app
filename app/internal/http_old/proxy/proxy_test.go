package proxy_test

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/app/internal/docker"
	"gh.tarampamp.am/indocker-app/app/internal/http_old/proxy"
)

type mockDockerRouter struct {
	routes map[string]docker.Route
	err    error
}

func (m *mockDockerRouter) RouteToContainerByHostname(hostname string) (*docker.Route, error) {
	if m.err != nil {
		return nil, m.err
	}

	if len(m.routes) == 0 {
		return nil, docker.ErrNoRegisteredRoutes
	}

	if route, ok := m.routes[hostname]; ok {
		return &route, nil
	}

	return nil, docker.ErrNoRouteFound
}

func (m *mockDockerRouter) Routes() map[string]docker.Route {
	return m.routes
}

func TestProxy_ServeHTTP(t *testing.T) {
	defer goleak.VerifyNone(t)

	var (
		mux    = http.NewServeMux()
		server = &http.Server{Handler: mux} //nolint:gosec // create a test server
	)

	// register test handlers
	mux.Handle("/ping", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))
	mux.Handle("/nested/path/to/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete { //nolint:nestif
			if r.URL.Query().Get("foo") == "bar" {
				if r.Header.Get("X-Foo") == "x-bar" {
					body, _ := io.ReadAll(r.Body)

					if string(body) == "secret payload" {
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte("pong"))

						return
					}
				}
			}
		}

		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	mux.Handle("/error", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("foo error 123"))
	}))
	mux.Handle("/flusher", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("foo"))

		w.(http.Flusher).Flush()
		_, _ = w.Write([]byte("_bar"))

		delay := time.NewTimer(time.Millisecond * 5)
		<-delay.C
		delay.Stop()

		w.(http.Flusher).Flush()
		_, _ = w.Write([]byte("_baz"))
	}))

	l, port := newListener(t)
	defer func() { _ = l.Close() }()

	// start the server in a goroutine
	go func() { assert.ErrorIs(t, server.Serve(l), http.ErrServerClosed) }()

	// shutdown the server at the end of the test
	defer func() { assert.NoError(t, server.Shutdown(context.Background())) }()

	var serverRoute = docker.Route{
		Scheme: "http",
		Port:   port,
		IPAddr: "127.0.0.1",
	}

	for name, testCase := range map[string]struct {
		giveRoutes    map[string]docker.Route
		giveRouterErr error
		giveRequest   func() *http.Request
		checkResponse func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		"route not found": {
			giveRoutes: map[string]docker.Route{"foo": serverRoute},
			giveRequest: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "https://bar.indocker.app/", http.NoBody)

				return r
			},
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)

				assert.Contains(t, rr.Body.String(), "<html")
				assert.Contains(t, rr.Body.String(), "<head>")
				assert.Contains(t, rr.Body.String(), "<body>")
				assert.Contains(t, rr.Body.String(), "404")
				assert.Contains(t, rr.Body.String(), "No route found")
				assert.Contains(t, rr.Body.String(), "</body>")
				assert.Contains(t, rr.Body.String(), "</html>")
			},
		},
		"no routes": {
			giveRequest: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "https://foo.indocker.app/", http.NoBody)

				return r
			},
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

				assert.Contains(t, rr.Body.String(), "<html")
				assert.Contains(t, rr.Body.String(), "<head>")
				assert.Contains(t, rr.Body.String(), "<body>")
				assert.Contains(t, rr.Body.String(), "422")
				assert.Contains(t, rr.Body.String(), "No registered routes")
				assert.Contains(t, rr.Body.String(), "</body>")
				assert.Contains(t, rr.Body.String(), "</html>")
			},
		},
		"success": {
			giveRoutes: map[string]docker.Route{"foo": serverRoute},
			giveRequest: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "https://foo.indocker.app/ping", http.NoBody)

				return r
			},
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)

				assert.Equal(t, "pong", rr.Body.String())
			},
		},
		"success without domain postfix": {
			giveRoutes: map[string]docker.Route{"foo": serverRoute},
			giveRequest: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "https://foo/ping", http.NoBody)

				return r
			},
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)

				assert.Equal(t, "pong", rr.Body.String())
			},
		},
		"success with query and headers": {
			giveRoutes: map[string]docker.Route{"foo": serverRoute},
			giveRequest: func() *http.Request {
				r, _ := http.NewRequest(http.MethodDelete,
					"https://foo.indocker.app/nested/path/to/ping?foo=bar",
					bytes.NewBuffer([]byte("secret payload")),
				)
				r.Header.Set("X-Foo", "x-bar")

				return r
			},
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)

				assert.Equal(t, "pong", rr.Body.String())
			},
		},
		"error from backend": {
			giveRoutes: map[string]docker.Route{"foo": serverRoute},
			giveRequest: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "https://foo.indocker.app/error", http.NoBody)

				return r
			},
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)

				assert.Equal(t, "foo error 123", rr.Body.String())
			},
		},
		"flusher": {
			giveRoutes: map[string]docker.Route{"foo": serverRoute},
			giveRequest: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "https://foo.indocker.app/flusher", http.NoBody)

				return r
			},
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)

				assert.Equal(t, "foo_bar_baz", rr.Body.String())
			},
		},
	} {
		tt := testCase

		t.Run(name, func(t *testing.T) {
			var router = &mockDockerRouter{
				routes: tt.giveRoutes,
				err:    tt.giveRouterErr,
			}

			var (
				rr      = httptest.NewRecorder()
				handler = proxy.NewProxy(zap.NewNop(), router)
			)

			handler.ServeHTTP(rr, tt.giveRequest())

			tt.checkResponse(t, rr)
		})
	}
}

func newListener(t *testing.T) (net.Listener, uint16) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0") // let the OS choose a free port
	require.NoError(t, err)

	return l, uint16(l.Addr().(*net.TCPAddr).Port)
}
