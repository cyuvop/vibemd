//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// spawnWindow launches a new vibemd window in its own session so it
// survives if the MCP server process is later killed.
func spawnWindow(exe string, args ...string) error {
	cmd := exec.Command(exe, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd.Start()
}
