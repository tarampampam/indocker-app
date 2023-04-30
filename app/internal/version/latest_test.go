package version_test

import (
	"bytes"
	"context"
	_ "embed"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/app/internal/version"
)

type httpClientFunc func(*http.Request) (*http.Response, error)

func (f httpClientFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

//go:embed testdata/github_releases.json
var githubReleases []byte

func TestLatest_Fetch(t *testing.T) {
	t.Parallel()

	var httpMock httpClientFunc = func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t,
			"https://api.github.com/repos/foo/bar/releases?per_page=99&page=1", req.URL.String(),
		)
		assert.Equal(t, "application/vnd.github.v3+json", req.Header.Get("Accept"))
		assert.Equal(t, "2022-11-28", req.Header.Get("X-GitHub-Api-Version"))
		assert.Equal(t, "Bearer FOO-TOKEN", req.Header.Get("Authorization"))

		return &http.Response{
			Header: http.Header{
				"content-type":          {"application/json; charset=utf-8"},
				"x-ratelimit-limit":     {"60"},
				"x-ratelimit-remaining": {"53"},
				"x-ratelimit-reset":     {"1678029063"},
				"x-ratelimit-resource":  {"core"},
				"x-ratelimit-used":      {"7"},
			},
			Body:       io.NopCloser(bytes.NewReader(githubReleases)),
			StatusCode: http.StatusOK,
		}, nil
	}

	latest := version.NewLatest(
		version.WithGithubAPIKey("FOO-TOKEN"),
		version.WithHTTPClient(httpMock),
		version.WithGithubRepo("foo/bar"),
	)

	release, err := latest.Fetch()
	assert.NoError(t, err)

	assert.Equal(t, "v1.2.0", release.Version)
	assert.Equal(t, "https://github.com/tarampampam/indocker-app/releases/tag/v1.2.0", release.URL)
	assert.Equal(t, "v1.2.0", release.Name)
	assert.Equal(t, "## What's Changed\r\n\r\n* Added possibility", release.Body)

	createdAt, err := time.Parse(time.RFC3339, "2023-02-06T17:09:10Z")
	assert.NoError(t, err)
	assert.Equal(t, createdAt, release.CreatedAt)
}

//go:embed testdata/github_releases_one.json
var githubReleasesOne []byte

func TestLatest_Fetch_One(t *testing.T) {
	t.Parallel()

	var httpMock httpClientFunc = func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t,
			"https://api.github.com/repos/tarampampam/indocker-app/releases?per_page=99&page=1", req.URL.String(),
		)
		assert.Equal(t, "application/vnd.github.v3+json", req.Header.Get("Accept"))
		assert.Equal(t, "2022-11-28", req.Header.Get("X-GitHub-Api-Version"))
		assert.Equal(t, "", req.Header.Get("Authorization"))

		return &http.Response{
			Header: http.Header{
				"content-type":      {"application/json; charset=utf-8"},
				"x-ratelimit-limit": {"60"},
			},
			Body:       io.NopCloser(bytes.NewReader(githubReleasesOne)),
			StatusCode: http.StatusOK,
		}, nil
	}

	latest := version.NewLatest(version.WithHTTPClient(httpMock))

	release, err := latest.Fetch()
	assert.NoError(t, err)

	assert.Equal(t, "v1.1.1", release.Version)
	assert.Equal(t, "https://github.com/tarampampam/indocker-app/releases/tag/v1.1.1", release.URL)
	assert.Equal(t, "Foo bar", release.Name)
	assert.Equal(t, "**Full Changelog**: https://github.com/tarampampam/indocker-app/compare/v1.1.0...v1.1.1", release.Body) //nolint:lll

	createdAt, err := time.Parse(time.RFC3339, "2023-02-06T08:45:16Z")
	assert.NoError(t, err)
	assert.Equal(t, createdAt, release.CreatedAt)
}

func TestLatest_Fetch_Cancel(t *testing.T) {
	t.Parallel()

	var ctx, cancel = context.WithCancel(context.Background())

	latest := version.NewLatest(version.WithContext(ctx))

	cancel() // cancel context

	release, err := latest.Fetch()
	assert.Nil(t, release)
	assert.ErrorIs(t, err, context.Canceled)
}
