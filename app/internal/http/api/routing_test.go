package api_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/app/internal/http/api"
)

var okHandler = api.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusOK)

	return nil
})

var errHandler = api.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
	return errors.New("some error")
})

func TestRouter_ServeHTTP(t *testing.T) {
	t.Parallel()

	for name, testCase := range map[string]struct {
		givePrefix   string
		giveRoutes   map[[2]string]api.Handler
		giveFallback http.Handler
		giveRequest  *http.Request
		checkResult  func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		"empty": {
			giveRequest: httptest.NewRequest(http.MethodGet, "https://unit:123/foo/bar", http.NoBody),
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
				assertNoCorsHeaders(t, rr)
				assert.JSONEq(t, `{"error":"not found"}`, rr.Body.String())
			},
		},
		"with prefix only": {
			givePrefix:  "/prefix",
			giveRequest: httptest.NewRequest(http.MethodGet, "https://unit/foo/bar", http.NoBody),
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
				assertNoCorsHeaders(t, rr)
				assert.JSONEq(t, `{"error":"fallback handler is not set"}`, rr.Body.String())
			},
		},
		"with fallback only": {
			giveFallback: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			}),
			giveRequest: httptest.NewRequest(http.MethodGet, "https://unit/foo/bar", http.NoBody),
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
				assertNoCorsHeaders(t, rr)
				assert.JSONEq(t, `{"error":"not found"}`, rr.Body.String())
			},
		},
		"with fallback and prefix": {
			givePrefix: "/prefix",
			giveFallback: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			}),
			giveRequest: httptest.NewRequest(http.MethodGet, "https://unit/foo/bar", http.NoBody),
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusTeapot, rr.Code)
				assertNoCorsHeaders(t, rr)
				assert.Empty(t, rr.Body.String())
			},
		},
		"ok handler with prefix": {
			givePrefix:  "/prefix",
			giveRoutes:  map[[2]string]api.Handler{{http.MethodGet, "/foo"}: okHandler},
			giveRequest: httptest.NewRequest(http.MethodGet, "https://unit:123/prefix/foo", http.NoBody),
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)

				assert.Equal(t, "https://unit:123", rr.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Headers"))

				assert.Empty(t, rr.Body.String())
			},
		},
		"error handler with prefix": {
			givePrefix:  "/prefix",
			giveRoutes:  map[[2]string]api.Handler{{http.MethodGet, "/foo"}: errHandler},
			giveRequest: httptest.NewRequest(http.MethodGet, "http://unit:123/prefix/foo", http.NoBody),
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)

				assert.Equal(t, "http://unit:123", rr.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Headers"))
				assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

				assert.JSONEq(t, `{"error":"some error"}`, rr.Body.String())
			},
		},
		// TODO: add more tests
	} {
		tt := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var (
				router = api.NewRouter(tt.givePrefix, tt.giveFallback)
				rr     = httptest.NewRecorder()
			)

			for s, handler := range tt.giveRoutes {
				router.Register(s[0], s[1], handler)
			}

			router.ServeHTTP(rr, tt.giveRequest)

			tt.checkResult(t, rr)
		})
	}
}

func assertNoCorsHeaders(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rr.Header().Get("Access-Control-Allow-Methods"))
	assert.Empty(t, rr.Header().Get("Access-Control-Allow-Headers"))
}
