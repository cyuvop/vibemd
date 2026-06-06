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

// OpenFile in headless mode just sets the path — no Wails context exists
// to emit events to, and get_current_file reads fresh from disk each call.
func (h *headlessMCPState) OpenFile(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if _, err = os.Stat(abs); err != nil {
		return err
	}
	h.app.filePath = abs
	return nil
}

func runMCPServer() {
	app := NewApp()
	mcp.RunStdio(&headlessMCPState{app: app})
}
