# vibemd — Design Document

> A cross-platform, secure, 8-bit Markdown reader built for AI tools.

---

## 1. Overview

**vibemd** is a dedicated Markdown *viewer* (not editor) for macOS and Windows. It renders `.md` files with beautiful typography in a retro 8-bit aesthetic, integrates natively with AI coding tools via MCP, and supports OS-level "Open With" / default reader registration so any Markdown file can be opened instantly.

**Core philosophy:**
- Reader-first, zero cruft — one file, beautiful rendering, nothing else
- Built for AI tools — vibemd is what Claude, Cursor, and Windsurf open when they surface a `.md` file
- 8-bit nerd soul — Press Start 2P font, pixelated borders, scanline texture, CRT glow accents
- Security-first — no CDN scripts, no relay servers, local rendering only
- Light and dark modes that still look retro

---

## 2. Market Gap (Research-backed)

| App | Platform | Open With | AI MCP | 8-bit |
|-----|----------|-----------|--------|-------|
| Marked 2 | macOS only | ✓ | ✗ | ✗ |
| Obsidian | Cross-platform | ✗ (vault-based) | Plugin only | ✗ |
| Typora | Cross-platform | ✓ | ✗ | ✗ |
| SoloMD | Cross-platform | ✗ | Partial | ✗ |
| Noteriv | Cross-platform | ✗ | ✓ (22 tools) | ✗ |
| **vibemd** | **macOS + Windows** | **✓** | **✓ (MCP)** | **✓** |

There is no macOS + Windows app that combines native file association, MCP integration, and a distinct visual identity. vibemd fills this gap.

---

## 3. Tech Stack

### Framework: Wails v2 (Go + system WebView)
- Go backend handles all logic, rendering, file I/O, and MCP — zero JavaScript business logic
- Uses system WebView (WKWebView on macOS, WebView2 on Windows) — small binary, no Chromium bundle
- Go IPC bindings (`wails.Bind`) expose Go structs directly to the frontend — no hand-written JS glue
- File associations registered via OS-specific installer config (macOS `.plist` UTI, Windows NSIS registry)
- Single `wails build` command produces a signed, self-contained binary

**Why Go over Rust/Tauri:**
- No npm, no `node_modules` — eliminates the entire JS supply chain attack surface (OWASP A06)
- `go.sum` provides cryptographic hash verification for every module dependency
- `govulncheck ./...` scans all transitive deps against the Go vuln database — standard CI step
- Rendering pipeline runs entirely in Go: HTML is sanitized before it ever reaches the WebView
- Single binary output, `go build` reproducibility, no runtime to exploit

### Frontend: Vanilla HTML + CSS only
- No JavaScript framework, no npm packages — the frontend is a pure display layer
- Go pre-renders the full sanitized HTML string and pushes it via Wails IPC
- The only JS is ~50 lines of Wails runtime glue (auto-generated) + theme class toggling
- NES.css (locally bundled CSS file) provides all 8-bit component styles
- Press Start 2P font locally bundled — no Google Fonts CDN call

### 8-bit Design System
- **NES.css** — established retro component library (pixel borders, dialog boxes, progress bars)
- **Press Start 2P** (locally bundled) — canonical 8-bit typeface
- **Custom CSS** — scanline overlay, CRT glow on headings, pixelated scrollbar
- Dark mode: dark phosphor green (`#00FF41`) accents on near-black (`#0d0d0d`)
- Light mode: cream paper (`#FFF9E6`) with black pixel borders

### Markdown Rendering (Go-side, server-side)
- **goldmark** (`github.com/yuin/goldmark`) — GFM-compliant Markdown → HTML in Go; extensible, actively maintained
- **bluemonday** (`github.com/microcosm-cc/bluemonday`) — HTML sanitization in Go; Go's battle-tested equivalent of DOMPurify; allowlist defined in Go, not JS config
- **chroma** (`github.com/alecthomas/chroma`) — syntax highlighting in Go; outputs pre-styled HTML, no client-side JS needed
- HTML is fully sanitized in Go *before* being passed to the WebView — the browser never sees unsanitized content

