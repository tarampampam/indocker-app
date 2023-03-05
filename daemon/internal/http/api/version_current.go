package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

func VersionCurrent(ver string) http.HandlerFunc {
	var (
		once  sync.Once
		cache []byte
	)

	return func(w http.ResponseWriter, _ *http.Request) {
		once.Do(func() {
			cache, _ = json.Marshal(struct {
				Version string `json:"version"`
			}{
				Version: ver,
			})
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write(cache)
	}
}
