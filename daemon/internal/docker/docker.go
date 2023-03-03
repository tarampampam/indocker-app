package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const (
	labelsPrefix = "indocker."

	LabelHost    = labelsPrefix + "host"
	LabelPort    = labelsPrefix + "port"
	LabelNetwork = labelsPrefix + "network"
	LabelScheme  = labelsPrefix + "scheme"
)

type Docker struct {
	frequency time.Duration
	client    *client.Client

	aliveMu sync.RWMutex
	alive   map[string]types.Container

	hostsMu sync.RWMutex
	hosts   map[string]string // map[hostname]container_id

	snapshotsMu sync.RWMutex
	snapshots   []Snapshot
}

type (
	Snapshot struct {
		CreatedAt  time.Time             `json:"created_at"`
		Containers map[string]*Container `json:"containers"`
	}

	// Container contains the Docker container information.
	Container struct {
		Inspect types.ContainerJSON `json:"inspect"`
		Stats   types.StatsJSON     `json:"stats"`
	}
)

func NewDocker(frequency time.Duration, opts ...client.Opt) (*Docker, error) {
	c, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	return &Docker{
		frequency: frequency,
		client:    c,
		alive:     make(map[string]types.Container),
		hosts:     make(map[string]string),
		snapshots: make([]Snapshot, 0, 20),
	}, nil
}

func (d *Docker) Watch(pCtx context.Context) {
	var (
		ctx, cancel = context.WithCancel(pCtx)
		alive       = make(chan []types.Container, 10)
	)

	defer func() {
		cancel()
		close(alive)
	}()

	go d.watchAliveContainers(ctx, alive)

	for {
		select {
		case <-ctx.Done():
			return

		case list := <-alive:
			d.aliveMu.Lock()
			for k := range d.alive {
				delete(d.alive, k)
			}
			for _, c := range list {
				d.alive[c.ID] = c
			}
			d.aliveMu.Unlock()

			d.hostsMu.Lock()
			for k := range d.hosts {
				delete(d.hosts, k)
			}
			for _, c := range list {
				if value, ok := c.Labels[LabelHost]; ok {
					d.hosts[value] = c.ID
					break
				}
			}
			d.hostsMu.Unlock()

			var (
				snapshotMu sync.Mutex
				snapshot   = Snapshot{CreatedAt: time.Now(), Containers: make(map[string]*Container, len(list))}
			)

			for _, c := range list {
				snapshot.Containers[c.ID] = &Container{}
			}

			var wg sync.WaitGroup

			for _, c := range list {
				wg.Add(1)
				go func(id string) { // inspect
					defer wg.Done()

					if inspect, err := d.client.ContainerInspect(ctx, id); err != nil {
						return
					} else {
						snapshotMu.Lock()
						snapshot.Containers[id].Inspect = inspect
						snapshotMu.Unlock()
					}
				}(c.ID)

				wg.Add(1)
				go func(id string) { // stats
					defer wg.Done()

					if resp, err := d.client.ContainerStatsOneShot(ctx, id); err != nil {
						return
					} else {
						var stats = types.StatsJSON{}

						if decodingErr := json.NewDecoder(resp.Body).Decode(&stats); decodingErr != nil {
							return
						}

						snapshotMu.Lock()
						snapshot.Containers[id].Stats = stats
						snapshotMu.Unlock()
					}
				}(c.ID)
			}

			wg.Wait()

			d.snapshotsMu.Lock()
			if len(d.snapshots) >= cap(d.snapshots) {
				d.snapshots = append(d.snapshots[1:len(d.snapshots)], snapshot) // remove first element and append new one
			} else {
				d.snapshots = append(d.snapshots, snapshot) // append new one
			}
			d.snapshotsMu.Unlock()
		}
	}
}

func (d *Docker) watchAliveContainers(ctx context.Context, out chan<- []types.Container, outErr ...chan<- error) {
	var f = filters.NewArgs()

	// https://docs.docker.com/engine/api/v1.42/#tag/Container/operation/ContainerList
	// status=(created|restarting|running|removing|paused|exited|dead)
	for _, s := range []string{"created", "restarting", "running"} {
		f.Add("status", s)
	}

	var opt = types.ContainerListOptions{Filters: f}

	var t = time.NewTicker(d.frequency)
	defer t.Stop()

	for {
		list, err := d.client.ContainerList(ctx, opt)
		if err != nil {
			if len(outErr) > 0 {
				select {
				case <-ctx.Done():
					return

				case outErr[0] <- err:
				}
			}
		} else {
			select {
			case <-ctx.Done():
				return

			case out <- list:
			}
		}

		t.Reset(d.frequency)

		select {
		case <-ctx.Done():
			return

		case <-t.C:
		}
	}
}

func (d *Docker) FindRoute(hostname string) (string, error) {
	d.hostsMu.RLock()
	var (
		hostsLen        = len(d.hosts)
		id, routeExists = d.hosts[hostname] // get the container ID for the hostname
	)
	d.hostsMu.RUnlock()

	if hostsLen == 0 {
		return "", errors.New("no routes registered")
	}

	if routeExists {
		d.aliveMu.RLock()
		defer d.aliveMu.RUnlock()

		if container, found := d.alive[id]; found {
			var scheme, port, netName = "http", "80", "bridge" // defaults

			if v, ok := container.Labels[LabelScheme]; ok {
				scheme = v
			}

			if v, ok := container.Labels[LabelPort]; ok {
				port = v
			}

			if v, ok := container.Labels[LabelNetwork]; ok {
				netName = v
			}

			if container.NetworkSettings != nil && len(container.NetworkSettings.Networks) > 0 {
				var net *network.EndpointSettings

				if namedNet, ok := container.NetworkSettings.Networks[netName]; ok {
					net = namedNet
				} else {
					for _, rndNet := range container.NetworkSettings.Networks { // take first (random) value from the map
						net = rndNet

						break
					}
				}

				if net != nil {
					return scheme + "://" + net.IPAddress + ":" + port, nil
				}

				return "", fmt.Errorf("no network for the container %s found", id)
			}
		}
	}

	return "", errors.New("not found")
}

// Snapshots returns the current snapshots.
func (d *Docker) Snapshots() []Snapshot {
	d.snapshotsMu.RLock()
	defer d.snapshotsMu.RUnlock()

	var s = make([]Snapshot, len(d.snapshots))
	copy(s, d.snapshots) // copy slice

	return s
}
