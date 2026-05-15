package cli

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"godesk/internal/health"
)

func newHealthCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "health <project>",
		Short: "Show project health status",
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
			if len(p.HealthURLs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "health_urls not configured; add health_urls to .godesk.yaml")
				return nil
			}

			checker := health.Checker{}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "URL\tSTATUS\tLATENCY\tERROR")
			failed := false
			for _, url := range p.HealthURLs {
				result := checker.Check(context.Background(), url)
				state := "ok"
				if !result.OK {
					state = "fail"
					failed = true
				}
				status := "-"
				if result.StatusCode != 0 {
					status = fmt.Sprintf("%d", result.StatusCode)
				}
				fmt.Fprintf(w, "%s\t%s/%s\t%s\t%s\n", result.URL, state, status, result.Latency.Round(1e6), result.Error)
			}
			if err := w.Flush(); err != nil {
				return err
			}
			if failed {
				return fmt.Errorf("one or more health checks failed")
			}
			return nil
		},
	}
}
