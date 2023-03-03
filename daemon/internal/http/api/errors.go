package api

import "net/http"

func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		_, _ = w.Write([]byte(`{"error": "not found"}`))
	}
}
