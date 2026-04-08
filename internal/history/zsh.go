package history

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ZshParser parses zsh history files.
// Format: ": <timestamp>:0;<command>" or plain "<command>"
type ZshParser struct{}

// Parse reads a zsh history file and returns entries.
func (p *ZshParser) Parse(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening zsh history: %w", err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		entry, ok := parseZshLine(line)
		if ok {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading zsh history: %w", err)
	}

	return entries, nil
}

// parseZshLine parses a single zsh history line.
// Extended format: ": 1234567890:0;command"
// Plain format: "command"
func parseZshLine(line string) (Entry, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return Entry{}, false
	}

	// Extended format: ": timestamp:0;command"
	if strings.HasPrefix(line, ": ") {
		parts := strings.SplitN(line[2:], ";", 2)
		if len(parts) == 2 {
			cmd := strings.TrimSpace(parts[1])
			if cmd == "" {
				return Entry{}, false
			}

			// Parse timestamp from "timestamp:0"
			tsPart := strings.SplitN(parts[0], ":", 2)
			ts, err := strconv.ParseInt(strings.TrimSpace(tsPart[0]), 10, 64)
			if err == nil {
				return Entry{
					Command:   cmd,
					Timestamp: time.Unix(ts, 0),
				}, true
			}

			// Fallback: timestamp parse failed but command exists
			return Entry{Command: cmd}, true
		}
	}

	// Plain format
	return Entry{Command: line}, true
}
