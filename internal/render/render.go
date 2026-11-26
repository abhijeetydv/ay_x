package render

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"portfolio/internal/model"
)

// PageSpec describes a single generated HTML page.
type PageSpec struct {
	Key      string
	Template string
	Output   string
	Title    func(*model.HomepageData) string
}

// PageContext wraps HomepageData with the active page key and title.
type PageContext struct {
	*model.HomepageData
	Page      string
	PageTitle string
}

// DefaultPages returns the configured pages to generate for the site.
func DefaultPages() []PageSpec {
	return []PageSpec{
		{
			Key:      "home",
			Template: "home.html",
			Output:   "index.html",
			Title:    func(d *model.HomepageData) string { return d.Site.Title },
		},
		{
			Key:      "about",
			Template: "about.html",
			Output:   "about.html",
			Title:    func(d *model.HomepageData) string { return d.AboutPage.Title },
		},
		{
			Key:      "projects",
			Template: "projects.html",
			Output:   "projects.html",
			Title:    func(d *model.HomepageData) string { return d.Projects.Page.Title },
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

	for _, p := range pages {
		title := data.Site.Title
		if p.Title != nil {
			if t := p.Title(data); t != "" {
				title = t
			}
		}
		tmplPath := filepath.Join(templateDir, p.Template)
		outHTML := filepath.Join(outDir, p.Output)
		if err := renderPage(tmplPath, outHTML, PageContext{
			HomepageData: data,
			Page:         p.Key,
			PageTitle:    title,
		}); err != nil {
			return err
		}
	}

	return copyStaticAssets(rootDir, outDir)
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
