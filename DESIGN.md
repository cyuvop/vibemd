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

### Framework: Tauri 2 (Rust + WebView)
- Uses system WebView (WKWebView on macOS, WebView2 on Windows) — small binary, no Chromium bundle
- Rust backend provides memory safety for file I/O
- CSP is enforced at the framework level with locally-bundled assets only (no CDN scripts — Tauri's documented best practice)
- `fileAssociations` config in `tauri.conf.json` handles "Open With" registration on both platforms

### UI: React 18 + TypeScript + Vite
- `@tauri-apps/api` for IPC (file open, theme detection, keychain access)
- `nativeTheme` events via Tauri for system light/dark detection

### 8-bit Design System
- **NES.css** — established retro component library (pixel borders, dialog boxes, progress bars)
- **Press Start 2P** (Google Fonts, locally bundled) — canonical 8-bit typeface
- **Custom CSS** — scanline overlay, CRT glow on headings, pixelated scrollbar
- Dark mode: dark phosphor green (`#00FF41`) accents on near-black (`#0d0d0d`)
- Light mode: cream paper (`#FFF9E6`) with black pixel borders

### Markdown Rendering
- **marked.js** (not markdown-to-jsx — see CVE-2024-21535 below)
- **DOMPurify 3.x** — sanitize rendered HTML before injection
- **Mermaid.js** — diagram rendering (locally bundled, not CDN)
- **KaTeX** — math rendering (locally bundled)
- **Prism.js** — syntax highlighting with 8-bit color theme

### AI / MCP Integration
- **Built-in MCP server** (stdio transport) — exposes the currently-open file and rendering state to AI tools
- API keys stored in OS keychain only: macOS Keychain / Windows Credential Manager via `keyring` crate
- No relay server — all AI requests are direct from user machine to provider

---

## 4. Feature Set

### MVP (v0.1)
- [x] Open a `.md` file from CLI: `vibemd README.md`
- [x] Render Markdown with GFM (tables, task lists, strikethrough)
- [x] Light / dark mode (system-following, toggle override)
- [x] 8-bit visual theme (NES.css + Press Start 2P)
- [x] "Open With" registration on macOS and Windows
- [x] Syntax highlighting (Prism.js, locally bundled)
- [x] File-watching: auto-refresh when the `.md` file changes on disk
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

### Mitigations

| Threat | Mitigation |
|--------|------------|
| XSS in rendered HTML (CVE-2024-21535 class) | DOMPurify 3.x with strict config (deny `<iframe>`, `<script>`, `javascript:` URIs) |
| Remote script injection | CSP: `script-src 'self'` — no CDN, no inline scripts |
| Malicious `<iframe src="javascript:...">` | DOMPurify blocks `javascript:` in all attributes |
| AI API key exposure | OS keychain only (macOS Keychain / Windows Credential Manager via Rust `keyring` crate) |
| Remote relay eavesdropping | No relay — AI requests go directly from user machine to provider |
| Electron-style Node.js RCE | Tauri: renderer has no Node.js / no `nodeIntegration` equivalent |
| Mermaid/KaTeX script injection | Both loaded as locally-bundled ESM, not CDN `<script>` tags |

### DOMPurify Config
```js
DOMPurify.sanitize(html, {
  ALLOWED_TAGS: ['p','h1','h2','h3','h4','h5','h6','ul','ol','li','blockquote',
                 'pre','code','em','strong','del','table','thead','tbody','tr',
                 'th','td','a','img','hr','br','details','summary','span','div'],
  ALLOWED_ATTR: ['href','src','alt','title','class','id','target','rel',
                 'data-language','checked','disabled'],
  ALLOW_DATA_ATTR: false,
  FORCE_BODY: true,
  FORBID_TAGS: ['script','iframe','object','embed','form','input','button'],
  FORBID_ATTR: ['onerror','onload','onclick','style'],
})
```

### CSP Header (tauri.conf.json)
```json
"security": {
  "csp": "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'none'"
}
```

---

## 7. File Association ("Open With")

### macOS
In `tauri.conf.json`:
```json
"macOS": {
  "fileAssociations": [
    { "ext": ["md", "markdown", "mdown", "mkd"], "name": "Markdown Document", "role": "Viewer" }
  ]
}
```
This registers the app in the macOS UTI system. After install, right-click any `.md` → "Open With" → vibemd. Can be set as default with "Always Open With."

### Windows
In `tauri.conf.json`:
```json
"windows": {
  "fileAssociations": [
    { "ext": ["md", "markdown"], "name": "Markdown Document" }
  ]
}
```
NSIS/WiX installer registers the file extension. First run prompts to set as default. Windows SmartScreen bypass is expected on first run for unsigned builds — release builds will be code-signed.

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
- [ ] `npm create tauri-app@latest vibemd` — React + TypeScript template
- [ ] Add NES.css, Press Start 2P (locally bundled)
- [ ] Basic window chrome with 8-bit title bar
- [ ] File open via CLI arg + `tauri::Manager` file association handler
- [ ] Wire `nativeTheme` → CSS class toggle

### Phase 1 — Markdown Rendering (Week 2)
- [ ] marked.js integration with GFM extensions
- [ ] DOMPurify sanitization pipeline
- [ ] 8-bit styled CSS for all Markdown elements (headings, code, tables, blockquotes)
- [ ] Prism.js syntax highlighting (locally bundled, dark/light themes)
- [ ] File watcher (`notify` crate in Tauri backend) → IPC → re-render

### Phase 2 — Polish & Distribution (Week 3)
- [ ] Keyboard shortcuts (`Cmd+W`, `Cmd+T`, `Cmd+K`)
- [ ] Status bar (filename, word count, theme indicator, MCP status)
- [ ] macOS "Open With" + Windows file association testing
- [ ] Code signing setup (macOS notarization, Windows Authenticode)
- [ ] GitHub Actions CI: build matrix (macOS-latest, windows-latest)

### Phase 3 — MCP Server (Week 4)
- [ ] Rust MCP server over stdio (`mcp-server` crate or hand-rolled JSON-RPC)
- [ ] Implement 6 tools (see §8)
- [ ] `AGENTS.md` + `CLAUDE.md` in repo
- [ ] Test with `claude mcp add` + Claude Code session
- [ ] Document one-liner in README

### Phase 4 — Power Features (Week 5+)
- [ ] Mermaid.js rendering (locally bundled)
- [ ] KaTeX math rendering (locally bundled)
- [ ] Recent files (stored in Tauri app data dir)
- [ ] Theme switcher UI
- [ ] PDF export via Tauri print API
- [ ] Custom per-file CSS (`.vibemd.css` sidecar)

---

## 10. Repo Structure

```
vibemd/
├── src/                    # React frontend
│   ├── components/
│   │   ├── TitleBar.tsx    # 8-bit window title
│   │   ├── MarkdownView.tsx # Rendered content area
│   │   ├── StatusBar.tsx   # Bottom status strip
│   │   └── CommandPalette.tsx
│   ├── styles/
│   │   ├── nes-overrides.css  # NES.css customizations
│   │   ├── dark.css           # Dark phosphor theme
│   │   ├── light.css          # Light paper theme
│   │   └── markdown.css       # Rendered MD element styles
│   ├── lib/
│   │   ├── renderer.ts    # marked.js + DOMPurify pipeline
│   │   └── mcp.ts         # MCP client for status
│   └── main.tsx
├── src-tauri/
│   ├── src/
│   │   ├── main.rs         # Tauri app entry, file association handler
│   │   ├── watcher.rs      # File watch → emit event
│   │   └── mcp_server.rs   # MCP stdio server
│   └── tauri.conf.json     # File associations, CSP, window config
├── assets/                 # Locally bundled fonts, NES.css
├── AGENTS.md               # AI tool integration docs
├── CLAUDE.md               # Claude Code context
└── DESIGN.md               # This document
```

---

## 11. Open Questions (from Research)

1. **Tauri file associations**: Needs testing on both platforms — the `fileAssociations` field may have edge cases on Windows 10 vs 11 and macOS 12 vs 14.
2. **MCP server crate**: Evaluate `mcp-server` (Rust) vs hand-rolled JSON-RPC over stdio. The ecosystem is young (2025).
3. **DOMPurify in Tauri**: DOMPurify requires a DOM — confirm it runs cleanly in Tauri's WebView renderer process (expected: yes, it runs in the web layer).
4. **Press Start 2P readability**: Heading-only use of the font should be readable, but needs visual testing with real-world `.md` files of varying heading density.
5. **Windows SmartScreen**: First-run bypass UX for unsigned builds — document clearly in README for early adopters.

---

## 12. Sources

- Marked 2 App: https://marked2app.com/
- SoloMD security model: https://github.com/zhitongblog/solomd
- Noteriv MCP pattern: https://github.com/thejacedev/Noteriv
- Tauri CSP: https://v2.tauri.app/security/csp/
- Electron dark mode: https://www.electronjs.org/docs/latest/tutorial/dark-mode
- 8bitcn/ui: https://github.com/TheOrcDev/8bitcn-ui
- NES.css: https://github.com/nostalgic-css/NES.css
- CVE-2024-21535: https://security.snyk.io/vuln/SNYK-JS-MARKDOWNTOJSX-6258886

---

*Generated 2026-06-05. Maintained alongside the codebase — update as decisions are made.*
