package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"godesk/internal/compose"
	"godesk/internal/envfile"
	portscan "godesk/internal/ports"
)

func newPortsCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "ports <project>",
		Short: "Show project port occupancy",
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
			candidates := []portscan.Candidate{}
			if p.EnvFile != "" {
				file, err := os.Open(filepath.Join(p.Path, p.EnvFile))
				if err != nil {
					return err
				}
				defer file.Close()
				entries, err := envfile.Parse(file)
				if err != nil {
					return err
				}
				candidates = append(candidates, portscan.FromEnv(entries)...)
			}
			if p.ComposeFile != "" {
				file, err := compose.Load(filepath.Join(p.Path, p.ComposeFile))
				if err != nil {
					return err
				}
				candidates = append(candidates, portscan.FromCompose(file)...)
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SOURCE\tNAME\tPORT\tSTATUS\tPROCESS")
			for _, candidate := range candidates {
				status := portscan.Check(context.Background(), candidate)
				state := "free"
				if status.Listening {
					state = "busy"
				}
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n", status.Source, status.Name, status.Port, state, status.Process)
			}
			return w.Flush()
		},
	}
}
