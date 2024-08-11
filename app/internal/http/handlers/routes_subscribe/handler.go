package routes_subscribe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

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

// New is a constructor for the [Handler] structure.
func New(sub routesSub) *Handler { return &Handler{sub: sub} }

// Handle is a function that handles the WebSocket connection. It reads messages from the client and sends routing
// updates to the client in [openapi.ContainerRoutesList] format.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	// upgrade the connection to the WebSocket
	ws, upgErr := h.upgrader.Upgrade(w, r, http.Header{})
	if upgErr != nil {
		return fmt.Errorf("failed to upgrade the connection: %w", upgErr)
	}

	defer func() { _ = ws.Close() }()

	// create a new context for the request
	var ctx, cancel = context.WithCancel(r.Context())
	defer cancel()

	// uncomment to debug the ping/pong messages
	// ws.SetPongHandler(func(appData string) error { fmt.Println(">>> pong", appData); return nil })

	// read messages from the client in a separate goroutine and cancel the context when the connection is closed or
	// an error occurs
	go func() { defer cancel(); _ = h.reader(ctx, ws) }()

	// run a loop that sends routing updates to the client and pings the client periodically
	return h.writer(ctx, ws)
}

// reader is a function that reads messages from the client. It must be run in a separate goroutine to prevent
// blocking. This function will exit when the context is canceled, the client closes the connection, or an error
// during the reading occurs.
func (h *Handler) reader(ctx context.Context, ws *websocket.Conn) error {
	for {
		if ctx.Err() != nil { // check if the context is canceled
			return nil
		}

		var messageType, msgReader, msgErr = ws.NextReader() // TODO: is there any way to avoid locking without context?
		if msgErr != nil {
			return msgErr
		}

		if msgReader != nil {
			_, _ = io.Copy(io.Discard, msgReader) // ignore the message body but read it to prevent potential memory leaks
		}

		if messageType == websocket.CloseMessage {
			return nil // client closed the connection
		}
	}
}

// writer is a function that writes messages to the client. It may NOT be run in a separate goroutine because it
// will block until the context is canceled, the client closes the connection, or an error during the writing occurs.
//
// This function sends the routing updates to the client and pings the client periodically.
func (h *Handler) writer(ctx context.Context, ws *websocket.Conn) error {
	// subscribe for routing updates
	var sub, stop = h.sub.SubscribeForRoutingUpdates()
	defer stop()

	// create a ticker for the ping messages
	var pingTicker = time.NewTicker(10 * time.Second) //nolint:mnd
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done(): // check if the context is canceled
			return nil

		case routes, isOpened := <-sub: // wait for the routing updates
			if !isOpened {
				return nil // this should never happen, but just in case
			}

			// write the response to the client
			if err := ws.WriteJSON(h.routesToResponse(routes)); err != nil {
				return fmt.Errorf("failed to write the message: %w", err)
			}

		case <-pingTicker.C: // send ping messages to the client
			if err := ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil { //nolint:mnd
				return fmt.Errorf("failed to send the ping message: %w", err)
			}
		}
	}
}

// routesToResponse is a helper function that converts the routing data to the response format.
func (*Handler) routesToResponse(routes map[string][]url.URL) openapi.ContainerRoutesList {
	var response = openapi.ContainerRoutesList{Routes: make([]openapi.ContainerRoute, 0, len(routes))}

	for hostname, urls := range routes {
		var route = openapi.ContainerRoute{Hostname: hostname, Urls: make([]string, 0, len(urls))}

		for _, u := range urls {
			route.Urls = append(route.Urls, u.String())
		}

		response.Routes = append(response.Routes, route)
	}

	return response
}
