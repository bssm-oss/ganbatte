package history

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// BashParser parses bash history files.
// Format: one command per line (no timestamps by default).
type BashParser struct{}

// Parse reads a bash history file and returns entries.
func (p *BashParser) Parse(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening bash history: %w", err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, Entry{Command: line})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading bash history: %w", err)
	}

	return entries, nil
}
