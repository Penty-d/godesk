package config

import (
	"testing"

	"godesk/internal/project"
)

func TestApplyOverride(t *testing.T) {
	base := project.Project{
		Name:        "base",
		Path:        "/tmp/base",
		EnvFile:     ".env",
		ComposeFile: "compose.yaml",
		TestCmd:     "go test ./...",
	}
	got := ApplyOverride(base, ProjectOverride{
		Name:        "custom",
		ComposeFile: "docker-compose.dev.yml",
		LintCmd:     "golangci-lint run",
		HealthURLs:  []string{"http://localhost:8080/health"},
	})
	if got.Name != "custom" {
		t.Fatalf("name was not overridden: %#v", got)
	}
	if got.EnvFile != ".env" {
		t.Fatalf("empty env override should keep auto value: %#v", got)
	}
	if got.ComposeFile != "docker-compose.dev.yml" || got.LintCmd == "" || len(got.HealthURLs) != 1 {
		t.Fatalf("override not applied: %#v", got)
	}
}
