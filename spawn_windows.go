//go:build windows

package main

import "os/exec"

// spawnWindow launches a new vibemd window process.
func spawnWindow(exe string, args ...string) error {
	return exec.Command(exe, args...).Start()
}
