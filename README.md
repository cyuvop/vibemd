# VIBE·MD

> A cross-platform, secure Markdown viewer built for AI coding tools — with an 8-bit soul.

**vibemd** renders `.md` files beautifully in a retro 8-bit aesthetic. It is a dedicated *viewer* (not an editor), designed to be the app that Claude Code, Cursor, Windsurf, and other AI tools reach for when they surface a Markdown file. It exposes a built-in MCP server so AI agents can read, navigate, and control the viewer programmatically.

---

## Features

- **Instant rendering** — goldmark GFM renderer with syntax highlighting, tables, task lists, and strikethrough
- **8-bit aesthetic** — Press Start 2P headings, NES.css pixel borders, phosphor green dark mode, cream paper light mode
- **Live auto-refresh** — fsnotify watches the open file; changes from any editor appear immediately
- **Search** — `Cmd+F` full-text search with highlighted matches and forward/backward navigation
- **Multi-window** — `vibemd file.md` switches the running instance; `--new-window` opens an independent second window
- **Built-in welcome** — opens with the README when no file is specified
- **MCP server** — 6 tools over stdio; registers with Claude Code in one command
- **Secure by design** — all rendering happens in Go (bluemonday sanitization), no CDN scripts, no relay servers, no npm
- **Single-file distribution** — macOS universal DMG and Windows zip, both self-contained

---

## Interface

```
┌────────────────────────────────────────────────────────┐
│ ▓▒░ VIBE·MD ░▒▓   README.md (/path/to/README.md)   [?]│  ← title bar (drag to move)
├────────────────────────────────────────────────────────┤
│ SEARCH...                        3 / 12  ◄  ►  ✕      │  ← Cmd+F search bar
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
│ README.md │ 412 words │ DARK │ # LN │ ↺ SYNC          │  ← status bar
└────────────────────────────────────────────────────────┘
```

**Dark mode** — phosphor green (`#00FF41`) on near-black (`#0d0d0d`)  
**Light mode** — cream paper (`#FFF9E6`) with black pixel borders

---

## Installation

### macOS

