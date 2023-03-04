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

type Docker struct {
	frequency time.Duration
	client    *client.Client

	aliveMu sync.RWMutex
	alive   map[string]types.Container // TODO remove?

	hostsMu sync.RWMutex
	hosts   map[string]string // map[hostname]container_id

	subsMu sync.Mutex
	subs   map[chan<- map[string]*ContainerDetails]chan struct{} // map[subscription]cancel
}

type (
	// ContainerDetails contains the Docker container information.
	ContainerDetails struct {
		Inspect *Inspect `json:"inspect"`
		Stats   *Stats   `json:"stats"`
	}

	// Inspect is the Docker inspect output.
	Inspect struct {
		Cmd      []string          `json:"cmd"`      // config
		Env      []string          `json:"env"`      // config
		Hostname string            `json:"hostname"` // config
		Labels   map[string]string `json:"labels"`   // config
		User     string            `json:"user"`     // config

		Created      string `json:"created"`
		ID           string `json:"id"`
		Image        string `json:"image"`
		Name         string `json:"name"`
		RestartCount uint32 `json:"restart_count"`

		ExitCode      int    `json:"exit_code"`      // state
		HealthStatus  string `json:"health_status"`  // health state
		FailingStreak uint32 `json:"failing_streak"` // health state
		OOMKilled     bool   `json:"oom_killed"`     // state
		Dead          bool   `json:"dead"`           // state
		Paused        bool   `json:"paused"`         // state
		Restarting    bool   `json:"restarting"`     // state
		Running       bool   `json:"running"`        // state
		PID           int    `json:"pid"`            // state
		Status        string `json:"status"`         // state
	}

	// Stats is the Docker stats output.
	Stats struct {
		Read            time.Time `json:"read"`
		NumProcs        uint32    `json:"num_procs"`
		CPUUsage        uint64    `json:"cpu_usage"`
		MemoryUsage     uint64    `json:"memory_usage"`
		MemoryMaxUsage  uint64    `json:"memory_max_usage"`
		MemoryLimit     uint64    `json:"memory_limit"`
		NetworkRxBytes  uint64    `json:"network_rx_bytes"`
		NetworkRxErrors uint64    `json:"network_rx_errors"`
		NetworkTxBytes  uint64    `json:"network_tx_bytes"`
		NetworkTxErrors uint64    `json:"network_tx_errors"`
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
		subs:      make(map[chan<- map[string]*ContainerDetails]chan struct{}),
	}, nil
}

const (
	labelsPrefix = "indocker."

	LabelHost    = labelsPrefix + "host"
	LabelPort    = labelsPrefix + "Port"
	LabelNetwork = labelsPrefix + "Network"
	LabelScheme  = labelsPrefix + "Scheme"
)

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

			d.subsMu.Lock()
			var subsCount = len(d.subs)
			d.subsMu.Unlock()

			if subsCount > 0 {
				var (
					details = make(map[string]*ContainerDetails, len(list))
					wg      sync.WaitGroup
				)

				for _, c := range list {
					details[c.ID] = &ContainerDetails{} // init map values
				}

				for _, c := range list {
					wg.Add(1)
					go func(id string) { // inspect
						defer wg.Done()

						if inspect, err := d.inspectForSnapshot(ctx, id); err != nil {
							return
						} else {
							details[id].Inspect = inspect
						}
					}(c.ID)

					wg.Add(1)
					go func(id string) { // stats
						defer wg.Done()

						if stats, err := d.statsForSnapshot(ctx, id); err != nil {
							return
						} else {
							details[id].Stats = stats
						}
					}(c.ID)
				}

				wg.Wait()

				d.subsMu.Lock()
				for ch, stop := range d.subs {
					go func(ch chan<- map[string]*ContainerDetails, stop <-chan struct{}) {
						select {
						case <-stop:
							return

						case <-ctx.Done():
							return

						case ch <- details:
						}
					}(ch, stop)
				}
				d.subsMu.Unlock()
			}
		}
	}
}

func (d *Docker) Subscribe(ch chan<- map[string]*ContainerDetails) error {
	d.subsMu.Lock()
	defer d.subsMu.Unlock()

	if _, ok := d.subs[ch]; ok {
		return errors.New("already subscribed")
	}

	d.subs[ch] = make(chan struct{})

	return nil
}

func (d *Docker) Unsubscribe(ch chan map[string]*ContainerDetails) error {
	d.subsMu.Lock()
	defer d.subsMu.Unlock()

	if stop, ok := d.subs[ch]; !ok {
		return errors.New("not subscribed")
	} else {
		close(stop)
	}

	delete(d.subs, ch)

	return nil
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

func (d *Docker) inspectForSnapshot(ctx context.Context, id string) (*Inspect, error) {
	resp, err := d.client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	var inspect = Inspect{
		Created:      resp.Created,
		ID:           resp.ID,
		Image:        resp.Image,
		Name:         resp.Name,
		RestartCount: uint32(resp.RestartCount),
	}

	if config := resp.Config; config != nil {
		inspect.Cmd = config.Cmd
		inspect.Env = config.Env
		inspect.Hostname = config.Hostname
		inspect.Labels = config.Labels
		inspect.User = config.User
	}

	if state := resp.State; state != nil {
		inspect.ExitCode = state.ExitCode
		inspect.Status = state.Status
		inspect.OOMKilled = state.OOMKilled
		inspect.Dead = state.Dead
		inspect.Paused = state.Paused
		inspect.Restarting = state.Restarting
		inspect.Running = state.Running
		inspect.PID = state.Pid

		if health := state.Health; health != nil {
			inspect.HealthStatus = health.Status
			inspect.FailingStreak = uint32(health.FailingStreak)
		}
	}

	return &inspect, nil
}

func (d *Docker) statsForSnapshot(ctx context.Context, id string) (*Stats, error) {
	resp, err := d.client.ContainerStatsOneShot(ctx, id)
	if err != nil {
		return nil, err
	}

	var data = types.StatsJSON{}
	if decodingErr := json.NewDecoder(resp.Body).Decode(&data); decodingErr != nil {
		return nil, decodingErr
	}

	var stats = Stats{
		Read:           data.Read,
		NumProcs:       data.NumProcs,
		CPUUsage:       data.CPUStats.CPUUsage.TotalUsage,
		MemoryUsage:    data.MemoryStats.Usage,
		MemoryMaxUsage: data.MemoryStats.MaxUsage,
		MemoryLimit:    data.MemoryStats.Limit,
	}

	for _, netStat := range data.Networks {
		stats.NetworkRxBytes += netStat.RxBytes
		stats.NetworkRxErrors += netStat.RxErrors
		stats.NetworkTxBytes += netStat.TxBytes
		stats.NetworkTxErrors += netStat.TxErrors
	}

	return &stats, nil
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

				return "", fmt.Errorf("no Network for the container %s found", id)
			}
		}
	}

	return "", errors.New("not found")
}
