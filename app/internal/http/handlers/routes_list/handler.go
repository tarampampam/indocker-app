package routes_list

import (
	"net/url"

	"gh.tarampamp.am/indocker-app/app/internal/http/openapi"
)

type (
	dockerRouter interface{ AllContainerURLs() map[string][]url.URL }

	Handler struct{ router dockerRouter }
)

func New(router dockerRouter) *Handler { return &Handler{router: router} }

func (h *Handler) Handle() (resp openapi.RegisteredRoutesListResponse) {
	var routes = h.router.AllContainerURLs()

	resp.Routes = make([]openapi.ContainerRoute, 0, len(routes))

	for hostname, urls := range routes {
		var route = openapi.ContainerRoute{Hostname: hostname, Urls: make([]string, 0, len(urls))}

		for _, u := range urls {
			route.Urls = append(route.Urls, u.String())
		}

		resp.Routes = append(resp.Routes, route)
	}

	return
}
