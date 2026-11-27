package validate_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"portfolio/internal/loader"
	"portfolio/internal/model"
	"portfolio/internal/posts"
	"portfolio/internal/validate"
)

func TestValidContentDirectory(t *testing.T) {
	data, err := loader.LoadAll("../../content")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	if err := validate.Homepage(data); err != nil {
		t.Fatalf("Homepage validation failed: %v", err)
	}
	if len(data.OrderedSections) == 0 {
		t.Fatal("expected at least 1 ordered section, got 0")
	}
	if len(data.FeaturedProjects) == 0 {
		t.Fatal("expected at least 1 featured project, got 0")
	}
	seenIDs := make(map[string]bool, len(data.AllProjects))
	for _, item := range data.AllProjects {
		if seenIDs[item.Slug] {
			t.Fatalf("duplicate project slug %q", item.Slug)
		}
		seenIDs[item.Slug] = true
	}
}

func TestFixtureContentCounts(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll fixture failed: %v", err)
	}
	if err := validate.Homepage(data); err != nil {
		t.Fatalf("fixture validation failed: %v", err)
	}
	if len(data.OrderedSections) != 6 {
		t.Fatalf("expected 6 ordered sections, got %d", len(data.OrderedSections))
	}
	if len(data.FeaturedProjects) != 1 {
		t.Fatalf("expected 1 featured project, got %d", len(data.FeaturedProjects))
	}
}

func TestInvalidURLRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.Social.HeroLinks[0].URL = "javascript:alert(1)"
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for javascript: URL, got nil")
	}
}

func TestDuplicateProjectIDRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.AllProjects[1].Slug = data.AllProjects[0].Slug
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for duplicate project slug, got nil")
	}
}

func TestNonHTTPSBaseURLRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.Site.URL = "http://example.com/"
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for non-https site.url, got nil")
	}
}

func TestMissingOGImageRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.Site.OGImage = "assets/does-not-exist.png"
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for missing og_image, got nil")
	}
}

func TestBadPostDateRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.SortedPosts[0].Date = "2026/01/10"
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for bad post date, got nil")
	}
}

func TestDuplicatePostSlugRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.SortedPosts[1].Slug = data.SortedPosts[0].Slug
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for duplicate post slug, got nil")
	}
}

func TestInvalidDockEmailRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.Social.Dock.Email.Address = "not-an-email"
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for invalid dock email, got nil")
	}
}

func TestUnknownDockItemRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.Social.Dock.Items = append(data.Social.Dock.Items, "mastodon")
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for unknown dock item, got nil")
	}
}

func TestAllowedTitleColorsMatchCSS(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "styles.css"))
	if err != nil {
		t.Fatalf("reading styles.css: %v", err)
	}
	css := string(raw)

	allowed := make(map[string]bool, len(model.AllowedTitleColors))
	for _, c := range model.AllowedTitleColors {
		allowed[c] = true
		if !strings.Contains(css, ".title-"+c) {
			t.Errorf("styles.css missing CSS class .title-%s for allowlisted colour %q", c, c)
		}
		if !strings.Contains(css, "--color-"+c+":") {
			t.Errorf("styles.css missing CSS variable --color-%s for allowlisted colour %q", c, c)
		}
	}

	titleClassRe := regexp.MustCompile(`\.title-([a-z0-9-]+)`)
	for _, m := range titleClassRe.FindAllStringSubmatch(css, -1) {
		name := m[1]
		if name == "color" {
			continue
		}
		if !allowed[name] {
			t.Errorf("styles.css contains unallowed .title-%s class", name)
		}
	}
}

func TestBadPostSlugRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.SortedPosts[0].Slug = "Bad_Slug"
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for bad post slug, got nil")
	}
}

func TestTitleColorNotInTitleRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.SortedPosts[0].TitleColors = []model.TitleColor{{Text: "MissingWord", Color: "blue"}}
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for title_colors text not in title, got nil")
	}
}

func TestTitleColorOverlapRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.SortedPosts[0].Title = "Building Scalable APIs"
	data.SortedPosts[0].TitleColors = []model.TitleColor{
		{Text: "Building Scalable", Color: "cyan"},
		{Text: "Scalable APIs", Color: "blue"},
	}
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for overlapping title_colors, got nil")
	}
}

func TestTitleColorDuplicateTextFirstOccurrenceAllowed(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.SortedPosts[0].Title = "Go and Go Again"
	data.SortedPosts[0].TitleColors = []model.TitleColor{{Text: "Go", Color: "blue"}}
	if err := validate.Homepage(data); err != nil {
		t.Fatalf("expected duplicate text first-occurrence rule to succeed, got: %v", err)
	}
}

func TestUnknownTitleColorRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.SortedPosts[0].TitleColors = []model.TitleColor{{Text: data.SortedPosts[0].Title, Color: "teal"}}
	err = validate.Homepage(data)
	if err == nil {
		t.Fatal("expected validation error for unknown title color, got nil")
	}
	if !strings.Contains(err.Error(), "allowed:") || !strings.Contains(err.Error(), "blue") {
		t.Errorf("expected error listing allowed colours, got: %v", err)
	}
}

func TestUnknownTitleColorKeyRejected(t *testing.T) {
	raw := []byte("---\ntitle: \"Hello\"\ntitle_colors:\n  - text: \"Hello\"\n    color: \"blue\"\n    extra_key: 1\ndate: \"2026-01-01\"\nsummary: \"Sum\"\n---\nBody\n")
	if _, _, err := posts.SplitFrontMatter(raw); err == nil {
		t.Fatal("expected error for unknown key inside title_colors entry, got nil")
	}
}

func TestLongPostSummaryRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.SortedPosts[0].Summary = strings.Repeat("a", 201)
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for summary > 200 chars, got nil")
	}
}

func TestNewTabWithoutLabelRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.Site.Nav.NewTabLabel = ""
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error when new_tab_label is empty for new_tab link, got nil")
	}
}

func TestMissingNavAssetRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	for i := range data.Site.Nav.Links {
		if data.Site.Nav.Links[i].NewTab {
			data.Site.Nav.Links[i].Href = "assets/does-not-exist.pdf"
		}
	}
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for missing nav asset file, got nil")
	}
}

func TestDuplicateDockItemRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.Social.Dock.Items = append(data.Social.Dock.Items, data.Social.Dock.Items[0])
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for duplicate dock item, got nil")
	}
}
