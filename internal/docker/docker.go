package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"godesk/internal/project"
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
		resolved, err := project.NormalizeProjectPath(dir, composeFile)
		if err != nil {
			return err
		}
		args = append(args, "-f", resolved)
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

func (c Compose) Logs(ctx context.Context, dir string, composeFile string, services []string, tail int) error {
	if _, err := exec.LookPath("docker"); err != nil {
		return errors.New("docker CLI not found")
	}
	args := []string{"compose"}
	if composeFile != "" {
		resolved, err := project.NormalizeProjectPath(dir, composeFile)
		if err != nil {
			return err
		}
		args = append(args, "-f", resolved)
	}
	args = append(args, "logs", "-f", "--tail", fmt.Sprintf("%d", tail))
	args = append(args, services...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = dir
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil
		}
		return classifyDockerError(err, "")
	}
	return nil
}

func (c Compose) Ps(ctx context.Context, dir string, composeFile string) error {
	if _, err := exec.LookPath("docker"); err != nil {
		return errors.New("docker CLI not found")
	}
	args := []string{"compose"}
	if composeFile != "" {
		resolved, err := project.NormalizeProjectPath(dir, composeFile)
		if err != nil {
			return err
		}
		args = append(args, "-f", resolved)
	}
	args = append(args, "ps")

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = dir
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
	if err := cmd.Run(); err != nil {
		return classifyDockerError(err, "")
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
