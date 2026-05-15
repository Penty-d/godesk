package cli

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	dockercmd "godesk/internal/docker"
	"godesk/internal/runner"
)

func newUpCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "up <project>",
		Short: "Start project dependency services",
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
			ctx := context.Background()
			if p.UpCmd != "" {
				return runner.Runner{Stdout: os.Stdout, Stderr: os.Stderr}.Run(ctx, p.Path, p.UpCmd)
			}
			return dockercmd.Compose{Stdout: os.Stdout, Stderr: os.Stderr}.Up(ctx, p.Path, p.ComposeFile)
		},
	}
}
