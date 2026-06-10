package main

import (
	"os"
	"path/filepath"

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

// OpenFile always opens the file in a NEW vibemd window so MCP clients
// (Claude Code, Cursor, etc.) can open multiple docs side by side without
// replacing an existing view. Use the CLI directly for single-window mode.
//
// The headless MCP state is also updated so get_current_file /
// get_rendered_html return the right content after this call.
func (h *headlessMCPState) OpenFile(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if _, err = os.Stat(abs); err != nil {
		return err
	}
	h.app.filePath = abs

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	// Always --new-window: MCP callers expect independent windows.
	// Single-window mode is available via the CLI: vibemd file.md
	return spawnWindow(exe, "--new-window", abs)
}

func runMCPServer() {
	app := NewApp()
	mcp.RunStdio(&headlessMCPState{app: app})
}
