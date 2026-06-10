package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/cyuvop/vibemd/mcp"
)

// headlessMCPState satisfies mcp.State without a Wails window context.
// File operations work normally; theme/scroll events are no-ops.
type headlessMCPState struct {
	app *App
}

func (h *headlessMCPState) GetCurrentFile() map[string]interface{} { return h.app.GetCurrentFile() }
func (h *headlessMCPState) GetRenderedHTML() string                 { return h.app.GetRenderedHTML() }
func (h *headlessMCPState) GetTOC() []map[string]interface{}        { return h.app.GetTOC() }
func (h *headlessMCPState) GetFilePath() string                     { return h.app.GetFilePath() }
func (h *headlessMCPState) SetTheme(_ string)                       {} // no window in MCP mode

// OpenFile opens the file in a visible vibemd window:
//  1. If a window is already running, send the path via IPC socket.
//  2. Otherwise launch a new vibemd window process with the file.
//
// In both cases the headless MCP state is also updated so subsequent
// get_current_file / get_rendered_html calls return the right content.
func (h *headlessMCPState) OpenFile(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if _, err = os.Stat(abs); err != nil {
		return err
	}
	h.app.filePath = abs

	// Try to hand off to an already-running vibemd window.
	if tryDelegate(abs) {
		return nil
	}

	// No running window — spawn one.
	// Setsid puts the window in its own session so it survives if the
	// MCP server process is later killed (e.g. Claude Code disconnects).
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, abs)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func runMCPServer() {
	app := NewApp()
	mcp.RunStdio(&headlessMCPState{app: app})
}