### AI / MCP Integration
- **Built-in MCP server** (stdio transport) in pure Go — hand-rolled JSON-RPC 2.0, no third-party MCP crate
- API keys stored in OS keychain: macOS Keychain / Windows Credential Manager via `github.com/zalando/go-keyring`
- No relay server — all AI requests direct from user machine to provider

---

## 4. Feature Set

### MVP (v0.1)
- [x] Open a `.md` file from CLI: `vibemd README.md`
- [x] Render Markdown with GFM (tables, task lists, strikethrough) via goldmark
- [x] Light / dark mode (system-following, toggle override)
- [x] 8-bit visual theme (NES.css + Press Start 2P)
- [x] "Open With" registration on macOS and Windows
- [x] Syntax highlighting via chroma (Go-side, zero client JS)
- [x] File-watching: auto-refresh when the `.md` file changes on disk (fsnotify)
- [x] Keyboard shortcut: `Cmd/Ctrl+W` to close, `Cmd/Ctrl+T` to toggle theme

### v0.2 — AI Tools
- [ ] MCP server (stdio) with tools:
  - `get_current_file` — returns path + raw markdown
  - `get_rendered_html` — returns sanitized HTML
  - `get_toc` — returns table of contents as JSON
  - `scroll_to_heading` — scroll the view to a heading
  - `set_theme` — switch light/dark from AI tool
- [ ] `AGENTS.md` / `CLAUDE.md` context file for AI coding tools
- [ ] Claude Code one-liner: `claude mcp add vibemd -- vibemd --mcp`

### v0.3 — Power Features
- [ ] Recent files list (8-bit styled menu)
- [ ] Print / export to PDF
- [ ] Mermaid diagram rendering
- [ ] KaTeX math rendering
- [ ] Custom CSS snippets (drop a `.vibemd.css` next to the `.md` file)
- [ ] Command palette (`Cmd/Ctrl+K`)

### v0.4 — Themes & Community
- [ ] Theme switcher: Dark Phosphor, Dark Amber, Dark Dracula, Light Paper, Light Sepia
- [ ] Custom theme import/export
- [ ] Plugin system via WASM

---

## 5. UI/UX Design

### Layout
```
┌─────────────────────────────────────────────────┐
│ ░░░░░░░░░ VIBEMD ░░░░░░░░░  [◐] [?]  ░░░░░░░░░ │  ← 8-bit title bar
├─────────────────────────────────────────────────┤
│                                                 │
│  ████ H1 Heading                                │  ← Pixelated heading marker
│                                                 │
│  ░░ H2 Section                                  │
│                                                 │
│  Body text in system-ui (readable, not 8-bit)  │
│                                                 │
│  ┌──────────────────────────────┐               │
│  │ code block with 8-bit theme │               │  ← NES.css container
│  └──────────────────────────────┘               │
│                                                 │
└─────────────────────────────────────────────────┘
│ filename.md ░ 1234 words ░ [DARK] ░ MCP: ON    │  ← Status bar
└─────────────────────────────────────────────────┘
```

### Typography Rules
- **Headings only**: Press Start 2P (8-bit) — visual anchor, kept short
- **Body text**: system-ui / -apple-system — readable at length (not 8-bit)
- **Code blocks**: "Courier Prime" or JetBrains Mono — monospace, pixel-friendly
- This avoids the readability trap of forcing 8-bit font on prose

### Color Palettes

**Dark Mode (Phosphor Green)**
```
Background: #0d0d0d
Surface:    #1a1a1a
Border:     #333333
Text:       #e0e0e0
Accent:     #00FF41  (phosphor green)
Heading:    #00FF41
Code BG:    #111111
Link:       #39FF14
```

**Light Mode (Cream Paper)**
```
Background: #FFF9E6
Surface:    #FFFDF5
Border:     #1a1a1a
Text:       #1a1a1a
Accent:     #C41230  (8-bit red)
Heading:    #000000
Code BG:    #F0ECD8
Link:       #0000CC
```

