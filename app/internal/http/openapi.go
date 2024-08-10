package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	pingHandler "gh.tarampamp.am/indocker-app/app/internal/http/handlers/ping"
	routesListHandler "gh.tarampamp.am/indocker-app/app/internal/http/handlers/routes_list"
	routesSubscribeHandler "gh.tarampamp.am/indocker-app/app/internal/http/handlers/routes_subscribe"
	versionHandler "gh.tarampamp.am/indocker-app/app/internal/http/handlers/version"
	latestVersionHandler "gh.tarampamp.am/indocker-app/app/internal/http/handlers/version_latest"
	"gh.tarampamp.am/indocker-app/app/internal/http/openapi"
	"gh.tarampamp.am/indocker-app/app/internal/version"
)

type (
	dockerRouter interface {
		AllContainerURLs() map[string][]url.URL
		SubscribeForRoutingUpdates() (sub <-chan map[string][]url.URL, stop func())
	}

	OpenAPI struct {
		log *zap.Logger

		handlers struct {
			ping            func() openapi.PingResponse
			version         func() openapi.AppVersionResponse
			latestVersion   func(http.ResponseWriter) (*openapi.AppVersionResponse, error)
			routesList      func() openapi.RegisteredRoutesListResponse
			routesSubscribe func(http.ResponseWriter, *http.Request) error
		}
	}
)

var _ openapi.ServerInterface = (*OpenAPI)(nil) // verify interface implementation

func NewOpenAPI(ctx context.Context, log *zap.Logger, dockerRouter dockerRouter) *OpenAPI {
	var si = &OpenAPI{log: log}

	si.handlers.ping = pingHandler.New().Handle
	si.handlers.version = versionHandler.New(version.Version()).Handle
	si.handlers.latestVersion = latestVersionHandler.New(func() (string, error) { return version.Latest(ctx) }).Handle
	si.handlers.routesList = routesListHandler.New(dockerRouter).Handle
	si.handlers.routesSubscribe = routesSubscribeHandler.New(dockerRouter).Handle

	return si
}

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json; charset=utf-8"
)

// --------------------------------------------------- API handlers ---------------------------------------------------

func (o *OpenAPI) Ping(w http.ResponseWriter, _ *http.Request) {
	o.respToJson(w, o.handlers.ping())
}

func (o *OpenAPI) GetAppVersion(w http.ResponseWriter, _ *http.Request) {
	o.respToJson(w, o.handlers.version())
}

func (o *OpenAPI) GetLatestAppVersion(w http.ResponseWriter, _ *http.Request) {
	if resp, err := o.handlers.latestVersion(w); err != nil {
		o.errorToJson(w, err, http.StatusInternalServerError)
	} else {
		o.respToJson(w, resp)
	}
}

func (o *OpenAPI) ListRoutes(w http.ResponseWriter, _ *http.Request) {
	o.respToJson(w, o.handlers.routesList())
}

func (o *OpenAPI) SubscribeRoutes(w http.ResponseWriter, r *http.Request, _ openapi.SubscribeRoutesParams) {
	if err := o.handlers.routesSubscribe(w, r); err != nil {
		o.errorToJson(w, err, http.StatusInternalServerError)
	}
}

// -------------------------------------------------- Error handlers --------------------------------------------------

// HandleInternalError is a default error handler for internal server errors (e.g. query parameters binding
// errors, and so on).
func (o *OpenAPI) HandleInternalError(w http.ResponseWriter, _ *http.Request, err error) {
	o.errorToJson(w, err, http.StatusBadRequest)
}

// HandleNotFoundError is a default error handler for "404: not found" errors.
func (o *OpenAPI) HandleNotFoundError(w http.ResponseWriter, _ *http.Request) {
	o.errorToJson(w, errors.New("not found"), http.StatusNotFound)
}

// ------------------------------------------------- Internal helpers -------------------------------------------------

func (o *OpenAPI) respToJson(w http.ResponseWriter, resp any) {
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if resp == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		o.log.Error("failed to encode/write response", zap.Error(err))
	}
}

func (o *OpenAPI) errorToJson(w http.ResponseWriter, err error, status int) {
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(status)

	if err == nil {
		return
	}

	if encErr := json.NewEncoder(w).Encode(openapi.ErrorResponse{Error: err.Error()}); encErr != nil {
		o.log.Error("failed to encode/write error response", zap.Error(encErr))
	}
}
