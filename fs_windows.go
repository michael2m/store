// +build windows

package main

import "fmt"

func (f *file) Lock() error {
	fmt.Println("** WINDOWS FILE LOCK (FAKE!)")
	return nil
}

func (f *file) Unlock() error {
	fmt.Println("** WINDOWS FILE UNLOCK (FAKE!)")
	return nil
}
