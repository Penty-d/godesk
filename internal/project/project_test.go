package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanRootsFindsGoProjectsAndSkipsVendor(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "app")
	vendor := filepath.Join(root, "vendor", "ignored")
	hidden := filepath.Join(root, ".hidden", "ignored")
	for _, dir := range []string{app, vendor, hidden} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(app, ".env"), []byte("PORT=8080\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "compose.yaml"), []byte("services: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	projects, err := ScanRoots([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d: %#v", len(projects), projects)
	}
	if projects[0].Name != "app" {
		t.Fatalf("unexpected project name: %s", projects[0].Name)
	}
	if projects[0].EnvFile != ".env" || projects[0].ComposeFile != "compose.yaml" {
		t.Fatalf("unexpected discovered files: %#v", projects[0])
	}
}

func TestScanRootsFindsComposeInParentDirectory(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	app := filepath.Join(repo, "services", "api")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "go.mod"), []byte("module api\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "docker-compose.yml"), []byte("services: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	projects, err := ScanRoots([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d: %#v", len(projects), projects)
	}
	if projects[0].ComposeFile != "../../docker-compose.yml" {
		t.Fatalf("expected parent compose path, got %#v", projects[0])
	}
}
