package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/daemon/internal/http/api"
)

func TestVersionCurrent(t *testing.T) {
	t.Parallel()

	var (
		rr     = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, "http://test/foo", http.NoBody)
	)

	api.VersionCurrent("v1.2.3")(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
	assert.JSONEq(t, `{"version":"v1.2.3"}`, rr.Body.String())
}
