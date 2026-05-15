package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"godesk/internal/config"
)

func newRootsCommand(app *appContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roots",
		Short: "Manage default scan roots",
	}
	cmd.AddCommand(
		newRootsListCommand(app),
		newRootsAddCommand(app),
		newRootsRemoveCommand(app),
	)
	return cmd
}

func newRootsListCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured scan roots",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.store.LoadGlobal()
			if err != nil {
				return err
			}
			if len(cfg.Roots) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no roots configured")
				return nil
			}
			for _, root := range cfg.Roots {
				fmt.Fprintln(cmd.OutOrStdout(), root)
			}
			return nil
		},
	}
}

func newRootsAddCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "add <path>",
		Short: "Add a default scan root",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := config.NormalizeRoot(args[0])
			if err != nil {
				return err
			}
			cfg, err := app.store.LoadGlobal()
			if err != nil {
				return err
			}
			if containsString(cfg.Roots, root) {
				fmt.Fprintf(cmd.OutOrStdout(), "already configured: %s\n", root)
				return nil
			}
			cfg.Roots = append(cfg.Roots, root)
			if err := app.store.SaveGlobal(cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "added root: %s\n", root)
			return nil
		},
	}
}

func newRootsRemoveCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <path>",
		Short: "Remove a default scan root",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := config.NormalizeRoot(args[0])
			if err != nil {
				return err
			}
			cfg, err := app.store.LoadGlobal()
			if err != nil {
				return err
			}
			next := make([]string, 0, len(cfg.Roots))
			removed := false
			for _, existing := range cfg.Roots {
				if existing == root {
					removed = true
					continue
				}
				next = append(next, existing)
			}
			if !removed {
				fmt.Fprintf(cmd.OutOrStdout(), "not configured: %s\n", root)
				return nil
			}
			cfg.Roots = next
			if err := app.store.SaveGlobal(cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed root: %s\n", root)
			return nil
		},
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
