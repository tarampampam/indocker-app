package collector

import (
	"context"
	"sync"
	"time"
)

type Collector interface {
	// Schedule schedules event to be sent.
	Schedule(...Event)

	// Stop stops collector.
	Stop()
}

// NoopCollector is a collector that does nothing.
type NoopCollector struct{}

func (*NoopCollector) Schedule(...Event) {}
func (*NoopCollector) Stop()             {}

// RealCollector is a collector that sends events to the remote server.
type RealCollector struct {
	ctx       context.Context
	initDelay time.Duration
	interval  time.Duration
	sender    Sender
	uid       string

	queueMu sync.Mutex
	queue   []Event

	stopOnce sync.Once
	stop     chan struct{}
}

// NewCollector creates new collector. It starts worker goroutine, if user ID resolved and hashed.
// Otherwise, it returns NoopCollector. If user ID can't be resolved, it uses "unknown" as user ID.
//
// Note: Do fot forget to call Stop() method to stop worker goroutine.
func NewCollector(ctx context.Context, initDelay, interval time.Duration, sender Sender, uid UIDResolver) Collector {
	id, err := uid.Resolve()
	if err != nil || id == "" {
		id = "unknown"
	} else {
		if hashed, hErr := HashParts(id); hErr == nil {
			id = hashed
		} else {
			return &NoopCollector{} // do not send events if we can't hash uid
		}
	}

	var c = RealCollector{
		ctx:       ctx,
		initDelay: initDelay,
		interval:  interval,
		sender:    sender,
		uid:       id,

		queue: make([]Event, 0),
		stop:  make(chan struct{}),
	}

	go c.worker() // start worker

	return &c
}

// worker periodically sends events to the remote server.
func (c *RealCollector) worker() {
	defer c.Stop()

	var t = time.NewTimer(c.initDelay)
	defer t.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return

		case <-c.stop:
			return

		case <-t.C:
			c.queueMu.Lock()
			if len(c.queue) > 0 {
				const chunkSize = 49

				var chunks [][]Event

				for i := 0; i < len(c.queue); i += chunkSize {
					end := i + chunkSize

					if end > len(c.queue) {
						end = len(c.queue)
					}

					chunks = append(chunks, c.queue[i:end])
				}

				for _, chunk := range chunks {
					go func(events []Event) {
						_ = c.sender.Send(c.ctx, c.uid, events...) // TODO: log error
					}(chunk)
				}

				c.queue = make([]Event, 0, cap(c.queue)) // reset queue
			}
			c.queueMu.Unlock()

			t.Reset(c.interval)
		}
	}
}

// Schedule schedules event to be sent.
func (c *RealCollector) Schedule(events ...Event) {
	if len(events) > 0 {
		for i := range events {
			if events[i].Timestamp.IsZero() {
				events[i].Timestamp = time.Now()
			}
		}

		c.queueMu.Lock()
		c.queue = append(c.queue, events...)
		c.queueMu.Unlock()
	}
}

// Stop stops collector.
func (c *RealCollector) Stop() {
	c.stopOnce.Do(func() {
		close(c.stop)
	})
}
