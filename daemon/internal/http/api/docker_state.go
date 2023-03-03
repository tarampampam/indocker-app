package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"gh.tarampamp.am/indocker-app/daemon/internal/docker"
)

type DockerState struct {
	docker *docker.Docker
}

func NewDockerState(docker *docker.Docker) *DockerState {
	return &DockerState{docker: docker}
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

	var interval = 1 * time.Second // default refresh interval

	if passedInterval := r.URL.Query().Get("interval"); passedInterval != "" {
		if d, err := time.ParseDuration(passedInterval); err != nil {
			return false, errors.New("invalid interval: " + err.Error())
		} else {
			interval = d
		}
	}

	var t = time.NewTicker(interval)
	defer t.Stop()

	var buf bytes.Buffer // reuse buffer to reduce allocations

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	var j = json.NewEncoder(&buf)

	for {
		buf.WriteString("data: ")

		if err := j.Encode(h.docker.Snapshots()); err != nil {
			return true, err
		}

		buf.WriteRune('\n')

		if _, err := buf.WriteTo(w); err != nil { // writing automatically resets the buffer
			return true, err
		}

		flusher.Flush()

		select {
		case <-r.Context().Done(): // received browser disconnection
			return true, nil

		case <-t.C:
		}
	}
}
