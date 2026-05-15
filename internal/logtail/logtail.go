package logtail

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Tailer struct {
	Stdout       io.Writer
	Stderr       io.Writer
	PollInterval time.Duration
}

func (t Tailer) TailFiles(ctx context.Context, root string, paths []string, tail int) error {
	if len(paths) == 0 {
		return nil
	}
	interval := t.PollInterval
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	if tail < 0 {
		tail = 0
	}

	var wg sync.WaitGroup
	for _, path := range paths {
		path := path
		wg.Add(1)
		go func() {
			defer wg.Done()
			t.tailFile(ctx, root, path, tail, interval)
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		<-done
		return nil
	case <-done:
		return nil
	}
}

func (t Tailer) tailFile(ctx context.Context, root string, path string, tail int, interval time.Duration) {
	resolved := resolvePath(root, path)
	file, err := os.Open(resolved)
	if err != nil {
		fmt.Fprintf(t.Stderr, "[file:%s] %v\n", path, err)
		return
	}
	defer file.Close()

	lines, err := lastLines(file, tail)
	if err != nil {
		fmt.Fprintf(t.Stderr, "[file:%s] %v\n", path, err)
		return
	}
	for _, line := range lines {
		fmt.Fprintf(t.Stdout, "[file:%s] %s\n", path, line)
	}
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		fmt.Fprintf(t.Stderr, "[file:%s] %v\n", path, err)
		return
	}

	reader := bufio.NewReader(file)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := reader.ReadString('\n')
		if line != "" {
			fmt.Fprintf(t.Stdout, "[file:%s] %s", path, line)
			if line[len(line)-1] != '\n' {
				fmt.Fprintln(t.Stdout)
			}
		}
		if err == nil {
			continue
		}
		if errors.Is(err, io.EOF) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				continue
			}
		}
		fmt.Fprintf(t.Stderr, "[file:%s] %v\n", path, err)
		return
	}
}

func lastLines(file *os.File, count int) ([]string, error) {
	if count == 0 {
		return nil, nil
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	lines := make([]string, 0, count)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if len(lines) < count {
			lines = append(lines, line)
			continue
		}
		copy(lines, lines[1:])
		lines[len(lines)-1] = line
	}
	return lines, scanner.Err()
}

func resolvePath(root string, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
