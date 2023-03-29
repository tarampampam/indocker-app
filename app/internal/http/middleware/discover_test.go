package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/app/internal/http/middleware"
)

func TestDiscoverMiddleware(t *testing.T) {
	var (
		rr     = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodTrace, "http://test/discover", http.NoBody)
	)

	req.Header.Set("X-InDocker", "true")

	middleware.DiscoverMiddleware("foobar", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler must not be called")
	})).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"base_url":"http://foobar.indocker.app"}`, rr.Body.String())
}
