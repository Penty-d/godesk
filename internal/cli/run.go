package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"godesk/internal/runner"
)

func newLintCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "lint <project>",
		Short: "Run project lint command",
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
			if p.LintCmd == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "lint command not configured; add lint_cmd to .godesk.yaml")
				return nil
			}
			return runner.Runner{Stdout: os.Stdout, Stderr: os.Stderr}.Run(context.Background(), p.Path, p.LintCmd)
		},
	}
}
