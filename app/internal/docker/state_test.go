package docker_test

import (
	"context"
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gh.tarampamp.am/indocker-app/app/internal/docker"
)

func TestNewState(t *testing.T) {
	t.Parallel()

	dc, dcErr := client.NewClientWithOpts(client.WithHost("unix:///var/run/docker.sock"))
	require.NoError(t, dcErr)

	var state = docker.NewState(dc)

	assert.NoError(t, state.Update(context.Background()))

	t.Logf("%+v", state.AllContainerURLs())
}
