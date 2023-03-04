package docker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// ContainersWatch is a running docker containers watcher. It allows you to subscribe to changes in the status of
// running docker containers.
type ContainersWatch struct {
	interval time.Duration
	client   *client.Client

	mu   sync.Mutex
	subs map[ContainersSubscription]chan struct{}
}

type (
	ContainersSubscription chan map[string]types.Container // map[container_id]container_data

	ContainersWatcher interface {
		// Subscribe subscribes to changes in the status of running docker containers.
		Subscribe(ch ContainersSubscription) error

		// Unsubscribe unsubscribes from changes in the status of running docker containers.
		Unsubscribe(ch ContainersSubscription) error
	}
)

// NewContainersWatch creates a new ContainersWatch.
func NewContainersWatch(interval time.Duration, opts ...client.Opt) (*ContainersWatch, error) {
	c, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	return &ContainersWatch{
		interval: interval,
		client:   c,
		subs:     make(map[ContainersSubscription]chan struct{}),
	}, nil
}

// Watch starts watching for changes in the status of running docker containers.
func (w *ContainersWatch) Watch(ctx context.Context) error {
	var f = filters.NewArgs()

	// https://docs.docker.com/engine/api/v1.42/#tag/Container/operation/ContainerList
	// status=(created|restarting|running|removing|paused|exited|dead)
	for _, s := range []string{"created", "restarting", "running", "removing", "paused"} {
		f.Add("status", s)
	}

	var opt = types.ContainerListOptions{Filters: f}

	var t = time.NewTicker(w.interval)
	defer t.Stop()

	for {
		w.mu.Lock()
		var subsCount = len(w.subs)
		w.mu.Unlock()

		if subsCount > 0 {
			list, err := w.client.ContainerList(ctx, opt)
			if err == nil { // note: this error is not logged anywhere
				listMap := make(map[string]types.Container, len(list))
				for _, c := range list {
					listMap[c.ID] = c
				}

				w.mu.Lock()
				for subscriber, stop := range w.subs {
					if ctx.Err() != nil {
						return ctx.Err()
					}

					go func(subscriber ContainersSubscription, stop <-chan struct{}) {
						select {
						case <-ctx.Done():
							return

						case <-stop:
							return

						case subscriber <- listMap:
						}
					}(subscriber, stop)
				}
				w.mu.Unlock()
			}
		}

		t.Reset(w.interval)

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-t.C:
		}
	}
}

// Subscribe subscribes to changes in the status of running docker containers.
// Note: do not forget to Unsubscribe when you are done.
func (w *ContainersWatch) Subscribe(ch ContainersSubscription) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.subs[ch]; ok {
		return errors.New("already subscribed")
	}

	w.subs[ch] = make(chan struct{})

	return nil
}

// Unsubscribe unsubscribes from changes in the status of running docker containers.
func (w *ContainersWatch) Unsubscribe(ch ContainersSubscription) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if stop, ok := w.subs[ch]; !ok {
		return errors.New("not subscribed")
	} else {
		close(stop)
	}

	delete(w.subs, ch)

	return nil
}
