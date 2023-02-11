package docker_info_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"docker_info"
)

func TestPlugin_ServeHTTP(t *testing.T) {
	t.Parallel()

	var (
		req, _ = http.NewRequest(http.MethodGet, "http://test/docker-info/containers/list", http.NoBody)
		rr     = httptest.NewRecorder()
		next   = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) { rw.WriteHeader(http.StatusOK) })
	)

	plugin, err := docker_info.New(
		context.Background(),
		next,
		docker_info.CreateConfig(),
		"",
	)

	if err != nil {
		t.Fatal(err)
	}

	plugin.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	t.Log(rr.Body.String()) // TODO: for debugging, remove later
}
