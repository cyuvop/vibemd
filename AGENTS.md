# vibemd — AI Tool Integration

vibemd is a cross-platform Markdown viewer with a built-in MCP server.
AI coding tools can read, navigate, and control the viewer over stdio.

## Register with Claude Code

```bash
claude mcp add vibemd -- vibemd --mcp
```

## Register with other MCP clients (Cursor, Cline, Continue, Windsurf)

Add to your MCP config:
```json
{
  "mcpServers": {
    "vibemd": {
      "command": "vibemd",
      "args": ["--mcp"]
    }
  }
}
```

## Available Tools

| Tool | Description |
|------|-------------|
| `get_current_file` | Path, raw Markdown, word count, last-modified timestamp |
| `get_rendered_html` | Sanitized HTML of the current view |
| `get_toc` | Table of contents as `[{level, text, anchor}]` |
| `scroll_to_heading` | Scroll viewer to a heading (fuzzy match) |
| `set_theme` | Switch between `"light"` and `"dark"` |
| `open_file` | Open any `.md` file in the viewer |

## Security

- API keys are stored in the OS keychain only — never in config files or env vars
- MCP server communicates over stdio only — no network port is opened
- All Markdown rendering is sanitized by bluemonday before the WebView sees it
