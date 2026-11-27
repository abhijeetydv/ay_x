package markdown_test

import (
	"strings"
	"testing"

	"portfolio/internal/markdown"
)

func TestRenderMarkdownFeatures(t *testing.T) {
	md := `## Section One

Paragraph with ` + "`inline`" + `, **bold**, *italic*, ~~strike~~, and [a link](https://example.com).

- item one
- item two

> A quoted line.

| Key | Value |
| --- | --- |
| go  | fast  |
`
	out, err := markdown.RenderMarkdown(md, markdown.CopyLabels{Copy: "Copy code", Copied: "Copied"})
	if err != nil {
		t.Fatalf("RenderMarkdown failed: %v", err)
	}
	html := string(out)

	checks := []string{
		`<h2 id="section-one">Section One</h2>`,
		`<code>inline</code>`,
		`<strong>bold</strong>`,
		`<em>italic</em>`,
		`<del>strike</del>`,
		`<a href="https://example.com">a link</a>`,
		`<ul>`,
		`<blockquote>`,
		`<table>`,
	}
	for _, want := range checks {
		if !strings.Contains(html, want) {
			t.Errorf("rendered HTML missing %q\ngot:\n%s", want, html)
		}
	}
}

func TestRenderMarkdownCodeBlocks(t *testing.T) {
	md := "```go\npackage main\nfunc main() {}\n```\n\n```\nplain text\n```\n\n```madeuplang\nhello\n```\n"
	out, err := markdown.RenderMarkdown(md, markdown.CopyLabels{Copy: "Copy snippet", Copied: "Copied!"})
	if err != nil {
		t.Fatalf("RenderMarkdown failed: %v", err)
	}
	html := string(out)

	if !strings.Contains(html, `<div class="code-block"><span class="code-lang">go</span>`) {
		t.Errorf("expected language label for go block, got:\n%s", html)
	}
	if !strings.Contains(html, `class="code-copy" aria-label="Copy snippet" data-copy-label="Copy snippet" data-copied-label="Copied!"`) {
		t.Errorf("expected copy button with YAML labels, got:\n%s", html)
	}
	if !strings.Contains(html, `tabindex="0"`) {
		t.Errorf("expected tabindex=\"0\" on <pre>, got:\n%s", html)
	}
	if !strings.Contains(html, `<span class="kn">package</span>`) {
		t.Errorf("expected Chroma CSS class spans in go block, got:\n%s", html)
	}
	if count := strings.Count(html, `class="code-block"`); count != 3 {
		t.Errorf("expected 3 code-block wrappers, got %d", count)
	}
	if count := strings.Count(html, `class="code-lang"`); count != 2 {
		t.Errorf("expected 2 code-lang labels (none for plain fence), got %d", count)
	}
	if !strings.Contains(html, "plain text") || !strings.Contains(html, "hello") {
		t.Errorf("expected plain and unknown-language fences to render content")
	}
}

func TestRenderMarkdownNeutralisesRawScript(t *testing.T) {
	md := "Hello\n\n<script>alert(1)</script>\n\nWorld"
	out, err := markdown.RenderMarkdown(md, markdown.CopyLabels{Copy: "Copy", Copied: "Copied"})
	if err != nil {
		t.Fatalf("RenderMarkdown failed: %v", err)
	}
	html := string(out)
	if strings.Contains(html, "<script>") {
		t.Fatalf("raw <script> tag leaked into rendered HTML:\n%s", html)
	}
}

func TestRenderMarkdownWithTOC(t *testing.T) {
	md := `### Orphan Subheading

## Section with ` + "`code`" + ` and [link](https://example.com)

### Nested Child

## Duplicate

### Duplicate
`
	out, toc, err := markdown.RenderMarkdownWithTOC(md, markdown.CopyLabels{Copy: "Copy", Copied: "Copied"})
	if err != nil {
		t.Fatalf("RenderMarkdownWithTOC failed: %v", err)
	}
	html := string(out)

	if len(toc) != 3 {
		t.Fatalf("expected 3 top-level TOC items, got %d: %+v", len(toc), toc)
	}
	if toc[0].Level != 3 || toc[0].Text != "Orphan Subheading" {
		t.Errorf("unexpected first TOC item: %+v", toc[0])
	}
	if toc[1].Text != "Section with code and link" {
		t.Errorf("expected plain text stripped of inline markup, got %q", toc[1].Text)
	}
	if len(toc[1].Children) != 1 || toc[1].Children[0].Text != "Nested Child" {
		t.Errorf("expected Nested Child under Section, got %+v", toc[1].Children)
	}
	if len(toc[2].Children) != 1 || toc[2].ID == toc[2].Children[0].ID {
		t.Errorf("expected duplicate headings to have distinct IDs, got parent=%q child=%+v", toc[2].ID, toc[2].Children)
	}

	var allIDs []string
	for _, item := range toc {
		allIDs = append(allIDs, item.ID)
		for _, ch := range item.Children {
			allIDs = append(allIDs, ch.ID)
		}
	}
	for _, id := range allIDs {
		if !strings.Contains(html, `id="`+id+`"`) {
			t.Errorf("TOC id %q not found in rendered HTML:\n%s", id, html)
		}
	}
}
