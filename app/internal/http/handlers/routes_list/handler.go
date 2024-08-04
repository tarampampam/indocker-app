package routes_list

import (
	"net/url"

	"gh.tarampamp.am/indocker-app/app/internal/http/openapi"
)

type (
	dockerRouter interface{ AllContainerURLs() map[string]url.URL }

	Handler struct{ router dockerRouter }
)

func New(router dockerRouter) *Handler { return &Handler{router: router} }

func (h *Handler) Handle() (resp openapi.RegisteredRoutesListResponse) {
	var routes = h.router.AllContainerURLs()

	resp.Routes = make([]openapi.ContainerRoute, 0, len(routes))

	for hostname, u := range routes {
		resp.Routes = append(resp.Routes, openapi.ContainerRoute{Hostname: hostname, Url: u.String()})
	}

	return
}
