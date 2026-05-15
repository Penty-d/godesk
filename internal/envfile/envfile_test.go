package envfile

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	input := `
# comment
APP_PORT=8080
export DB_PORT="5432"
EMPTY=
IGNORED
`
	entries, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0] != (Entry{Key: "APP_PORT", Value: "8080"}) {
		t.Fatalf("unexpected first entry: %#v", entries[0])
	}
	if entries[1] != (Entry{Key: "DB_PORT", Value: "5432"}) {
		t.Fatalf("unexpected second entry: %#v", entries[1])
	}
}
