package docker_test

import (
	"bytes"
	"context"
	_ "embed"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"gh.tarampamp.am/indocker-app/daemon/internal/docker"
)

//go:embed testdata/list_containers.json
var listContainers []byte

func TestContainersWatch_Subscribe(t *testing.T) {
	defer goleak.VerifyNone(t)

	w, err := docker.NewContainersWatch(time.Millisecond, client.WithHTTPClient(
		newMockClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(listContainers)),
			}, nil
		}),
	))
	assert.NoError(t, err)

	// create a context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// run the watcher in a goroutine
	go func() { assert.ErrorIs(t, w.Watch(ctx), context.Canceled) }()

	var sub = make(docker.ContainersSubscription)
	defer close(sub)

	// subscribe to the updates
	assert.NoError(t, w.Subscribe(sub))
	assert.ErrorContains(t, w.Subscribe(sub), "already subscribed")

	defer func() {
		assert.NoError(t, w.Unsubscribe(sub))
		assert.ErrorContains(t, w.Unsubscribe(sub), "not subscribed")
	}()

	for i := 0; i < 10; i++ {
		update := <-sub
		assert.Len(t, update, 4)
		assert.Equal(t, "8dfafdbc3a40", update["8dfafdbc3a40"].ID)
		assert.Equal(t, "9cd87474be90", update["9cd87474be90"].ID)
		assert.Equal(t, "3176a2479c92", update["3176a2479c92"].ID)
		assert.Equal(t, "4cb07b47f9fb", update["4cb07b47f9fb"].ID)
	}
}