### Signature Visual Details
- 1px pixel borders (no border-radius) — everything is square
- Checkerboard scrollbar thumb
- Scanline CSS overlay on the document (subtle, ~5% opacity)
- Heading markers: `█ H1`, `▓ H2`, `░ H3` (block characters as decoration)
- Status bar uses NES.css `nes-badge` style chips
- Window title: `VIBEMD v0.1.0 > filename.md`

---

## 6. Security Architecture

### Threat Model
vibemd opens arbitrary `.md` files from disk — some may be untrusted (cloned repos, AI-generated). The rendering pipeline must not execute attacker-controlled code.

### Why Go improves on the previous Rust/JS design (OWASP alignment)

| OWASP Top 10 | Previous (Tauri + JS) | Go approach |
|---|---|---|
| A03 Injection / XSS | DOMPurify in browser JS | bluemonday sanitizes in Go *before* WebView sees HTML |
| A06 Vulnerable Components | npm audit (node_modules) | `govulncheck ./...` against Go vuln DB; `go.sum` hash pinning |
| A08 Software Integrity | npm lockfile (weaker) | `go.sum` SHA-256 per module, verified at build time |
| A09 Logging Failures | ad-hoc | Go `log/slog` structured logging, errors surfaced explicitly |

### Mitigations

| Threat | Mitigation |
|--------|------------|
| XSS in rendered HTML (CVE-2024-21535 class) | bluemonday allowlist in Go — HTML never reaches WebView unsanitized |
| Remote script injection | CSP: `script-src 'self'`; Wails embeds assets locally, no CDN |
| Malicious `<iframe src="javascript:...">` | bluemonday strips `<iframe>` entirely; Go-side, not browser-side |
| AI API key exposure | OS keychain only via `go-keyring` — never in env vars, config files, or logs |
| Remote relay eavesdropping | No relay — direct HTTP from user machine to provider |
| WebView RCE via Node.js | No Node.js — Wails frontend has no runtime; Go handles all I/O |
| Supply chain compromise | No npm; `go.sum` pins every transitive dep by SHA-256 |
| Known vuln in deps | `govulncheck ./...` in CI catches CVEs across all Go modules |

### bluemonday Policy (Go)
```go
p := bluemonday.NewPolicy()
p.AllowElements("p","h1","h2","h3","h4","h5","h6",
    "ul","ol","li","blockquote","pre","code",
    "em","strong","del","table","thead","tbody",
    "tr","th","td","hr","br","details","summary",
    "span","div")
p.AllowAttrs("href").OnElements("a")
p.AllowAttrs("src","alt","title").OnElements("img")
p.AllowAttrs("class").OnElements("code","pre","span","div")
p.AllowAttrs("checked","disabled").OnElements("input")
p.RequireNoFollowOnLinks(true)
p.RequireNoReferrerOnLinks(true)
// No style, no onerror, no onclick, no iframe, no script — ever
```

### CSP (Wails wails.json)
```json
"security": {
  "csp": "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'none'"
}
```

### Dependency Audit CI Step
```yaml
- name: govulncheck
  run: go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./...
```

---

## 7. File Association ("Open With")

### macOS
Wails embeds an `Info.plist` in the `.app` bundle. Add to `wails.json`:
```json
"mac": {
  "bundleId": "com.cyuvop.vibemd",
  "info": {
    "CFBundleDocumentTypes": [{
      "CFBundleTypeName": "Markdown Document",
      "CFBundleTypeRole": "Viewer",
      "LSHandlerRank": "Alternate",
      "LSItemContentTypes": ["net.daringfireball.markdown"],
      "CFBundleTypeExtensions": ["md","markdown","mdown","mkd"]
    }]
  }
}
```
Go receives the file path via the Wails `OnFileOpen` event. Right-click any `.md` → "Open With" → vibemd, or "Always Open With" to set as default.

