package docker_test

import (
	"context"
	"runtime"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"gh.tarampamp.am/indocker-app/daemon/internal/docker"
)

func TestContainersRoute_RouteToContainer(t *testing.T) {
	defer goleak.VerifyNone(t)

	var (
		router  = docker.NewContainersRoute()
		watcher = &watcherMock{}
	)

	// create a context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// run the watcher in a goroutine
	go func() { assert.ErrorIs(t, router.Watch(ctx, watcher), context.Canceled) }()

	runtime.Gosched() // wait for the router to start

	route, err := router.RouteToContainerByHostname("foo")
	assert.Empty(t, route)
	assert.ErrorContains(t, err, "no routes registered")

	watcher.Push(map[string]types.Container{
		"3176a2479c92": {
			ID: "3176a2479c92",
			Labels: map[string]string{
				"indocker.host":    "foo",
				"indocker.Port":    "123",
				"indocker.Network": "bar-Network",
				"indocker.Scheme":  "ftp",
			},
			NetworkSettings: &types.SummaryNetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"bar-Network": {
						IPAddress: "1.2.3.4",
					},
				},
			},
		},
		"4cb07b47f9fb": {
			ID: "4cb07b47f9fb",
			Labels: map[string]string{
				"indocker.host": "bar",
			},
			NetworkSettings: &types.SummaryNetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"blah-blah-Network": {
						IPAddress: "3.4.5.6",
					},
				},
			},
		},
	})

	runtime.Gosched() // wait for the router to process the update

	route, err = router.RouteToContainerByHostname("foo")
	assert.NoError(t, err)
	assert.Equal(t, "ftp://1.2.3.4:123", route)

	// test the default values
	route, err = router.RouteToContainerByHostname("bar")
	assert.NoError(t, err)
	assert.Equal(t, "http://3.4.5.6:80", route)

	route, err = router.RouteToContainerByHostname("baz")
	assert.Empty(t, route)
	assert.ErrorContains(t, err, "no Route found")
}
