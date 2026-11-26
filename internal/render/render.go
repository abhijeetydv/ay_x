package render

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"portfolio/internal/model"
)

// PageSpec describes a single generated HTML page.
type PageSpec struct {
	Key           string
	Template      string
	Output        string
	CanonicalPath string
	PathSegments  func(*model.HomepageData) []string
	Title         func(*model.HomepageData) string
	Description   func(*model.HomepageData) string
}

// PageContext wraps HomepageData with the active page metadata.
type PageContext struct {
	*model.HomepageData
	Page            string
	PageTitle       string
	PageDescription string
	CanonicalURL    string
	OGImageURL      string
	FaviconType     string
	PathSegments    []string
	LogoText        string
}

// DefaultPages returns the configured pages to generate for the site.
func DefaultPages() []PageSpec {
	return []PageSpec{
		{
			Key:           "home",
			Template:      "home.html",
			Output:        "index.html",
			CanonicalPath: "",
			Title:         func(d *model.HomepageData) string { return d.Site.Title },
			Description:   func(d *model.HomepageData) string { return d.Site.Description },
		},
		{
			Key:           "about",
			Template:      "about.html",
			Output:        "about.html",
			CanonicalPath: "about.html",
			Title:         func(d *model.HomepageData) string { return d.AboutPage.Title },
			Description:   func(d *model.HomepageData) string { return d.AboutPage.Description },
		},
		{
			Key:           "projects",
			Template:      "projects.html",
			Output:        "projects.html",
			CanonicalPath: "projects.html",
			Title:         func(d *model.HomepageData) string { return d.Projects.Page.Title },
			Description:   func(d *model.HomepageData) string { return d.Projects.Page.Description },
		},
		{
			Key:           "posts",
			Template:      "posts.html",
			Output:        "posts.html",
			CanonicalPath: "posts.html",
			Title:         func(d *model.HomepageData) string { return d.Posts.Page.Title },
			Description:   func(d *model.HomepageData) string { return d.Posts.Page.Description },
		},
		{
			Key:           "contact",
			Template:      "contact.html",
			Output:        "contact.html",
			CanonicalPath: "contact.html",
			Title:         func(d *model.HomepageData) string { return d.Contact.Page.Title },
			Description:   func(d *model.HomepageData) string { return d.Contact.Page.Description },
		},
	}
}

// BuildSite renders all configured pages from a single loaded HomepageData
// and copies static assets (styles.css, script.js, and optional assets/) once.
func BuildSite(templateDir, rootDir, outDir string, data *model.HomepageData) error {
	return BuildPages(templateDir, rootDir, outDir, data, DefaultPages())
}

// BuildPages renders the given pages from a single loaded HomepageData
// and copies static assets once into outDir.
func BuildPages(templateDir, rootDir, outDir string, data *model.HomepageData, pages []PageSpec) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", outDir, err)
	}

	preparePosts(data)

	for _, p := range pages {
		title := data.Site.Title
		if p.Title != nil {
			if t := p.Title(data); t != "" {
				title = t
			}
		}
		desc := data.Site.Description
		if p.Description != nil {
			if d := p.Description(data); d != "" {
				desc = d
			}
		}
		tmplPath := filepath.Join(templateDir, p.Template)
		outHTML := filepath.Join(outDir, p.Output)
		segments := resolvePathSegments(data, p)
		if err := renderPage(tmplPath, outHTML, PageContext{
			HomepageData:    data,
			Page:            p.Key,
			PageTitle:       title,
			PageDescription: desc,
			CanonicalURL:    data.Site.URL + p.CanonicalPath,
			OGImageURL:      data.Site.URL + strings.TrimPrefix(filepath.ToSlash(data.Site.OGImage), "/"),
			FaviconType:     faviconMIMEType(data.Site.Favicon),
			PathSegments:    segments,
			LogoText:        data.Site.Nav.Logo + strings.Join(segments, "/"),
		}); err != nil {
			return err
		}
	}

	if err := writeSitemap(filepath.Join(outDir, "sitemap.xml"), data.Site.URL, pages); err != nil {
		return err
	}
	if err := writeRobots(filepath.Join(outDir, "robots.txt"), data.Site.URL); err != nil {
		return err
	}

	return copyStaticAssets(rootDir, outDir)
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc string `xml:"loc"`
}

