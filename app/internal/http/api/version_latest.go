package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"gh.tarampamp.am/indocker-app/app/internal/version"
)

type latestVersionFetcher func() (*version.LatestVersion, error)

// TODO: remove this?
func VersionLatest(fetcher latestVersionFetcher, invalidateCacheAfter time.Duration) http.HandlerFunc {
	var (
		mu        sync.Mutex
		updatedAt time.Time
		cache     []byte
		lastError error
	)

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		mu.Lock()
		defer mu.Unlock()

		if time.Since(updatedAt) > invalidateCacheAfter || lastError != nil {
			if l, err := fetcher(); err == nil {
				cache, _ = json.Marshal(struct {
					Version   string `json:"version"`
					URL       string `json:"url"`
					Name      string `json:"name"`
					Body      string `json:"body"`
					CreatedAt string `json:"created_at"`
				}{
					Version:   l.Version,
					URL:       l.URL,
					Name:      l.Name,
					Body:      l.Body,
					CreatedAt: l.CreatedAt.Format(time.RFC3339),
				})

				updatedAt, lastError = time.Now(), nil
			} else {
				lastError, cache = err, []byte(`{"error":"`+err.Error()+`"}`)
			}
		}

		if lastError == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		_, _ = w.Write(cache)
	}
}
