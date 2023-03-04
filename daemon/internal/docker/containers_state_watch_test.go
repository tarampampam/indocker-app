package docker_test

import (
	"context"
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"gh.tarampamp.am/indocker-app/daemon/internal/docker"
)

type watcherMock2 struct {
	sync.Mutex
	ch docker.ContainersSubscription
}

var _ docker.ContainersWatcher = (*watcherMock2)(nil) // verify interface implementation

func (wm *watcherMock2) Push(d map[string]types.Container) { // this is a helper method for testing
	wm.Lock()
	if wm.ch != nil {
		wm.ch <- d
	}
	wm.Unlock()
}

func (wm *watcherMock2) Subscribe(ch docker.ContainersSubscription) error {
	wm.Lock()
	wm.ch = ch
	wm.Unlock()

	return nil
}

func (wm *watcherMock2) Unsubscribe(docker.ContainersSubscription) error {
	wm.Lock()
	wm.ch = nil
	wm.Unlock()

	return nil
}

func TestNewContainerStateWatch_Subscribe(t *testing.T) {
	defer goleak.VerifyNone(t)

	w, err := docker.NewContainerStateWatch(client.WithHTTPClient(
		newMockClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			}, nil
		}),
	))
	assert.NoError(t, err)

	// create a context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cw = &watcherMock2{}

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

	// for i := 0; i < 10; i++ {
	go cw.Push(map[string]types.Container{
		"3176a2479c92": {ID: "3176a2479c92"},
		"4cb07b47f9fb": {ID: "4cb07b47f9fb"},
	})
	// }

	runtime.Gosched()
	<-time.After(10 * time.Millisecond)

	update := <-sub
	t.Log(update)
}