func writeSitemap(outPath, baseURL string, pages []PageSpec) error {
	urls := make([]sitemapURL, 0, len(pages))
	for _, p := range pages {
		urls = append(urls, sitemapURL{Loc: baseURL + p.CanonicalPath})
	}
	payload := sitemapURLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}
	body, err := xml.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling sitemap.xml: %w", err)
	}
	content := append([]byte(xml.Header), body...)
	content = append(content, '\n')
	if err := os.WriteFile(outPath, content, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}
	return nil
}

func writeRobots(outPath, baseURL string) error {
	content := fmt.Sprintf("User-agent: *\nAllow: /\n\nSitemap: %ssitemap.xml\n", baseURL)
	if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}
	return nil
}

func preparePosts(data *model.HomepageData) {
	items := make([]model.PostEntry, len(data.Posts.Items))
	copy(items, data.Posts.Items)

	dateFormat := data.Posts.Page.DateFormat
	if dateFormat == "" {
		dateFormat = "Jan 2, 2006"
	}

	parsedDates := make(map[string]time.Time, len(items))
	for i := range items {
		t, err := time.Parse("2006-01-02", items[i].Date)
		if err == nil {
			parsedDates[items[i].Date] = t
			items[i].FormattedDate = t.Format(dateFormat)
		} else {
			items[i].FormattedDate = items[i].Date
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		ti, oki := parsedDates[items[i].Date]
		tj, okj := parsedDates[items[j].Date]
		if oki && okj && !ti.Equal(tj) {
			return ti.After(tj)
		}
		if items[i].Date != items[j].Date {
			return items[i].Date > items[j].Date
		}
		return items[i].ID < items[j].ID
	})

	data.SortedPosts = items
	if len(items) > 0 {
		latest := items[0]
		data.LatestPost = &latest
	} else {
		data.LatestPost = nil
	}
}

func resolvePathSegments(data *model.HomepageData, p PageSpec) []string {
	if p.PathSegments != nil {
		return p.PathSegments(data)
	}
	if seg := data.Site.Nav.PagePaths[p.Key]; seg != "" {
		return []string{seg}
	}
	return nil
}

func faviconMIMEType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		return "image/png"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
}

// BuildHomepage renders templates/home.html using data into outDir/index.html
// and copies static assets (styles.css, script.js, and optional assets/) into outDir.
func BuildHomepage(templatePath, rootDir, outDir string, data *model.HomepageData) error {
	return BuildPages(filepath.Dir(templatePath), rootDir, outDir, data, []PageSpec{
		{
			Key:      "home",
			Template: filepath.Base(templatePath),
			Output:   "index.html",
			Title:    func(d *model.HomepageData) string { return d.Site.Title },
		},
	})
}

func renderPage(templatePath, outHTML string, ctx PageContext) error {
	funcs := template.FuncMap{
		"contribStyle": func(width, color string) template.CSS {
			return template.CSS(fmt.Sprintf("width: %s; background: %s;", width, color))
		},
		"lower": strings.ToLower,
		"isExternal": func(href string) bool {
			l := strings.ToLower(href)
			return strings.HasPrefix(l, "http://") || strings.HasPrefix(l, "https://")
		},
		"navHref": func(page, href string) string {
			if page != "home" && strings.HasPrefix(href, "#") {
				return "index.html" + href
			}
			return href
		},
	}

	files := []string{templatePath}
	partialsGlob := filepath.Join(filepath.Dir(templatePath), "partials", "*.html")
	if partials, err := filepath.Glob(partialsGlob); err == nil {
		files = append(files, partials...)
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(funcs).ParseFiles(files...)
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", templatePath, err)
	}

	f, err := os.Create(outHTML)
	if err != nil {
		return fmt.Errorf("creating %s: %w", outHTML, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, ctx); err != nil {
		return fmt.Errorf("executing template %s: %w", templatePath, err)
	}
	return f.Close()
}

func copyStaticAssets(rootDir, outDir string) error {
	// Copy core static assets.
	for _, asset := range []string{"styles.css", "script.js"} {
		src := filepath.Join(rootDir, asset)
		dst := filepath.Join(outDir, asset)
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("copying %s: %w", asset, err)
		}
	}

	// Copy optional assets/ directory if present.
	assetsDir := filepath.Join(rootDir, "assets")
	if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
		if err := copyDir(assetsDir, filepath.Join(outDir, "assets")); err != nil {
			return fmt.Errorf("copying assets directory: %w", err)
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func copyDir(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dstDir, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}
