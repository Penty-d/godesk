package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"godesk/internal/config"
	"godesk/internal/project"
)

func newScanCommand(app *appContext) *cobra.Command {
	var roots []string
	cmd := &cobra.Command{
		Use:   "scan [root...]",
		Short: "Scan roots for Go projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, err := app.store.LoadGlobal()
			if err != nil {
				return err
			}
			allRoots := append([]string{}, global.Roots...)
			allRoots = append(allRoots, roots...)
			allRoots = append(allRoots, args...)
			if len(allRoots) == 0 {
				return fmt.Errorf("no roots configured; pass a root path or add roots to %s", app.store.ConfigPath())
			}

			projects, err := project.ScanRoots(allRoots)
			if err != nil {
				return err
			}
			for i, p := range projects {
				override, ok, err := config.LoadProjectOverride(p.Path)
				if err != nil {
					return err
				}
				if ok {
					projects[i], err = config.ApplyOverride(p, override)
					if err != nil {
						return err
					}
				}
			}
			if err := app.store.SaveIndex(config.ProjectIndex{Projects: projects}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "saved %d projects to %s\n", len(projects), app.store.ProjectsPath())
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&roots, "root", nil, "root directory to scan")
	return cmd
}
