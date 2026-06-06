package renderer

import (
	"bytes"
	"regexp"
	"strings"

	highlighting "github.com/yuin/goldmark-highlighting/v2"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Table,
		extension.Strikethrough,
		extension.TaskList,
		highlighting.NewHighlighting(
			highlighting.WithStyle("monokai"),
			highlighting.WithCSSWriter(nil),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithXHTML(),
	),
)

var policy = buildPolicy()

func buildPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements(
		"p", "h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li", "blockquote", "pre", "code",
		"em", "strong", "del", "table", "thead", "tbody",
		"tr", "th", "td", "hr", "br", "details", "summary",
		"span", "div",
	)
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src", "alt", "title").OnElements("img")
	p.AllowAttrs("class", "id").OnElements(
		"code", "pre", "span", "div",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"li", "ul", "ol",
	)
	p.AllowAttrs("checked", "disabled", "type").OnElements("input")
	p.RequireNoFollowOnLinks(true)
	return p
}

// RenderMarkdown converts Markdown to sanitized HTML.
func RenderMarkdown(src string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return "<p>Error rendering markdown</p>"
	}
	return policy.Sanitize(buf.String())
}

// headingRE matches ATX headings (# through ######).
var headingRE = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)

// TOCEntry represents a single heading in the table of contents.
type TOCEntry struct {
	Level  int    `json:"level"`
	Text   string `json:"text"`
	Anchor string `json:"anchor"`
}

// ExtractTOC returns a list of headings from the Markdown source.
func ExtractTOC(src string) []map[string]interface{} {
	matches := headingRE.FindAllStringSubmatch(src, -1)
	result := make([]map[string]interface{}, 0, len(matches))
	for _, m := range matches {
		level := len(m[1])
		text := strings.TrimSpace(m[2])
		anchor := slugify(text)
		result = append(result, map[string]interface{}{
			"level":  level,
			"text":   text,
			"anchor": anchor,
		})
	}
	return result
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlnum.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
