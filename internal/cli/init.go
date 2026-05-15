package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"godesk/internal/config"
	"godesk/internal/project"
)

func newInitCommand(app *appContext) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init <project>",
		Short: "Create a .godesk.yaml for an indexed project",
		Args: func(cmd *cobra.Command, args []string) error {
			_, err := requireProjectName(cmd, args)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := requireProjectName(cmd, args)
			p, err := initIndexedProject(app, name)
			if err != nil {
				return err
			}
			override := config.ProjectOverride{
				Name:        p.Name,
				EnvFile:     p.EnvFile,
				ComposeFile: p.ComposeFile,
				LintCmd:     p.LintCmd,
				UpCmd:       p.UpCmd,
				HealthURLs:  p.HealthURLs,
				LogFiles:    p.LogFiles,
			}
			path, err := config.SaveProjectOverride(p.Path, override, force)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "created %s\n", path)
			fmt.Fprintf(out, "project: %s\n", p.Name)
			fmt.Fprintf(out, "root: %s\n", p.Path)
			fmt.Fprintf(out, "env: %s\n", marker(p.EnvFile))
			fmt.Fprintf(out, "compose: %s\n", marker(p.ComposeFile))
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing .godesk.yaml")
	return cmd
}

func newInitLocalCommand(app *appContext) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init-local",
		Short: "Create a .godesk.yaml for the current Go project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := discoverCurrentProject()
			if err != nil {
				return err
			}
			override := config.ProjectOverride{
				Name:        p.Name,
				EnvFile:     p.EnvFile,
				ComposeFile: p.ComposeFile,
				LintCmd:     p.LintCmd,
				UpCmd:       p.UpCmd,
				HealthURLs:  p.HealthURLs,
				LogFiles:    p.LogFiles,
			}
			path, err := config.SaveProjectOverride(p.Path, override, force)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "created %s\n", path)
			fmt.Fprintf(out, "project: %s\n", p.Name)
			fmt.Fprintf(out, "root: %s\n", p.Path)
			fmt.Fprintf(out, "env: %s\n", marker(p.EnvFile))
			fmt.Fprintf(out, "compose: %s\n", marker(p.ComposeFile))
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing .godesk.yaml")
	return cmd
}

func initIndexedProject(app *appContext, name string) (project.Project, error) {
	p, err := app.store.FindProject(name)
	if err != nil {
		return project.Project{}, err
	}
	discovered, err := project.Discover(p.Path)
	if err != nil {
		return project.Project{}, err
	}
	p.EnvFile = discovered.EnvFile
	p.ComposeFile = discovered.ComposeFile
	return p, nil
}

func discoverCurrentProject() (project.Project, error) {
	root, err := project.FindRoot(".")
	if err != nil {
		return project.Project{}, err
	}
	p, err := project.Discover(root)
	if err != nil {
		return project.Project{}, err
	}
	return p, nil
}
