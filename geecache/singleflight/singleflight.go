package singleflight

import "sync"

type call struct {
	wg    sync.WaitGroup
	value any
	err   error
}

type Group struct {
	mu      sync.Mutex
	callers map[string]*call
}

// Do
// If this key has a call struct mapped to it, wait and return
// else execute passed funtion
func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()

	if g.callers == nil {
		g.callers = make(map[string]*call)
	}

	if c, ok := g.callers[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.value, c.err
	}

	c := &call{}
	g.callers[key] = c
	// add before unlock
	c.wg.Add(1)
	g.mu.Unlock()
	c.value, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	// delete mapping
	delete(g.callers, key)
	g.mu.Unlock()

	return c.value, c.err
}
