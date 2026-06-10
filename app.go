package main

import (
	"context"
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/cyuvop/vibemd/renderer"
	"github.com/cyuvop/vibemd/watcher"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed README.md
var builtinReadme []byte

// App holds runtime state and is bound to the Wails frontend.
type App struct {
	ctx          context.Context
	filePath     string
	watchCancel  context.CancelFunc
	pendingRender map[string]interface{} // set during startup before frontend is ready
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Use a sentinel so emitRender knows the frontend isn't ready yet.
	// Ready() clears this and flushes the cached payload.
	a.pendingRender = map[string]interface{}{}
	for _, arg := range os.Args[1:] {
		if arg != "--mcp" && arg != "--new-window" && !strings.HasPrefix(arg, "-") {
			_ = a.OpenFile(arg)
			return
		}
	}
	// No file argument — show the built-in README as the welcome screen.
	a.emitRender("README.md", builtinReadme)
}

// Ready is called by the frontend once its event listeners are registered.
// It flushes any render that happened during startup before JS was ready.
func (a *App) Ready() {
	if a.pendingRender != nil {
		runtime.EventsEmit(a.ctx, "markdown:rendered", a.pendingRender)
		a.pendingRender = nil
	}
}

func (a *App) shutdown(_ context.Context) {
	if a.watchCancel != nil {
		a.watchCancel()
	}
}

// OpenFile reads a Markdown file, renders it, starts the file watcher,
// and emits a markdown:rendered event to the frontend.
func (a *App) OpenFile(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return err
	}

	// Cancel any existing watcher
	if a.watchCancel != nil {
		a.watchCancel()
	}

	a.filePath = abs
	a.emitRender(abs, data)

	// Start watching for changes
	watchCtx, cancel := context.WithCancel(a.ctx)
	a.watchCancel = cancel
	go func() {
		_ = watcher.Watch(watchCtx, abs, func(p string, d []byte) {
			a.emitRender(p, d)
		})
	}()

	return nil
}

func (a *App) emitRender(path string, data []byte) {
	// Signal busy state before the (potentially slow) render.
	if a.pendingRender == nil {
		runtime.EventsEmit(a.ctx, "render:start", nil)
	}
	html := renderer.RenderMarkdown(string(data))
	payload := map[string]interface{}{
		"html":      html,
		"path":      path,
		"filename":  filepath.Base(path),
		"wordCount": len(strings.Fields(string(data))),
	}
	// During startup the WebView hasn't loaded yet — cache and flush in Ready().
	// After startup, emit directly (watcher updates, MCP open_file, etc.).
	if a.pendingRender != nil {
		a.pendingRender = payload
		return
	}
	runtime.EventsEmit(a.ctx, "markdown:rendered", payload)
}

// GetCurrentFile returns metadata about the open file (used by MCP + frontend).
func (a *App) GetCurrentFile() map[string]interface{} {
	if a.filePath == "" {
		return map[string]interface{}{"path": "", "filename": "", "wordCount": 0}
	}
	data, err := os.ReadFile(a.filePath)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	info, _ := os.Stat(a.filePath)
	return map[string]interface{}{
		"path":         a.filePath,
		"filename":     filepath.Base(a.filePath),
		"rawMarkdown":  string(data),
		"wordCount":    len(strings.Fields(string(data))),
		"lastModified": info.ModTime().Unix(),
	}
}

// GetRenderedHTML returns sanitized HTML for the current file (used by MCP).
func (a *App) GetRenderedHTML() string {
	if a.filePath == "" {
		return ""
	}
	data, err := os.ReadFile(a.filePath)
	if err != nil {
		return ""
	}
	return renderer.RenderMarkdown(string(data))
}

// GetFilePath returns the currently open file path.
func (a *App) GetFilePath() string { return a.filePath }

// GetVersion returns the current app version string.
func (a *App) GetVersion() string { return AppVersion }

// Refresh re-reads and re-renders the current file on demand.
func (a *App) Refresh() error {
	if a.filePath == "" {
		return nil
	}
	data, err := os.ReadFile(a.filePath)
	if err != nil {
		return err
	}
	a.emitRender(a.filePath, data)
	return nil
}

// SetTheme emits a theme-change event (callable from MCP and frontend).
func (a *App) SetTheme(theme string) {
	if theme != "light" && theme != "dark" {
		return
	}
	runtime.EventsEmit(a.ctx, "theme:changed", theme)
}

// GetTOC extracts headings from the current file as a structured list.
func (a *App) GetTOC() []map[string]interface{} {
	if a.filePath == "" {
		return nil
	}
	data, err := os.ReadFile(a.filePath)
	if err != nil {
		return nil
	}
	return renderer.ExtractTOC(string(data))
}
