package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"portfolio/internal/markdown"
	"portfolio/internal/model"
	"portfolio/internal/posts"
)

type projectFrontMatter struct {
	Title        string                `yaml:"title"`
	TitleColors  []markdown.TitleColor `yaml:"title_colors,omitempty"`
	Summary      string                `yaml:"summary"`
	TerminalDesc string                `yaml:"terminal_desc,omitempty"`
	Owner        string                `yaml:"owner"`
	Repo         string                `yaml:"repo"`
	Stars        int                   `yaml:"stars"`
	Tags         []string              `yaml:"tags,omitempty"`
	Featured     bool                  `yaml:"featured"`
	Order        *int                  `yaml:"order"`
	GitHubURL    string                `yaml:"github_url"`
	LiveURL      string                `yaml:"live_url,omitempty"`
	Updated      string                `yaml:"updated,omitempty"`
}

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

	projFiles, err := markdown.ListDir(filepath.Join(contentDir, "projects"))
	if err != nil {
		return nil, err
	}
	data.AllProjects = make([]model.ProjectItem, 0, len(projFiles))
	for _, pf := range projFiles {
		var fm projectFrontMatter
		body, err := markdown.SplitFrontMatterInto(pf.Raw, &fm)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", pf.SourcePath, err)
		}
		colors := make([]model.TitleColor, 0, len(fm.TitleColors))
		for _, tc := range fm.TitleColors {
			colors = append(colors, model.TitleColor{
				Text:  tc.Text,
				Color: tc.Color,
			})
		}
		orderVal := 0
		hasOrder := fm.Order != nil
		if hasOrder {
			orderVal = *fm.Order
		}
		data.AllProjects = append(data.AllProjects, model.ProjectItem{
			Slug:         pf.Slug,
			SourcePath:   pf.SourcePath,
			Title:        fm.Title,
			TitleColors:  colors,
			Summary:      fm.Summary,
			TerminalDesc: fm.TerminalDesc,
			Owner:        fm.Owner,
			Repo:         fm.Repo,
			Stars:        fm.Stars,
			Tags:         fm.Tags,
			Featured:     fm.Featured,
			Order:        orderVal,
			HasOrder:     hasOrder,
			GitHubURL:    fm.GitHubURL,
			LiveURL:      fm.LiveURL,
			Updated:      fm.Updated,
			RawBody:      body,
			URL:          "projects/" + pf.Slug + "/",
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
