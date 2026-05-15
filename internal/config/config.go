package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"godesk/internal/project"
)

const (
	appDirName       = "godesk"
	configFileName   = "config.yaml"
	projectsFileName = "projects.yaml"
	projectFileName  = ".godesk.yaml"
)

type GlobalConfig struct {
	Roots []string `yaml:"roots"`
}

type ProjectOverride struct {
	Name        string   `yaml:"name"`
	EnvFile     string   `yaml:"env_file"`
	ComposeFile string   `yaml:"compose_file"`
	LintCmd     string   `yaml:"lint_cmd"`
	UpCmd       string   `yaml:"up_cmd"`
	HealthURLs  []string `yaml:"health_urls"`
}

type ProjectIndex struct {
	Projects []project.Project `yaml:"projects"`
}

type Store struct {
	baseDir string
}

func NewStore() *Store {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr == nil && home != "" {
			dir = filepath.Join(home, ".config")
		}
	}
	return &Store{baseDir: filepath.Join(dir, appDirName)}
}

func (s *Store) ConfigPath() string {
	return filepath.Join(s.baseDir, configFileName)
}

func (s *Store) ProjectsPath() string {
	return filepath.Join(s.baseDir, projectsFileName)
}

func (s *Store) LoadGlobal() (GlobalConfig, error) {
	var cfg GlobalConfig
	err := readYAML(s.ConfigPath(), &cfg)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	for i, root := range cfg.Roots {
		expanded, err := expandPath(root)
		if err != nil {
			return cfg, err
		}
		cfg.Roots[i] = expanded
	}
	return cfg, nil
}

func (s *Store) LoadIndex() (ProjectIndex, error) {
	var idx ProjectIndex
	err := readYAML(s.ProjectsPath(), &idx)
	if errors.Is(err, os.ErrNotExist) {
		return idx, nil
	}
	if err != nil {
		return idx, err
	}
	sortProjects(idx.Projects)
	return idx, nil
}

func (s *Store) SaveIndex(idx ProjectIndex) error {
	sortProjects(idx.Projects)
	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return err
	}
	return writeYAML(s.ProjectsPath(), idx)
}

func (s *Store) FindProject(name string) (project.Project, error) {
	idx, err := s.LoadIndex()
	if err != nil {
		return project.Project{}, err
	}
	for _, p := range idx.Projects {
		if p.Name == name {
			return p, nil
		}
	}
	return project.Project{}, fmt.Errorf("project %q not found; run godesk scan --root <dir>", name)
}

func LoadProjectOverride(root string) (ProjectOverride, bool, error) {
	path := filepath.Join(root, projectFileName)
	var override ProjectOverride
	err := readYAML(path, &override)
	if errors.Is(err, os.ErrNotExist) {
		return override, false, nil
	}
	if err != nil {
		return override, false, err
	}
	return override, true, nil
}

func SaveProjectOverride(root string, override ProjectOverride, overwrite bool) (string, error) {
	path := filepath.Join(root, projectFileName)
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return "", fmt.Errorf("%s already exists; pass --force to overwrite", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	data, err := yaml.Marshal(override)
	if err != nil {
		return "", err
	}
	data = append([]byte("# godesk project config\n"), data...)
	return path, os.WriteFile(path, data, 0o644)
}

func ApplyOverride(p project.Project, override ProjectOverride) project.Project {
	if override.Name != "" {
		p.Name = override.Name
	}
	if override.EnvFile != "" {
		p.EnvFile = override.EnvFile
	}
	if override.ComposeFile != "" {
		p.ComposeFile = override.ComposeFile
	}
	if override.LintCmd != "" {
		p.LintCmd = override.LintCmd
	}
	if override.UpCmd != "" {
		p.UpCmd = override.UpCmd
	}
	if len(override.HealthURLs) > 0 {
		p.HealthURLs = override.HealthURLs
	}
	return p
}

func readYAML(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}

func writeYAML(path string, in any) error {
	data, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func expandPath(path string) (string, error) {
	if path == "" || path[0] != '~' {
		return filepath.Abs(path)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" {
		return home, nil
	}
	if len(path) > 1 && os.IsPathSeparator(path[1]) {
		return filepath.Join(home, path[2:]), nil
	}
	return "", fmt.Errorf("unsupported path %q", path)
}

func sortProjects(projects []project.Project) {
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})
}
