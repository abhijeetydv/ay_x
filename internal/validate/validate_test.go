package validate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"portfolio/internal/loader"
	"portfolio/internal/model"
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
	seenIDs := make(map[string]bool, len(data.Projects.Items))
	for _, item := range data.Projects.Items {
		if seenIDs[item.ID] {
			t.Fatalf("duplicate project ID %q", item.ID)
		}
		seenIDs[item.ID] = true
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
	if len(data.FeaturedProjects) != 2 {
		t.Fatalf("expected 2 featured projects, got %d", len(data.FeaturedProjects))
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
	data.Projects.Items[1].ID = data.Projects.Items[0].ID
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for duplicate project id, got nil")
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

func TestInvalidContactEmailRejected(t *testing.T) {
	data, err := loader.LoadAll("testdata/valid")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.Contact.Email.Address = "not-an-email"
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for invalid contact email, got nil")
	}
}

func TestAllowedTitleStylesExistInCSS(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "styles.css"))
	if err != nil {
		t.Fatalf("reading styles.css: %v", err)
	}
	css := string(raw)
	for _, st := range model.AllowedTitleStyles {
		cls := ".title-" + st
		if !strings.Contains(css, cls) {
			t.Errorf("styles.css missing CSS class %s for allowlisted title style %q", cls, st)
		}
	}
}
