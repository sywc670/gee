package lru

import (
	"container/list"
)

type Cache struct {
	maxBytes int64
	// infinity capacity when set 0
	usedBytes  int64
	list       *list.List
	cache      map[string]*list.Element
	OnEviction func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEviction func(string, Value)) *Cache {
	return &Cache{
		maxBytes:   maxBytes,
		OnEviction: onEviction,
		list:       list.New(),
		cache:      make(map[string]*list.Element),
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if elem, ok := c.cache[key]; ok {
		c.list.MoveToFront(elem)
		e := elem.Value.(*entry)
		return e.value, true
	}
	return
}

func (c *Cache) RemoveOldest() {
	elem := c.list.Back()
	if elem != nil {
		c.list.Remove(elem)
		e := elem.Value.(*entry)
		c.usedBytes -= int64(e.value.Len()) + int64(len(e.key))
		delete(c.cache, e.key)
		if c.OnEviction != nil {
			c.OnEviction(e.key, e.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if elem, ok := c.cache[key]; ok {
		c.list.MoveToFront(elem)
		e := elem.Value.(*entry)
		c.usedBytes += int64(value.Len()) - int64(e.value.Len())
		e.value = value
	} else {
		elem := c.list.PushFront(&entry{key, value})
		c.cache[key] = elem
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.usedBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.list.Len()
}
