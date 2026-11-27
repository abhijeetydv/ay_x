package render_test

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"portfolio/internal/loader"
	"portfolio/internal/model"
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
	data.Hero.Greeting = marker
	data.AboutPage.Paragraphs = append(data.AboutPage.Paragraphs, aboutMarker)
	data.Projects.Page.Heading = projectsMarker
	data.Posts.Page.Heading = postsMarker

	outDir := t.TempDir()
	templateDir := filepath.Join("..", "..", "templates")
	rootDir := filepath.Join("..", "..")
	if err := render.BuildSite(templateDir, rootDir, outDir, data); err != nil {
		t.Fatalf("BuildSite failed: %v", err)
	}

	for _, name := range []string{"index.html", "about.html", "projects.html", "posts.html", "styles.css", "script.js"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("expected output file %s: %v", name, err)
		}
	}

	expectedMarkers := map[string]string{
		"index.html":    marker,
		"about.html":    aboutMarker,
		"projects.html": projectsMarker,
		"posts.html":    postsMarker,
	}
	for _, page := range render.SitePages(data) {
		raw, err := os.ReadFile(filepath.Join(outDir, filepath.FromSlash(page.Output)))
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

	htmlFiles := map[string]string{}
	if err := filepath.WalkDir(outDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".html") {
			return nil
		}
		relPath, err := filepath.Rel(outDir, path)
		if err != nil {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		htmlFiles[filepath.ToSlash(relPath)] = string(raw)
		return nil
	}); err != nil {
		t.Fatalf("walking outDir: %v", err)
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

		pathPart, frag, hasFrag := strings.Cut(ref, "#")
		sourceDir := filepath.Dir(filepath.FromSlash(sourceFile))
		targetPath := filepath.Clean(filepath.Join(outDir, sourceDir, filepath.FromSlash(pathPart)))
		relToOut, err := filepath.Rel(outDir, targetPath)
		if err != nil || strings.HasPrefix(relToOut, "..") {
			t.Errorf("%s: %s=%q escapes output directory", sourceFile, attr, ref)
			return
		}
		info, err := os.Stat(targetPath)
		if err != nil {
			t.Errorf("%s: %s=%q points to missing file %q", sourceFile, attr, ref, relToOut)
			return
		}
		if info.IsDir() {
			indexPath := filepath.Join(targetPath, "index.html")
			if _, err := os.Stat(indexPath); err != nil {
				t.Errorf("%s: %s=%q directory missing index.html", sourceFile, attr, ref)
				return
			}
			relToOut = filepath.Join(relToOut, "index.html")
		}
		if hasFrag && frag != "" {
			targetIDs, ok := idsByFile[filepath.ToSlash(relToOut)]
			if !ok {
				t.Errorf("%s: %s=%q references fragment #%s on non-HTML target", sourceFile, attr, ref, frag)
				return
			}
			if !targetIDs[frag] {
				t.Errorf("%s: %s=%q references missing id %q in %s", sourceFile, attr, ref, frag, relToOut)
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

	for _, page := range render.SitePages(data) {
		raw, err := os.ReadFile(filepath.Join(outDir, filepath.FromSlash(page.Output)))
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
		} else if wantCanon := data.Site.URL + page.CanonicalPath; canonMatches[0][1] != wantCanon {
			t.Errorf("%s: canonical = %q, want %q", page.Output, canonMatches[0][1], wantCanon)
		}

		wantOGType := "website"
		if page.OGType != "" {
			wantOGType = page.OGType
		}
		if !strings.Contains(html, `<meta property="og:type" content="`+wantOGType+`">`) {
			t.Errorf("%s: missing og:type=%q", page.Output, wantOGType)
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
		} else {
			favResolved := filepath.Join(outDir, filepath.Dir(filepath.FromSlash(page.Output)), filepath.FromSlash(favMatch[1]))
			if _, err := os.Stat(favResolved); err != nil {
				t.Errorf("%s: favicon file %q missing at %s: %v", page.Output, favMatch[1], favResolved, err)
			}
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
	for _, page := range render.SitePages(data) {
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
	for _, page := range render.SitePages(data) {
		raw, err := os.ReadFile(filepath.Join(outDir, filepath.FromSlash(page.Output)))
		if err != nil {
			t.Fatalf("reading %s: %v", page.Output, err)
		}
		m := logoRe.FindStringSubmatch(string(raw))
		if m == nil {
			t.Fatalf("%s: nav-logo link not found", page.Output)
		}
		wantHref := data.Site.Nav.LogoHref
		wantLogo := data.Site.Nav.Logo + data.Site.Nav.PagePaths[page.Key]
		if page.Post != nil {
			wantHref = "../../" + wantHref
			wantLogo = data.Site.Nav.Logo + "posts/" + page.Post.Slug
		}
		if m[1] != wantHref {
			t.Errorf("%s: logo href = %q, want %q", page.Output, m[1], wantHref)
		}
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

	wantConnectHref := "mailto:" + data.Social.Dock.Email.Address
	connectRe := regexp.MustCompile(`<a href="` + regexp.QuoteMeta(wantConnectHref) + `" class="connect-btn">\s*` + regexp.QuoteMeta(data.About.Connect.CTALabel) + `\s*</a>`)
	if !connectRe.Match(raw) {
		t.Errorf("index.html missing connect-btn link to %s with label %q", wantConnectHref, data.About.Connect.CTALabel)
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
	data.SortedPosts = nil
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

func TestPostPagesEndToEnd(t *testing.T) {
	data, err := loader.LoadAll(filepath.Join("..", "validate", "testdata", "valid"))
	if err != nil {
		t.Fatalf("LoadAll fixture failed: %v", err)
	}
	// Inject a raw <script> tag into one fixture post body to verify end-to-end neutralisation.
	for i := range data.SortedPosts {
		if data.SortedPosts[i].Slug == "older-fixture-post" {
			data.SortedPosts[i].RawBody += "\n\n<script>alert(1)</script>\n"
		}
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

	newestRaw, err := os.ReadFile(filepath.Join(outDir, "posts", "newest-fixture-post", "index.html"))
	if err != nil {
		t.Fatalf("reading newest-fixture-post/index.html: %v", err)
	}
	newestHTML := string(newestRaw)

	// Coloured title spans in order
	wantH1 := `<h1 class="post-title-display"><span class="title-color title-cyan">Newest</span> Fixture <span class="title-color title-blue">Post</span></h1>`
	if !strings.Contains(newestHTML, wantH1) {
		t.Errorf("newest-fixture-post missing expected coloured h1:\nwant: %s", wantH1)
	}

	// Special characters in title are HTML-escaped
	data.SortedPosts[0].Title = `Go & "Quotes" <Safe>`
	data.SortedPosts[0].TitleColors = []model.TitleColor{{Text: "Go", Color: "blue"}}
	escOutDir := t.TempDir()
	if err := render.BuildSite(templateDir, rootDir, escOutDir, data); err != nil {
		t.Fatalf("BuildSite escaped title failed: %v", err)
	}
	escRaw, err := os.ReadFile(filepath.Join(escOutDir, "posts", data.SortedPosts[0].Slug, "index.html"))
	if err != nil {
		t.Fatalf("reading escaped post: %v", err)
	}
	wantEscH1 := `<span class="title-color title-blue">Go</span> &amp; &#34;Quotes&#34; &lt;Safe&gt;`
	if !strings.Contains(string(escRaw), wantEscH1) {
		t.Errorf("escaped title h1 mismatch, want %q", wantEscH1)
	}

	// Code-block wrapper and YAML copy labels
	if !strings.Contains(newestHTML, `<div class="code-block">`) ||
		!strings.Contains(newestHTML, `data-copy-label="`+data.Posts.PostPage.CopyLabel+`"`) ||
		!strings.Contains(newestHTML, `data-copied-label="`+data.Posts.PostPage.CopiedLabel+`"`) {
		t.Errorf("newest-fixture-post missing code-block wrapper or copy button labels")
	}

	// Depth-aware ../../ asset prefixes
	for _, wantRef := range []string{`href="../../styles.css"`, `src="../../script.js"`} {
		if !strings.Contains(newestHTML, wantRef) {
			t.Errorf("newest-fixture-post missing %s", wantRef)
		}
	}

	// Tag palette consistency between posts.html and post page
	postsListRaw, err := os.ReadFile(filepath.Join(outDir, "posts.html"))
	if err != nil {
		t.Fatalf("reading posts.html: %v", err)
	}
	postsListHTML := string(postsListRaw)
	tagRe := regexp.MustCompile(`<span class="tag (tag-c[0-4])">([^<]+)</span>`)
	listTags := map[string]string{}
	for _, m := range tagRe.FindAllStringSubmatch(postsListHTML, -1) {
		listTags[m[2]] = m[1]
	}
	for _, m := range tagRe.FindAllStringSubmatch(newestHTML, -1) {
		if wantCls, ok := listTags[m[2]]; ok && m[1] != wantCls {
			t.Errorf("tag %q color mismatch: list=%s post=%s", m[2], wantCls, m[1])
		}
	}

	// Links from posts.html and home tile to post pages
	if !strings.Contains(postsListHTML, `href="posts/newest-fixture-post/"`) {
		t.Errorf("posts.html missing link to posts/newest-fixture-post/")
	}
	homeRaw, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("reading index.html: %v", err)
	}
	if !strings.Contains(string(homeRaw), `<a href="posts/newest-fixture-post/" class="post-item">`) {
		t.Errorf("index.html latest-post tile does not link to posts/newest-fixture-post/")
	}

	// Raw <script> neutralised end-to-end
	olderRaw, err := os.ReadFile(filepath.Join(outDir, "posts", "older-fixture-post", "index.html"))
	if err != nil {
		t.Fatalf("reading older-fixture-post/index.html: %v", err)
	}
	if strings.Contains(string(olderRaw), "<script>alert(1)</script>") {
		t.Fatalf("older-fixture-post/index.html leaked raw <script> tag")
	}
}

func TestPostTOCAndDividerFreeLayout(t *testing.T) {
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

	// Post with >= toc_min_items headings has nav.toc with nested items in order.
	newestRaw, err := os.ReadFile(filepath.Join(outDir, "posts", "newest-fixture-post", "index.html"))
	if err != nil {
		t.Fatalf("reading newest-fixture-post: %v", err)
	}
	newestHTML := string(newestRaw)
	if count := strings.Count(newestHTML, `<nav class="toc"`); count != 1 {
		t.Fatalf("expected 1 nav.toc on newest-fixture-post, got %d", count)
	}
	if !strings.Contains(newestHTML, `aria-label="`+data.Posts.PostPage.ContentsAriaLabel+`"`) ||
		!strings.Contains(newestHTML, `<p class="toc-label">`+data.Posts.PostPage.ContentsLabel+`</p>`) {
		t.Errorf("newest-fixture-post missing configured TOC aria-label or label")
	}
	if !strings.Contains(newestHTML, `<ol class="toc-sublist">`) {
		t.Errorf("newest-fixture-post missing nested .toc-sublist")
	}
	i1 := strings.Index(newestHTML, `href="#heading-one"`)
	i2 := strings.Index(newestHTML, `href="#sub-heading"`)
	i3 := strings.Index(newestHTML, `href="#heading-two"`)
	if i1 == -1 || i2 == -1 || i3 == -1 || !(i1 < i2 && i2 < i3) {
		t.Errorf("TOC links missing or out of order: %d, %d, %d", i1, i2, i3)
	}
	for _, id := range []string{"heading-one", "sub-heading", "heading-two"} {
		if !strings.Contains(newestHTML, `id="`+id+`"`) {
			t.Errorf("newest-fixture-post missing heading id %q", id)
		}
	}
	if strings.Contains(newestHTML, "<hr") {
		t.Errorf("newest-fixture-post contains unexpected <hr>")
	}

	// Post with fewer than toc_min_items headings has no nav.toc.
	olderRaw, err := os.ReadFile(filepath.Join(outDir, "posts", "older-fixture-post", "index.html"))
	if err != nil {
		t.Fatalf("reading older-fixture-post: %v", err)
	}
	if strings.Contains(string(olderRaw), `class="toc"`) {
		t.Errorf("older-fixture-post should not render nav.toc")
	}
	if strings.Contains(string(olderRaw), "<hr") {
		t.Errorf("older-fixture-post contains unexpected <hr>")
	}

	// Check CSS rules for .post-header, .toc, .post-article, .post-layout contain no border-top/border-bottom.
	cssRaw, err := os.ReadFile(filepath.Join("..", "..", "styles.css"))
	if err != nil {
		t.Fatalf("reading styles.css: %v", err)
	}
	css := string(cssRaw)
	for _, sel := range []string{".post-article", ".post-header", ".post-layout.has-toc", ".toc"} {
		ruleRe := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(sel) + `\s*\{([^}]*)\}`)
		for _, m := range ruleRe.FindAllStringSubmatch(css, -1) {
			if strings.Contains(m[1], "border-top") || strings.Contains(m[1], "border-bottom") {
				t.Errorf("CSS rule %s must not contain border-top or border-bottom", sel)
			}
		}
	}
}

func TestNavPanelDockAndResumeEndToEnd(t *testing.T) {
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

	// contact.html must not exist in output or sitemap.xml
	if _, err := os.Stat(filepath.Join(outDir, "contact.html")); !os.IsNotExist(err) {
		t.Errorf("contact.html must not be generated")
	}
	sitemapRaw, err := os.ReadFile(filepath.Join(outDir, "sitemap.xml"))
	if err != nil {
		t.Fatalf("reading sitemap.xml: %v", err)
	}
	if strings.Contains(string(sitemapRaw), "contact.html") {
		t.Errorf("sitemap.xml must not reference contact.html")
	}

	// Check root index.html and nested post page
	indexRaw, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("reading index.html: %v", err)
	}
	indexHTML := string(indexRaw)

	for _, want := range []string{
		`document.documentElement.classList.add('js')`,
		`href="assets/resume.pdf"`,
		`target="_blank" rel="noopener noreferrer"`,
		`class="nav-toggle"`,
		`id="nav-panel-backdrop"`,
		`id="nav-panel"`,
		`class="footer-socials"`,
		`data-dock-state="docked"`,
		`class="dock-sockets"`,
		`class="dock-socket"`,
		`class="dock-link"`,
		`class="dock-bubble"`,
	} {
		if !strings.Contains(indexHTML, want) {
			t.Errorf("index.html missing %q", want)
		}
	}

	// #nav-panel must be a sibling after </nav>, not nested inside .sticky-nav
	navCloseIdx := strings.Index(indexHTML, "</nav>")
	panelIdx := strings.Index(indexHTML, `id="nav-panel"`)
	if navCloseIdx == -1 || panelIdx == -1 || panelIdx < navCloseIdx {
		t.Errorf("expected #nav-panel after closing </nav> of .sticky-nav")
	}

	postRaw, err := os.ReadFile(filepath.Join(outDir, "posts", data.SortedPosts[0].Slug, "index.html"))
	if err != nil {
		t.Fatalf("reading post page: %v", err)
	}
	if !strings.Contains(string(postRaw), `href="../../assets/resume.pdf"`) {
		t.Errorf("post page missing relative ../../assets/resume.pdf link")
	}

	// Verify CSS tuning variables and view-transition-name rules
	cssRaw, err := os.ReadFile(filepath.Join(outDir, "styles.css"))
	if err != nil {
		t.Fatalf("reading built styles.css: %v", err)
	}
	css := string(cssRaw)
	for _, cssToken := range []string{
		"--dock-size:",
		"--dock-gap:",
		"--dock-edge:",
		"--dock-fly-duration:",
		"--dock-stagger:",
		"--dock-pop-scale:",
		"--dock-ease:",
		"--panel-width:",
		"--panel-duration:",
		"view-transition-name: dock-link-0;",
		"text-transform: uppercase;",
	} {
		if !strings.Contains(css, cssToken) {
			t.Errorf("styles.css missing %q", cssToken)
		}
	}
}
