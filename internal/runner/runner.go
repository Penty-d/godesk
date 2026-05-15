package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Runner struct {
	Stdout io.Writer
	Stderr io.Writer
}

func (r Runner) Run(ctx context.Context, dir string, command string) error {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return errors.New("empty command")
	}
	cmd := exec.CommandContext(ctx, fields[0], fields[1:]...)
	cmd.Dir = dir
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", command, err)
	}
	return nil
}
