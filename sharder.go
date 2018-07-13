package main

import (
	"crypto/sha1"
	"fmt"
	"path"
)

// Sharder of keys.
type Sharder interface {
	NumShards() uint16
	LockPath(shard uint16) string
	ShardOf(key string) uint16
	Acquire(shard uint16) error
	Release(shard uint16) error
}

// NewSharder instance.
func NewSharder(numShards uint16, fs FileSystem, basePath string) Sharder {
	sharder := &sharder{
		fs:        fs,
		basePath:  basePath,
		numShards: numShards,
		locks:     map[uint16]File{},
	}

	return sharder
}

type sharder struct {
	numShards uint16 // readonly
	basePath  string // readonly
	fs        FileSystem

	locks map[uint16]File
}

func (s *sharder) NumShards() uint16 { return s.numShards }

func (s *sharder) LockPath(shard uint16) string {
	return path.Join(s.basePath, fmt.Sprintf(".shards/%d", shard))
}

func (s *sharder) ShardOf(key string) uint16 {
	hash := sha1.Sum([]byte(key))
	value := uint16(hash[0]) | uint16(hash[1])<<8 // binary.LittleEndian.Uint16(...)
	shard := value % s.numShards
	return shard
}

func (s *sharder) Acquire(shard uint16) error {
	// done if already acquired (cached)
	if _, ok := s.locks[shard]; ok {
		return nil
	}

	// try acquire lock on shard
	lockPath := s.LockPath(shard)
	f, err := s.fs.OpenFile(lockPath, fileFlags, fileMode)
	if err != nil {
		return err
	}

	if err = f.Lock(); err != nil {
		f.Close()
		return err
	}

	// cache lock
	s.locks[shard] = f
	return nil
}

func (s *sharder) Release(shard uint16) error {
	// done if not acquired (cached)
	f, ok := s.locks[shard]
	if !ok {
		return nil
	}

	// try unlock shard
	if err := f.Unlock(); err != nil {
		return err
	}

	// remove lock from cache
	f.Close()
	delete(s.locks, shard)
	return nil
}
