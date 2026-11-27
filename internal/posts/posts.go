package posts

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"gopkg.in/yaml.v3"
)

// TOCItem represents a heading entry extracted from Markdown for the table of contents.
type TOCItem struct {
	ID       string
	Text     string
	Level    int
	Children []TOCItem
}

// CopyLabels configures accessible labels for fenced code-block copy buttons.
type CopyLabels struct {
	Copy   string
	Copied string
}

// TitleColor represents a coloured substring entry in a post's front matter title_colors.
type TitleColor struct {
	Text  string `yaml:"text"`
	Color string `yaml:"color"`
}

// FrontMatter represents the strict YAML front matter of a Markdown post.
type FrontMatter struct {
	Title       string       `yaml:"title"`
	TitleColors []TitleColor `yaml:"title_colors,omitempty"`
	Date        string       `yaml:"date"`
	Updated     string       `yaml:"updated,omitempty"`
	Summary     string       `yaml:"summary"`
	Tags        []string     `yaml:"tags,omitempty"`
}

// PostFile holds the parsed front matter, slug, and raw Markdown body of a post.
type PostFile struct {
	Slug        string
	SourcePath  string
	FrontMatter FrontMatter
	Body        string
}

// LoadDir reads all *.md files in dir, parses strict YAML front matter and Markdown body,
// and returns them sorted by filename. Non-.md files and subdirectories are ignored.
// A missing directory returns an empty slice without error.
func LoadDir(dir string) ([]PostFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading posts directory %s: %w", dir, err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	posts := make([]PostFile, 0, len(names))
	for _, name := range names {
		path := filepath.Join(dir, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading post %s: %w", path, err)
		}
		fm, body, err := SplitFrontMatter(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		slug := strings.TrimSuffix(name, ".md")
		posts = append(posts, PostFile{
			Slug:        slug,
			SourcePath:  path,
			FrontMatter: fm,
			Body:        body,
		})
	}
	return posts, nil
}

// SplitFrontMatter splits leading YAML front matter (delimited by "---" lines)
// from the Markdown body and decodes it with KnownFields(true).
func SplitFrontMatter(raw []byte) (FrontMatter, string, error) {
	normalized := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return FrontMatter{}, "", fmt.Errorf("missing leading '---' front matter delimiter")
	}
	rest := normalized[len("---\n"):]
	idx := strings.Index(rest, "\n---\n")
	endOffset := len("\n---\n")
	if idx == -1 {
		if strings.HasSuffix(rest, "\n---") {
			idx = len(rest) - len("\n---")
			endOffset = len("\n---")
		} else {
			return FrontMatter{}, "", fmt.Errorf("missing closing '---' front matter delimiter")
		}
	}

	fmRaw := rest[:idx]
	body := rest[idx+endOffset:]

	var fm FrontMatter
	dec := yaml.NewDecoder(strings.NewReader(fmRaw))
	dec.KnownFields(true)
	if err := dec.Decode(&fm); err != nil {
		return FrontMatter{}, "", fmt.Errorf("decoding front matter: %w", err)
	}
	return fm, body, nil
}

type preWrapper struct{}

func (preWrapper) Start(code bool, styleAttr string) string {
	if code {
		return `<pre tabindex="0"` + styleAttr + `><code>`
	}
	return `<pre tabindex="0"` + styleAttr + `>`
}

func (preWrapper) End(code bool) string {
	if code {
		return `</code></pre>`
	}
	return `</pre>`
}

// RenderMarkdown renders owner-authored Markdown into sanitized HTML.
func RenderMarkdown(markdown string, labels CopyLabels) (template.HTML, error) {
	body, _, err := RenderMarkdownWithTOC(markdown, labels)
	return body, err
}

// RenderMarkdownWithTOC renders owner-authored Markdown into sanitized HTML and extracts
// a nested table of contents (h2 and h3 headings with their goldmark-assigned IDs).
//
// Security note: this is the ONLY place in the codebase that constructs template.HTML.
// It is safe because goldmark runs with its default unsafe=false setting (raw HTML
// tags and unsafe URLs in Markdown are stripped/neutralised rather than passed through)
// and all post files are owner-authored local files.
func RenderMarkdownWithTOC(markdown string, labels CopyLabels) (template.HTML, []TOCItem, error) {
	copyLabel := labels.Copy
	if copyLabel == "" {
		copyLabel = "Copy code"
	}
	copiedLabel := labels.Copied
	if copiedLabel == "" {
		copiedLabel = "Copied"
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
					chromahtml.TabWidth(4),
					chromahtml.WithPreWrapper(preWrapper{}),
				),
				highlighting.WithWrapperRenderer(func(w util.BufWriter, c highlighting.CodeBlockContext, entering bool) {
					if entering {
						_, _ = w.WriteString(`<div class="code-block">`)
						if lang, ok := c.Language(); ok && len(bytes.TrimSpace(lang)) > 0 {
							_, _ = w.WriteString(`<span class="code-lang">`)
							_, _ = w.WriteString(html.EscapeString(string(bytes.TrimSpace(lang))))
							_, _ = w.WriteString(`</span>`)
						}
						_, _ = fmt.Fprintf(
							w,
							`<button type="button" class="code-copy" aria-label="%s" data-copy-label="%s" data-copied-label="%s"><span class="code-copy-icon" aria-hidden="true"></span><span class="sr-only">%s</span></button>`,
							html.EscapeString(copyLabel),
							html.EscapeString(copyLabel),
							html.EscapeString(copiedLabel),
							html.EscapeString(copyLabel),
						)
						if !c.Highlighted() {
							_, _ = w.WriteString(`<pre tabindex="0"><code>`)
						}
					} else {
						if !c.Highlighted() {
							_, _ = w.WriteString(`</code></pre>`)
						}
						_, _ = w.WriteString("</div>\n")
					}
				}),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		// Leave goldmark's default unsafe=false: raw HTML in Markdown is NOT passed through.
	)

	source := []byte(markdown)
	doc := md.Parser().Parse(text.NewReader(source))

	var toc []TOCItem
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok || (h.Level != 2 && h.Level != 3) {
			return ast.WalkContinue, nil
		}
		idVal, hasID := h.AttributeString("id")
		if !hasID {
			return ast.WalkContinue, nil
		}
		var id string
		switch v := idVal.(type) {
		case []byte:
			id = string(v)
		case string:
			id = v
		}
		if id == "" {
			return ast.WalkContinue, nil
		}
		headingText := strings.TrimSpace(extractPlainText(h, source))
		if headingText == "" {
			return ast.WalkContinue, nil
		}
		item := TOCItem{
			ID:    id,
			Text:  headingText,
			Level: h.Level,
		}
		if h.Level == 3 && len(toc) > 0 && toc[len(toc)-1].Level == 2 {
			toc[len(toc)-1].Children = append(toc[len(toc)-1].Children, item)
		} else {
			toc = append(toc, item)
		}
		return ast.WalkSkipChildren, nil
	})

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, source, doc); err != nil {
		return "", nil, fmt.Errorf("rendering markdown: %w", err)
	}
	return template.HTML(buf.String()), toc, nil
}

func extractPlainText(node ast.Node, source []byte) string {
	var b strings.Builder
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Text:
			b.Write(v.Segment.Value(source))
			if v.SoftLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(v.Value)
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}
