package indocker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type Docker struct {
	client *http.Client

	containers struct {
		sync.RWMutex
		m map[string]Container
	}
}

type (
	Container struct {
		Inspect Inspect
		Stats   Stats
	}

	Inspect struct {
		Cmd      []string          // config
		Env      []string          // config
		Hostname string            // config
		Labels   map[string]string // config
		User     string            // config

		Created      time.Time
		ID           string
		Image        string
		Name         string
		RestartCount uint32

		ExitCode      int    // state
		HealthStatus  string // health state
		FailingStreak uint32 // health state
		OOMKilled     bool   // state
		Dead          bool   // state
		Paused        bool   // state
		Restarting    bool   // state
		Running       bool   // state
		PID           int    // state
		Status        string // state
	}

	Stats struct {
		Read            time.Time
		NumProcs        uint
		CPUUsage        uint64
		MemoryUsage     uint64
		MemoryMaxUsage  uint64
		MemoryLimit     uint64
		NetworkRxBytes  uint64
		NetworkRxErrors uint64
		NetworkTxBytes  uint64
		NetworkTxErrors uint64
	}
)

func NewDocker(unixSocket string) *Docker {
	var d = Docker{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, "unix", unixSocket)
				},
				DisableKeepAlives: true,
			},
		},
	}

	d.containers.m = make(map[string]Container) // init map

	return &d
}

func (d *Docker) Watch(ctx context.Context, interval time.Duration) {
	var wg sync.WaitGroup

	for {
		ids, rErr := d.RunningContainerIDs(ctx)
		if rErr != nil {
			_, _ = os.Stderr.WriteString(rErr.Error())

			if d.pause(ctx, interval) {
				return
			}

			continue
		}

		{ // sync the containers map
			var tmp = make(map[string]struct{}, len(ids))
			for _, id := range ids { // convert slice into temporary map
				tmp[id] = struct{}{}
			}

			d.containers.Lock()

			for id := range d.containers.m { // remove non-existent containers
				if _, ok := tmp[id]; !ok {
					delete(d.containers.m, id)
				}
			}

			for id := range tmp { // add new containers
				if _, ok := d.containers.m[id]; !ok {
					d.containers.m[id] = Container{} // but without useful data (yet)
				}
			}

			d.containers.Unlock()
		}

		d.containers.RLock()
		for id := range d.containers.m {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()

				inspect, err := d.Inspect(ctx, id)
				if err != nil {
					return
				}

				d.containers.Lock()
				if v, ok := d.containers.m[id]; ok {
					v.Inspect = *inspect
					d.containers.m[id] = v
				}
				d.containers.Unlock()
			}(id)

			wg.Add(1)
			go func(id string) {
				defer wg.Done()

				stats, err := d.Stats(ctx, id)
				if err != nil {
					return
				}

				d.containers.Lock()
				if v, ok := d.containers.m[id]; ok {
					v.Stats = *stats
					d.containers.m[id] = v
				}
				d.containers.Unlock()
			}(id)
		}
		d.containers.RUnlock()

		wg.Wait()

		fmt.Printf("%+v\n\n", d.containers.m) // TODO only for debug

		if d.pause(ctx, interval) {
			return
		}
	}
}

func (d *Docker) pause(ctx context.Context, interval time.Duration) (canceled bool) {
	var t = time.NewTimer(interval)
	defer t.Stop()

	select {
	case <-ctx.Done():
		canceled = true

	case <-t.C:
	}

	return
}

func (d *Docker) RunningContainerIDs(ctx context.Context) ([]string, error) {
	// https://docs.docker.com/engine/api/v1.42/#tag/Container/operation/ContainerList
	req, _ := http.NewRequestWithContext(ctx,
		http.MethodGet,
		// allowed statuses is: created|restarting|running|removing|paused|exited|dead
		`http://docker/containers/json?filters={"status":["created","restarting","running","removing","paused"]}`,
		http.NoBody,
	)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code: " + resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	_ = resp.Body.Close() // better sooner than later

	// decode the response
	var payload = make([]struct {
		ID string `json:"Id"`
	}, 0, 16)
	if err = json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	if len(payload) == 0 {
		return nil, nil
	}

	ids := make([]string, len(payload))
	for i := 0; i < len(payload); i++ {
		ids[i] = payload[i].ID
	}

	return ids, nil
}

