package shell

import (
	"os"
	"path/filepath"
)

// Detect returns the current shell name (zsh, bash, fish, etc.).
// Returns "unknown" if the shell cannot be determined.
func Detect() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "unknown"
	}
	return filepath.Base(shell)
}

// HistoryPath returns the default history file path for the given shell.
// Returns empty string if the shell is not supported.
func HistoryPath(shellName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch shellName {
	case "zsh":
		if histFile := os.Getenv("HISTFILE"); histFile != "" {
			return histFile
		}
		return filepath.Join(home, ".zsh_history")
	case "bash":
		if histFile := os.Getenv("HISTFILE"); histFile != "" {
			return histFile
		}
		return filepath.Join(home, ".bash_history")
	case "fish":
		dataDir := os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			dataDir = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(dataDir, "fish", "fish_history")
	default:
		return ""
	}
}
