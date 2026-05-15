package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"godesk/internal/config"
)

type appContext struct {
	store *config.Store
}

func Execute() error {
	ctx := &appContext{store: config.NewStore()}

	rootCmd := &cobra.Command{
		Use:           "godesk",
		Short:         "Manage local Go backend workspaces",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(
		newInitCommand(ctx),
		newInitLocalCommand(ctx),
		newRootsCommand(ctx),
		newScanCommand(ctx),
		newListCommand(ctx),
		newInspectCommand(ctx),
		newUpCommand(ctx),
		newPortsCommand(ctx),
		newHealthCommand(ctx),
		newLogsCommand(ctx),
		newLintCommand(ctx),
	)

	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	return rootCmd.Execute()
}

func requireProjectName(cmd *cobra.Command, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("expected exactly one project name")
	}
	return args[0], nil
}
