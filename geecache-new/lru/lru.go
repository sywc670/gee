package lru

import "container/list"

type Cache struct {
	maxBytes   int64
	usedBytes  int64
	list       *list.List
	cache      map[string]*list.Element
	OnEviction func(key string, value Value)
}

type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.list.Len()
}

func New(cacheMaxBytes int64, oneviction func(string, Value)) *Cache {
	return &Cache{
		maxBytes:   cacheMaxBytes,
		list:       list.New(),
		cache:      make(map[string]*list.Element),
		OnEviction: oneviction,
	}
}

func (c *Cache) Add(key string, value Value) {
	if elem, ok := c.cache[key]; ok {
		c.list.MoveToFront(elem)
		entry := elem.Value.(*entry)
		c.usedBytes += int64(value.Len()) - int64(entry.value.Len())
		entry.value = value
	} else {
		elem := c.list.PushFront(&entry{key, value})
		c.cache[key] = elem
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}
	for c.usedBytes > c.maxBytes && c.maxBytes != 0 {
		c.RemoveOldest()
	}

}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if elem, ok := c.cache[key]; ok {
		c.list.MoveToFront(elem)
		entry := elem.Value.(*entry)
		return entry.value, ok
	}
	return nil, false
}

func (c *Cache) RemoveOldest() {
	elem := c.list.Back()
	if elem != nil {
		entry := elem.Value.(*entry)
		c.list.Remove(elem)
		c.usedBytes -= int64(entry.value.Len()) + int64(len(entry.key))
		delete(c.cache, entry.key)
		if c.OnEviction != nil {
			c.OnEviction(entry.key, entry.value)
		}
	}
}
