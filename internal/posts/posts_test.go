package posts_test

import (
	"testing"

	"portfolio/internal/posts"
)

func TestSplitFrontMatterRejectsUnknownKey(t *testing.T) {
	raw := []byte("---\ntitle: \"Hello\"\ndate: \"2026-01-01\"\nsummary: \"Sum\"\nextra_unknown: true\n---\n\nBody\n")
	if _, _, err := posts.SplitFrontMatter(raw); err == nil {
		t.Fatal("expected error for unknown front matter key, got nil")
	}
}
