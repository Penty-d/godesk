package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Compose struct {
	Stdout io.Writer
	Stderr io.Writer
}

func (c Compose) Up(ctx context.Context, dir string, composeFile string) error {
	if _, err := exec.LookPath("docker"); err != nil {
		return errors.New("docker CLI not found")
	}
	args := []string{"compose"}
	if composeFile != "" {
		args = append(args, "-f", composeFile)
	}
	args = append(args, "up", "-d")

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		if _, writeErr := io.Copy(c.Stdout, bytes.NewReader(output)); writeErr != nil {
			return writeErr
		}
	}
	if err != nil {
		return classifyDockerError(err, string(output))
	}
	return nil
}

func classifyDockerError(err error, output string) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(output + "\n" + err.Error())
	if strings.Contains(msg, "permission denied") {
		return fmt.Errorf("docker socket permission denied: %w", err)
	}
	if strings.Contains(msg, "cannot connect") || strings.Contains(msg, "connect: no such file") {
		return fmt.Errorf("docker daemon is not running: %w", err)
	}
	return fmt.Errorf("docker compose failed: %w", err)
}
