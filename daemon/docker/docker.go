package docker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const (
	prefix       = "indocker"
	hostLabel    = prefix + ".host"
	portLabel    = prefix + ".port"
	networkLabel = prefix + ".network"
	schemeLabel  = prefix + ".scheme"
)

type Docker struct {
	frequency time.Duration
	client    *client.Client

	aliveMu sync.RWMutex
	alive   map[string]types.Container

	hostsMu sync.RWMutex
	hosts   map[string]string // map[hostname]container_id
}

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
			for _, container := range list {
				d.alive[container.ID] = container
			}
			d.aliveMu.Unlock()

			d.hostsMu.Lock()
			for k := range d.hosts {
				delete(d.hosts, k)
			}
			for _, container := range list {
				if value, ok := container.Labels[hostLabel]; ok {
					d.hosts[value] = container.ID
					break
				}
			}
			d.hostsMu.Unlock()
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

			if v, ok := container.Labels[schemeLabel]; ok {
				scheme = v
			}

			if v, ok := container.Labels[portLabel]; ok {
				port = v
			}

			if v, ok := container.Labels[networkLabel]; ok {
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

				return scheme + "://" + net.IPAddress + ":" + port, nil
			}
		}
	}

	return "", errors.New("not found")
}
