package envfile

import (
	"bufio"
	"io"
	"strings"
)

type Entry struct {
	Key   string
	Value string
}

func Parse(r io.Reader) ([]Entry, error) {
	scanner := bufio.NewScanner(r)
	entries := []Entry{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key != "" {
			entries = append(entries, Entry{Key: key, Value: value})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
