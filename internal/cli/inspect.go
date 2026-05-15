package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"godesk/internal/compose"
	"godesk/internal/envfile"
)

func newInspectCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <project>",
		Short: "Inspect resolved project configuration",
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
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "name: %s\npath: %s\nenv: %s\ncompose: %s\nlint: %s\nup: %s\n",
				p.Name, p.Path, marker(p.EnvFile), marker(p.ComposeFile), marker(p.LintCmd), marker(p.UpCmd))
			if p.EnvFile != "" {
				file, err := os.Open(filepath.Join(p.Path, p.EnvFile))
				if err == nil {
					defer file.Close()
					entries, err := envfile.Parse(file)
					if err != nil {
						return err
					}
					fmt.Fprintln(out, "\nenv entries:")
					for _, entry := range entries {
						fmt.Fprintf(out, "  %s=%s\n", entry.Key, entry.Value)
					}
				}
			}
			if p.ComposeFile != "" {
				file, err := compose.Load(filepath.Join(p.Path, p.ComposeFile))
				if err != nil {
					return err
				}
				fmt.Fprintln(out, "\ncompose services:")
				for service, spec := range file.Services {
					fmt.Fprintf(out, "  %s", service)
					for _, port := range spec.Ports {
						fmt.Fprintf(out, " %d:%d", port.Published, port.Target)
					}
					fmt.Fprintln(out)
				}
			}
			return nil
		},
	}
}
