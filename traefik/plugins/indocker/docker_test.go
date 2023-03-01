package indocker_test

import (
	"context"
	"testing"
	"time"

	"indocker"
)

func TestDocker_Watch(t *testing.T) {
	var (
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*8)
		d           = indocker.NewDocker("/var/run/docker.sock", indocker.WithSnapshotsCapacity(3))
	)

	defer cancel()

	go d.Watch(ctx, time.Millisecond*499)

loop:
	for {
		select {
		case <-ctx.Done():
			break loop

		case <-time.After(time.Millisecond * 500):
			t.Log(d.Snapshots())
		}
	}
}
