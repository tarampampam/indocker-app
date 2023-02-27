package indocker_test

import (
	"context"
	"testing"
	"time"

	"indocker"
)

func TestDocker_Watch(t *testing.T) {
	docker := indocker.NewDocker("/var/run/docker.sock")

	docker.Watch(context.Background(), time.Millisecond*3500)
}
