package cli

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"godesk/internal/compose"
	"godesk/internal/envfile"
	"godesk/internal/project"
)

type doctorCheck struct {
	Name    string
	Status  string
	Message string
}

func newDoctorCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor <project>",
		Short: "Check project configuration and local tools",
		Args: func(cmd *cobra.Command, args []string) error {
			_, err := requireProjectName(cmd, args)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := requireProjectName(cmd, args)
			p, err := app.store.FindProject(name)
			if err != nil {
				return err
			}
			checks := runDoctorChecks(p)
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "CHECK\tSTATUS\tMESSAGE")
			failed := false
			for _, check := range checks {
				if check.Status == "fail" {
					failed = true
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", check.Name, check.Status, check.Message)
			}
			if err := w.Flush(); err != nil {
				return err
			}
			if failed {
				return fmt.Errorf("doctor found configuration issues")
			}
			return nil
		},
	}
}

func runDoctorChecks(p project.Project) []doctorCheck {
	checks := []doctorCheck{
		checkProjectRoot(p),
		checkGoMod(p),
		checkEnvFile(p),
		checkComposeFile(p),
	}
	checks = append(checks, checkLogFiles(p)...)
	checks = append(checks, checkHealthURLs(p)...)
	checks = append(checks, checkTool("docker", "Docker CLI for up and logs"))
	checks = append(checks, checkTool("lsof", "lsof for ports"))
	return checks
}

func checkProjectRoot(p project.Project) doctorCheck {
	info, err := os.Stat(p.Path)
	if err != nil {
		return doctorCheck{"project_root", "fail", err.Error()}
	}
	if !info.IsDir() {
		return doctorCheck{"project_root", "fail", "path is not a directory"}
	}
	return doctorCheck{"project_root", "ok", p.Path}
}

func checkGoMod(p project.Project) doctorCheck {
	path := filepath.Join(p.Path, "go.mod")
	if _, err := os.Stat(path); err != nil {
		return doctorCheck{"go_mod", "fail", err.Error()}
	}
	return doctorCheck{"go_mod", "ok", "go.mod found"}
}

func checkEnvFile(p project.Project) doctorCheck {
	if p.EnvFile == "" {
		return doctorCheck{"env_file", "warn", "not configured"}
	}
	path, err := project.ResolveProjectPath(p.Path, p.EnvFile)
	if err != nil {
		return doctorCheck{"env_file", "fail", err.Error()}
	}
	file, err := os.Open(path)
	if err != nil {
		return doctorCheck{"env_file", "fail", err.Error()}
	}
	defer file.Close()
	entries, err := envfile.Parse(file)
	if err != nil {
		return doctorCheck{"env_file", "fail", err.Error()}
	}
	return doctorCheck{"env_file", "ok", fmt.Sprintf("%s (%d entries)", p.EnvFile, len(entries))}
}

func checkComposeFile(p project.Project) doctorCheck {
	if p.ComposeFile == "" {
		if p.UpCmd != "" {
			return doctorCheck{"compose_file", "warn", "not configured; up_cmd is configured"}
		}
		return doctorCheck{"compose_file", "warn", "not configured"}
	}
	path, err := project.ResolveProjectPath(p.Path, p.ComposeFile)
	if err != nil {
		return doctorCheck{"compose_file", "fail", err.Error()}
	}
	file, err := compose.Load(path)
	if err != nil {
		return doctorCheck{"compose_file", "fail", err.Error()}
	}
	return doctorCheck{"compose_file", "ok", fmt.Sprintf("%s (%d services)", p.ComposeFile, len(file.Services))}
}

func checkLogFiles(p project.Project) []doctorCheck {
	if len(p.LogFiles) == 0 {
		return []doctorCheck{{"log_files", "warn", "not configured"}}
	}
	checks := make([]doctorCheck, 0, len(p.LogFiles))
	for _, logFile := range p.LogFiles {
		path, err := project.ResolveProjectPath(p.Path, logFile)
		if err != nil {
			checks = append(checks, doctorCheck{"log_file", "fail", err.Error()})
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			checks = append(checks, doctorCheck{"log_file", "warn", fmt.Sprintf("%s: %v", logFile, err)})
			continue
		}
		if info.IsDir() {
			checks = append(checks, doctorCheck{"log_file", "fail", logFile + " is a directory"})
			continue
		}
		checks = append(checks, doctorCheck{"log_file", "ok", logFile})
	}
	return checks
}

func checkHealthURLs(p project.Project) []doctorCheck {
	if len(p.HealthURLs) == 0 {
		return []doctorCheck{{"health_urls", "warn", "not configured"}}
	}
	checks := make([]doctorCheck, 0, len(p.HealthURLs))
	for _, rawURL := range p.HealthURLs {
		parsed, err := url.ParseRequestURI(rawURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			checks = append(checks, doctorCheck{"health_url", "fail", rawURL + " is not a valid absolute URL"})
			continue
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			checks = append(checks, doctorCheck{"health_url", "fail", rawURL + " must use http or https"})
			continue
		}
		checks = append(checks, doctorCheck{"health_url", "ok", rawURL})
	}
	return checks
}

func checkTool(name string, purpose string) doctorCheck {
	path, err := exec.LookPath(name)
	if err != nil {
		return doctorCheck{name, "warn", purpose + " not available"}
	}
	return doctorCheck{name, "ok", path}
}
