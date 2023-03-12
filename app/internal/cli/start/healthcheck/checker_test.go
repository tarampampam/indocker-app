package healthcheck_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/app/internal/cli/start/healthcheck"
)

type httpClientFunc func(*http.Request) (*http.Response, error)

func (f httpClientFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestHealthChecker_CheckSuccess(t *testing.T) {
	t.Parallel()

	var httpMock httpClientFunc = func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case "http://127.0.0.1:80/healthz", "https://127.0.0.1:443/healthz":
			// ok

		default:
			t.Error("unexpected URL")
		}

		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "HealthChecker/indocker", req.Header.Get("User-Agent"))

		return &http.Response{
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
			StatusCode: http.StatusOK,
		}, nil
	}

	checker := healthcheck.NewHealthChecker(context.Background(), httpMock)

	assert.NoError(t, checker.Check(80, 443))
}

func TestHealthChecker_CheckFail(t *testing.T) {
	t.Parallel()

	var httpMock httpClientFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
			StatusCode: http.StatusBadGateway,
		}, nil
	}

	checker := healthcheck.NewHealthChecker(context.Background(), httpMock)

	err := checker.Check(80, 81)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wrong status code")
}