### Windows
NSIS installer script registers file extensions in the registry:
```nsi
WriteRegStr HKCR ".md"    "" "vibemd.Document"
WriteRegStr HKCR ".markdown" "" "vibemd.Document"
WriteRegStr HKCR "vibemd.Document" "" "Markdown Document"
WriteRegStr HKCR "vibemd.Document\shell\open\command" "" '"$INSTDIR\vibemd.exe" "%1"'
```
First run prompts to set as default. Release builds will be Authenticode-signed to avoid SmartScreen on first run.

### CLI Entry Point
```bash
vibemd path/to/file.md          # open file
vibemd --mcp                    # start MCP server on stdio
vibemd --list-themes            # list available themes
```

---

## 8. MCP Integration

vibemd exposes a stdio MCP server for AI coding tools (Claude Code, Cursor, Cline, Continue, Windsurf).

### Adding to Claude Code
```bash
claude mcp add vibemd -- vibemd --mcp
```

### Tools Exposed (v0.2)

| Tool | Description |
|------|-------------|
| `get_current_file` | Returns `{path, raw_markdown, word_count, last_modified}` |
| `get_rendered_html` | Returns sanitized HTML of current view |
| `get_toc` | Returns `[{level, text, anchor}]` table of contents |
| `scroll_to_heading` | Scrolls viewer to matching heading (fuzzy match) |
| `set_theme` | Accepts `"light"` or `"dark"` |
| `open_file` | Opens a new `.md` file path in vibemd |

### AGENTS.md (placed in repo root)
```markdown
# vibemd AI Integration

vibemd exposes an MCP server. If Claude Code is running alongside vibemd,
AI tools can read, navigate, and control the Markdown viewer.

Register: `claude mcp add vibemd -- vibemd --mcp`

Available tools: get_current_file, get_rendered_html, get_toc,
scroll_to_heading, set_theme, open_file
```

---

## 9. Implementation Plan

### Phase 0 — Scaffold (Week 1)
- [ ] `wails init -n vibemd -t vanilla` — Go + vanilla HTML template (no JS framework)
- [ ] Add NES.css, Press Start 2P as static assets (no npm)
- [ ] Basic window chrome with 8-bit title bar in HTML/CSS
- [ ] File open via CLI arg + Wails `OnFileOpen` handler in `main.go`
- [ ] Wire system theme detection (`runtime.WindowGetSystemTheme`) → CSS class on `<body>`

### Phase 1 — Markdown Rendering (Week 2)
- [ ] goldmark with GFM extension (`github.com/yuin/goldmark`)
- [ ] bluemonday sanitization policy (see §6)
- [ ] chroma syntax highlighting — Go outputs pre-colored `<span>` HTML, no client JS
- [ ] 8-bit CSS for all rendered Markdown elements (headings, code, tables, blockquotes)
- [ ] fsnotify file watcher (`github.com/fsnotify/fsnotify`) → Wails event → frontend re-renders

### Phase 2 — Polish & Distribution (Week 3)
- [ ] Keyboard shortcuts (`Cmd+W`, `Cmd+T`, `Cmd+K`) via Wails menu API
- [ ] Status bar (filename, word count, theme indicator, MCP status)
- [ ] macOS `wails.json` UTI + Windows NSIS file association testing
- [ ] Code signing (macOS notarization via `wails build --sign`, Windows Authenticode)
- [ ] GitHub Actions CI matrix (macOS-latest, windows-latest) + `govulncheck`

### Phase 3 — MCP Server (Week 4)
- [ ] Pure Go MCP server over stdio (JSON-RPC 2.0, stdlib `encoding/json`)
- [ ] Implement 6 tools (see §8) — all state held in Go, no JS involved
- [ ] `AGENTS.md` + `CLAUDE.md` in repo root
- [ ] Test with `claude mcp add vibemd -- vibemd --mcp`
- [ ] Document one-liner in README

