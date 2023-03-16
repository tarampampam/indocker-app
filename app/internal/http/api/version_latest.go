package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"gh.tarampamp.am/indocker-app/app/internal/version"
)

type versionFetcher interface {
	Fetch() (*version.Release, error)
}

func VersionLatest(vf versionFetcher, cacheTTL time.Duration) Handler {
	var (
		mu        sync.Mutex
		updatedAt time.Time
		cache     []byte
	)

	return HandlerFunc(func(w http.ResponseWriter, _ *http.Request) error {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		mu.Lock()
		defer mu.Unlock()

		// check if cache is not expired
		if time.Since(updatedAt) < cacheTTL { // cache is not expired
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(http.StatusOK)

			_, _ = w.Write(cache)

			return nil
		}

		w.Header().Set("X-Cache", "MISS")

		// cache is expired, fetch new version
		latest, err := vf.Fetch()
		if err != nil {
			return err
		}

		// cache new version
		cache, _ = json.Marshal(struct {
			Version   string `json:"version"`
			URL       string `json:"url"`
			Name      string `json:"name"`
			Body      string `json:"body"`
			CreatedAt string `json:"created_at"`
		}{
			Version:   latest.Version,
			URL:       latest.URL,
			Name:      latest.Name,
			Body:      latest.Body,
			CreatedAt: latest.CreatedAt.Format(time.RFC3339),
		})

		// update timestamp
		updatedAt = time.Now()

		w.WriteHeader(http.StatusOK)

		_, err = w.Write(cache)

		return err
	})
}
