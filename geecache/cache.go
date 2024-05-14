package geecache

import (
	"sync"

	"github.com/sywc670/gee/geecache/lru"
)

// support concurrent
type cache struct {
	mu         sync.Mutex
	cacheBytes int64
	lru        *lru.Cache
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	v, ok := c.lru.Get(key)
	if ok {
		return v.(ByteView), ok
	}
	return
}
