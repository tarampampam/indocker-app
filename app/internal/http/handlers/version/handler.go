package version

import "gh.tarampamp.am/indocker-app/app/internal/http/openapi"

type Handler struct{ ver string }

func New(ver string) *Handler { return &Handler{ver: ver} }

func (h *Handler) Handle() openapi.AppVersionResponse {
	return openapi.AppVersionResponse{Version: h.ver}
}
