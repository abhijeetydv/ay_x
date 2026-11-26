package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"portfolio/internal/model"
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
		{"footer.yaml", &data.Footer},
	}

	for _, f := range files {
		path := filepath.Join(contentDir, f.name)
		if err := decodeYAMLFile(path, f.dest); err != nil {
			return nil, fmt.Errorf("loading %s: %w", f.name, err)
		}
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
