// +build linux

package main

import "golang.org/x/sys/unix"

func (f *file) Lock() error {
	return unix.Flock(int(f.File.Fd()), unix.LOCK_EX|unix.LOCK_NB)
}

func (f *file) Unlock() error {
	return unix.Flock(int(f.File.Fd()), unix.LOCK_UN)
}
