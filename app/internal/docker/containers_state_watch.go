package docker

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type ContainerStateWatch struct {
	client *client.Client

	mu   sync.Mutex
	subs map[ContainersStateSubscription]chan struct{}
}

type (
	ContainersStateSubscription chan map[string]*ContainerState // map[container_id]container_state

	ContainerState struct {
		Inspect *types.ContainerJSON `json:"inspect"`
		Stats   *types.StatsJSON     `json:"stats"`
	}

	ContainersStateWatcher interface {
		Subscribe(ch ContainersStateSubscription) error
		Unsubscribe(ch ContainersStateSubscription) error
	}
)

// NewContainerStateWatch creates a new ContainerStateWatch.
func NewContainerStateWatch(opts ...client.Opt) (*ContainerStateWatch, error) {
	c, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	return &ContainerStateWatch{
		client: c,
		subs:   make(map[ContainersStateSubscription]chan struct{}),
	}, nil
}

// Watch starts watching for containers state changes.
func (w *ContainerStateWatch) Watch(ctx context.Context, cw ContainersWatcher) error { //nolint:funlen,gocognit,gocyclo,lll
	// create a subscription channel
	var sub = make(ContainersSubscription)
	defer close(sub)

	// subscribe to updates
	if err := cw.Subscribe(sub); err != nil {
		return err
	}

	// unsubscribe from updates
	defer func() { _ = cw.Unsubscribe(sub) }()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case containers := <-sub:
			w.mu.Lock()
			var subsCount = len(w.subs)
			w.mu.Unlock()

			if subsCount > 0 { //nolint:nestif
				var (
					details = make(map[string]*ContainerState, len(containers))
					wg      sync.WaitGroup
				)

				for _, c := range containers {
					details[c.ID] = &ContainerState{} // init map values
				}

				for _, c := range containers {
					wg.Add(1)

					go func(id string) { // inspect
						defer wg.Done()

						if inspect, err := w.client.ContainerInspect(ctx, id); err != nil {
							return
						} else {
							details[id].Inspect = &inspect
						}
					}(c.ID)

					wg.Add(1)

					go func(id string) { // stats
						defer wg.Done()

						if stats, err := w.client.ContainerStatsOneShot(ctx, id); err != nil {
							return
						} else {
							var data = types.StatsJSON{}
							if decodingErr := json.NewDecoder(stats.Body).Decode(&data); decodingErr != nil {
								return
							}

							details[id].Stats = &data
						}
					}(c.ID)
				}

				wg.Wait()

				w.mu.Lock()
				for subscriber, stop := range w.subs {
					if ctx.Err() != nil {
						return ctx.Err()
					}

					go func(subscriber ContainersStateSubscription, stop <-chan struct{}) {
						select {
						case <-ctx.Done():
							return

						case <-stop:
							return

						case subscriber <- details:
						}
					}(subscriber, stop)
				}
				w.mu.Unlock()
			}
		}
	}
}

// Subscribe adds a new subscription.
// Note: do not forget to Unsubscribe when you are done.
func (w *ContainerStateWatch) Subscribe(ch ContainersStateSubscription) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.subs[ch]; ok {
		return errors.New("already subscribed")
	}

	w.subs[ch] = make(chan struct{})

	return nil
}

// Unsubscribe removes the subscription.
func (w *ContainerStateWatch) Unsubscribe(ch ContainersStateSubscription) error {
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
