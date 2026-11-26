package validate_test

import (
	"testing"

	"portfolio/internal/loader"
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
	if len(data.OrderedSections) != 6 {
		t.Fatalf("expected 6 ordered sections, got %d", len(data.OrderedSections))
	}
	if len(data.FeaturedProjects) != 2 {
		t.Fatalf("expected 2 featured projects, got %d", len(data.FeaturedProjects))
	}
}

func TestInvalidURLRejected(t *testing.T) {
	data, err := loader.LoadAll("../../content")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.Social.HeroLinks[0].URL = "javascript:alert(1)"
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for javascript: URL, got nil")
	}
}

func TestDuplicateProjectIDRejected(t *testing.T) {
	data, err := loader.LoadAll("../../content")
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	data.Projects.Items[1].ID = data.Projects.Items[0].ID
	if err := validate.Homepage(data); err == nil {
		t.Fatal("expected validation error for duplicate project id, got nil")
	}
}
