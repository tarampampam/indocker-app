package collector_test

import (
	"context"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/app/internal/collector"
)

type senderMock struct {
	sync.Mutex
	sent []collector.Event
}

func (s *senderMock) Send(_ context.Context, _ string, e ...collector.Event) error {
	s.Lock()
	s.sent = append(s.sent, e...)
	s.Unlock()

	return nil
}

func (s *senderMock) Sent() []collector.Event {
	s.Lock()
	defer s.Unlock()

	return s.sent
}

type uidResolverMock struct{}

func (uidResolverMock) Resolve() (string, error) {
	return "test", nil
}

func TestRealCollector_Schedule(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		sender = senderMock{}
		c      = collector.NewCollector(ctx, zap.NewNop(), time.Millisecond, time.Millisecond*10, &sender, &uidResolverMock{})
	)

	defer c.Stop()

	assert.Empty(t, sender.Sent())

	c.Schedule(collector.Event{Name: "test"})
	pause(t, time.Millisecond*20)

	assert.Len(t, sender.Sent(), 1)

	c.Schedule(collector.Event{Name: "test2"})
	c.Schedule(collector.Event{Name: "test3"})

	assert.Len(t, sender.Sent(), 1) // not sent yet

	pause(t, time.Millisecond*20)

	assert.Len(t, sender.Sent(), 3) // sent

	var (
		batch = make([]collector.Event, 800)
		now   = time.Now()
	)

	for i := 0; i < len(batch); i++ {
		batch[i] = collector.Event{Name: strconv.Itoa(i), Timestamp: now}
	}

	c.Schedule(batch...)

	pause(t, time.Millisecond*20)

	assert.Len(t, sender.Sent(), 803)

	sent := sender.Sent()[3:]

	sort.SliceStable(sent, func(i, j int) bool { // sort by names
		numA, _ := strconv.Atoi(sent[i].Name)
		numB, _ := strconv.Atoi(sent[j].Name)

		return numA < numB
	})

	assert.EqualValues(t, batch, sent)
}

func pause(t *testing.T, d time.Duration) {
	t.Helper()

	defer runtime.Gosched()

	timer := time.NewTimer(d)
	defer timer.Stop()

	<-timer.C
}
