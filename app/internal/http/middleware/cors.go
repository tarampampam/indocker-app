package middleware

import (
	"fmt"
	"net/http"
)

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var scheme = "http"

		if r.TLS != nil {
			scheme = "https"
		}

		w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("%s://%s", scheme, r.Host))
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		next.ServeHTTP(w, r)
	})
}
