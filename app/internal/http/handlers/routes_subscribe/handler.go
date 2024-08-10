package routes_subscribe

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"

	"gh.tarampamp.am/indocker-app/app/internal/http/openapi"
)

type (
	routesSub interface {
		SubscribeForRoutingUpdates() (sub <-chan map[string][]url.URL, stop func())
	}

	Handler struct {
		sub      routesSub
		upgrader websocket.Upgrader
	}
)

func New(sub routesSub) *Handler { return &Handler{sub: sub} }

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error { //nolint:funlen
	// upgrade the connection to the WebSocket
	ws, upgErr := h.upgrader.Upgrade(w, r, http.Header{})
	if upgErr != nil {
		return fmt.Errorf("failed to upgrade the connection: %w", upgErr)
	} else {
		defer func() { _ = ws.Close() }()
	}

	// create a new context for the request
	var ctx, cancel = context.WithCancel(r.Context())
	defer cancel()

	// subscribe for routing updates
	var sub, stop = h.sub.SubscribeForRoutingUpdates()
	defer stop()

	// read messages from the client in separate goroutine
	go func() {
		defer cancel() // cancel the context when the function exits

		for {
			if ctx.Err() != nil { // check if the context is canceled
				return
			}

			if messageType, _, err := ws.NextReader(); err != nil || messageType == websocket.CloseMessage {
				return // client closed the connection or an error occurred
			}
		}
	}()

	// send messages to the client with the routing updates
	for {
		select {
		case <-ctx.Done(): // check if the context is canceled
			return nil
		case routes, isOpened := <-sub: // wait for the routing updates
			if !isOpened {
				return nil // this should never happen, but just in case
			}

			// create a new response
			var response = openapi.ContainerRoutesList{Routes: make([]openapi.ContainerRoute, 0, len(routes))}

			// fill the response with the current routing data (what we got from the router)
			for hostname, urls := range routes {
				var route = openapi.ContainerRoute{Hostname: hostname, Urls: make([]string, 0, len(urls))}

				for _, u := range urls {
					route.Urls = append(route.Urls, u.String())
				}

				response.Routes = append(response.Routes, route)
			}

			// write the response to the client
			if err := ws.WriteJSON(response); err != nil {
				return fmt.Errorf("failed to write the message: %w", err)
			}
		}
	}
}
