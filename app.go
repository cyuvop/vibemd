package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/cyuvop/vibemd/renderer"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App holds runtime state and is bound to the Wails frontend.
type App struct {
	ctx      context.Context
	filePath string
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	for _, arg := range os.Args[1:] {
		if arg != "--mcp" && !strings.HasPrefix(arg, "-") {
			_ = a.OpenFile(arg)
			return
		}
	}
}

// OpenFile reads a Markdown file, renders it, and emits an event to the frontend.
func (a *App) OpenFile(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return err
	}
	a.filePath = abs
	a.emitRender(abs, data)
	return nil
}

func (a *App) emitRender(path string, data []byte) {
	html := renderer.RenderMarkdown(string(data))
	wordCount := len(strings.Fields(string(data)))
	runtime.EventsEmit(a.ctx, "markdown:rendered", map[string]interface{}{
		"html":      html,
		"path":      path,
		"filename":  filepath.Base(path),
		"wordCount": wordCount,
	})
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

// GetRenderedHTML returns the sanitized HTML for the current file.
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

// SetTheme emits a theme-change event (also callable from MCP).
func (a *App) SetTheme(theme string) {
	if theme != "light" && theme != "dark" {
		return
	}
	runtime.EventsEmit(a.ctx, "theme:changed", theme)
}

// GetTOC parses headings from the current file and returns a structured list.
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
