package history

import "time"

// Entry represents a single command from shell history.
type Entry struct {
	Command   string
	Timestamp time.Time
}

// Parser defines the interface for parsing shell history files.
type Parser interface {
	// Parse reads a history file and returns entries.
	Parse(path string) ([]Entry, error)
}
