package favicon

import (
	"image"
	"sync"
)

type cache struct {
	mu   sync.Mutex // protects data
	data map[string]image.Image
}

func newCache() *cache { return &cache{data: make(map[string]image.Image)} }

func (c *cache) Get(key string) (img image.Image, hit bool) {
	c.mu.Lock()
	img, hit = c.data[key]
	c.mu.Unlock()

	return
}

func (c *cache) Put(key string, img image.Image) {
	c.mu.Lock()
	c.data[key] = img
	c.mu.Unlock()
}

func (c *cache) Clear() {
	c.mu.Lock()
	clear(c.data)
	c.mu.Unlock()
}
