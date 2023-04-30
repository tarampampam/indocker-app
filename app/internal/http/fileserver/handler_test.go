package fileserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/app/internal/http/fileserver"
)

func TestHandler(t *testing.T) {
	var (
		root = http.FS(fstest.MapFS{
			"index.html": {
				Data: []byte("<html><body>index</body></html>"),
			},
			"404.html": {
				Data: []byte("<html><body>OLOLO 404</body></html>"),
			},
		})
	)

	for name, testCase := range map[string]struct {
		giveUrl               string
		giveMethod            string
		wantCode              int
		wantInBody            string
		wantEmptyResponseBody bool
	}{
		"root": {
			giveUrl:    "/",
			giveMethod: http.MethodGet,
			wantCode:   http.StatusOK,
			wantInBody: "<body>index</body>",
		},
		"root (head)": {
			giveUrl:               "/",
			giveMethod:            http.MethodHead,
			wantCode:              http.StatusOK,
			wantEmptyResponseBody: true,
		},
		"index": {
			giveUrl:               "/index.html",
			giveMethod:            http.MethodGet,
			wantCode:              http.StatusMovedPermanently,
			wantEmptyResponseBody: true,
		},
		"not found": {
			giveUrl:    "/foo",
			giveMethod: http.MethodGet,
			wantCode:   http.StatusOK,
			wantInBody: "<html><body>index</body></html>",
		},
		"not found (head)": {
			giveUrl:               "/foo",
			giveMethod:            http.MethodHead,
			wantCode:              http.StatusNotFound,
			wantEmptyResponseBody: true,
		},
		// TODO: add cases with GZipped files
	} {
		tt := testCase

		t.Run(name, func(t *testing.T) {
			var (
				req = httptest.NewRequest(tt.giveMethod, tt.giveUrl, http.NoBody)
				rec = httptest.NewRecorder()
			)

			fileserver.NewHandler(root).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantCode, rec.Code)

			if tt.wantEmptyResponseBody {
				assert.Empty(t, rec.Body.String())
			} else {
				assert.Contains(t, rec.Body.String(), tt.wantInBody)
			}
		})
	}
}
