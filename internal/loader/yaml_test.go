package loader_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"portfolio/internal/loader"
)

func TestUnknownYAMLKeyRejected(t *testing.T) {
	dir := t.TempDir()
	fixtureDir := filepath.Join("..", "validate", "testdata", "valid")
	entries, err := os.ReadDir(fixtureDir)
	if err != nil {
		t.Fatalf("reading fixture dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(fixtureDir, e.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", e.Name(), err)
		}
		if e.Name() == "site.yaml" {
			raw = append(raw, []byte("\nunknown_extra_key: \"should_fail\"\n")...)
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), raw, 0o644); err != nil {
			t.Fatalf("writing %s: %v", e.Name(), err)
		}
	}

	_, err = loader.LoadAll(dir)
	if err == nil {
		t.Fatal("expected error for unknown YAML key, got nil")
	}
	if !strings.Contains(err.Error(), "unknown_extra_key") {
		t.Fatalf("expected error mentioning unknown_extra_key, got: %v", err)
	}
}