### Phase 4 — Power Features (Week 5+)
- [ ] Mermaid diagram rendering — Go calls mermaid-go or pre-renders SVG server-side
- [ ] Recent files list (stored in OS config dir via `os.UserConfigDir()`)
- [ ] Theme switcher UI (NES.css styled)
- [ ] PDF export via Wails print / OS print dialog
- [ ] Custom per-file CSS sidecar (`.vibemd.css` alongside the `.md` file)

---

## 10. Repo Structure

```
vibemd/
├── main.go                 # Wails entry point, CLI arg handling, file open
├── app.go                  # App struct — Wails-bound Go API exposed to frontend
├── renderer/
│   ├── renderer.go         # goldmark → HTML pipeline
│   ├── sanitize.go         # bluemonday policy definition
│   └── highlight.go        # chroma syntax highlighting config
├── watcher/
│   └── watcher.go          # fsnotify wrapper → Wails event emit
├── mcp/
│   └── server.go           # JSON-RPC 2.0 stdio MCP server + 6 tools
├── keychain/
│   └── keychain.go         # go-keyring wrapper for AI API keys
├── frontend/               # Pure HTML/CSS — display layer only
│   ├── index.html          # Main window template
│   ├── style/
│   │   ├── nes.css         # NES.css (locally bundled, unmodified)
│   │   ├── overrides.css   # NES.css customizations
│   │   ├── dark.css        # Dark phosphor theme
│   │   ├── light.css       # Light paper theme
│   │   └── markdown.css    # Rendered MD element styles
│   ├── fonts/              # Press Start 2P, JetBrains Mono (locally bundled)
│   └── main.js             # ~50 lines: Wails event listeners + theme class toggle
├── wails.json              # Wails config: window, CSP, file associations
├── go.mod
├── go.sum                  # Cryptographic module hashes
├── AGENTS.md               # AI tool integration docs
├── CLAUDE.md               # Claude Code context
└── DESIGN.md               # This document
```

**Zero npm. Zero node_modules. `go build` is the only build command for logic.**

---

## 11. Open Questions

1. **Wails file associations on Windows**: Wails v2 does not have a built-in `fileAssociations` field like Tauri — requires custom NSIS script in the installer. Needs testing on Windows 10 vs 11.
2. **Mermaid server-side rendering**: `mermaid-go` (Go port) is less mature than the JS original. May need to ship a minimal sandboxed headless renderer or pre-render at install time.
3. **Press Start 2P readability**: Heading-only use of the font should be readable, but needs visual testing with real-world `.md` files of varying heading density.
4. **Windows SmartScreen**: First-run UX for unsigned builds — document clearly in README for early adopters until Authenticode signing is set up.
5. **Wails v3 timeline**: Wails v3 (alpha as of 2025) drops the WebView2 bundling requirement on Windows. Worth monitoring for the v0.3 milestone.

---

## 12. Key Go Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/wailsapp/wails/v2` | Desktop app framework (Go + system WebView) |
| `github.com/yuin/goldmark` | Markdown → HTML (GFM, extensible) |
| `github.com/microcosm-cc/bluemonday` | HTML sanitization (OWASP XSS defense) |
| `github.com/alecthomas/chroma/v2` | Syntax highlighting (Go-side, no client JS) |
| `github.com/fsnotify/fsnotify` | Cross-platform file watching |
| `github.com/zalando/go-keyring` | OS keychain (macOS / Windows Credential Manager) |
| `golang.org/x/vuln/cmd/govulncheck` | CVE scanning for all Go modules (CI) |

## 13. Sources

- Marked 2 App: https://marked2app.com/
- SoloMD security model: https://github.com/zhitongblog/solomd
- Noteriv MCP pattern: https://github.com/thejacedev/Noteriv
- Wails: https://wails.io/
- bluemonday: https://github.com/microcosm-cc/bluemonday
- NES.css: https://github.com/nostalgic-css/NES.css
- CVE-2024-21535: https://security.snyk.io/vuln/SNYK-JS-MARKDOWNTOJSX-6258886

---

*Generated 2026-06-05. Maintained alongside the codebase — update as decisions are made.*
