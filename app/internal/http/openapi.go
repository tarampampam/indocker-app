package http

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	pingHandler "gh.tarampamp.am/indocker-app/app/internal/http/handlers/ping"
	versionHandler "gh.tarampamp.am/indocker-app/app/internal/http/handlers/version"
	latestVersionHandler "gh.tarampamp.am/indocker-app/app/internal/http/handlers/version_latest"
	"gh.tarampamp.am/indocker-app/app/internal/http/openapi"
	"gh.tarampamp.am/indocker-app/app/internal/version"
)

type OpenAPI struct {
	log *zap.Logger

	handlers struct {
		ping          func() openapi.PingResponse
		version       func() openapi.AppVersionResponse
		latestVersion func(http.ResponseWriter) (*openapi.AppVersionResponse, error)
	}
}

var _ openapi.ServerInterface = (*OpenAPI)(nil) // verify interface implementation

func NewOpenAPI(ctx context.Context, log *zap.Logger) *OpenAPI {
	var si = &OpenAPI{log: log}

	si.handlers.ping = pingHandler.New().Handle
	si.handlers.version = versionHandler.New(version.Version()).Handle
	si.handlers.latestVersion = latestVersionHandler.New(func() (string, error) { return version.Latest(ctx) }).Handle

	return si
}

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json; charset=utf-8"
)

func (o *OpenAPI) Ping(w http.ResponseWriter, _ *http.Request) {
	o.respToJson(w, o.handlers.ping())
}

func (o *OpenAPI) GetAppVersion(w http.ResponseWriter, _ *http.Request) {
	o.respToJson(w, o.handlers.version())
}

func (o *OpenAPI) GetLatestAppVersion(w http.ResponseWriter, _ *http.Request) {
	if resp, err := o.handlers.latestVersion(w); err != nil {
		o.errorToJson(w, err)
	} else {
		o.respToJson(w, resp)
	}
}

func (o *OpenAPI) respToJson(w http.ResponseWriter, resp any) {
	if resp == nil {
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		o.log.Error("failed to encode/write response", zap.Error(err))
	}
}

func (o *OpenAPI) errorToJson(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusInternalServerError)

	if encErr := json.NewEncoder(w).Encode(openapi.ErrorResponse{Error: err.Error()}); encErr != nil {
		o.log.Error("failed to encode/write error response", zap.Error(encErr))
	}
}