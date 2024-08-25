package docker

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	dc "github.com/docker/docker/client"
)

type (
	routesMap = map[string]map[string]url.URL // map[hostname]map[container_id]url.URL

	State struct {
		dc *dc.Client

		routesMu sync.Mutex // protects routes
		routes   routesMap  // containers routing, map[hostname]url.URL

		routeChangesSubsMu sync.Mutex                       // protects subs
		routeChangesSubs   map[chan routesMap]chan struct{} // map[subscription]stop_channel
	}
)

func NewState(dc *dc.Client) *State {
	return &State{
		dc:               dc,
		routes:           make(routesMap),
		routeChangesSubs: make(map[chan routesMap]chan struct{}),
	}
}

// StartAutoUpdate starts an automatic update of the state of running containers, using the docker events API.
// It returns a function to stop the updating process.
func (s *State) StartAutoUpdate(ctx context.Context) (stop func()) { //nolint:gocognit
	var filter = filters.NewArgs()

	// https://docs.docker.com/engine/api/v1.46/#tag/System/operation/SystemEvents
	// all types = (container|image|volume|network|daemon|plugin|node|service|secret|config)
	for _, eventType := range []string{"container", "network", "daemon", "service"} {
		filter.Add("type", eventType)
	}

	var eventsCtx, cancel = context.WithCancel(ctx)

	go func() {
		var (
			eventsCh <-chan events.Message
			errorsCh <-chan error
		)

		for { // this loop is needed to re-open the stream in the event of an error
			if eventsCtx.Err() != nil {
				return
			}

			// (re)create the stream
			eventsCh, errorsCh = s.dc.Events(eventsCtx, events.ListOptions{Filters: filter})

		readingLoop:
			for {
				select {
				case <-eventsCtx.Done():
					return
				case <-eventsCh:
					for range 2 { // retry the update on error (max 2 times)
						if updErr := s.Update(eventsCtx); updErr == nil {
							break
						}

						if eventsCtx.Err() != nil {
							return
						}
					}
				case <-errorsCh:
					break readingLoop // re-open the stream
				}
			}
		}
	}()

	return sync.OnceFunc(cancel)
}

// Update updates the state of running containers immediately. It returns an error if something went wrong.
func (s *State) Update(ctx context.Context) error { //nolint:funlen,gocyclo
	var filter = filters.NewArgs()

	// we need to filter only certain statuses (alive containers)
	// all statuses = (created|restarting|running|removing|paused|exited|dead)
	for _, status := range []string{"created", "restarting", "running", "removing", "paused"} {
		filter.Add("status", status)
	}

	var options = container.ListOptions{Filters: filter}

	// get the list of containers
	// https://docs.docker.com/engine/api/v1.46/#tag/Container/operation/ContainerList
	list, listErr := s.dc.ContainerList(ctx, options)
	if listErr != nil {
		return listErr
	}

	var newRoutes = make(routesMap, len(list))

	for _, listedContainer := range list {
		// set the routing info, if possible
		if scheme, hostname, ipAddr, port, found := s.buildRouteToContainer(listedContainer); found {
			if scheme != "" && hostname != "" && ipAddr != "" && port != 0 { // an additional check
				var u = url.URL{
					Scheme: scheme,
					Host:   fmt.Sprintf("%s:%d", ipAddr, port),
				}

				if _, ok := newRoutes[hostname]; !ok {
					newRoutes[hostname] = make(map[string]url.URL)
				}

				newRoutes[hostname][listedContainer.ID] = u
			}
		}
	}

	var routesUpdated bool

	s.routesMu.Lock()
	routesUpdated = !reflect.DeepEqual(s.routes, newRoutes) // check if the routes have been updated
	clear(s.routes)                                         // care about the memory
	s.routes = newRoutes                                    // update the routes
	s.routesMu.Unlock()

	if routesUpdated {
		s.routeChangesSubsMu.Lock()
		for sub, stop := range s.routeChangesSubs {
			go func(sub chan<- routesMap, stop <-chan struct{}) {
				select { // first, check if the subscription and context are still alive
				case <-stop: // is the subscription stopped?
				case <-ctx.Done(): // is the context done?
				default:
					select { // then, notify the subscribers
					case <-stop: // is the subscription stopped?
					case <-ctx.Done(): //  is the context done?
					case sub <- maps.Clone(newRoutes): // notify the subscriber
					}
				}
			}(sub, stop)
		}
		s.routeChangesSubsMu.Unlock()
	}

	return nil
}

