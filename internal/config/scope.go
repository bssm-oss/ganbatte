package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ScopedConfig holds separate global and project configs with merge result.
type ScopedConfig struct {
	Global    *Config
	Project   *Config    // nil if no project config found
	Merged    *Config    // merged result, project overrides global
	Conflicts []Conflict // items that exist in both scopes
}

// Conflict represents a name collision between global and project scope.
type Conflict struct {
	Name       string // alias or workflow name
	Type       string // "alias" or "workflow"
	GlobalVal  string // command (alias) or description (workflow) in global
	ProjectVal string // command (alias) or description (workflow) in project
}

// LoadScoped loads global and project configs separately, then merges.
func LoadScoped() (*ScopedConfig, error) {
	global, err := loadFromScope("global")
	if err != nil {
		return nil, fmt.Errorf("loading global config: %w", err)
	}

	project, err := loadFromScope("project")
	if err != nil {
		return nil, fmt.Errorf("loading project config: %w", err)
	}

	merged, conflicts := Merge(global, project)

	return &ScopedConfig{
		Global:    global,
		Project:   project,
		Merged:    merged,
		Conflicts: conflicts,
	}, nil
}

// loadFromScope loads a config from a specific scope.
func loadFromScope(scope string) (*Config, error) {
	v := viper.New()

	switch scope {
	case "global":
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting home directory: %w", err)
		}
		configFile, format := detectGlobalConfig(filepath.Join(home, ".config", "ganbatte"))
		if _, err := os.Stat(configFile); err != nil {
			if os.IsNotExist(err) {
				return &Config{
					Version:   "0.1.0",
					Aliases:   make(map[string]Alias),
					Workflows: make(map[string]Workflow),
				}, nil
			}
			return nil, fmt.Errorf("checking global config: %w", err)
		}
		v.SetConfigFile(configFile)
		v.SetConfigType(format)
	case "project":
		path, err := findProjectConfig()
		if err != nil {
			return nil, err
		}
		if path == "" {
			// No project config found
			return nil, nil
		}
		v.SetConfigFile(path)
	default:
		return nil, fmt.Errorf("unknown scope: %s", scope)
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if scope == "global" {
				return &Config{
					Version:   "0.1.0",
					Aliases:   make(map[string]Alias),
					Workflows: make(map[string]Workflow),
				}, nil
			}
			return nil, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]Alias)
	}
	if cfg.Workflows == nil {
		cfg.Workflows = make(map[string]Workflow)
	}

	return &cfg, nil
}

// findProjectConfig walks from the current directory upward looking for .ganbatte.*.
func findProjectConfig() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	startDir := dir

	for {
		for _, ext := range []string{"toml", "yaml", "yml", "json"} {
			path := filepath.Join(dir, ".ganbatte."+ext)
			if _, err := os.Stat(path); err == nil {
				if dir != startDir && !hasVCSMarker(dir) {
					continue
				}
				if safe, err := isSafeProjectConfig(path); err != nil {
					return "", err
				} else if !safe {
					continue
				}
				return path, nil
			} else if !os.IsNotExist(err) {
				return "", fmt.Errorf("checking project config: %w", err)
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

func hasVCSMarker(dir string) bool {
	for _, marker := range []string{".git", ".hg", ".svn"} {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}

func isSafeProjectConfig(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, fmt.Errorf("checking project config safety: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false, nil
	}
	if !info.Mode().IsRegular() {
		return false, nil
	}
	if isOtherWritable(info.Mode()) {
		return false, nil
	}
	if owned, err := isOwnedByCurrentUser(path); err != nil {
		return false, err
	} else if !owned {
		return false, nil
	}

	dir := filepath.Dir(path)
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return false, fmt.Errorf("checking project config directory: %w", err)
	}
	if isOtherWritable(dirInfo.Mode()) {
		return false, nil
	}
	if owned, err := isOwnedByCurrentUser(dir); err != nil {
		return false, err
	} else if !owned {
		return false, nil
	}

	return true, nil
}
