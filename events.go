package main

import (
	"fmt"
	"sync"

	"github.com/hashicorp/memberlist"
)

// Join node
type Join string

// Leave node
type Leave string

// Events related membership.
type Events interface {
	memberlist.EventDelegate
}

// NewEvents returns an events handler instance.
func NewEvents(mutex sync.Locker, ring Ring, sharder Sharder, cache Cache) Events {
	return &events{
		mutex:   mutex,
		ring:    ring,
		sharder: sharder,
		cache:   cache,
	}
}

type events struct {
	mutex   sync.Locker
	ring    Ring
	sharder Sharder
	cache   Cache
}

func (e *events) update() {
	var waiter sync.WaitGroup

	waiter.Add(1)
	go func() {
		defer waiter.Done()
		e.release()
	}()

	waiter.Add(1)
	go func() {
		defer waiter.Done()
		e.evict()
	}()

	waiter.Wait()
}

func (e *events) release() {
	// release "old" shards for self (node)
	for shard := uint16(0); shard < e.sharder.NumShards(); shard++ {
		node := e.ring.NodeOf(shard)
		if node != e.ring.Self() {
			e.sharder.Release(shard)
		}
	}
}

func (e *events) evict() {
	// evict files not for shards of self (node)
	for _, path := range e.cache.List() {
		shard := e.sharder.ShardOf(path)
		node := e.ring.NodeOf(shard)
		if node != e.ring.Self() {
			if f, ok := e.cache.Get(path); ok {
				f.Close()
			}

			e.cache.Remove(path)
		}
	}
}

func (e *events) NotifyJoin(node *memberlist.Node) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	fmt.Printf("JOIN %s\n", node)

	e.ring.Add(node.Name)
	e.update()
}

func (e *events) NotifyLeave(node *memberlist.Node) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	fmt.Printf("LEAVE %s\n", node)

	e.ring.Remove(node.Name)
	e.update()
}

func (e *events) NotifyUpdate(node *memberlist.Node) {}
