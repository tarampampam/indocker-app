package api_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"gh.tarampamp.am/indocker-app/daemon/internal/http/api"
	"gh.tarampamp.am/indocker-app/daemon/internal/version"
)

func TestVersionLatest(t *testing.T) {
	t.Parallel()

	var (
		rr     = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, "http://test/foo", http.NoBody)

		fetched    = 0
		now        = time.Now()
		fetcherErr error
	)

	var handler = api.VersionLatest(func() (*version.LatestVersion, error) {
		fetched++

		return &version.LatestVersion{
			Version:   "v1.2.3",
			URL:       "https://example.com",
			Name:      "Test release",
			Body:      "## Description\n\nThis is a test release",
			CreatedAt: now,
		}, fetcherErr
	}, time.Millisecond*10)

	assert.Zero(t, fetched)
	handler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

	assert.JSONEq(t, `{
	"name":"Test release",
	"url":"https://example.com",
	"version":"v1.2.3",
	"body":"## Description\n\nThis is a test release",
	"created_at":"`+now.Format(time.RFC3339)+`"
}`, rr.Body.String())

	assert.Equal(t, 1, fetched)
	handler(rr, req)
	assert.Equal(t, 1, fetched) // not changed
	handler(rr, req)
	assert.Equal(t, 1, fetched) // not changed

	<-time.After(time.Millisecond * 11)

	handler(rr, req)
	assert.Equal(t, 2, fetched) // changed
	handler(rr, req)
	assert.Equal(t, 2, fetched) // not changed

	// work with errors
	<-time.After(time.Millisecond * 11)

	fetcherErr = errors.New("test error") // SET the fetcher error

	rr = httptest.NewRecorder()
	handler(rr, req)

	assert.Equal(t, 3, fetched) // changed
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.JSONEq(t, `{"error":"test error"}`, rr.Body.String())

	rr = httptest.NewRecorder()
	handler(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	assert.Equal(t, 4, fetched) // changed

	rr = httptest.NewRecorder()
	handler(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	assert.Equal(t, 5, fetched) // changed

	// unset the fetcher error
	fetcherErr = nil

	rr = httptest.NewRecorder()
	handler(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 6, fetched) // changed
}

func TestVersionLatest_Concurrent(t *testing.T) {
	defer goleak.VerifyNone(t)

	var (
		fetched = 0
		now     = time.Now()
	)

	var handler = api.VersionLatest(func() (*version.LatestVersion, error) {
		fetched++

		return &version.LatestVersion{
			Version:   "v1.2.3",
			URL:       "https://example.com",
			Name:      "Test release",
			Body:      "## Description\n\nThis is a test release",
			CreatedAt: now,
		}, nil
	}, time.Millisecond*10)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			var (
				rr     = httptest.NewRecorder()
				req, _ = http.NewRequest(http.MethodGet, "http://test/foo", http.NoBody)
			)

			handler(rr, req)
		}()
	}

	wg.Wait()

	assert.Equal(t, 1, fetched)
}
