package docker_test

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"gh.tarampamp.am/indocker-app/daemon/docker"
)

func TestDocker_WatchContainers(t *testing.T) {
	var (
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*8)
		w, err      = docker.NewDocker(time.Second, client.WithHost("unix:///var/run/docker.sock"))
	)

	defer cancel()

	require.NoError(t, err)

	go w.Watch(context.Background())

loop:
	for {
		select {
		case <-ctx.Done():
			break loop

		case <-time.After(time.Millisecond * 500):
			t.Log(w.Alive())
		}
	}
}
