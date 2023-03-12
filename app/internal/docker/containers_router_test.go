package docker_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"gh.tarampamp.am/indocker-app/app/internal/docker"
)

func TestContainersRoute_RouteToContainerByHostname(t *testing.T) {
	defer goleak.VerifyNone(t)

	for name, testCase := range map[string]struct {
		giveEvents   map[string]types.Container
		giveHostname string
		wantRoute    *docker.Route
		wantErr      error
	}{
		"no containers": {
			giveHostname: "foo",
			wantErr:      docker.ErrNoRegisteredRoutes,
		},
		"one container with all properties": {
			giveEvents: map[string]types.Container{
				"3176a2479c92": {
					ID: "3176a2479c92",
					Labels: map[string]string{
						"indocker.host":    "foo",
						"indocker.port":    "123",
						"indocker.network": "bar-Network",
						"indocker.scheme":  "ftp",
					},
					NetworkSettings: &types.SummaryNetworkSettings{
						Networks: map[string]*network.EndpointSettings{
							"bar-Network": {IPAddress: "1.2.3.4"},
							"bridge":      {IPAddress: "18.8.8.18"},
						},
					},
				},
			},
			giveHostname: "foo",
			wantRoute: &docker.Route{
				Scheme:  "ftp",
				Port:    123,
				Network: "bar-Network",
				IPAddr:  "1.2.3.4",
			},
		},
		"check defaults": {
			giveEvents: map[string]types.Container{
				"4cb07b47f9fb": {
					ID: "4cb07b47f9fb",
					Labels: map[string]string{
						"indocker.host": "bar",
					},
					NetworkSettings: &types.SummaryNetworkSettings{
						Networks: map[string]*network.EndpointSettings{
							"un-existent-network": {
								IPAddress: "3.4.5.6",
							},
						},
					},
				},
			},
			giveHostname: "bar",
			wantRoute: &docker.Route{
				Scheme:  "http",
				Port:    80,
				Network: "un-existent-network",
				IPAddr:  "3.4.5.6",
			},
		},
		"with postfix in docker label": {
			giveEvents: map[string]types.Container{
				"4cb07b47f9fb": {
					ID: "4cb07b47f9fb",
					Labels: map[string]string{
						"indocker.host": "foo.InDoCkEr.aPp",
					},
					NetworkSettings: &types.SummaryNetworkSettings{
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "3.4.5.6",
							},
						},
					},
				},
			},
			giveHostname: "foo",
			wantRoute: &docker.Route{
				Scheme:  "http",
				Port:    80,
				Network: "bridge",
				IPAddr:  "3.4.5.6",
			},
		},
		"with postfix in hostname": {
			giveEvents: map[string]types.Container{
				"4cb07b47f9fb": {
					ID: "4cb07b47f9fb",
					Labels: map[string]string{
						"indocker.host": "foo",
					},
					NetworkSettings: &types.SummaryNetworkSettings{
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "3.4.5.6",
							},
						},
					},
				},
			},
			giveHostname: "foo.InDoCkEr.aPp",
			wantRoute: &docker.Route{
				Scheme:  "http",
				Port:    80,
				Network: "bridge",
				IPAddr:  "3.4.5.6",
			},
		},
		"not found": {
			giveEvents: map[string]types.Container{
				"4cb07b47f9fb": {
					ID: "4cb07b47f9fb",
					Labels: map[string]string{
						"indocker.host": "foo",
					},
					NetworkSettings: &types.SummaryNetworkSettings{
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "3.4.5.6",
							},
						},
					},
				},
			},
			giveHostname: "bar",
			wantErr:      docker.ErrNoRouteFound,
		},
	} {
		tt := testCase

		t.Run(name, func(t *testing.T) {
			var (
				router  = docker.NewContainersRoute()
				watcher = &watcherMock{}
			)

			// create a context with cancel
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// run the watcher in a goroutine
			go func() { assert.ErrorIs(t, router.Watch(ctx, watcher), context.Canceled) }()

			pause(10 * time.Millisecond) // wait for the router to start

			watcher.Push(tt.giveEvents)

			pause(10 * time.Millisecond) // wait for the router to process the update

			route, err := router.RouteToContainerByHostname(tt.giveHostname)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}

			assert.Equal(t, tt.wantRoute, route)
		})
	}
}

func pause(d time.Duration) {
	timer := time.NewTicker(d)
	<-timer.C
	timer.Stop()

	runtime.Gosched() // wait for the router to start
}
