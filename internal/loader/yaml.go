package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"portfolio/internal/model"
	"portfolio/internal/posts"
)

// LoadAll reads all YAML content files from contentDir into a HomepageData struct.
func LoadAll(contentDir string) (*model.HomepageData, error) {
	data := &model.HomepageData{}

	files := []struct {
		name string
		dest any
	}{
		{"site.yaml", &data.Site},
		{"hero.yaml", &data.Hero},
		{"social.yaml", &data.Social},
		{"experience.yaml", &data.Experience},
		{"projects.yaml", &data.Projects},
		{"skills.yaml", &data.Skills},
		{"about.yaml", &data.About},
		{"about_page.yaml", &data.AboutPage},
		{"posts.yaml", &data.Posts},
		{"footer.yaml", &data.Footer},
	}

	for _, f := range files {
		path := filepath.Join(contentDir, f.name)
		if err := decodeYAMLFile(path, f.dest); err != nil {
			return nil, fmt.Errorf("loading %s: %w", f.name, err)
		}
	}

	postFiles, err := posts.LoadDir(filepath.Join(contentDir, "posts"))
	if err != nil {
		return nil, err
	}
	data.SortedPosts = make([]model.Post, 0, len(postFiles))
	for _, pf := range postFiles {
		colors := make([]model.TitleColor, 0, len(pf.FrontMatter.TitleColors))
		for _, tc := range pf.FrontMatter.TitleColors {
			colors = append(colors, model.TitleColor{
				Text:  tc.Text,
				Color: tc.Color,
			})
		}
		data.SortedPosts = append(data.SortedPosts, model.Post{
			Slug:        pf.Slug,
			SourcePath:  pf.SourcePath,
			Title:       pf.FrontMatter.Title,
			TitleColors: colors,
			Date:        pf.FrontMatter.Date,
			Updated:     pf.FrontMatter.Updated,
			Summary:     pf.FrontMatter.Summary,
			Tags:        pf.FrontMatter.Tags,
			RawBody:     pf.Body,
			URL:         "posts/" + pf.Slug + "/",
		})
	}

	return data, nil
}

func decodeYAMLFile(path string, dest any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	dec := yaml.NewDecoder(file)
	dec.KnownFields(true)
	if err := dec.Decode(dest); err != nil {
		return err
	}
	return nil
}
