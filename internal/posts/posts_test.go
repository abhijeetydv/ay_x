package posts_test

import (
	"strings"
	"testing"

	"portfolio/internal/posts"
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
	out, err := posts.RenderMarkdown(md, posts.CopyLabels{Copy: "Copy code", Copied: "Copied"})
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
	out, err := posts.RenderMarkdown(md, posts.CopyLabels{Copy: "Copy snippet", Copied: "Copied!"})
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
	// Fence without language must have no .code-lang span (only 2 .code-lang spans total: go and madeuplang, or 1 if madeuplang has it)
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
	out, err := posts.RenderMarkdown(md, posts.CopyLabels{Copy: "Copy", Copied: "Copied"})
	if err != nil {
		t.Fatalf("RenderMarkdown failed: %v", err)
	}
	html := string(out)
	if strings.Contains(html, "<script>") {
		t.Fatalf("raw <script> tag leaked into rendered HTML:\n%s", html)
	}
}

func TestSplitFrontMatterRejectsUnknownKey(t *testing.T) {
	raw := []byte("---\ntitle: \"Hello\"\ndate: \"2026-01-01\"\nsummary: \"Sum\"\nextra_unknown: true\n---\n\nBody\n")
	if _, _, err := posts.SplitFrontMatter(raw); err == nil {
		t.Fatal("expected error for unknown front matter key, got nil")
	}
}
