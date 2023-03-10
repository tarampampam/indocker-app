package collector

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

type UIDResolver interface {
	// Resolve returns a unique ID for the host.
	Resolve() (string, error)
}

// DockerIDResolver resolves the host ID from the Docker daemon.
type DockerIDResolver struct {
	ctx    context.Context
	client *client.Client
}

var _ UIDResolver = (*DockerIDResolver)(nil) // ensure interface is implemented

// NewDockerIDResolver creates a new DockerIDResolver.
func NewDockerIDResolver(ctx context.Context, opts ...client.Opt) (*DockerIDResolver, error) {
	c, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	return &DockerIDResolver{ctx: ctx, client: c}, nil
}

// Resolve returns a unique ID for the host.
func (r *DockerIDResolver) Resolve() (string, error) {
	// https://docs.docker.com/engine/api/v1.42/#tag/System/operation/SystemInfo
	info, err := r.client.Info(r.ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to get docker info")
	}

	return info.ID, nil
}
