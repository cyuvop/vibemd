package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// mockState implements State for testing.
type mockState struct {
	filePath    string
	markdown    string
	html        string
	toc         []map[string]interface{}
	themeSet    string
	openedFile  string
}

func (m *mockState) GetCurrentFile() map[string]interface{} {
	return map[string]interface{}{
		"path":     m.filePath,
		"filename": "test.md",
		"wordCount": 3,
	}
}
func (m *mockState) GetRenderedHTML() string                    { return m.html }
func (m *mockState) GetTOC() []map[string]interface{}           { return m.toc }
func (m *mockState) GetFilePath() string                        { return m.filePath }
func (m *mockState) OpenFile(path string) error                 { m.openedFile = path; return nil }
func (m *mockState) SetTheme(theme string)                      { m.themeSet = theme }

func rpcCall(t *testing.T, state State, method string, params interface{}) map[string]interface{} {
	t.Helper()
	paramBytes, _ := json.Marshal(params)
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  json.RawMessage(paramBytes),
	}
	reqLine, _ := json.Marshal(req)

	var out bytes.Buffer
	Serve(state, strings.NewReader(string(reqLine)+"\n"), &out)

	var resp map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response JSON: %v\nraw: %s", err, out.String())
	}
	return resp
}

func TestInitialize(t *testing.T) {
	resp := rpcCall(t, &mockState{}, "initialize", map[string]interface{}{})
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
	result := resp["result"].(map[string]interface{})
	if result["protocolVersion"] == nil {
		t.Error("expected protocolVersion in initialize response")
	}
}

func TestToolsList(t *testing.T) {
	resp := rpcCall(t, &mockState{}, "tools/list", map[string]interface{}{})
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
	result := resp["result"].(map[string]interface{})
	tools := result["tools"].([]interface{})
	if len(tools) != 6 {
		t.Errorf("expected 6 tools, got %d", len(tools))
	}
}

func TestGetCurrentFile(t *testing.T) {
	state := &mockState{filePath: "/tmp/test.md"}
	resp := rpcCall(t, state, "tools/call", map[string]interface{}{
		"name": "get_current_file", "arguments": map[string]interface{}{},
	})
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
}

func TestGetRenderedHTML(t *testing.T) {
	state := &mockState{html: "<h1>Test</h1>"}
	resp := rpcCall(t, state, "tools/call", map[string]interface{}{
		"name": "get_rendered_html", "arguments": map[string]interface{}{},
	})
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
}

func TestSetTheme_Valid(t *testing.T) {
	state := &mockState{}
	resp := rpcCall(t, state, "tools/call", map[string]interface{}{
		"name": "set_theme", "arguments": map[string]interface{}{"theme": "light"},
	})
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
	if state.themeSet != "light" {
		t.Errorf("expected theme 'light', got %q", state.themeSet)
	}
}

func TestSetTheme_Invalid(t *testing.T) {
	state := &mockState{}
	resp := rpcCall(t, state, "tools/call", map[string]interface{}{
		"name": "set_theme", "arguments": map[string]interface{}{"theme": "rainbow"},
	})
	if resp["error"] == nil {
		t.Error("expected error for invalid theme")
	}
}

func TestOpenFile(t *testing.T) {
	state := &mockState{}
	resp := rpcCall(t, state, "tools/call", map[string]interface{}{
		"name": "open_file", "arguments": map[string]interface{}{"path": "/tmp/foo.md"},
	})
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
	if state.openedFile != "/tmp/foo.md" {
		t.Errorf("expected openedFile '/tmp/foo.md', got %q", state.openedFile)
	}
}

func TestUnknownMethod(t *testing.T) {
	resp := rpcCall(t, &mockState{}, "no_such_method", map[string]interface{}{})
	if resp["error"] == nil {
		t.Error("expected error for unknown method")
	}
}

func TestUnknownTool(t *testing.T) {
	resp := rpcCall(t, &mockState{}, "tools/call", map[string]interface{}{
		"name": "delete_everything", "arguments": map[string]interface{}{},
	})
	if resp["error"] == nil {
		t.Error("expected error for unknown tool")
	}
}
