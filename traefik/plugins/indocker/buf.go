package indocker

import "sync"

type SizeLimitedBuf struct {
	mu   sync.RWMutex
	cap  int
	data []any
}

func NewSizeLimitedBuf(cap uint) *SizeLimitedBuf {
	return &SizeLimitedBuf{
		cap:  int(cap),
		data: make([]any, 0, cap),
	}
}

// Add adds a value to the buffer. If the buffer is full, the oldest value is removed.
func (b *SizeLimitedBuf) Add(v any) {
	b.mu.RLock()
	var l = len(b.data)
	b.mu.RUnlock()

	if l >= b.cap {
		b.mu.Lock()
		b.data = append(b.data[1:len(b.data)], v)
		b.mu.Unlock()

		return
	}

	b.data = append(b.data, v)
}

// Get returns a copy of the data.
func (b *SizeLimitedBuf) Get() []any {
	b.mu.RLock()
	var data = make([]any, len(b.data))
	copy(data, b.data)
	b.mu.RUnlock()

	return data
}
