package track

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/justn-hyeok/ganbatte/internal/history"
)

const maxLogSize = 10 * 1024 * 1024 // 10 MB

// LogPath returns the path to the passive tracking log.
func LogPath() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "ganbatte", "track.log"), nil
}

// Parse reads the track log and returns history entries.
// Rotates the log if it exceeds maxLogSize before reading.
// Line format: unix_timestamp\texit_code\tcommand
func Parse(path string) ([]history.Entry, error) {
	rotateIfNeeded(path)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening track log: %w", err)
	}
	defer f.Close()

	var entries []history.Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		ts, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}
		cmd := strings.TrimSpace(parts[2])
		if cmd == "" {
			continue
		}
		entries = append(entries, history.Entry{
			Command:   cmd,
			Timestamp: time.Unix(ts, 0),
		})
	}
	return entries, scanner.Err()
}

// Count returns the number of entries in the track log without loading all entries.
func Count(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	defer f.Close()

	n := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Text() != "" {
			n++
		}
	}
	return n, scanner.Err()
}

// rotateIfNeeded renames track.log → track.log.1 if it exceeds maxLogSize.
func rotateIfNeeded(path string) {
	info, err := os.Stat(path)
	if err != nil || info.Size() < maxLogSize {
		return
	}
	_ = os.Rename(path, path+".1")
}
