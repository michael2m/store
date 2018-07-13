package main

import (
	"fmt"
	"path"
	"sync"
)

// TODO: upon file close, ensure consistent FS cache!

// API represents the logical file system API.
type API interface {
	Open(path string) (File, error)
	Close(path string) error
	Remove(path string) error
}

// NewAPI returns a file system API instance.
func NewAPI(mutex sync.Locker, ring Ring, sharder Sharder, cache Cache, fs FileSystem) API {
	return &api{
		mutex:   mutex,
		ring:    ring,
		sharder: sharder,
		cache:   cache,
		fs:      fs,
	}
}

type api struct {
	mutex   sync.Locker
	ring    Ring
	sharder Sharder
	cache   Cache
	fs      FileSystem
}

func (a *api) Open(filePath string) (File, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	shard := a.sharder.ShardOf(filePath)
	node := a.ring.NodeOf(shard)
	if node != a.ring.Self() {
		return nil, fmt.Errorf("ErrOtherNode")
	}

	err := a.sharder.Acquire(shard)
	if err != nil {
		return nil, fmt.Errorf("ErrShardLock")
	}

	if item, ok := a.cache.Get(filePath); ok {
		return item.(File), nil
	}

	err = a.fs.MkdirAll(path.Dir(filePath), dirMode)
	if err != nil {
		return nil, err
	}

	f, err := a.fs.OpenFile(filePath, fileFlags, fileMode)
	if err != nil {
		return nil, err
	}

	a.cache.Add(filePath, f)
	return f, nil
}

func (a *api) Close(filePath string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	shard := a.sharder.ShardOf(filePath)
	node := a.ring.NodeOf(shard)
	if node != a.ring.Self() {
		return fmt.Errorf("ErrOtherNode")
	}

	err := a.sharder.Acquire(shard)
	if err != nil {
		return fmt.Errorf("ErrShardLock")
	}

	if item, ok := a.cache.Remove(filePath); ok {
		item.(File).Close()
	}
	return nil
}

func (a *api) Remove(filePath string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	shard := a.sharder.ShardOf(filePath)
	node := a.ring.NodeOf(shard)
	if node != a.ring.Self() {
		return fmt.Errorf("ErrOtherNode")
	}

	err := a.sharder.Acquire(shard)
	if err != nil {
		return fmt.Errorf("ErrShardLock")
	}

	if item, ok := a.cache.Remove(filePath); ok {
		item.(File).Close()
	}

	err = a.fs.RemoveAll(filePath)
	if err != nil {
		return fmt.Errorf("ErrRemoveFile")
	}
	return nil
}
