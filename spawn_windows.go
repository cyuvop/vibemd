//go:build windows

package main

import "os/exec"

// spawnWindow launches a new vibemd window process.
func spawnWindow(exe, file string) error {
	return exec.Command(exe, file).Start()
}
