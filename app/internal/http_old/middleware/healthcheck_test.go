package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/app/internal/http_old/middleware"
)

func TestHealthcheckMiddleware(t *testing.T) {
	var (
		rr     = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, "http://test/healthz", http.NoBody)
	)

	req.Header.Set("User-Agent", "HealthChecker/indocker")

	middleware.HealthcheckMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler must not be called")
	})).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}

func TestHealthcheckMiddlewareSkipping(t *testing.T) {
	var (
		rr     = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, "http://test/healthz", http.NoBody)
	)

	// req.Header.Set("User-Agent", "HealthChecker/indocker") // no UA header

	middleware.HealthcheckMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusContinue)
	})).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusContinue, rr.Code)
}
