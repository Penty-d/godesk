package ports

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"godesk/internal/compose"
	"godesk/internal/envfile"
)

type Candidate struct {
	Source string
	Name   string
	Port   int
}

type Status struct {
	Candidate
	Listening bool
	Process   string
}

func FromEnv(entries []envfile.Entry) []Candidate {
	var candidates []Candidate
	for _, entry := range entries {
		if !strings.Contains(strings.ToUpper(entry.Key), "PORT") {
			continue
		}
		port, err := strconv.Atoi(entry.Value)
		if err != nil || port <= 0 {
			continue
		}
		candidates = append(candidates, Candidate{
			Source: "env",
			Name:   entry.Key,
			Port:   port,
		})
	}
	return candidates
}

func FromCompose(file compose.File) []Candidate {
	var candidates []Candidate
	for name, svc := range file.Services {
		for _, port := range svc.Ports {
			if port.Published <= 0 {
				continue
			}
			candidates = append(candidates, Candidate{
				Source: "compose",
				Name:   name,
				Port:   port.Published,
			})
		}
	}
	return candidates
}

func Check(ctx context.Context, candidate Candidate) Status {
	status := Status{Candidate: candidate}
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(checkCtx, "lsof", "-nP", fmt.Sprintf("-iTCP:%d", candidate.Port), "-sTCP:LISTEN")
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return status
	}
	status.Listening = true
	status.Process = firstProcessLine(string(output))
	return status
}

func firstProcessLine(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) <= 1 {
		return ""
	}
	fields := strings.Fields(lines[1])
	if len(fields) >= 2 {
		return fields[0] + " " + fields[1]
	}
	return lines[1]
}
