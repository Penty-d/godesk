package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newListCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List indexed Go projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			idx, err := app.store.LoadIndex()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tENV\tCOMPOSE\tPATH")
			for _, p := range idx.Projects {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, marker(p.EnvFile), marker(p.ComposeFile), p.Path)
			}
			return w.Flush()
		},
	}
}

func marker(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
