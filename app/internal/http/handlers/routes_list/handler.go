package routes_list

import (
	"slices"
	"strings"

	"gh.tarampamp.am/indocker-app/app/internal/docker"
	"gh.tarampamp.am/indocker-app/app/internal/http/openapi"
)

type Handler struct {
	router docker.AllContainerURLsResolver
}

func New(router docker.AllContainerURLsResolver) *Handler { return &Handler{router: router} }

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
