package render_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"portfolio/internal/loader"
	"portfolio/internal/render"
	"portfolio/internal/validate"
)

func TestBuildSite(t *testing.T) {
	data, err := loader.LoadAll(filepath.Join("..", "validate", "testdata", "valid"))
	if err != nil {
		t.Fatalf("LoadAll fixture failed: %v", err)
	}
	if err := validate.Homepage(data); err != nil {
		t.Fatalf("validate.Homepage failed: %v", err)
	}

	const marker = "ZZMARK_RENDER_TEST_GREETING"
	const aboutMarker = "ZZMARK_RENDER_TEST_ABOUT_PARAGRAPH"
	const projectsMarker = "ZZMARK_RENDER_TEST_PROJECTS_HEADING"
	data.Hero.Greeting = marker
	data.AboutPage.Paragraphs = append(data.AboutPage.Paragraphs, aboutMarker)
	data.Projects.Page.Heading = projectsMarker

	outDir := t.TempDir()
	templateDir := filepath.Join("..", "..", "templates")
	rootDir := filepath.Join("..", "..")
	if err := render.BuildSite(templateDir, rootDir, outDir, data); err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	for _, name := range []string{"index.html", "about.html", "projects.html", "styles.css", "script.js"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("expected output file %s: %v", name, err)
		}
	}

	expectedMarkers := map[string]string{
		"index.html":    marker,
		"about.html":    aboutMarker,
		"projects.html": projectsMarker,
	}
	for _, page := range render.DefaultPages() {
		raw, err := os.ReadFile(filepath.Join(outDir, page.Output))
		if err != nil {
			t.Fatalf("reading generated %s: %v", page.Output, err)
		}
		html := string(raw)
		if m, ok := expectedMarkers[page.Output]; ok && !strings.Contains(html, m) {
			t.Errorf("generated %s does not contain marker %q", page.Output, m)
		}
		for _, bad := range []string{"<no value>", "{{", "%!"} {
			if strings.Contains(html, bad) {
				t.Errorf("generated %s contains unrendered artifact %q", page.Output, bad)
			}
		}
	}
}
