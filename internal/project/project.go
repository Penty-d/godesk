package project

import (
	"os"
	"path/filepath"
	"strings"
)

type Project struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	EnvFile     string   `yaml:"env_file,omitempty"`
	ComposeFile string   `yaml:"compose_file,omitempty"`
	TestCmd     string   `yaml:"test_cmd,omitempty"`
	LintCmd     string   `yaml:"lint_cmd,omitempty"`
	UpCmd       string   `yaml:"up_cmd,omitempty"`
	HealthURLs  []string `yaml:"health_urls,omitempty"`
}

func Discover(root string) (Project, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Project{}, err
	}
	return discover(abs, abs)
}

func discover(root string, searchRoot string) (Project, error) {
	envFile, err := firstExistingUpward(root, searchRoot, ".env")
	if err != nil {
		return Project{}, err
	}
	composeFile, err := firstExistingUpward(root, searchRoot, "docker-compose.yml", "docker-compose.yaml", "compose.yaml")
	if err != nil {
		return Project{}, err
	}
	return Project{
		Name:        filepath.Base(root),
		Path:        root,
		EnvFile:     envFile,
		ComposeFile: composeFile,
		TestCmd:     "go test ./...",
	}, nil
}

func ScanRoots(roots []string) ([]Project, error) {
	seen := map[string]Project{}
	for _, root := range roots {
		if root == "" {
			continue
		}
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return nil, err
		}
		err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() && shouldSkipDir(d.Name(), path != absRoot) {
				return filepath.SkipDir
			}
			if d.IsDir() {
				goMod := filepath.Join(path, "go.mod")
				if _, err := os.Stat(goMod); err == nil {
					p, err := discover(path, absRoot)
					if err != nil {
						return err
					}
					seen[p.Path] = p
					return filepath.SkipDir
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	projects := make([]Project, 0, len(seen))
	for _, p := range seen {
		projects = append(projects, p)
	}
	return projects, nil
}

func firstExisting(root string, names ...string) string {
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			return name
		}
	}
	return ""
}

func firstExistingUpward(root string, searchRoot string, names ...string) (string, error) {
	current := root
	for {
		for _, name := range names {
			candidate := filepath.Join(current, name)
			if _, err := os.Stat(candidate); err == nil {
				return filepath.Rel(root, candidate)
			}
		}
		if current == searchRoot {
			return "", nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", nil
		}
		rel, err := filepath.Rel(searchRoot, parent)
		if err != nil {
			return "", err
		}
		if strings.HasPrefix(rel, "..") {
			return "", nil
		}
		current = parent
	}
}

func shouldSkipDir(name string, nested bool) bool {
	if !nested {
		return false
	}
	if name == ".git" || name == "node_modules" || name == "vendor" {
		return true
	}
	return strings.HasPrefix(name, ".")
}
