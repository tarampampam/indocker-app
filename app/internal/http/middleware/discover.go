package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// DiscoverMiddleware is a middleware that returns information about the current service.
func DiscoverMiddleware(dashboardDomain string, next http.Handler) http.Handler {
	const (
		needRoute  = "/x/indocker/discover"
		needHeader = "X-Indocker"
		meedMethod = http.MethodGet
	)

	// getScheme returns the scheme of the request.
	var getScheme = func(r *http.Request) string {
		if r.TLS != nil {
			return "https"
		}

		return "http"
	}

	// writeResponse writes the response to the client.
	var writeResponse = func(w http.ResponseWriter, r *http.Request) {
		var data = struct {
			API struct {
				BaseUrl *string `json:"base_url"` // without trailing slash
			} `json:"api"`
		}{}

		if dashboardDomain != "" {
			u := fmt.Sprintf("%s://%s.indocker.app/api", getScheme(r), dashboardDomain)
			data.API.BaseUrl = &u
		}

		_ = json.NewEncoder(w).Encode(data)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if dashboardDomain != "" && r.URL.Path == needRoute {
			w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("%s://%s.indocker.app", getScheme(r), dashboardDomain))
			w.Header().Set("Access-Control-Allow-Methods", meedMethod)
			w.Header().Set("Access-Control-Allow-Headers", "*")

			switch r.Method {
			case http.MethodOptions: // https://developer.mozilla.org/en-US/docs/Glossary/Preflight_request
				w.WriteHeader(http.StatusNoContent)

			case meedMethod: // response with data
				if ok, _ := strconv.ParseBool(r.Header.Get(needHeader)); ok {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusOK)

					writeResponse(w, r)
				} else {
					w.WriteHeader(http.StatusBadRequest) // missing header
				}

			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}

			return
		}

		next.ServeHTTP(w, r)
	})
}
