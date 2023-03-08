package docker_test

import (
	"bytes"
	"context"
	_ "embed"
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"gh.tarampamp.am/indocker-app/app/internal/docker"
)

//go:embed testdata/inspect.json
var inspect []byte

//go:embed testdata/stats.json
var stats []byte

func TestNewContainerStateWatch_Subscribe(t *testing.T) {
	defer goleak.VerifyNone(t)

	w, err := docker.NewContainerStateWatch(client.WithHTTPClient(
		newMockClient(func(req *http.Request) (*http.Response, error) {
			if strings.HasSuffix(req.URL.Path, "/stats") {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(stats)),
				}, nil
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(inspect)),
			}, nil
		}),
	))
	assert.NoError(t, err)

	// create a context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cw = &watcherMock{}

	// run the state watcher in a goroutine
	go func() { assert.ErrorIs(t, w.Watch(ctx, cw), context.Canceled) }()

	var sub = make(docker.ContainersStateSubscription)
	defer close(sub)

	// subscribe to the updates
	assert.NoError(t, w.Subscribe(sub))
	assert.ErrorContains(t, w.Subscribe(sub), "already subscribed")

	defer func() {
		assert.NoError(t, w.Unsubscribe(sub))
		assert.ErrorContains(t, w.Unsubscribe(sub), "not subscribed")
	}()

	runtime.Gosched()

	for i := 0; i < 10; i++ {
		go cw.Push(map[string]types.Container{
			"3176a2479c92": {ID: "3176a2479c92"},
			"4cb07b47f9fb": {ID: "4cb07b47f9fb"},
		})

		<-time.After(time.Millisecond * 10)
		runtime.Gosched()

		update := <-sub
		assert.Len(t, update, 2)
		assert.Equal(t, "ba033ac4401106a3b513bc9d639eee123ad78ca3616b921167cd74b20e25ed39", update["3176a2479c92"].Inspect.ID)
		assert.Equal(t, uint64(3), update["3176a2479c92"].Stats.PidsStats.Current)
	}
}
