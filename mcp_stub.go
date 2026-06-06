package main

import (
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
func (h *headlessMCPState) OpenFile(path string) error              { return h.app.OpenFile(path) }
func (h *headlessMCPState) SetTheme(_ string)                       {} // no window in MCP mode

func runMCPServer() {
	app := NewApp()
	mcp.RunStdio(&headlessMCPState{app: app})
}
