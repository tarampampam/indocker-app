package docker

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/network"
)

// ContainersRoute is a docker containers router. It allows you to get the Route of a docker container by its labels.
type ContainersRoute struct {
	defaults Route
	builder  RouteBuilder

	mu    sync.RWMutex
	hosts map[string]Route // map[hostname]container_id
}

type (
	ContainersRouter interface {
		// RouteToContainerByHostname returns a URL to the container with the given hostname (from the docker
		// label, of course).
		RouteToContainerByHostname(hostname string) (string, error)
	}

	Route struct {
		Scheme  string // http, https or something like that (default: http)
		Port    uint16 // port number (default: 80)
		Network string // network name (default: bridge)
		IPAddr  string // IP address
	}

	RouteBuilder func(Route) string
)

// ContainersRouteOption is a function that configures a ContainersRoute.
type ContainersRouteOption func(*ContainersRoute)

// WithContainersRouteDefaultScheme sets the default Scheme for the containers' router.
func WithContainersRouteDefaultScheme(scheme string) ContainersRouteOption {
	return func(r *ContainersRoute) { r.defaults.Scheme = scheme }
}

// WithContainersRouteDefaultPort sets the default Port for the containers' router.
func WithContainersRouteDefaultPort(port uint16) ContainersRouteOption {
	return func(r *ContainersRoute) { r.defaults.Port = port }
}

// WithContainersRouteDefaultNetwork sets the default Network for the containers' router.
func WithContainersRouteDefaultNetwork(network string) ContainersRouteOption {
	return func(r *ContainersRoute) { r.defaults.Network = network }
}

// WithContainersRouteBuilder sets the RouteBuilder for the containers' router.
func WithContainersRouteBuilder(b RouteBuilder) ContainersRouteOption {
	return func(r *ContainersRoute) { r.builder = b }
}

// defaultRouteBuilder is used by default to build the route.
var defaultRouteBuilder RouteBuilder = func(r Route) string {
	var s strings.Builder
	s.Grow(len(r.Scheme) + 3 + len(r.IPAddr) + 1 + 5)

	s.WriteString(r.Scheme)
	s.WriteString("://")
	s.WriteString(r.IPAddr)
	s.WriteString(":")
	s.WriteString(strconv.FormatUint(uint64(r.Port), 10))

	return s.String()
}

// NewContainersRoute creates a new ContainersRoute.
func NewContainersRoute(opt ...ContainersRouteOption) *ContainersRoute {
	var router = ContainersRoute{
		hosts:   make(map[string]Route),
		builder: defaultRouteBuilder,
	}

	router.defaults.Scheme = "http"
	router.defaults.Port = 80
	router.defaults.Network = "bridge"

	for _, option := range opt {
		option(&router)
	}

	return &router
}

// Watch starts watching for changes in the docker containers and updates the internal routing table.
func (r *ContainersRoute) Watch(ctx context.Context, watcher ContainersWatcher) error {
	const (
		labelsPrefix = "indocker."

		labelHost    = labelsPrefix + "host"
		labelPort    = labelsPrefix + "Port"
		labelNetwork = labelsPrefix + "Network"
		labelScheme  = labelsPrefix + "Scheme"
	)

	// create a subscription channel
	var sub = make(ContainersSubscription)
	defer close(sub)

	// subscribe to updates
	if err := watcher.Subscribe(sub); err != nil {
		return err
	}

	// unsubscribe from updates
	defer func() { _ = watcher.Unsubscribe(sub) }()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case containers := <-sub:
			r.mu.Lock()

			// clear the map
			for k := range r.hosts {
				delete(r.hosts, k)
			}

			// fill the map with new values
			for _, c := range containers {
				// check if the container has a host label (this is the only required label)
				if host, found := c.Labels[labelHost]; found {
					const fullHostSuffix = ".indocker.app"

					if strings.HasSuffix(host, fullHostSuffix) {
						host = strings.TrimSuffix(host, fullHostSuffix)
					}

					var newRoute = r.defaults

					// check if the container has other labels
					if v, ok := c.Labels[labelScheme]; ok {
						newRoute.Scheme = v
					}

					if v, ok := c.Labels[labelPort]; ok {
						if p, err := strconv.ParseUint(v, 10, 16); err == nil {
							newRoute.Port = uint16(p)
						}
					}

					if v, ok := c.Labels[labelNetwork]; ok {
						newRoute.Network = v
					}

					// without this check, the container may not have any networks
					if c.NetworkSettings != nil && len(c.NetworkSettings.Networks) > 0 {
						var net *network.EndpointSettings

						// check if the container has the Network we need
						if namedNet, ok := c.NetworkSettings.Networks[newRoute.Network]; ok {
							net = namedNet
						} else {
							// pick random value from the map
							for _, rndNet := range c.NetworkSettings.Networks {
								net = rndNet

								break
							}
						}

						if net != nil {
							newRoute.IPAddr = net.IPAddress

							// add the Route to the map
							r.hosts[host] = newRoute
						}
					}
				}
			}

			r.mu.Unlock()
		}
	}
}

// RouteToContainerByHostname returns a URL to the container with the given hostname (from the docker label, of course).
func (r *ContainersRoute) RouteToContainerByHostname(hostname string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.hosts) == 0 {
		return "", errors.New("no routes registered")
	}

	if rt, exists := r.hosts[hostname]; exists {
		return r.builder(rt), nil
	}

	return "", errors.New("no Route found for " + hostname)
}