Download `vibemd-mac.dmg` from [Releases](https://github.com/cyuvop/vibemd/releases), open it, and drag **vibemd** to Applications.

**First launch — Gatekeeper:** vibemd is not yet notarized with Apple. macOS will show _"Apple could not verify vibemd is free of malware"_. Use one of these three bypass methods (one-time only):

**Option A — System Settings (recommended):**
1. Try to open vibemd — macOS blocks it
2. Open **System Settings → Privacy & Security**
3. Scroll down to the _"vibemd was blocked"_ message
4. Click **Open Anyway** → **Open**

**Option B — Terminal (permanent, no more prompts ever):**
```bash
xattr -rd com.apple.quarantine /Applications/vibemd.app
```

**Option C — Right-click bypass:**
Right-click `vibemd.app` in Finder → **Open** → click **Open** in the dialog

To set as the default Markdown reader:
1. Right-click any `.md` file in Finder
2. **Get Info** → **Open With** → select vibemd → **Change All**

### Windows

Download `vibemd-windows.zip` from [Releases](https://github.com/cyuvop/vibemd/releases), extract it, and run `vibemd.exe`.

To add vibemd to your PATH so you can run it from the terminal:
1. Move `vibemd.exe` to a folder like `C:\Tools\`
2. Add that folder to your system PATH via System Properties → Environment Variables

---

## Usage

### Opening a file

**From the terminal:**
```bash
vibemd README.md
vibemd /path/to/any/file.md
```

**No file specified** — vibemd opens with this README as the welcome screen.

**From Finder / File Explorer:** Double-click any `.md` file (if vibemd is set as default), or right-click → **Open With** → vibemd.

### Multiple windows

```bash
vibemd DESIGN.md               # opens / switches the primary window
vibemd AGENTS.md               # switches primary window to AGENTS.md
vibemd --new-window AGENTS.md  # opens AGENTS.md in a new independent window
```

`vibemd file.md` uses a Unix socket to hand the file to any running instance in milliseconds. `--new-window` bypasses this and always launches a fresh window.

### Search

Press **`Cmd+F`** (Mac) or **`Ctrl+F`** (Windows) to open the search bar:

- Type to highlight all matches — current match in accent colour, others in yellow
- **`Enter`** — next match
- **`Shift+Enter`** — previous match
- **`◄` / `►`** buttons — navigate with the mouse
- Counter shows `3 / 12` style, or `NO MATCH` when nothing is found
- **`Escape`** — close search and clear highlights
- Search re-applies automatically when the file reloads

### Keyboard shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl + F` | Open search |
| `Enter` (in search) | Next match |
| `Shift + Enter` (in search) | Previous match |
| `Escape` (in search) | Close search |
| `Cmd/Ctrl + T` | Toggle light / dark theme |
| `Cmd/Ctrl + W` | Close window |

### Status bar buttons

| Button | Action |
|--------|--------|
| `DARK` / `LIGHT` | Toggle theme — hover highlights in accent colour |
| `# LN` / `LN ON` | Toggle line numbers in left gutter |
| `↺ SYNC` | Manually reload the file from disk |

### Auto-refresh

vibemd watches the open file with fsnotify. Save the file in any editor and the viewer updates instantly — a brief loading spinner appears during re-render.

### Help dialog

Click **?** in the top-right of the title bar to see the app icon, version, and a link to this repository.

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

Add to your MCP config file:

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
go build -o /tmp/vibemd .

printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}\n' \
  | /tmp/vibemd --mcp

printf '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}\n' \
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

### Commands

```bash
go test ./...          # run all tests
~/go/bin/wails dev     # development mode with hot reload
make mac               # build macOS universal DMG
make windows           # build Windows zip (run on Windows)
make all               # both
```

### Release

```bash
git tag v1.0.0
git push --tags        # triggers CI → builds both platforms → GitHub Release
```

---

## Project Structure

```
vibemd/
├── main.go              # Entry point, --new-window, IPC, Wails bootstrap
├── app.go               # App struct, file open, watcher, embedded README
├── ipc.go               # Unix socket for multi-file switching
├── mcp_stub.go          # Headless MCP state (no Wails context)
├── renderer/            # goldmark → bluemonday → sanitized HTML
├── watcher/             # fsnotify file watcher
├── mcp/                 # JSON-RPC 2.0 stdio MCP server (6 tools)
├── frontend/
│   ├── index.html       # Search bar, busy overlay, help dialog
│   ├── main.js          # Search, drag, theme, line numbers, MCP glue
│   └── style/           # dark.css, light.css, markdown.css, base.css, nes.css
├── build/
│   ├── appicon.png      # 1024×1024 app icon
│   ├── darwin/          # Info.plist (UTI file associations)
│   └── windows/         # installer.nsi (NSIS script)
├── Makefile
├── AGENTS.md            # MCP registration for AI tools
├── CLAUDE.md            # Claude Code session context
└── DESIGN.md            # Architecture and design decisions
```

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| App framework | [Wails v2](https://wails.io) — Go + system WebView |
| Markdown rendering | [goldmark](https://github.com/yuin/goldmark) |
| HTML sanitization | [bluemonday](https://github.com/microcosm-cc/bluemonday) |
| Syntax highlighting | [goldmark-highlighting](https://github.com/yuin/goldmark-highlighting) (Chroma) |
| File watching | [fsnotify](https://github.com/fsnotify/fsnotify) |
| 8-bit CSS | [NES.css](https://github.com/nostalgic-css/NES.css) + Press Start 2P |
| Vulnerability scanning | [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) |

Zero npm. Zero node_modules. `go build` is the only build tool for logic.

---

## License

MIT — see [LICENSE](LICENSE)
