// Package mcp implements a JSON-RPC 2.0 stdio MCP server exposing
// vibemd's current file state to AI coding tools (Claude Code, Cursor, etc.).
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// State holds shared app state the MCP server reads from.
type State interface {
	GetCurrentFile() map[string]interface{}
	GetRenderedHTML() string
	GetTOC() []map[string]interface{}
	GetFilePath() string
	OpenFile(path string) error
	SetTheme(theme string)
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
)

// Serve reads JSON-RPC requests from r and writes responses to w until EOF.
func Serve(state State, r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)
	enc := json.NewEncoder(w)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			_ = enc.Encode(response{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &rpcError{Code: codeParseError, Message: "parse error"},
			})
			continue
		}

		result, rpcErr := dispatch(state, req.Method, req.Params)
		resp := response{JSONRPC: "2.0", ID: req.ID}
		if rpcErr != nil {
			resp.Error = rpcErr
		} else {
			resp.Result = result
		}
		_ = enc.Encode(resp)
	}
}

func dispatch(state State, method string, params json.RawMessage) (interface{}, *rpcError) {
	switch method {
	case "initialize":
		return map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
			"serverInfo":      map[string]interface{}{"name": "vibemd", "version": "0.1.0"},
		}, nil

	case "tools/list":
		return map[string]interface{}{"tools": toolList()}, nil

	case "tools/call":
		var p struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		return callTool(state, p.Name, p.Arguments)

	default:
		return nil, &rpcError{Code: codeMethodNotFound, Message: fmt.Sprintf("method not found: %s", method)}
	}
}

func callTool(state State, name string, args map[string]interface{}) (interface{}, *rpcError) {
	text := func(s string) interface{} {
		return map[string]interface{}{
			"content": []map[string]interface{}{{"type": "text", "text": s}},
		}
	}
	jsonText := func(v interface{}) interface{} {
		b, _ := json.MarshalIndent(v, "", "  ")
		return text(string(b))
	}

	switch name {
	case "get_current_file":
		return jsonText(state.GetCurrentFile()), nil

	case "get_rendered_html":
		return text(state.GetRenderedHTML()), nil

	case "get_toc":
		return jsonText(state.GetTOC()), nil

	case "scroll_to_heading":
		heading, _ := args["heading"].(string)
		if heading == "" {
			return nil, &rpcError{Code: codeInvalidParams, Message: "heading required"}
		}
		// Emit handled by frontend; return the heading so the AI knows it was received
		return text(fmt.Sprintf("scroll_to_heading: %s", heading)), nil

	case "set_theme":
		theme, _ := args["theme"].(string)
		if theme != "light" && theme != "dark" {
			return nil, &rpcError{Code: codeInvalidParams, Message: "theme must be 'light' or 'dark'"}
		}
		state.SetTheme(theme)
		return text("theme set to " + theme), nil

	case "open_file":
		path, _ := args["path"].(string)
		if path == "" {
			return nil, &rpcError{Code: codeInvalidParams, Message: "path required"}
		}
		if err := state.OpenFile(path); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		return text("opened " + path), nil

	default:
		return nil, &rpcError{Code: codeMethodNotFound, Message: fmt.Sprintf("unknown tool: %s", name)}
	}
}

func toolList() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "get_current_file",
			"description": "Returns the path, raw Markdown, word count, and last-modified time of the currently open file.",
			"inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		{
			"name":        "get_rendered_html",
			"description": "Returns the sanitized HTML rendered from the current Markdown file.",
			"inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		{
			"name":        "get_toc",
			"description": "Returns the table of contents as a JSON array of {level, text, anchor} objects.",
			"inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		{
			"name":        "scroll_to_heading",
			"description": "Scrolls the vibemd viewer to a heading matching the given text.",
			"inputSchema": map[string]interface{}{
				"type":     "object",
				"required": []string{"heading"},
				"properties": map[string]interface{}{
					"heading": map[string]interface{}{"type": "string", "description": "Heading text to scroll to"},
				},
			},
		},
		{
			"name":        "set_theme",
			"description": "Switches the viewer between light and dark mode.",
			"inputSchema": map[string]interface{}{
				"type":     "object",
				"required": []string{"theme"},
				"properties": map[string]interface{}{
					"theme": map[string]interface{}{"type": "string", "enum": []string{"light", "dark"}},
				},
			},
		},
		{
			"name":        "open_file",
			"description": "Opens a Markdown file at the given absolute or relative path in vibemd.",
			"inputSchema": map[string]interface{}{
				"type":     "object",
				"required": []string{"path"},
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Path to the .md file"},
				},
			},
		},
	}
}

// RunStdio is the entry point for `vibemd --mcp`. Uses stdin/stdout.
func RunStdio(state State) {
	Serve(state, os.Stdin, os.Stdout)
}