func (d *Docker) Inspect(ctx context.Context, id string) (*Inspect, error) {
	// https://docs.docker.com/engine/api/v1.42/#tag/Container/operation/ContainerInspect
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/containers/"+id+"/json", http.NoBody)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code: " + resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	_ = resp.Body.Close() // better sooner than later

	// decode the response
	var payload struct {
		Config struct {
			Cmd      []string          `json:"Cmd"`
			Env      []string          `json:"Env"`
			Hostname string            `json:"Hostname"`
			Image    string            `json:"Image"`
			Labels   map[string]string `json:"Labels"`
			User     string            `json:"User"`
		} `json:"Config"`
		Created      time.Time `json:"Created"`
		ID           string    `json:"Id"`
		Image        string    `json:"Image"`
		Name         string    `json:"Name"`
		RestartCount uint32    `json:"RestartCount"`
		State        struct {
			ExitCode int `json:"ExitCode"`
			Health   struct {
				Status        string `json:"Status"`
				FailingStreak uint32 `json:"FailingStreak"`
			} `json:"Health"`
			OOMKilled  bool   `json:"OOMKilled"`
			Dead       bool   `json:"Dead"`
			Paused     bool   `json:"Paused"`
			Restarting bool   `json:"Restarting"`
			Running    bool   `json:"Running"`
			PID        int    `json:"Pid"`
			Status     string `json:"Status"`
		} `json:"State"`
	}
	if err = json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	return &Inspect{
		Cmd:           payload.Config.Cmd,
		Env:           payload.Config.Env,
		Hostname:      payload.Config.Hostname,
		Labels:        payload.Config.Labels,
		User:          payload.Config.User,
		Created:       payload.Created,
		ID:            payload.ID,
		Image:         payload.Image,
		Name:          payload.Name,
		RestartCount:  payload.RestartCount,
		ExitCode:      payload.State.ExitCode,
		Status:        payload.State.Status,
		HealthStatus:  payload.State.Health.Status,
		FailingStreak: payload.State.Health.FailingStreak,
		OOMKilled:     payload.State.OOMKilled,
		Dead:          payload.State.Dead,
		Paused:        payload.State.Paused,
		Restarting:    payload.State.Restarting,
		Running:       payload.State.Running,
		PID:           payload.State.PID,
	}, nil
}

