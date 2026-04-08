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
		v.SetConfigName("config")
		v.AddConfigPath(filepath.Join(home, ".config", "ganbatte"))
	case "project":
		// Look for .ganbatte.* in current directory
		for _, ext := range []string{"toml", "yaml", "yml", "json"} {
			name := ".ganbatte." + ext
			if _, err := os.Stat(name); err == nil {
				v.SetConfigFile(name)
				break
			}
		}
		if v.ConfigFileUsed() == "" {
			// No project config found
			return nil, nil
		}
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
