package ping

import "gh.tarampamp.am/indocker-app/app/internal/http/openapi"

type Handler struct{}

func New() *Handler { return &Handler{} }

func (h *Handler) Handle() openapi.PingResponse { return "pong" }
