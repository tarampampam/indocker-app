package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

func VersionCurrent(ver string) Handler {
	var (
		once  sync.Once
		cache []byte
	)

	return HandlerFunc(func(w http.ResponseWriter, _ *http.Request) error {
		once.Do(func() {
			cache, _ = json.Marshal(struct {
				Version string `json:"version"`
			}{
				Version: ver,
			})
		})

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write(cache)

		return err
	})
}
