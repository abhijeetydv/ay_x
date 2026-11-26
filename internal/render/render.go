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

// PageContext wraps HomepageData with the active page key and title.
type PageContext struct {
	*model.HomepageData
	Page      string
	PageTitle string
}

// BuildHomepage renders templates/home.html using data into outDir/index.html
// and copies static assets (styles.css, script.js, and optional assets/) into outDir.
func BuildHomepage(templatePath, rootDir, outDir string, data *model.HomepageData) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", outDir, err)
	}

	if err := renderPage(templatePath, filepath.Join(outDir, "index.html"), PageContext{
		HomepageData: data,
		Page:         "home",
		PageTitle:    data.Site.Title,
	}); err != nil {
		return err
	}

	return copyStaticAssets(rootDir, outDir)
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
