package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// DiscoverMiddleware is a middleware that returns information about the current service.
func DiscoverMiddleware(baseUrl string, next http.Handler) http.Handler {
	const (
		needRoute  = "/discover"
		needMethod = http.MethodTrace
	)

	type response struct {
		BaseUrl *string `json:"base_url"` // without trailing slash
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == needMethod && r.URL.Path == needRoute && r.Header.Get("X-InDocker") == "true" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", needMethod)
			w.Header().Set("Access-Control-Allow-Headers", "X-InDocker")
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			w.WriteHeader(http.StatusOK)

			var (
				scheme = "http"
				data   response
			)

			if r.TLS != nil {
				scheme = "https"
			}

			if baseUrl != "" {
				u := fmt.Sprintf("%s://%s.indocker.app", scheme, baseUrl)
				data.BaseUrl = &u
			}

			_ = json.NewEncoder(w).Encode(data)

			return
		}

		next.ServeHTTP(w, r)
	})
}
