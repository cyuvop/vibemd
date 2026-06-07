package main

import (
	"context"
	"embed"
	"log"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	// Collect flags and the first non-flag argument as the file to open.
	var fileArg string
	var newWindow bool
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--mcp":
			runMCPServer()
			return
		case "--new-window":
			newWindow = true
		default:
			if !strings.HasPrefix(arg, "-") && fileArg == "" {
				fileArg = arg
			}
		}
	}

	// Without --new-window, hand the file to any running instance and exit.
	if !newWindow && tryDelegate(fileArg) {
		return
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "vibemd",
		Width:  960,
		Height: 720,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 13, G: 13, B: 13, A: 255},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			// Only the primary instance owns the IPC socket.
			// --new-window instances are independent and don't steal it.
			if !newWindow {
				listenForFiles(app)
			}
		},
		OnShutdown: app.shutdown,
		Bind:             []interface{}{app},
		Mac: &mac.Options{
			// Handle "Open With" / double-click from Finder
			OnFileOpen: func(filePath string) {
				_ = app.OpenFile(filePath)
			},
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
