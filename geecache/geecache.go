package geecache

import (
	"fmt"
	"log"
	"sync"

	"github.com/sywc670/gee/geecache/geecachepb"
	"github.com/sywc670/gee/geecache/singleflight"
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
	mainCache    *cache
	getter       Getter
	name         string
	peers        PeerPicker
	singleflight *singleflight.Group
}

func NewGroup(name string, cacheSize int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	g := &Group{
		name:         name,
		getter:       getter,
		mainCache:    &cache{cacheBytes: cacheSize},
		singleflight: &singleflight.Group{},
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

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	// Use singleflight package to reduce redundant request.
	view, err := g.singleflight.Do(key, func() (any, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})
	return view.(ByteView), err
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	in := &geecachepb.Request{
		Group: g.name,
		Key:   key,
	}
	out := &geecachepb.Response{}
	err := peer.Get(in, out)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: out.GetValue()}, nil
}
