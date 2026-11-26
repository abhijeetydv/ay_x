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

func TestBuildHomepage(t *testing.T) {
	data, err := loader.LoadAll(filepath.Join("..", "validate", "testdata", "valid"))
	if err != nil {
		t.Fatalf("LoadAll fixture failed: %v", err)
	}
	if err := validate.Homepage(data); err != nil {
		t.Fatalf("validate.Homepage failed: %v", err)
	}

	const marker = "ZZMARK_RENDER_TEST_GREETING"
	data.Hero.Greeting = marker

	outDir := t.TempDir()
	templatePath := filepath.Join("..", "..", "templates", "home.html")
	rootDir := filepath.Join("..", "..")
	if err := render.BuildHomepage(templatePath, rootDir, outDir, data); err != nil {
		t.Fatalf("BuildHomepage failed: %v", err)
	}

	for _, name := range []string{"index.html", "styles.css", "script.js"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("expected output file %s: %v", name, err)
		}
	}

	raw, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("reading generated index.html: %v", err)
	}
	html := string(raw)
	if !strings.Contains(html, marker) {
		t.Errorf("generated index.html does not contain marker %q", marker)
	}
	for _, bad := range []string{"<no value>", "{{", "%!"} {
		if strings.Contains(html, bad) {
			t.Errorf("generated index.html contains unrendered artifact %q", bad)
		}
	}
}
