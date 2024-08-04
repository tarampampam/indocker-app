package docker

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/network"
)

type ContainersRouter interface {
	// RouteToContainerByHostname returns a URL to the container with the given hostname (from the docker
	// label, of course).
	RouteToContainerByHostname(hostname string) (*Route, error)

	// Routes returns a copy of the internal routing table.
	Routes() map[string]Route
}

type (
	// ContainersRoute is a docker containers router. It allows you to get the Route of a docker container by its labels.
	ContainersRoute struct {
		defaults Route

		mu    sync.RWMutex
		hosts map[string]Route // map[hostname]container_id
	}

	Route struct {
		Scheme  string // http, https or something like that (default: http) // TODO: is this needed?
		Port    uint16 // port number (default: 80)
		Network string // network name (default: bridge)
		IPAddr  string // IP address
	}
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

// NewContainersRoute creates a new ContainersRoute.
func NewContainersRoute(opt ...ContainersRouteOption) *ContainersRoute {
	var router = ContainersRoute{
		hosts: make(map[string]Route),
	}

	router.defaults.Scheme = "http" //nolint:goconst
	router.defaults.Port = 80
	router.defaults.Network = "bridge"

	for _, option := range opt {
		option(&router)
	}

	return &router
}

const fullHostSuffix = ".indocker.app"

// Watch starts watching for changes in the docker containers and updates the internal routing table.
func (r *ContainersRoute) Watch(ctx context.Context, watcher ContainersWatcher) error { //nolint:funlen,gocognit,gocyclo,lll
	const (
		labelHost    = "indocker.host"
		labelPort    = "indocker.port"
		labelNetwork = "indocker.network"
		labelScheme  = "indocker.scheme"
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
				if host, found := c.Labels[labelHost]; found { //nolint:nestif
					// drop the ".indocker.app" if it exists
					if strings.HasSuffix(strings.ToLower(host), fullHostSuffix) {
						host = host[:len(host)-len(fullHostSuffix)]
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
							for name, rndNet := range c.NetworkSettings.Networks {
								newRoute.Network, net = name, rndNet

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

var (
	ErrNoRegisteredRoutes = errors.New("no routes registered")
	ErrNoRouteFound       = errors.New("no route found")
)

// RouteToContainerByHostname returns a URL to the container with the given hostname (from the docker label, of course).
func (r *ContainersRoute) RouteToContainerByHostname(hostname string) (*Route, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.hosts) == 0 {
		return nil, ErrNoRegisteredRoutes
	}

	// drop the ".indocker.app" if it exists
	if strings.HasSuffix(strings.ToLower(hostname), fullHostSuffix) {
		hostname = hostname[:len(hostname)-len(fullHostSuffix)]
	}

	if route, exists := r.hosts[hostname]; exists {
		return &route, nil
	}

	return nil, ErrNoRouteFound
}

// RoutesCount returns the number of registered routes (hosts).
func (r *ContainersRoute) RoutesCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.hosts)
}

// Routes returns a copy of the internal routing table.
func (r *ContainersRoute) Routes() map[string]Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var routes = make(map[string]Route, len(r.hosts))
	for host, route := range r.hosts { // copy the map
		routes[host] = route
	}

	return routes
}
