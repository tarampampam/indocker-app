package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// OwnGithubRepo is a GitHub repository name, where the application is hosted.
const OwnGithubRepo = "tarampampam/indocker-app"

type (
	// Latest is a struct, that can be used to fetch the latest release information from GitHub.
	Latest struct {
		http httpClient
		ctx  context.Context
		repo string

		githubAPIKey string
	}

	// Release represents GitHub release information.
	Release struct {
		Version   string    // release version (with "v" prefix)
		URL       string    // URL to release page
		Name      string    // release name
		Body      string    // release body (description)
		CreatedAt time.Time // release creation date
	}

	// LatestOption is a function, that can be used to configure the Latest instance.
	LatestOption func(*Latest)
)

// WithHTTPClient allows to set custom HTTP client.
func WithHTTPClient(c httpClient) LatestOption { return func(l *Latest) { l.http = c } }

// WithContext allows to set custom context (context.Background will be used by default).
func WithContext(ctx context.Context) LatestOption { return func(l *Latest) { l.ctx = ctx } }

// WithGithubAPIKey allows to set GitHub API key (without this key, GitHub API will be limited to 60 requests per hour).
func WithGithubAPIKey(key string) LatestOption { return func(l *Latest) { l.githubAPIKey = key } }

// WithGithubRepo allows to set custom GitHub repository name.
func WithGithubRepo(repo string) LatestOption { return func(l *Latest) { l.repo = repo } }

// NewLatest creates new Latest instance.
func NewLatest(opts ...LatestOption) *Latest {
	latest := &Latest{
		ctx:  context.Background(),
		repo: OwnGithubRepo,
	}

	for _, opt := range opts {
		opt(latest)
	}

	if latest.http == nil {
		latest.http = &http.Client{
			Timeout: time.Second * 30, //nolint:mnd
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}
	}

	return latest
}

// Fetch fetches the latest release information from GitHub.
func (l *Latest) Fetch() (*Release, error) { //nolint:funlen
	req, err := http.NewRequestWithContext(l.ctx,
		// https://docs.github.com/en/rest/releases/releases?apiVersion=2022-11-28#list-releases
		http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=99&page=1", l.repo), http.NoBody,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	if l.githubAPIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", l.githubAPIKey))
	}

	resp, err := l.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var payload []struct {
		HtmlUrl   string    `json:"html_url"`
		TagName   string    `json:"tag_name"`
		Name      string    `json:"name"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
	}

	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}

	if len(payload) == 0 {
		return nil, fmt.Errorf("no releases found")
	}

	// force "v" prefix for the tag name (this is required for semver sorting)
	for i := range payload {
		if !strings.HasPrefix(payload[i].TagName, "v") {
			payload[i].TagName = "v" + payload[i].TagName
		}
	}

	if len(payload) == 1 {
		return &Release{
			Version:   payload[0].TagName,
			URL:       payload[0].HtmlUrl,
			Name:      payload[0].Name,
			Body:      payload[0].Body,
			CreatedAt: payload[0].CreatedAt,
		}, nil
	}

	var allVersions = make(semver.ByVersion, len(payload))
	for i := range payload {
		// store tag names in the separate slice
		allVersions[i] = payload[i].TagName
	}

	// https://pkg.go.dev/golang.org/x/mod/semver
	semver.Sort(allVersions)

	for _, release := range payload { // search for the latest release
		if release.TagName == allVersions[len(allVersions)-1] {
			return &Release{
				Version:   release.TagName,
				URL:       release.HtmlUrl,
				Name:      release.Name,
				Body:      release.Body,
				CreatedAt: release.CreatedAt,
			}, nil
		}
	}

	return nil, fmt.Errorf("no latest release found") // never happens
}
