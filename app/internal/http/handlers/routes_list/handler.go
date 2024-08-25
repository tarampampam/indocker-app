package routes_list

import (
	"net/url"
	"slices"
	"strings"

	"gh.tarampamp.am/indocker-app/app/internal/http/openapi"
)

type (
	dockerRouter interface {
		AllContainerURLs() map[string]map[string]url.URL
	}

	Handler struct{ router dockerRouter }
)

func New(router dockerRouter) *Handler { return &Handler{router: router} }

func (h *Handler) Handle() (resp openapi.RegisteredRoutesListResponse) {
	var routes = h.router.AllContainerURLs()

	resp.Routes = make([]openapi.ContainerRoute, 0, len(routes))

	for hostname, urlsMap := range routes {
		var route = openapi.ContainerRoute{Hostname: hostname, Urls: make(map[string]string, len(urlsMap))}

		for containerID, u := range urlsMap {
			route.Urls[containerID] = u.String()
		}

		resp.Routes = append(resp.Routes, route)
	}

	// keep the list sorted
	slices.SortFunc(resp.Routes, func(a, b openapi.ContainerRoute) int { return strings.Compare(a.Hostname, b.Hostname) })

	return
}
