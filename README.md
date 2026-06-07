# VIBE·MD

> A cross-platform, secure Markdown viewer built for AI coding tools — with an 8-bit soul.

**vibemd** renders `.md` files beautifully in a retro 8-bit aesthetic. It is a dedicated *viewer* (not an editor), designed to be the app that Claude Code, Cursor, Windsurf, and other AI tools reach for when they surface a Markdown file. It exposes a built-in MCP server so AI agents can read, navigate, and control the viewer programmatically.

---

## Features

- **Instant rendering** — goldmark GFM renderer with syntax highlighting, tables, task lists, and strikethrough
- **8-bit aesthetic** — Press Start 2P headings, NES.css pixel borders, phosphor green dark mode, cream paper light mode
- **Live auto-refresh** — fsnotify watches the open file; changes from any editor appear immediately
- **Multi-file support** — `vibemd file.md` switches a running instance to a new file without opening a second window
- **MCP server** — 6 tools over stdio; registers with Claude Code in one command
- **Secure by design** — all rendering happens in Go (bluemonday sanitization), no CDN scripts, no relay servers, no npm
- **Single-file distribution** — macOS universal DMG and Windows NSIS installer, both self-contained

---

## Interface

```
┌────────────────────────────────────────────────────────┐
│ ▓▒░ VIBE·MD ░▒▓   README.md (/path/to/README.md)   [?]│  ← title bar
├────────────────────────────────────────────────────────┤
│                                                        │
│  █ H1 Heading                                          │  ← 8-bit heading
│  ▓ H2 Section                                          │
│                                                        │
│  Body text in system-ui (readable at length)           │
│                                                        │
│  ┌─────────────────────────────┐                       │
│  │  code block  (syntax hilit) │                       │
│  └─────────────────────────────┘                       │
│                                                        │
├────────────────────────────────────────────────────────┤
│ README.md │ 412 words │ DARK │ MCP: OFF │ # LN │ ↺ SYNC│  ← status bar
└────────────────────────────────────────────────────────┘
```

**Dark mode** — phosphor green (`#00FF41`) on near-black (`#0d0d0d`)  
**Light mode** — cream paper (`#FFF9E6`) with black pixel borders

---

## Installation

### macOS

