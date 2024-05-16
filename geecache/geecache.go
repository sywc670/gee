package geecache

import (
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(string) ([]byte, error)
}

type GetterFunc func(string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

type Group struct {
	mainCache *cache
	getter    Getter
	name      string
}

func NewGroup(name string, cacheSize int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: &cache{cacheBytes: cacheSize},
	}
	mu.Lock()
	defer mu.Unlock()
	groups[name] = g

	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]

	return g
}

func (g *Group) Get(key string) (value ByteView, err error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if value, ok := g.mainCache.get(key); ok {
		log.Println("[geecache] hit!")
		return value, nil
	}
	log.Println("[geecache] miss!")
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (value ByteView, err error) {
	b, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	// clone
	bv := ByteView{clonebytes(b)}
	g.populateCache(key, bv)
	return bv, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
