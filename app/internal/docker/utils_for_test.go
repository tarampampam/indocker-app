package docker_test

import (
	"net/http"
	"sync"

	"github.com/docker/docker/api/types"

	"gh.tarampamp.am/indocker-app/app/internal/docker"
)

// transportFunc allows us to inject mock transport for testing. We define it
// here, so we can detect the tlsconfig and return nil for only this type.
type transportFunc func(*http.Request) (*http.Response, error)

func (tf transportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return tf(req)
}

func newMockClient(doer func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: transportFunc(doer),
	}
}

type watcherMock struct {
	sync.Mutex
	ch docker.ContainersSubscription
}

var _ docker.ContainersWatcher = (*watcherMock)(nil) // verify interface implementation

func (wm *watcherMock) Push(d map[string]types.Container) { // this is a helper method for testing
	wm.Lock()
	if wm.ch != nil {
		wm.ch <- d
	}
	wm.Unlock()
}

func (wm *watcherMock) Subscribe(ch docker.ContainersSubscription) error {
	wm.Lock()
	wm.ch = ch
	wm.Unlock()

	return nil
}

func (wm *watcherMock) Unsubscribe(docker.ContainersSubscription) error {
	wm.Lock()
	wm.ch = nil
	wm.Unlock()

	return nil
}
