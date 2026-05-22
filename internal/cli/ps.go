package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	dockercmd "godesk/internal/docker"
)

func newPsCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "ps <project>",
		Short: "Show Docker Compose service status",
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
			if p.ComposeFile == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "compose_file not configured; add compose_file to .godesk.yaml")
				return nil
			}
			return dockercmd.Compose{Stdout: os.Stdout, Stderr: os.Stderr}.Ps(context.Background(), p.Path, p.ComposeFile)
		},
	}
}
