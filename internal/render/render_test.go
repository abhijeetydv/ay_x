package render_test

import (
	"encoding/xml"
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
	const postsMarker = "ZZMARK_RENDER_TEST_POSTS_HEADING"
	const contactMarker = "ZZMARK_RENDER_TEST_CONTACT_HEADING"
	data.Hero.Greeting = marker
	data.AboutPage.Paragraphs = append(data.AboutPage.Paragraphs, aboutMarker)
	data.Projects.Page.Heading = projectsMarker
	data.Posts.Page.Heading = postsMarker
	data.Contact.Page.Heading = contactMarker

	outDir := t.TempDir()
	templateDir := filepath.Join("..", "..", "templates")
	rootDir := filepath.Join("..", "..")
	if err := render.BuildSite(templateDir, rootDir, outDir, data); err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	for _, name := range []string{"index.html", "about.html", "projects.html", "posts.html", "contact.html", "styles.css", "script.js"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("expected output file %s: %v", name, err)
		}
	}

	expectedMarkers := map[string]string{
		"index.html":    marker,
		"about.html":    aboutMarker,
		"projects.html": projectsMarker,
		"posts.html":    postsMarker,
		"contact.html":  contactMarker,
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

func TestMetadataAndSitemap(t *testing.T) {
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

	titleRe := regexp.MustCompile(`<title>[^<]+</title>`)
	descRe := regexp.MustCompile(`<meta name="description" content="([^"]+)">`)
	canonRe := regexp.MustCompile(`<link rel="canonical" href="([^"]+)">`)
	ogTitleRe := regexp.MustCompile(`<meta property="og:title" content="([^"]+)">`)
	ogDescRe := regexp.MustCompile(`<meta property="og:description" content="([^"]+)">`)
	ogURLRe := regexp.MustCompile(`<meta property="og:url" content="([^"]+)">`)
	ogImageRe := regexp.MustCompile(`<meta property="og:image" content="([^"]+)">`)
	twCardRe := regexp.MustCompile(`<meta name="twitter:card" content="summary_large_image">`)
	faviconRe := regexp.MustCompile(`<link rel="icon"[^>]*href="([^"]+)"`)

	for _, page := range render.DefaultPages() {
		raw, err := os.ReadFile(filepath.Join(outDir, page.Output))
		if err != nil {
			t.Fatalf("reading %s: %v", page.Output, err)
		}
		html := string(raw)

		if matches := titleRe.FindAllString(html, -1); len(matches) != 1 {
			t.Errorf("%s: expected 1 <title>, got %d", page.Output, len(matches))
		}
		if matches := descRe.FindAllStringSubmatch(html, -1); len(matches) != 1 || strings.TrimSpace(matches[0][1]) == "" {
			t.Errorf("%s: expected 1 non-empty meta description, got %d", page.Output, len(matches))
		}
		canonMatches := canonRe.FindAllStringSubmatch(html, -1)
		if len(canonMatches) != 1 {
			t.Errorf("%s: expected 1 canonical link, got %d", page.Output, len(canonMatches))
		} else if !strings.HasPrefix(canonMatches[0][1], data.Site.URL) {
			t.Errorf("%s: canonical %q does not start with %q", page.Output, canonMatches[0][1], data.Site.URL)
		}

		if !ogTitleRe.MatchString(html) {
			t.Errorf("%s: missing og:title", page.Output)
		}
		if !ogDescRe.MatchString(html) {
			t.Errorf("%s: missing og:description", page.Output)
		}
		if !ogURLRe.MatchString(html) {
			t.Errorf("%s: missing og:url", page.Output)
		}
		if m := ogImageRe.FindStringSubmatch(html); m == nil || !strings.HasPrefix(m[1], data.Site.URL) {
			t.Errorf("%s: missing or relative og:image (must start with %q)", page.Output, data.Site.URL)
		}
		if !twCardRe.MatchString(html) {
			t.Errorf("%s: missing twitter:card", page.Output)
		}
		favMatch := faviconRe.FindStringSubmatch(html)
		if favMatch == nil {
			t.Errorf("%s: missing favicon link", page.Output)
		} else if _, err := os.Stat(filepath.Join(outDir, filepath.FromSlash(favMatch[1]))); err != nil {
			t.Errorf("%s: favicon file %q missing in output: %v", page.Output, favMatch[1], err)
		}
	}

	sitemapRaw, err := os.ReadFile(filepath.Join(outDir, "sitemap.xml"))
	if err != nil {
		t.Fatalf("reading sitemap.xml: %v", err)
	}
	var parsed struct {
		URLs []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(sitemapRaw, &parsed); err != nil {
		t.Fatalf("parsing sitemap.xml: %v", err)
	}
	gotLocs := map[string]bool{}
	for _, u := range parsed.URLs {
		gotLocs[u.Loc] = true
	}
	for _, page := range render.DefaultPages() {
		want := data.Site.URL + page.CanonicalPath
		if !gotLocs[want] {
			t.Errorf("sitemap.xml missing %q", want)
		}
	}

	robotsRaw, err := os.ReadFile(filepath.Join(outDir, "robots.txt"))
	if err != nil {
		t.Fatalf("reading robots.txt: %v", err)
	}
	wantSitemapLine := "Sitemap: " + data.Site.URL + "sitemap.xml"
	if !strings.Contains(string(robotsRaw), wantSitemapLine) {
		t.Errorf("robots.txt missing %q", wantSitemapLine)
	}
}

func TestHeaderLogoPaths(t *testing.T) {
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

	logoRe := regexp.MustCompile(`<a href="([^"]+)" class="nav-logo">([^<]+)</a>`)
	for _, page := range render.DefaultPages() {
		raw, err := os.ReadFile(filepath.Join(outDir, page.Output))
		if err != nil {
			t.Fatalf("reading %s: %v", page.Output, err)
		}
		m := logoRe.FindStringSubmatch(string(raw))
		if m == nil {
			t.Fatalf("%s: nav-logo link not found", page.Output)
		}
		if m[1] != data.Site.Nav.LogoHref {
			t.Errorf("%s: logo href = %q, want %q", page.Output, m[1], data.Site.Nav.LogoHref)
		}
		wantLogo := data.Site.Nav.Logo + data.Site.Nav.PagePaths[page.Key]
		if m[2] != wantLogo {
			t.Errorf("%s: logo text = %q, want %q", page.Output, m[2], wantLogo)
		}
	}
}

func TestHomeMoreAboutLink(t *testing.T) {
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

	raw, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("reading index.html: %v", err)
	}
	re := regexp.MustCompile(`<a href="about\.html" class="social-link more-link">\s*` + regexp.QuoteMeta(data.Hero.MoreLink.Label) + `\s*</a>`)
	if !re.Match(raw) {
		t.Errorf("index.html missing about.html link with label %q", data.Hero.MoreLink.Label)
	}
}

func TestFooterTimerAndLocationClock(t *testing.T) {
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

	timerRe := regexp.MustCompile(`<span id="footer-time" role="timer" aria-label="` + regexp.QuoteMeta(data.Footer.TimerLabel) + `">00:00</span>`)
	for _, page := range render.DefaultPages() {
		raw, err := os.ReadFile(filepath.Join(outDir, page.Output))
		if err != nil {
			t.Fatalf("reading %s: %v", page.Output, err)
		}
		html := string(raw)
		if !timerRe.MatchString(html) {
			t.Errorf("%s: missing footer timer element with label %q and 00:00 placeholder", page.Output, data.Footer.TimerLabel)
		}
		for _, cls := range []string{`class="sticky-nav"`, `class="scroll-fade-wrapper"`, `class="container main-content"`, `class="footer-panel"`} {
			if !strings.Contains(html, cls) {
				t.Errorf("%s: missing shared sticky-footer structure %s", page.Output, cls)
			}
		}
	}

	homeRaw, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("reading index.html: %v", err)
	}
	if !strings.Contains(string(homeRaw), `<span class="location-time" id="current-time"></span>`) {
		t.Errorf("index.html: missing location clock #current-time")
	}
}

