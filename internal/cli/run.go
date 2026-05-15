package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"godesk/internal/runner"
)

func newTestCommand(app *appContext) *cobra.Command {
	return runProjectCommand(app, "test <project>", "Run project tests", func(testCmd, lintCmd string) (string, bool) {
		if testCmd == "" {
			return "go test ./...", true
		}
		return testCmd, true
	})
}

func newLintCommand(app *appContext) *cobra.Command {
	return runProjectCommand(app, "lint <project>", "Run project lint command", func(testCmd, lintCmd string) (string, bool) {
		if lintCmd == "" {
			return "", false
		}
		return lintCmd, true
	})
}

func runProjectCommand(app *appContext, use string, short string, pick func(testCmd, lintCmd string) (string, bool)) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
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
			command, ok := pick(p.TestCmd, p.LintCmd)
			if !ok {
				fmt.Fprintln(cmd.OutOrStdout(), "lint command not configured; add lint_cmd to .godesk.yaml")
				return nil
			}
			return runner.Runner{Stdout: os.Stdout, Stderr: os.Stderr}.Run(context.Background(), p.Path, command)
		},
	}
}
