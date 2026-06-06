package renderer

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_Heading(t *testing.T) {
	html := RenderMarkdown("# Hello World")
	if !strings.Contains(html, "<h1") || !strings.Contains(html, "Hello World") {
		t.Errorf("expected h1 with text, got: %s", html)
	}
}

func TestRenderMarkdown_Paragraph(t *testing.T) {
	html := RenderMarkdown("Just a paragraph.")
	if !strings.Contains(html, "<p>") {
		t.Errorf("expected paragraph, got: %s", html)
	}
}

func TestRenderMarkdown_GFMTable(t *testing.T) {
	src := "| Col A | Col B |\n|-------|-------|\n| 1     | 2     |"
	html := RenderMarkdown(src)
	if !strings.Contains(html, "<table") {
		t.Errorf("expected table, got: %s", html)
	}
}

func TestRenderMarkdown_TaskList(t *testing.T) {
	src := "- [x] Done\n- [ ] Not done"
	html := RenderMarkdown(src)
	if !strings.Contains(html, "checked") {
		t.Errorf("expected checked input, got: %s", html)
	}
}

func TestRenderMarkdown_XSSScriptStripped(t *testing.T) {
	html := RenderMarkdown("<script>alert('xss')</script>")
	if strings.Contains(html, "<script") {
		t.Errorf("XSS script tag not stripped: %s", html)
	}
}

func TestRenderMarkdown_XSSIframeBlocked(t *testing.T) {
	// CVE-2024-21535 class: iframe with javascript: src
	html := RenderMarkdown(`<iframe src="javascript:alert()"></iframe>`)
	if strings.Contains(html, "<iframe") {
		t.Errorf("iframe not stripped: %s", html)
	}
	if strings.Contains(html, "javascript:") {
		t.Errorf("javascript: URI not stripped: %s", html)
	}
}

func TestRenderMarkdown_XSSOnclick(t *testing.T) {
	html := RenderMarkdown(`<p onclick="evil()">text</p>`)
	if strings.Contains(html, "onclick") {
		t.Errorf("onclick not stripped: %s", html)
	}
}

func TestRenderMarkdown_CodeBlock(t *testing.T) {
	src := "```go\nfmt.Println(\"hello\")\n```"
	html := RenderMarkdown(src)
	if !strings.Contains(html, "<code") {
		t.Errorf("expected code block, got: %s", html)
	}
}

func TestRenderMarkdown_Strikethrough(t *testing.T) {
	html := RenderMarkdown("~~deleted~~")
	if !strings.Contains(html, "<del") {
		t.Errorf("expected del tag, got: %s", html)
	}
}

func TestRenderMarkdown_Empty(t *testing.T) {
	html := RenderMarkdown("")
	// Should not panic; empty or whitespace output is fine
	if strings.Contains(html, "<script") {
		t.Errorf("unexpected script in empty render: %s", html)
	}
}

func TestExtractTOC_Basic(t *testing.T) {
	src := "# Title\n## Section\n### Sub"
	toc := ExtractTOC(src)
	if len(toc) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(toc))
	}
	if toc[0]["level"] != 1 || toc[0]["text"] != "Title" {
		t.Errorf("wrong first entry: %v", toc[0])
	}
	if toc[1]["level"] != 2 || toc[1]["text"] != "Section" {
		t.Errorf("wrong second entry: %v", toc[1])
	}
}

func TestExtractTOC_Anchor(t *testing.T) {
	toc := ExtractTOC("## Hello World")
	if len(toc) == 0 {
		t.Fatal("expected toc entry")
	}
	if toc[0]["anchor"] != "hello-world" {
		t.Errorf("expected anchor 'hello-world', got %v", toc[0]["anchor"])
	}
}

func TestExtractTOC_Empty(t *testing.T) {
	toc := ExtractTOC("no headings here")
	if len(toc) != 0 {
		t.Errorf("expected empty toc, got %v", toc)
	}
}

func TestSlugify(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Hello World", "hello-world"},
		{"Foo & Bar", "foo-bar"},
		{"  spaces  ", "spaces"},
		{"CamelCase", "camelcase"},
	}
	for _, c := range cases {
		got := slugify(c.in)
		if got != c.want {
			t.Errorf("slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
