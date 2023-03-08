package ws

import (
	"net/http"

	"golang.org/x/net/websocket"

	"gh.tarampamp.am/indocker-app/app/internal/docker"
)

func DockerState(dsw docker.ContainersStateWatcher) http.Handler {
	return websocket.Server{
		Handshake: func(_ *websocket.Config, _ *http.Request) (err error) {
			return nil // disable origin checking
		},
		Handler: websocket.Handler(func(ws *websocket.Conn) {
			var sub = make(docker.ContainersStateSubscription)
			if err := dsw.Subscribe(sub); err != nil {
				return
			}

			defer func() { _ = dsw.Unsubscribe(sub) }()

			for {
				select {
				case state := <-sub:
					if err := websocket.JSON.Send(ws, state); err != nil {
						break
					}
				}
			}
		}),
	}
}
