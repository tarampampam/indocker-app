package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"gh.tarampamp.am/indocker-app/daemon/internal/docker"
)

type DockerState struct {
	dsw docker.ContainersStateWatcher
}

func NewDockerState(dsw docker.ContainersStateWatcher) *DockerState {
	return &DockerState{dsw: dsw}
}

func (h *DockerState) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if headersSent, err := h.handle(w, r); err != nil {
		if !headersSent {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
		}

		_ = json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
	}
}

func (h *DockerState) handle(w http.ResponseWriter, r *http.Request) (bool, error) {
	flusher, isFlusher := w.(http.Flusher)
	if !isFlusher {
		return false, errors.New("streaming unsupported")
	}

	var sub = make(docker.ContainersStateSubscription)
	if err := h.dsw.Subscribe(sub); err != nil {
		return false, err
	}

	defer func() { _ = h.dsw.Unsubscribe(sub) }()

	var buf bytes.Buffer // reuse buffer to reduce allocations

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	var enc = json.NewEncoder(&buf)

	for {
		select {
		case details := <-sub:
			buf.WriteString("data: ")

			if err := enc.Encode(details); err != nil {
				return true, err
			}

			buf.WriteRune('\n')

			if _, err := buf.WriteTo(w); err != nil { // writing automatically resets the buffer
				return true, err
			}

			flusher.Flush()

		case <-r.Context().Done(): // received browser disconnection
			return true, nil
		}
	}
}
