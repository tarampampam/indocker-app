package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/daemon/internal/http/middleware"
)

func TestCors(t *testing.T) {
	var (
		rr     = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, "http://test:123/healthz", http.NoBody)
	)

	middleware.Cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusContinue)
	})).ServeHTTP(rr, req)

	assert.Equal(t, "http://test:123", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Headers"))
}
