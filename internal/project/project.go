package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Project struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	EnvFile     string   `yaml:"env_file,omitempty"`
	ComposeFile string   `yaml:"compose_file,omitempty"`
	LintCmd     string   `yaml:"lint_cmd,omitempty"`
	UpCmd       string   `yaml:"up_cmd,omitempty"`
	HealthURLs  []string `yaml:"health_urls,omitempty"`
	LogFiles    []string `yaml:"log_files,omitempty"`
}

func Discover(root string) (Project, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Project{}, err
	}
	return discover(abs, abs)
}

func FindRoot(start string) (string, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		abs = filepath.Dir(abs)
	}
	for {
		if _, err := os.Stat(filepath.Join(abs, "go.mod")); err == nil {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", fmt.Errorf("go.mod not found from %s upward", start)
		}
		abs = parent
	}
}

func discover(root string, searchRoot string) (Project, error) {
	files, err := findProjectFiles(root, searchRoot)
	if err != nil {
		return Project{}, err
	}
	return Project{
		Name:        filepath.Base(root),
		Path:        root,
		EnvFile:     files.env,
		ComposeFile: files.compose,
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

type discoveredFiles struct {
	env     string
	compose string
}

func findProjectFiles(root string, searchRoot string) (discoveredFiles, error) {
	files, err := findProjectFilesUpward(root, searchRoot)
	if err != nil {
		return discoveredFiles{}, err
	}
	if files.env != "" && files.compose != "" {
		return files, nil
	}
	downward, err := findProjectFilesDownward(root, 3)
	if err != nil {
		return discoveredFiles{}, err
	}
	if files.env == "" {
		files.env = downward.env
	}
	if files.compose == "" {
		files.compose = downward.compose
	}
	return files, nil
}

func findProjectFilesUpward(root string, searchRoot string) (discoveredFiles, error) {
	files := discoveredFiles{}
	current := root
	for {
		if files.env == "" {
			env, err := firstExistingInDir(root, current, envFileNames...)
			if err != nil {
				return discoveredFiles{}, err
			}
			files.env = env
		}
		if files.compose == "" {
			compose, err := firstExistingInDir(root, current, composeFileNames...)
			if err != nil {
				return discoveredFiles{}, err
			}
			files.compose = compose
		}
		if files.env != "" && files.compose != "" {
			return files, nil
		}
		if current == searchRoot {
			return files, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return files, nil
		}
		rel, err := filepath.Rel(searchRoot, parent)
		if err != nil {
			return discoveredFiles{}, err
		}
		if strings.HasPrefix(rel, "..") {
			return files, nil
		}
		current = parent
	}
}

func findProjectFilesDownward(root string, maxDepth int) (discoveredFiles, error) {
	var envMatches []string
	var composeMatches []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		depth := pathDepth(rel)
		if d.IsDir() {
			if shouldSkipDir(d.Name(), true) || depth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		if depth > maxDepth {
			return nil
		}
		if hasName(d.Name(), envFileNames...) {
			envMatches = append(envMatches, rel)
		}
		if hasName(d.Name(), composeFileNames...) {
			composeMatches = append(composeMatches, rel)
		}
		return nil
	})
	if err != nil {
		return discoveredFiles{}, err
	}
	return discoveredFiles{
		env:     bestMatch(envMatches),
		compose: bestMatch(composeMatches),
	}, nil
}

func firstExistingInDir(root string, dir string, names ...string) (string, error) {
	for _, name := range names {
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Rel(root, candidate)
		}
	}
	return "", nil
}

func bestMatch(matches []string) string {
	if len(matches) == 0 {
		return ""
	}
	sort.Slice(matches, func(i, j int) bool {
		leftDepth := pathDepth(matches[i])
		rightDepth := pathDepth(matches[j])
		if leftDepth != rightDepth {
			return leftDepth < rightDepth
		}
		return matches[i] < matches[j]
	})
	return matches[0]
}

func hasName(target string, names ...string) bool {
	for _, name := range names {
		if target == name {
			return true
		}
	}
	return false
}

func pathDepth(path string) int {
	if path == "." || path == "" {
		return 0
	}
	return len(strings.Split(filepath.Clean(path), string(os.PathSeparator)))
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

var composeFileNames = []string{"docker-compose.yml", "docker-compose.yaml", "compose.yaml"}

var envFileNames = []string{".env"}
