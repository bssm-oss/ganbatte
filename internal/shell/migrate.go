package shell

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ShellAlias represents a parsed alias from shell config.
type ShellAlias struct {
	Name    string
	Command string
	Source  string // file path where alias was found
}

// zsh/bash: alias name='command' or alias name="command" or alias name=command
var bashAliasRe = regexp.MustCompile(`^\s*alias\s+([a-zA-Z_][a-zA-Z0-9_-]*)=["']?(.+?)["']?\s*$`)

// fish: alias name 'command' or alias name "command" or alias name command
var fishAliasRe = regexp.MustCompile(`^\s*alias\s+([a-zA-Z_][a-zA-Z0-9_-]*)\s+["']?(.+?)["']?\s*$`)

// ParseShellAliases reads a shell config file and extracts alias definitions.
func ParseShellAliases(path, shellType string) ([]ShellAlias, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	var aliases []ShellAlias
	scanner := bufio.NewScanner(f)

	re := bashAliasRe
	if shellType == "fish" {
		re = fishAliasRe
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			name := matches[1]
			cmd := matches[2]
			// Remove trailing quotes that regex might have left
			cmd = strings.TrimRight(cmd, "'\"")
			if name != "" && cmd != "" {
				aliases = append(aliases, ShellAlias{
					Name:    name,
					Command: cmd,
					Source:  path,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	return aliases, nil
}

// FindShellConfigs returns candidate config files for the given shell.
func FindShellConfigs(shellName string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var candidates []string

	switch shellName {
	case "zsh":
		candidates = []string{
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".zsh_aliases"),
			filepath.Join(home, ".aliases"),
		}
	case "bash":
		candidates = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".bash_aliases"),
			filepath.Join(home, ".bash_profile"),
			filepath.Join(home, ".aliases"),
		}
	case "fish":
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(home, ".config")
		}
		candidates = []string{
			filepath.Join(configDir, "fish", "config.fish"),
		}
	}

	var existing []string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			existing = append(existing, c)
		}
	}
	return existing
}
