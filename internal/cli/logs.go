package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"

	"github.com/spf13/cobra"

	"godesk/internal/docker"
	"godesk/internal/logtail"
)

func newLogsCommand(app *appContext) *cobra.Command {
	var filesOnly bool
	var composeOnly bool
	var tail int

	cmd := &cobra.Command{
		Use:   "logs <project> [service...]",
		Short: "Tail project file and Docker Compose logs",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("expected a project name")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if tail < 0 {
				return fmt.Errorf("tail must be greater than or equal to 0")
			}
			if filesOnly && composeOnly {
				return fmt.Errorf("--files-only and --compose-only cannot be used together")
			}
			if filesOnly && len(args) > 1 {
				return fmt.Errorf("service filters require Compose logs")
			}

			p, err := app.store.FindProject(args[0])
			if err != nil {
				return err
			}
			services := args[1:]
			if filesOnly && len(p.LogFiles) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "log_files not configured; add log_files to .godesk.yaml")
				return nil
			}
			if composeOnly && p.ComposeFile == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "compose_file not configured; add compose_file to .godesk.yaml")
				return nil
			}
			if len(services) > 0 && p.ComposeFile == "" {
				return fmt.Errorf("service filters require Compose logs")
			}
			runFiles := !composeOnly && len(p.LogFiles) > 0
			runCompose := !filesOnly && p.ComposeFile != ""
			if !runFiles && !runCompose {
				fmt.Fprintln(cmd.OutOrStdout(), "logs not configured; add log_files or compose_file to .godesk.yaml")
				return nil
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			var mu sync.Mutex
			errs := make(chan error, 2)
			var wg sync.WaitGroup
			if runFiles {
				wg.Add(1)
				go func() {
					defer wg.Done()
					errs <- logtail.Tailer{
						Stdout: &lockedWriter{mu: &mu, writer: cmd.OutOrStdout()},
						Stderr: &lockedWriter{mu: &mu, writer: cmd.ErrOrStderr()},
					}.TailFiles(ctx, p.Path, p.LogFiles, tail)
				}()
			}
			if runCompose {
				wg.Add(1)
				go func() {
					defer wg.Done()
					errs <- docker.Compose{
						Stdout: newPrefixWriter(&mu, cmd.OutOrStdout(), "[compose] "),
						Stderr: newPrefixWriter(&mu, cmd.ErrOrStderr(), "[compose] "),
					}.Logs(ctx, p.Path, p.ComposeFile, services, tail)
				}()
			}

			go func() {
				wg.Wait()
				close(errs)
			}()

			var firstErr error
			for err := range errs {
				if err != nil && !errors.Is(err, context.Canceled) && firstErr == nil {
					firstErr = err
					stop()
				}
			}
			return firstErr
		},
	}
	cmd.Flags().BoolVar(&filesOnly, "files-only", false, "tail configured log files only")
	cmd.Flags().BoolVar(&composeOnly, "compose-only", false, "tail Docker Compose logs only")
	cmd.Flags().IntVar(&tail, "tail", 200, "number of existing log lines to show")
	return cmd
}

type lockedWriter struct {
	mu     *sync.Mutex
	writer io.Writer
}

func (w *lockedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(p)
}

type prefixWriter struct {
	mu          *sync.Mutex
	writer      io.Writer
	prefix      string
	atLineStart bool
}

func newPrefixWriter(mu *sync.Mutex, writer io.Writer, prefix string) *prefixWriter {
	return &prefixWriter{
		mu:          mu,
		writer:      writer,
		prefix:      prefix,
		atLineStart: true,
	}
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	written := 0
	for written < len(p) {
		if w.atLineStart {
			if _, err := io.WriteString(w.writer, w.prefix); err != nil {
				return written, err
			}
			w.atLineStart = false
		}
		next := written
		for next < len(p) && p[next] != '\n' {
			next++
		}
		if next > written {
			n, err := w.writer.Write(p[written:next])
			written += n
			if err != nil {
				return written, err
			}
		}
		if next < len(p) && p[next] == '\n' {
			if _, err := w.writer.Write([]byte{'\n'}); err != nil {
				return written, err
			}
			written++
			w.atLineStart = true
		}
	}
	return len(p), nil
}