Download `vibemd-mac.dmg` from [Releases](https://github.com/cyuvop/vibemd/releases), open it, and drag **vibemd** to Applications.

**First launch — Gatekeeper prompt:** vibemd is not yet notarized with Apple (requires a paid Developer ID). macOS will block it on first open. To allow it:

> **System Settings → Privacy & Security → scroll down → "vibemd was blocked" → Open Anyway**

Or from the terminal (one-time, permanent):
```bash
xattr -rd com.apple.quarantine /Applications/vibemd.app
```

To set as the default Markdown reader:
1. Right-click any `.md` file in Finder
2. **Get Info** → **Open With** → select vibemd → **Change All**

### Windows

Download `vibemd-setup.exe` from [Releases](https://github.com/cyuvop/vibemd/releases) and run it. The installer:
- Places vibemd in `%PROGRAMFILES%\vibemd\`
- Registers `.md` and `.markdown` file associations
- Creates a Start Menu shortcut
- Installs WebView2 if not already present (Windows 11 ships with it)

---

## Usage

### Opening a file

**From the terminal:**
```bash
vibemd README.md
vibemd /path/to/any/file.md
```

**From Finder:** Double-click any `.md` file (if vibemd is set as default), or right-click → **Open With** → vibemd.

**From another app:** Use "Open With" in any file browser or editor.

### Switching files in a running instance

vibemd is single-window by design. When an instance is already open, calling it with a new file switches the content without opening a second window:

```bash
vibemd DESIGN.md        # opens vibemd
vibemd AGENTS.md        # switches to AGENTS.md — same window, instant
vibemd CHANGELOG.md     # switches again
```

This uses a Unix socket (`$TMPDIR/vibemd.sock`) so the second invocation exits in milliseconds after handing off the path.

### Keyboard shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl + T` | Toggle light / dark theme |
| `Cmd/Ctrl + W` | Close window |

### Status bar buttons

| Button | Action |
|--------|--------|
| `DARK` / `LIGHT` | Toggle theme |
| `# LN` / `LN ON` | Toggle line numbers |
| `↺ SYNC` | Manually reload the file |

Line numbers are non-selectable, float in the left gutter (under the macOS traffic-light buttons), and cause no layout shift when toggled.

### Auto-refresh

vibemd watches the open file with fsnotify. Save the file in any editor and the viewer updates instantly — a loading spinner appears during the re-render.

### Help dialog

Click **?** in the top-right of the title bar to see the app icon, version, and link to this repository.

---

## MCP Integration

vibemd ships a built-in [Model Context Protocol](https://modelcontextprotocol.io) server. AI coding tools can open files, read content, extract structure, and control the viewer programmatically — without any separate install.

### Register with Claude Code

```bash
claude mcp add --scope user vibemd -- /Applications/vibemd.app/Contents/MacOS/vibemd --mcp
```

Verify it connected:
```bash
claude mcp list
# vibemd: ... ✓ Connected
```

### Register with other MCP clients (Cursor, Cline, Continue, Windsurf)

Add to your MCP config file (`~/.cursor/mcp.json`, `.clinerules`, etc.):

```json
{
  "mcpServers": {
    "vibemd": {
      "command": "/Applications/vibemd.app/Contents/MacOS/vibemd",
      "args": ["--mcp"]
    }
  }
}
```

### Available Tools

| Tool | Description |
|------|-------------|
| `open_file` | Open a `.md` file at any path in the viewer |
| `get_current_file` | Returns `{path, filename, rawMarkdown, wordCount, lastModified}` |
| `get_rendered_html` | Returns the sanitized HTML of the current view |
| `get_toc` | Returns `[{level, text, anchor}]` — the document's heading tree |
| `scroll_to_heading` | Scrolls the viewer to a heading matching the given text |
| `set_theme` | Switches between `"light"` and `"dark"` |

### MCP User Flows

#### AI opens a file and reads its structure

```
User: "Summarise the architecture section of DESIGN.md"

Claude Code:
  → open_file({path: "/path/to/DESIGN.md"})
  → get_toc()                          # find the architecture heading
  → get_current_file()                 # get rawMarkdown for that section
  → summarises and replies
```

#### AI navigates the viewer while explaining

```
User: "Walk me through the security section"

Claude Code:
  → scroll_to_heading({heading: "Security Architecture"})
  → get_rendered_html()                # read what's on screen
  → explains each threat/mitigation
  → scroll_to_heading({heading: "Mitigations"})
```

#### AI switches themes for a screenshot or presentation

```
User: "Switch to light mode for the demo"

Claude Code:
  → set_theme({theme: "light"})
```

#### AI opens the right doc based on context

```
User: "Open the MCP integration docs"

Claude Code:
  → open_file({path: "/path/to/project/AGENTS.md"})
  → get_current_file()                 # confirm what's loaded
```

### Quick smoke test (no Claude required)

```bash
# Build a test binary
go build -o /tmp/vibemd .

# Check protocol handshake
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}\n' \
  | /tmp/vibemd --mcp

# List all tools
printf '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}\n' \
  | /tmp/vibemd --mcp

# Open a file and read it back
printf '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"open_file","arguments":{"path":"/path/to/file.md"}}}\n' \
  | /tmp/vibemd --mcp

printf '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_current_file","arguments":{}}}\n' \
  | /tmp/vibemd --mcp
```

---

## Security

vibemd opens arbitrary `.md` files — including AI-generated and untrusted content. The rendering pipeline is hardened:

| Threat | Mitigation |
|--------|------------|
| XSS / script injection | HTML rendered in Go by goldmark; sanitized by bluemonday before reaching the WebView |
| `<iframe src="javascript:...">` (CVE-2024-21535 class) | bluemonday strips `<iframe>` entirely, server-side |
| Remote script loading | CSP: `script-src 'self'` — no CDN, no inline scripts |
| Supply chain attacks | Zero npm; `go.sum` SHA-256 pins every Go module |
| Dependency CVEs | `govulncheck ./...` runs in CI on every release |
| AI API key exposure | Keys stored in OS keychain only (macOS Keychain / Windows Credential Manager) |
| MCP relay eavesdropping | No relay server — MCP runs over stdio only, direct from client to vibemd |

---

## Building from Source

### Prerequisites

- Go 1.22+
- Wails v2: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- macOS packaging: `brew install create-dmg`
- Windows packaging: NSIS (for the installer script)

### Commands

```bash
# Run tests
go test ./...

# Development mode (hot reload)
~/go/bin/wails dev

# Build macOS universal binary + DMG
make mac

# Build Windows NSIS installer (run on Windows or Windows CI)
make windows

# Build both
make all
```

Outputs:
- `dist/vibemd-mac.dmg` — drag-to-Applications, runs on Intel + Apple Silicon
- `dist/vibemd-setup.exe` — one-click Windows installer with WebView2 bootstrapper

### Release

Tag a commit to trigger the GitHub Actions release workflow:

```bash
git tag v0.1.0
git push --tags
```

CI runs `govulncheck`, builds on native macOS and Windows runners, and attaches both distributables to a GitHub Release automatically.

---

## Project Structure

```
vibemd/
├── main.go              # Entry point, single-instance IPC, Wails bootstrap
├── app.go               # App struct, file open, watcher, Wails bindings
├── ipc.go               # Unix socket IPC for multi-file switching
├── mcp_stub.go          # Headless MCP state (no Wails context needed)
├── renderer/            # goldmark → bluemonday → sanitized HTML
├── watcher/             # fsnotify file watcher
├── mcp/                 # JSON-RPC 2.0 stdio MCP server
├── frontend/            # Vanilla HTML + CSS (no JS framework, no npm)
│   ├── index.html
│   ├── main.js          # ~100 lines of display-only Wails glue
│   └── style/           # dark.css, light.css, markdown.css, nes.css
├── build/
│   ├── appicon.png      # 1024×1024 app icon
│   ├── darwin/          # Info.plist (macOS file associations)
│   └── windows/         # installer.nsi (NSIS script)
├── Makefile
├── AGENTS.md            # MCP registration instructions for AI tools
└── CLAUDE.md            # Claude Code session context
```

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| App framework | [Wails v2](https://wails.io) (Go + system WebView) |
| Markdown rendering | [goldmark](https://github.com/yuin/goldmark) |
| HTML sanitization | [bluemonday](https://github.com/microcosm-cc/bluemonday) |
| Syntax highlighting | [goldmark-highlighting](https://github.com/yuin/goldmark-highlighting) (Chroma) |
| File watching | [fsnotify](https://github.com/fsnotify/fsnotify) |
| 8-bit CSS | [NES.css](https://github.com/nostalgic-css/NES.css) + Press Start 2P |
| Vulnerability scanning | [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) |

---

## License

MIT — see [LICENSE](LICENSE)
