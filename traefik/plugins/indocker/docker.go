package indocker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Docker is a simple Docker watcher.
type Docker struct {
	client httpClient

	snapshots struct {
		sync.RWMutex
		data []*Snapshot
		cap  int // readonly
	}
}

type (
	Snapshot struct {
		CreatedAt  time.Time             `json:"created_at"`
		Containers map[string]*Container `json:"containers"`
	}

	// Container contains the Docker container information.
	Container struct {
		Inspect Inspect `json:"inspect"`
		Stats   Stats   `json:"stats"`
	}

	// Inspect is the Docker inspect output.
	Inspect struct {
		Cmd      []string          `json:"cmd"`      // config
		Env      []string          `json:"env"`      // config
		Hostname string            `json:"hostname"` // config
		Labels   map[string]string `json:"labels"`   // config
		User     string            `json:"user"`     // config

		Created      time.Time `json:"created"`
		ID           string    `json:"id"`
		Image        string    `json:"image"`
		Name         string    `json:"name"`
		RestartCount uint32    `json:"restart_count"`

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
		NumProcs        uint      `json:"num_procs"`
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

// DockerOption is a function that configures a Docker.
type DockerOption func(*Docker)

// WithHTTPClient allows to configure the HTTP client used to communicate with the Docker API.
func WithHTTPClient(httpClient httpClient) DockerOption {
	return func(d *Docker) { d.client = httpClient }
}

// WithSnapshotsCapacity allows to configure the maximal number of snapshots.
func WithSnapshotsCapacity(cap uint) DockerOption {
	return func(d *Docker) {
		d.snapshots.cap = int(cap)
		d.snapshots.data = make([]*Snapshot, 0, cap)
	}
}

// NewDocker creates a new Docker.
func NewDocker(unixSocket string, opt ...DockerOption) *Docker {
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

	for _, option := range opt {
		option(&d)
	}

	if d.snapshots.cap == 0 || d.snapshots.data == nil {
		const defaultCap = 10

		d.snapshots.cap = defaultCap
		d.snapshots.data = make([]*Snapshot, 0, defaultCap)
	}

	return &d
}

// Watch should be run in a goroutine and will watch for container changes.
func (d *Docker) Watch(ctx context.Context, interval time.Duration) {
	var wg sync.WaitGroup

	for {
		ids, rErr := d.RunningContainerIDs(ctx)
		if rErr != nil || len(ids) == 0 {
			if d.pause(ctx, interval) {
				return
			}

			continue
		}

		var (
			snapshotMu sync.Mutex
			snapshot   = Snapshot{CreatedAt: time.Now(), Containers: make(map[string]*Container, len(ids))}
		)

		for _, id := range ids {
			snapshot.Containers[id] = &Container{} // make init
		}

		for _, id := range ids {
			wg.Add(1)
			go func(id string) { // inspect
				defer wg.Done()

				if inspect, err := d.Inspect(ctx, id); err != nil {
					return
				} else {
					snapshotMu.Lock()
					snapshot.Containers[id].Inspect = *inspect
					snapshotMu.Unlock()
				}
			}(id)

			wg.Add(1)
			go func(id string) { // stats
				defer wg.Done()

				if stats, err := d.Stats(ctx, id); err != nil {
					return
				} else {
					snapshotMu.Lock()
					snapshot.Containers[id].Stats = *stats
					snapshotMu.Unlock()
				}
			}(id)
		}

		wg.Wait()

		{ // sync the snapshots slice
			d.snapshots.Lock()
			if len(d.snapshots.data) >= d.snapshots.cap {
				d.snapshots.data = append(d.snapshots.data[1:len(d.snapshots.data)], &snapshot) // remove first element and append new one
			} else {
				d.snapshots.data = append(d.snapshots.data, &snapshot) // append new one
			}
			d.snapshots.Unlock()
		}

		if d.pause(ctx, interval) {
			return
		}
	}
}

// Snapshots returns the current snapshots.
func (d *Docker) Snapshots() []*Snapshot {
	d.snapshots.RLock()
	defer d.snapshots.RUnlock()

	var s = make([]*Snapshot, len(d.snapshots.data))
	copy(s, d.snapshots.data) // copy slice

	return s
}

// pause pauses the current goroutine for the given interval.
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

// RunningContainerIDs returns the IDs of all running containers.
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

// Inspect returns the inspect data of the given container.
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

// Stats returns the stats of the given container.
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
				// ActiveAnon              uint64 `json:"active_anon"`
				// ActiveFile              uint64 `json:"active_file"`
				// Cache                   uint64 `json:"cache"`
				// Dirty                   uint64 `json:"dirty"`
				// HierarchicalMemoryLimit uint64 `json:"hierarchical_memory_limit"`
				// HierarchicalMemSWLimit  uint64 `json:"hierarchical_memsw_limit"`
				// InactiveAnon            uint64 `json:"inactive_anon"`
				// InactiveFile            uint64 `json:"inactive_file"`
				// MappedFile              uint64 `json:"mapped_file"`
				// PgFault                 uint64 `json:"pgfault"`
				// PGMAJFault              uint64 `json:"pgmajfault"`
				// PGPGin                  uint64 `json:"pgpgin"`
				// PGPGout                 uint64 `json:"pgpgout"`
				RSS uint64 `json:"rss"`
				// RSSHuge                 uint64 `json:"rss_huge"`
				// TotalActiveAnon         uint64 `json:"total_active_anon"`
				// TotalActiveFile         uint64 `json:"total_active_file"`
				// TotalCache              uint64 `json:"total_cache"`
				// TotalDirty              uint64 `json:"total_dirty"`
				// TotalInactiveAnon       uint64 `json:"total_inactive_anon"`
				// TotalInactiveFile       uint64 `json:"total_inactive_file"`
				// TotalMappedFile         uint64 `json:"total_mapped_file"`
				// TotalPGFault            uint64 `json:"total_pgfault"`
				// TotalPGMAJFault         uint64 `json:"total_pgmajfault"`
				// TotalPGPGin             uint64 `json:"total_pgpgin"`
				// TotalPGPGout            uint64 `json:"total_pgpgout"`
				// TotalRSS                uint64 `json:"total_rss"`
				// TotalRSSHuge            uint64 `json:"total_rss_huge"`
				// TotalUnEvictable        uint64 `json:"total_unevictable"`
				// TotalWriteBack          uint64 `json:"total_writeback"`
				// UnEvictable             uint64 `json:"unevictable"`
				// WriteBack               uint64 `json:"writeback"`
			} `json:"stats"`
			Limit uint64 `json:"limit"`
		} `json:"memory_stats"`
		Networks map[string]struct { // network stats of one container
			RxBytes   uint64 `json:"rx_bytes"`   // bytes received
			RxPackets uint64 `json:"rx_packets"` // packets received
			RxErrors  uint64 `json:"rx_errors"`  // received errors
			RxDropped uint64 `json:"rx_dropped"` // incoming packets dropped
			TxBytes   uint64 `json:"tx_bytes"`   // bytes sent
			TxPackets uint64 `json:"tx_packets"` // packets sent
			TxErrors  uint64 `json:"tx_errors"`  // sent errors
			TxDropped uint64 `json:"tx_dropped"` // outgoing packets dropped
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
