package middleware

import (
	"net/http"

	"gh.tarampamp.am/indocker-app/app/internal/cli/start/healthcheck"
)

// HealthcheckMiddleware is a middleware that handles healthcheck requests.
// It is used to check if the application is alive. Only for internal usage.
func HealthcheckMiddleware(next http.Handler) http.Handler {
	const (
		needUa     = healthcheck.UserAgent
		needRoute  = healthcheck.Route
		meedMethod = healthcheck.Method
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == meedMethod && r.URL.Path == needRoute && r.Header.Get("User-Agent") == needUa {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))

			return
		}

		next.ServeHTTP(w, r)
	})
}
