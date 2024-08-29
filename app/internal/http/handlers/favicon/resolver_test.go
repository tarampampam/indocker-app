package favicon_test

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"image"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/app/internal/http/handlers/favicon"
)

var (
	//go:embed testdata/github.ico
	githubIco []byte

	githubIcoImage = func() image.Image {
		img, _, err := image.Decode(bytes.NewReader(githubIco))
		if err != nil {
			panic(err)
		}

		return img
	}()
)

type httpClientFunc func(*http.Request) (*http.Response, error)

func (f httpClientFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestResolver_Resolve(t *testing.T) {
	for name, tc := range map[string]struct {
		giveClient    httpClientFunc
		giveBaseURL   string
		wantImage     image.Image
		wantErrSubstr string
	}{
		"success, favicon.ico": {
			giveClient: func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "/favicon.ico", req.URL.String())

				return &http.Response{
					Body:       io.NopCloser(bytes.NewReader(githubIco)),
					StatusCode: http.StatusOK,
				}, nil
			},
			wantImage: githubIcoImage,
		},
		"success, from HTML (rel = icon)": {
			giveClient: func(req *http.Request) (*http.Response, error) {
				switch req.URL.String() {
				case "https://example.com/favicon.ico":
					return &http.Response{Body: http.NoBody, StatusCode: http.StatusNotFound}, nil
				case "https://example.com":
					return &http.Response{
						Body:       io.NopCloser(strings.NewReader(`<html><head><link rel="icon" href="another-favicon.png"></head></html>`)),
						StatusCode: http.StatusOK,
					}, nil
				case "https://example.com/another-favicon.png":
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader(githubIco)),
						StatusCode: http.StatusOK,
					}, nil
				default:
					return nil, fmt.Errorf("unexpected URL: %s", req.URL.String())
				}
			},
			giveBaseURL: "https://example.com",
			wantImage:   githubIcoImage,
		},
		"success, from HTML (rel = apple-touch-icon)": {
			giveClient: func(req *http.Request) (*http.Response, error) {
				switch req.URL.String() {
				case "https://example.com/favicon.ico":
					return &http.Response{Body: http.NoBody, StatusCode: http.StatusNotFound}, nil
				case "https://example.com":
					return &http.Response{
						Body: io.NopCloser(strings.NewReader(`<html><head>
<link rel="icon" href="favicon-404.png">
<link rel="icon" href="ok-favicon.png">
</head></html>`)),
						StatusCode: http.StatusOK,
					}, nil
				case "https://example.com/favicon-404.png":
					return &http.Response{Body: http.NoBody, StatusCode: http.StatusNotFound}, nil
				case "https://example.com/ok-favicon.png":
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader(githubIco)),
						StatusCode: http.StatusOK,
					}, nil
				default:
					return nil, fmt.Errorf("unexpected URL: %s", req.URL.String())
				}
			},
			giveBaseURL: "https://example.com",
			wantImage:   githubIcoImage,
		},

		"error, favicon.ico and HTML not found": {
			giveClient: func(req *http.Request) (*http.Response, error) {
				switch req.URL.String() {
				case "https://example.com/favicon.ico":
					return &http.Response{Body: http.NoBody, StatusCode: http.StatusNotFound}, nil
				case "https://example.com":
					return &http.Response{Body: http.NoBody, StatusCode: http.StatusNotFound}, nil
				default:
					return nil, fmt.Errorf("unexpected URL: %s", req.URL.String())
				}
			},
			giveBaseURL: "https://example.com",
			wantErrSubstr: "unexpected status code (404) while fetching the image (https://example.com/favicon.ico); " +
				"failed to fetch the list of favicons from the HTML",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			img, err := favicon.
				NewResolver(favicon.WithHTTPClient(tc.giveClient)).
				Resolve(context.Background(), tc.giveBaseURL, time.Second)

			if tc.wantErrSubstr != "" {
				assert.Nil(t, img)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrSubstr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantImage, img)
			}
		})
	}
}
