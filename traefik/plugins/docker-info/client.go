package docker_info

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
)

type Client struct {
	io.Closer

	conn net.Conn
	http *http.Client
}

// NewClient creates a new client for the Docker API.
func NewClient(unixSocket string) (*Client, error) {
	return &Client{
		http: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, "unix", unixSocket)
				},
			},
		},
	}, nil
}

// Version returns the Docker version.
func (c *Client) Version(ctx context.Context) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/version", http.NoBody)
	if err != nil {
		return nil, 0, err
	}

	return c.makeRequest(req)
}

// ContainersList returns the list of containers.
func (c *Client) ContainersList(ctx context.Context) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/containers/json", http.NoBody)
	if err != nil {
		return nil, 0, err
	}

	return c.makeRequest(req)
}

// ContainerInspect returns the container details.
func (c *Client) ContainerInspect(ctx context.Context, containerID string) ([]byte, int, error) {
	if err := c.validateContainerID(containerID); err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"http://docker/containers/"+containerID+"/json",
		http.NoBody,
	)
	if err != nil {
		return nil, 0, err
	}

	return c.makeRequest(req)
}

// ContainerStats returns the container stats.
func (c *Client) ContainerStats(ctx context.Context, containerID string) ([]byte, int, error) {
	if err := c.validateContainerID(containerID); err != nil {
		return nil, 0, err
	}

	// https://github.com/moby/moby/blob/f34567bf41ad65102a8b1f05496dc92b500a3056/client/container_stats.go#L12
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		"http://docker/containers/"+containerID+"/stats?stream=0&one-shot=1",
		http.NoBody,
	)
	if err != nil {
		return nil, 0, err
	}

	return c.makeRequest(req)
}

// ContainerLogs returns the container logs.
func (c *Client) ContainerLogs(ctx context.Context, containerID string) ([]byte, int, error) {
	if err := c.validateContainerID(containerID); err != nil {
		return nil, 0, err
	}

	// https://github.com/moby/moby/blob/f34567bf41ad65102a8b1f05496dc92b500a3056/client/container_logs.go#L36
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		"http://docker/containers/"+containerID+"/logs?stdout=1&stderr=1&timestamps=0&details=1&follow=0",
		http.NoBody,
	)
	if err != nil {
		return nil, 0, err
	}

	return c.makeRequest(req)
}

// validateContainerID checks if the container ID is valid.
func (*Client) validateContainerID(id string) error {
	if len(id) < 2 {
		return errors.New("container ID is too short")
	} else if len(id) > 64 {
		return errors.New("container ID is too long")
	}

	for _, r := range id {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return errors.New("incorrect ID characters")
		}
	}

	return nil
}

// makeRequest makes a request to the Docker socket.
func (c *Client) makeRequest(req *http.Request) ([]byte, int, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return data, resp.StatusCode, nil
}

// Close closes the connection to the Docker socket.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}
