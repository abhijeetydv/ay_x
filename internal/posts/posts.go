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
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
	"gopkg.in/yaml.v3"
)

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
//
// Security note: this is the ONLY place in the codebase that constructs template.HTML.
// It is safe because goldmark runs with its default unsafe=false setting (raw HTML
// tags and unsafe URLs in Markdown are stripped/neutralised rather than passed through)
// and all post files are owner-authored local files.
func RenderMarkdown(markdown string, labels CopyLabels) (template.HTML, error) {
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

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return "", fmt.Errorf("rendering markdown: %w", err)
	}
	return template.HTML(buf.String()), nil
}
