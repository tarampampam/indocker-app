package middleware_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/app/internal/http_old/middleware"
)

func TestDiscoverMiddlewareCORS(t *testing.T) {
	t.Parallel()

	var (
		mw = middleware.DiscoverMiddleware("foobar", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler must not be called")
		}))

		rr     = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodOptions, "http://test/x/indocker/discover", http.NoBody)
	)

	mw.ServeHTTP(rr, req) // CORS request

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Empty(t, rr.Body.String())
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "http://foobar.indocker.app", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET", rr.Header().Get("Access-Control-Allow-Methods"))
}

func TestDiscoverMiddlewareRealRequestWithTLS(t *testing.T) {
	t.Parallel()

	var (
		mw = middleware.DiscoverMiddleware("foobar", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler must not be called")
		}))

		rr     = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, "https://test/x/indocker/discover", http.NoBody)
	)

	req.TLS = &tls.ConnectionState{}
	req.Header.Set("X-InDocker", "True")

	mw.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
	assert.JSONEq(t, `{"api":{"base_url":"https://foobar.indocker.app/api"}}`, rr.Body.String())
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "https://foobar.indocker.app", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET", rr.Header().Get("Access-Control-Allow-Methods"))
}

func TestDiscoverMiddlewareWrongMethod(t *testing.T) {
	t.Parallel()

	var (
		mw = middleware.DiscoverMiddleware("foobar", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler must not be called")
		}))

		rr     = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodPost, "http://test/x/indocker/discover", http.NoBody)
	)

	mw.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestDiscoverMiddlewareWithoutHeader(t *testing.T) {
	t.Parallel()

	var (
		mw = middleware.DiscoverMiddleware("foobar", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler must not be called")
		}))

		rr     = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, "http://test/x/indocker/discover", http.NoBody)
	)

	mw.ServeHTTP(rr, req) // without required header

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