func TestHomeConnectAndPostsLinks(t *testing.T) {
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

	raw, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("reading index.html: %v", err)
	}

	connectRe := regexp.MustCompile(`<a href="contact\.html" class="connect-btn">\s*` + regexp.QuoteMeta(data.About.Connect.CTALabel) + `\s*</a>`)
	if !connectRe.Match(raw) {
		t.Errorf("index.html missing connect-btn link to contact.html with label %q", data.About.Connect.CTALabel)
	}
	postsRe := regexp.MustCompile(`<a href="posts\.html" class="posts-external">\s*` + regexp.QuoteMeta(data.Posts.HomeTile.AllPostsLabel) + `\s*</a>`)
	if !postsRe.Match(raw) {
		t.Errorf("index.html missing posts-external link to posts.html with label %q", data.Posts.HomeTile.AllPostsLabel)
	}
}

func TestPostsSortingAndEmptyState(t *testing.T) {
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

	postsRaw, err := os.ReadFile(filepath.Join(outDir, "posts.html"))
	if err != nil {
		t.Fatalf("reading posts.html: %v", err)
	}
	postsHTML := string(postsRaw)
	newestIdx := strings.Index(postsHTML, "Newest Fixture Post")
	olderIdx := strings.Index(postsHTML, "Older Fixture Post")
	if newestIdx == -1 || olderIdx == -1 || newestIdx > olderIdx {
		t.Fatalf("expected Newest Fixture Post (%d) before Older Fixture Post (%d)", newestIdx, olderIdx)
	}

	// Empty items list must render empty_message on posts.html and index.html.
	data.Posts.Items = nil
	emptyOutDir := t.TempDir()
	if err := render.BuildSite(templateDir, rootDir, emptyOutDir, data); err != nil {
		t.Fatalf("BuildSite empty posts failed: %v", err)
	}
	for _, fn := range []string{"posts.html", "index.html"} {
		raw, err := os.ReadFile(filepath.Join(emptyOutDir, fn))
		if err != nil {
			t.Fatalf("reading %s: %v", fn, err)
		}
		if !strings.Contains(string(raw), data.Posts.Page.EmptyMessage) {
			t.Errorf("%s: missing empty_message %q", fn, data.Posts.Page.EmptyMessage)
		}
	}
}

func TestContactMailtoHref(t *testing.T) {
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

	raw, err := os.ReadFile(filepath.Join(outDir, "contact.html"))
	if err != nil {
		t.Fatalf("reading contact.html: %v", err)
	}
	wantHref := `href="mailto:` + data.Contact.Email.Address + `"`
	if !strings.Contains(string(raw), wantHref) {
		t.Errorf("contact.html missing %s", wantHref)
	}
}
