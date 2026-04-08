package history

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// FishParser parses fish shell history files.
// Format:
//
//	- cmd: <command>
//	  when: <timestamp>
type FishParser struct{}

// Parse reads a fish history file and returns entries.
func (p *FishParser) Parse(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening fish history: %w", err)
	}
	defer f.Close()

	var entries []Entry
	var current Entry
	hasCmd := false
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "- cmd: ") {
			// Save previous entry if exists
			if hasCmd {
				entries = append(entries, current)
			}
			current = Entry{Command: strings.TrimPrefix(line, "- cmd: ")}
			hasCmd = true
		} else if strings.HasPrefix(line, "  when: ") && hasCmd {
			tsStr := strings.TrimSpace(strings.TrimPrefix(line, "  when: "))
			ts, err := strconv.ParseInt(tsStr, 10, 64)
			if err == nil {
				current.Timestamp = time.Unix(ts, 0)
			}
		}
	}

	// Don't forget the last entry
	if hasCmd {
		entries = append(entries, current)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading fish history: %w", err)
	}

	return entries, nil
}
