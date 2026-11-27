package posts

import (
	"fmt"
	"html/template"

	"portfolio/internal/markdown"
)

// TOCItem aliases markdown.TOCItem for backward compatibility.
type TOCItem = markdown.TOCItem

// CopyLabels aliases markdown.CopyLabels for backward compatibility.
type CopyLabels = markdown.CopyLabels

// TitleColor aliases markdown.TitleColor for backward compatibility.
type TitleColor = markdown.TitleColor

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
// and returns them sorted by filename.
func LoadDir(dir string) ([]PostFile, error) {
	files, err := markdown.ListDir(dir)
	if err != nil {
		return nil, err
	}
	posts := make([]PostFile, 0, len(files))
	for _, f := range files {
		fm, body, err := SplitFrontMatter(f.Raw)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", f.SourcePath, err)
		}
		posts = append(posts, PostFile{
			Slug:        f.Slug,
			SourcePath:  f.SourcePath,
			FrontMatter: fm,
			Body:        body,
		})
	}
	return posts, nil
}

// SplitFrontMatter splits leading YAML front matter (delimited by "---" lines)
// from the Markdown body and decodes it with KnownFields(true).
func SplitFrontMatter(raw []byte) (FrontMatter, string, error) {
	var fm FrontMatter
	body, err := markdown.SplitFrontMatterInto(raw, &fm)
	if err != nil {
		return FrontMatter{}, "", err
	}
	return fm, body, nil
}

// RenderMarkdown renders owner-authored Markdown into sanitized HTML.
func RenderMarkdown(src string, labels CopyLabels) (template.HTML, error) {
	return markdown.RenderMarkdown(src, labels)
}

// RenderMarkdownWithTOC renders owner-authored Markdown into sanitized HTML and extracts a TOC.
func RenderMarkdownWithTOC(src string, labels CopyLabels) (template.HTML, []TOCItem, error) {
	return markdown.RenderMarkdownWithTOC(src, labels)
}
