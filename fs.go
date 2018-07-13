package main

import (
	"io"
	"os"
)

// File interface.
type File interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer

	Sync() error
	Stat() (os.FileInfo, error)

	Lock() error
	Unlock() error
}

type file struct {
	*os.File
}

// FileSystem interface.
type FileSystem interface {
	MkdirAll(path string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
	RemoveAll(path string) error
	Rename(oldPath, newPath string) error
}

// NewFileSystem returns the default OS-backed file system.
func NewFileSystem() FileSystem {
	return &osFileSystem{}
}

type osFileSystem struct{}

func (osFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (osFileSystem) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return &file{File: f}, nil
}

func (osFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (osFileSystem) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}
