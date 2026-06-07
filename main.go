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
	// Collect the first non-flag argument as the file to open.
	var fileArg string
	for _, arg := range os.Args[1:] {
		if arg == "--mcp" {
			runMCPServer()
			return
		}
		if !strings.HasPrefix(arg, "-") && fileArg == "" {
			fileArg = arg
		}
	}

	// If another vibemd is already running, hand the file to it and exit.
	if tryDelegate(fileArg) {
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
			listenForFiles(app)
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