// SubscribeForRoutingUpdates will return a subscription channel and a stop function. The subscription channel will
// receive a message when the routing info is updated. The stop function will stop the subscription.
// The subscription channel will be closed when the stop function is called.
func (s *State) SubscribeForRoutingUpdates() (sub <-chan routesMap, stop func()) {
	var ch, cancel = make(chan routesMap, 1), make(chan struct{})

	sub, stop = ch, sync.OnceFunc(func() {
		close(cancel) // close the stop channel will notify the subscription to stop

		s.routeChangesSubsMu.Lock()
		delete(s.routeChangesSubs, ch) // remove the subscription from the list
		s.routeChangesSubsMu.Unlock()

		// empty the channel
	emptyLoop:
		for {
			select {
			case <-ch:
			default:
				close(ch)

				break emptyLoop
			}
		}
	})

	s.routeChangesSubsMu.Lock()
	s.routeChangesSubs[ch] = cancel // add the subscription to the list
	s.routeChangesSubsMu.Unlock()

	return
}

// URLToContainerByHostname returns a URL to the container with the given hostname. It returns false if the container
// with the given hostname is not found.
func (s *State) URLToContainerByHostname(hostname string) (map[string]url.URL, bool) { // map[container_id]url.URL
	{ // normalize the hostname
		hostname = strings.ToLower(strings.TrimSpace(hostname))

		// drop the ".indocker.app" if it exists
		if withoutPostfix, cut := strings.CutSuffix(hostname, ".indocker.app"); cut {
			hostname = withoutPostfix
		}
	}

	s.routesMu.Lock()
	u, ok := s.routes[hostname]
	s.routesMu.Unlock()

	if ok {
		return u, true
	}

	return nil, false
}

// AllContainerURLs returns a map of all container URLs.
func (s *State) AllContainerURLs() (routes map[string]map[string]url.URL) { // map[hostname]map[container_id]url.URL
	s.routesMu.Lock()
	routes = maps.Clone(s.routes)
	s.routesMu.Unlock()

	return
}

//nolint:gochecknoglobals
var (
	hostLabels        = []string{"indocker.host", "indocker.hostname", "host", "hostname"}
	schemeLabels      = []string{"indocker.scheme", "indocker.schema", "scheme", "schema"}
	portLabels        = []string{"indocker.port", "port"}
	networkNameLabels = []string{"indocker.network", "indocker.net", "network", "net"}
)

// buildRouteToContainer returns the routing info to the container, if possible. It returns false if the container
// does not have the required labels or the network settings.
func (s *State) buildRouteToContainer(info types.Container) ( //nolint:funlen,gocognit,gocyclo
	scheme, host, ipAddr string, port uint16, found bool,
) {
	scheme, port = "http", uint16(80) //nolint:mnd // defaults

	// determine the host
	for _, wantHostLabel := range hostLabels {
		if v, ok := info.Labels[wantHostLabel]; ok {
			v = strings.ToLower(strings.TrimSpace(v))

			if v == "" {
				continue
			}

			// drop the ".indocker.app" if it exists
			if withoutPostfix, cut := strings.CutSuffix(v, ".indocker.app"); cut {
				host = withoutPostfix
			} else {
				host = v
			}

			break
		}
	}

	// determine the scheme
	for _, wantSchemeLabel := range schemeLabels {
		if v, ok := info.Labels[wantSchemeLabel]; ok {
			v = strings.ToLower(strings.TrimSpace(v))

			if v == "" {
				continue
			} else {
				if v == "https" {
					port = 443 // in case of https, set the default port to 443
				}

				scheme = v
			}

			break
		}
	}

	// determine the port
	for _, wantPortLabel := range portLabels {
		if v, ok := info.Labels[wantPortLabel]; ok {
			v = strings.TrimSpace(v)

			if v == "" {
				continue
			}

			// parse the port
			if parsed, parseErr := strconv.ParseUint(v, 10, 16); parseErr == nil {
				port = uint16(parsed) //nolint:gosec
			}

			break
		}
	}

	var netName = "bridge" // defaults

	// determine the network name
	for _, wantNetLabel := range networkNameLabels {
		if v, ok := info.Labels[wantNetLabel]; ok {
			v = strings.TrimSpace(v)

			if v == "" {
				continue
			}

			netName = v

			break
		}
	}

	// only if the host is set
	if host != "" { //nolint:nestif
		// check if the container has the networks at all
		if info.NetworkSettings != nil && len(info.NetworkSettings.Networks) > 0 {
			var net *network.EndpointSettings

			// check if the container has the required network
			if namedNet, ok := info.NetworkSettings.Networks[netName]; ok {
				net = namedNet // pick it
			} else {
				// if the container has multiple networks, but the required one is not found - pick a random one
				for _, rndNet := range info.NetworkSettings.Networks {
					net = rndNet

					break
				}
			}

			// and only if the network is set
			if net != nil && net.IPAddress != "" {
				// we can determine the IP address of the container
				ipAddr = net.IPAddress

				// and return the result
				return scheme, host, ipAddr, port, true
			}
		}
	}

	return "", "", "", 0, false
}
