# vibemd — Claude Code Context

## What this repo is

vibemd is a cross-platform (macOS + Windows) Markdown viewer with an 8-bit aesthetic,
built for AI coding tools. It is a **viewer only** — not an editor.

## Tech stack

- **Go 1.22** — all business logic, rendering, MCP server
- **Wails v2** — desktop app framework (Go + system WebView)
- **goldmark** — Markdown → HTML rendering (GFM)
- **bluemonday** — HTML sanitization (XSS defense)
- **fsnotify** — file watching
- **Frontend** — vanilla HTML + CSS only (no JS framework, no npm)

## Project structure

```
main.go          Wails entry, CLI args, --mcp dispatch
app.go           App struct bound to Wails frontend
renderer/        goldmark + bluemonday pipeline + TOC
watcher/         fsnotify wrapper
mcp/             JSON-RPC 2.0 stdio MCP server
frontend/        HTML/CSS/minimal JS (display layer only)
build/           Platform-specific build config (Info.plist, NSIS)
```

## Running / testing

```bash
go test ./...                           # run all tests
~/go/bin/wails dev                      # dev mode (hot reload)
make mac                                # build macOS DMG
make windows                            # build Windows installer
vibemd README.md                        # open a file
vibemd --mcp                            # run as MCP server
```

## MCP registration

```bash
claude mcp add vibemd -- vibemd --mcp
```

## Security rules

- Never add CDN script tags — all assets must be locally bundled
- Never store API keys outside the OS keychain
- All HTML must pass through bluemonday before reaching the WebView
- Run `govulncheck ./...` before any release
