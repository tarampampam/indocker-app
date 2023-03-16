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

	"gh.tarampamp.am/indocker-app/app/internal/http/api"
	"gh.tarampamp.am/indocker-app/app/internal/version"
)

type mockVersionFetcher struct {
	calls int
	err   error
}

func (m *mockVersionFetcher) Fetch() (*version.Release, error) {
	m.calls++

	return &version.Release{
		Version:   "v1.2.3",
		URL:       "https://example.com",
		Name:      "Test release",
		Body:      "## Description\n\nThis is a test release",
		CreatedAt: time.Now(),
	}, m.err
}

func TestVersionLatest(t *testing.T) {
	t.Parallel()

	var (
		rr     = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, "http://test/foo", http.NoBody)

		fetcher = &mockVersionFetcher{}
		handler = api.VersionLatest(fetcher, time.Millisecond*10)
	)

	assert.Zero(t, fetcher.calls)
	assert.NoError(t, handler.Handle(rr, req))

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
	assert.Equal(t, "MISS", rr.Header().Get("X-Cache"))

	assert.JSONEq(t, `{
	"name":"Test release",
	"url":"https://example.com",
	"version":"v1.2.3",
	"body":"## Description\n\nThis is a test release",
	"created_at":"`+time.Now().Format(time.RFC3339)+`"
}`, rr.Body.String())
	assert.Equal(t, 1, fetcher.calls)

	assert.NoError(t, handler.Handle(rr, req))
	assert.Equal(t, "HIT", rr.Header().Get("X-Cache"))
	assert.Equal(t, 1, fetcher.calls) // not changed

	assert.NoError(t, handler.Handle(rr, req))
	assert.Equal(t, "HIT", rr.Header().Get("X-Cache"))
	assert.Equal(t, 1, fetcher.calls) // not changed

	<-time.After(time.Millisecond * 11) // invalidate cache

	assert.NoError(t, handler.Handle(rr, req))
	assert.Equal(t, "MISS", rr.Header().Get("X-Cache"))
	assert.Equal(t, 2, fetcher.calls) // changed

	assert.NoError(t, handler.Handle(rr, req))
	assert.Equal(t, "HIT", rr.Header().Get("X-Cache"))
	assert.Equal(t, 2, fetcher.calls) // not changed

	// work with errors
	<-time.After(time.Millisecond * 11) // invalidate cache

	fetcher.err = errors.New("test error") // SET the fetcher error

	rr = httptest.NewRecorder()
	assert.ErrorContains(t, handler.Handle(rr, req), "test error")
	assert.Equal(t, "MISS", rr.Header().Get("X-Cache"))
	assert.Equal(t, 3, fetcher.calls) // changed
	assert.Empty(t, rr.Body.String()) // nothing written

	rr = httptest.NewRecorder()
	assert.ErrorContains(t, handler.Handle(rr, req), "test error")
	assert.Equal(t, "MISS", rr.Header().Get("X-Cache"))
	assert.Equal(t, 4, fetcher.calls) // changed

	fetcher.err = nil // unset the fetcher error

	rr = httptest.NewRecorder()
	assert.NoError(t, handler.Handle(rr, req))
	assert.Equal(t, "MISS", rr.Header().Get("X-Cache"))
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 5, fetcher.calls) // changed
}

func TestVersionLatest_Concurrent(t *testing.T) {
	defer goleak.VerifyNone(t)

	var (
		fetcher = &mockVersionFetcher{}
		handler = api.VersionLatest(fetcher, time.Millisecond*10)
	)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			assert.NoError(t, handler.Handle(
				httptest.NewRecorder(),
				httptest.NewRequest(http.MethodGet, "http://test/foo", http.NoBody)),
			)
		}()
	}

	wg.Wait()

	assert.Equal(t, 1, fetcher.calls)
}
