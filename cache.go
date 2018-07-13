package main

import (
	"sync"
)

// Cache for files by their paths.
type Cache interface {
	List() []string
	Get(path string) (file File, ok bool)
	Add(path string, file File)
	Remove(path string) (File, bool)
}

// NewCache instance.
func NewCache() Cache {
	return &cache{
		files: map[string]File{},
	}
}

type cache struct {
	mutex sync.RWMutex
	files map[string]File
}

func (c *cache) List() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	list := make([]string, 0, len(c.files))
	for path := range c.files {
		list = append(list, path)
	}
	return list
}

func (c *cache) Get(path string) (File, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if val, ok := c.files[path]; ok {
		return val.(File), true
	}
	return nil, false
}

func (c *cache) Add(path string, file File) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.files[path] = file
}

func (c *cache) Remove(path string) (File, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if val, ok := c.files[path]; ok {
		delete(c.files, path)
		return val.(File), true
	}
	return nil, false
}
