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

// LatestVersion represents the latest version from GitHub releases.
type LatestVersion struct {
	Version   string    // release version (with "v" prefix)
	URL       string    // URL to release page
	Name      string    // release name
	Body      string    // release body
	CreatedAt time.Time // release creation date
}

// GetLatestVersion returns the latest version from GitHub releases. If GitHub API key provided, it will be used for
// authentication. If not, then GitHub API will be used with anonymous access.
func GetLatestVersion(ctx context.Context, client httpClient, gitHubApiKey ...string) (*LatestVersion, error) {
	const repo = "tarampampam/indocker-app"

	req, err := http.NewRequestWithContext(ctx,
		// https://docs.github.com/en/rest/releases/releases?apiVersion=2022-11-28#list-releases
		http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=99&page=1", repo), http.NoBody,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	if len(gitHubApiKey) > 0 && gitHubApiKey[0] != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gitHubApiKey[0]))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var payload []struct {
		Url         string    `json:"url"`
		HtmlUrl     string    `json:"html_url"`
		TagName     string    `json:"tag_name"`
		Name        string    `json:"name"`
		Body        string    `json:"body"`
		Draft       bool      `json:"draft"`
		Prerelease  bool      `json:"prerelease"`
		CreatedAt   time.Time `json:"created_at"`
		PublishedAt time.Time `json:"published_at"`
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
		return &LatestVersion{
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
			return &LatestVersion{
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