func (d *Docker) Stats(ctx context.Context, id string) (*Stats, error) {
	// https://docs.docker.com/engine/api/v1.42/#tag/Container/operation/ContainerStats
	req, _ := http.NewRequestWithContext(ctx,
		http.MethodGet,
		"http://docker/containers/"+id+"/stats?stream=0&one-shot=1",
		http.NoBody,
	)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code: " + resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	_ = resp.Body.Close() // better sooner than later

	// decode the response
	var payload struct {
		Read     time.Time `json:"read"`
		NumProcs uint      `json:"num_procs"`
		CPUStats struct {
			CPUUsage struct { // CPU Usage
				TotalUsage        uint64 `json:"total_usage"`         // total CPU time consumed (nanoseconds)
				UsageInKernelMode uint64 `json:"usage_in_kernelmode"` // time spent in kernel mode (nanoseconds)
				UsageInUserMode   uint64 `json:"usage_in_usermode"`   // time spent in user mode (nanoseconds)
			} `json:"cpu_usage"`
			SystemUsage uint64 `json:"system_cpu_usage"` // system Usage
			OnlineCPUs  uint32 `json:"online_cpus"`      // online CPUs
		} `json:"cpu_stats"`
		MemoryStats struct { // memory Usage
			Usage    uint64   `json:"usage"`     // current res_counter usage for memory
			MaxUsage uint64   `json:"max_usage"` // maximum usage ever recorded
			Stats    struct { // all the stats exported via memory.stat (https://bit.ly/3Z5FpN8)
				ActiveAnon              uint64 `json:"active_anon"`
				ActiveFile              uint64 `json:"active_file"`
				Cache                   uint64 `json:"cache"`
				Dirty                   uint64 `json:"dirty"`
				HierarchicalMemoryLimit uint64 `json:"hierarchical_memory_limit"`
				HierarchicalMemSWLimit  uint64 `json:"hierarchical_memsw_limit"`
				InactiveAnon            uint64 `json:"inactive_anon"`
				InactiveFile            uint64 `json:"inactive_file"`
				MappedFile              uint64 `json:"mapped_file"`
				PgFault                 uint64 `json:"pgfault"`
				PGMAJFault              uint64 `json:"pgmajfault"`
				PGPGin                  uint64 `json:"pgpgin"`
				PGPGout                 uint64 `json:"pgpgout"`
				RSS                     uint64 `json:"rss"`
				RSSHuge                 uint64 `json:"rss_huge"`
				TotalActiveAnon         uint64 `json:"total_active_anon"`
				TotalActiveFile         uint64 `json:"total_active_file"`
				TotalCache              uint64 `json:"total_cache"`
				TotalDirty              uint64 `json:"total_dirty"`
				TotalInactiveAnon       uint64 `json:"total_inactive_anon"`
				TotalInactiveFile       uint64 `json:"total_inactive_file"`
				TotalMappedFile         uint64 `json:"total_mapped_file"`
				TotalPGFault            uint64 `json:"total_pgfault"`
				TotalPGMAJFault         uint64 `json:"total_pgmajfault"`
				TotalPGPGin             uint64 `json:"total_pgpgin"`
				TotalPGPGout            uint64 `json:"total_pgpgout"`
				TotalRSS                uint64 `json:"total_rss"`
				TotalRSSHuge            uint64 `json:"total_rss_huge"`
				TotalUnEvictable        uint64 `json:"total_unevictable"`
				TotalWriteBack          uint64 `json:"total_writeback"`
				UnEvictable             uint64 `json:"unevictable"`
				WriteBack               uint64 `json:"writeback"`
			} `json:"stats"`
			Limit uint64 `json:"limit"`
		} `json:"memory_stats"`
		Networks map[string]struct { // network stats of one container
			RxBytes    uint64 `json:"rx_bytes"`              // bytes received
			RxPackets  uint64 `json:"rx_packets"`            // packets received
			RxErrors   uint64 `json:"rx_errors"`             // received errors
			RxDropped  uint64 `json:"rx_dropped"`            // incoming packets dropped
			TxBytes    uint64 `json:"tx_bytes"`              // bytes sent
			TxPackets  uint64 `json:"tx_packets"`            // packets sent
			TxErrors   uint64 `json:"tx_errors"`             // sent errors
			TxDropped  uint64 `json:"tx_dropped"`            // outgoing packets dropped
			EndpointID string `json:"endpoint_id,omitempty"` // endpoint ID (not used on Linux)
			InstanceID string `json:"instance_id,omitempty"` // instance ID (not used on Linux)
		} `json:"networks"`
	}
	if err = json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	var stats = Stats{
		Read:           payload.Read,
		NumProcs:       payload.NumProcs,
		CPUUsage:       payload.CPUStats.CPUUsage.TotalUsage,
		MemoryUsage:    payload.MemoryStats.Usage,
		MemoryMaxUsage: payload.MemoryStats.MaxUsage,
		MemoryLimit:    payload.MemoryStats.Limit,
	}

	for _, s := range payload.Networks {
		stats.NetworkRxBytes += s.RxBytes
		stats.NetworkRxErrors += s.RxErrors
		stats.NetworkTxBytes += s.TxBytes
		stats.NetworkTxErrors += s.TxErrors
	}

	return &stats, nil
}
