package render_test

import (
	"os"
	"path/filepath"
	"regexp"
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

func TestLinkAndAssetIntegrity(t *testing.T) {
	data, err := loader.LoadAll(filepath.Join("..", "validate", "testdata", "valid"))
	if err != nil {
		t.Fatalf("LoadAll fixture failed: %v", err)
	}
	if err := validate.Homepage(data); err != nil {
		t.Fatalf("validate.Homepage failed: %v", err)
	}

	outDir := t.TempDir()
	templateDir := filepath.Join("..", "..", "templates")
	rootDir := filepath.Join("..", "..")
	if err := render.BuildSite(templateDir, rootDir, outDir, data); err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("reading outDir: %v", err)
	}

	htmlFiles := map[string]string{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".html") {
			raw, err := os.ReadFile(filepath.Join(outDir, e.Name()))
			if err != nil {
				t.Fatalf("reading %s: %v", e.Name(), err)
			}
			htmlFiles[e.Name()] = string(raw)
		}
	}
	if len(htmlFiles) == 0 {
		t.Fatal("no .html files generated")
	}

	attrRe := regexp.MustCompile(`\b(href|src|srcset)="([^"]*)"`)
	imgRe := regexp.MustCompile(`<img\b[^>]*>`)
	altRe := regexp.MustCompile(`\balt="([^"]*)"`)
	idRe := regexp.MustCompile(`\bid="([^"]+)"`)

	idsByFile := make(map[string]map[string]bool, len(htmlFiles))
	for name, body := range htmlFiles {
		ids := map[string]bool{}
		for _, m := range idRe.FindAllStringSubmatch(body, -1) {
			ids[m[1]] = true
		}
		idsByFile[name] = ids
	}

	checkRef := func(sourceFile, attr, ref string) {
		if ref == "" {
			t.Errorf("%s: empty %s attribute", sourceFile, attr)
			return
		}
		lower := strings.ToLower(ref)
		if strings.HasPrefix(lower, "http://") ||
			strings.HasPrefix(lower, "https://") ||
			strings.HasPrefix(lower, "mailto:") ||
			strings.HasPrefix(lower, "tel:") ||
			strings.HasPrefix(ref, "#") {
			return
		}
		if strings.HasPrefix(ref, "/") {
			t.Errorf("%s: %s=%q must not start with '/'", sourceFile, attr, ref)
			return
		}
		if strings.Contains(ref, "..") {
			t.Errorf("%s: %s=%q must not contain '..'", sourceFile, attr, ref)
			return
		}

		pathPart, frag, hasFrag := strings.Cut(ref, "#")
		targetPath := filepath.Join(outDir, filepath.FromSlash(pathPart))
		info, err := os.Stat(targetPath)
		if err != nil || info.IsDir() {
			t.Errorf("%s: %s=%q points to missing file %q", sourceFile, attr, ref, pathPart)
			return
		}
		if hasFrag && frag != "" {
			targetIDs, ok := idsByFile[filepath.ToSlash(pathPart)]
			if !ok {
				t.Errorf("%s: %s=%q references fragment #%s on non-HTML target", sourceFile, attr, ref, frag)
				return
			}
			if !targetIDs[frag] {
				t.Errorf("%s: %s=%q references missing id %q in %s", sourceFile, attr, ref, frag, pathPart)
			}
		}
	}

	for name, body := range htmlFiles {
		for _, m := range attrRe.FindAllStringSubmatch(body, -1) {
			attr, val := m[1], m[2]
			if attr == "srcset" {
				for _, cand := range strings.Split(val, ",") {
					fields := strings.Fields(strings.TrimSpace(cand))
					if len(fields) > 0 {
						checkRef(name, attr, fields[0])
					}
				}
				continue
			}
			checkRef(name, attr, val)
		}

		for _, imgTag := range imgRe.FindAllString(body, -1) {
			altMatch := altRe.FindStringSubmatch(imgTag)
			if altMatch == nil || strings.TrimSpace(altMatch[1]) == "" {
				t.Errorf("%s: <img> missing non-empty alt: %s", name, imgTag)
			}
		}
	}
}
