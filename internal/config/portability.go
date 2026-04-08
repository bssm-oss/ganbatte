package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// ExportOptions configures export behavior.
type ExportOptions struct {
	Names  []string // empty = export all
	Format string   // "toml", "yaml", "json"
}

// Export serializes selected aliases/workflows into a config file content.
func Export(cfg *Config, opts ExportOptions) ([]byte, error) {
	format := strings.ToLower(opts.Format)
	if format == "" {
		format = "toml"
	}

	// Build subset config
	subset := &Config{
		Version:   cfg.Version,
		Aliases:   make(map[string]Alias),
		Workflows: make(map[string]Workflow),
	}

	if len(opts.Names) == 0 {
		subset.Aliases = cfg.Aliases
		subset.Workflows = cfg.Workflows
	} else {
		for _, name := range opts.Names {
			if alias, ok := cfg.Aliases[name]; ok {
				subset.Aliases[name] = alias
			}
			if wf, ok := cfg.Workflows[name]; ok {
				subset.Workflows[name] = wf
			}
		}
	}

	v := viper.New()
	v.SetConfigType(format)
	v.Set("version", subset.Version)
	v.Set("alias", subset.Aliases)
	v.Set("workflow", subset.Workflows)

	tmpFile := fmt.Sprintf("/tmp/gnb-export-%d.%s", os.Getpid(), format)
	v.SetConfigFile(tmpFile)
	if err := v.WriteConfigAs(tmpFile); err != nil {
		return nil, fmt.Errorf("serializing config: %w", err)
	}
	defer os.Remove(tmpFile)

	return os.ReadFile(tmpFile)
}

// ImportResult contains the result of an import operation.
type ImportResult struct {
	Added    []string
	Skipped  []string
	Replaced []string
}

// Import reads a config file and merges items into the target config.
// strategy: "merge" (skip conflicts) or "replace" (overwrite conflicts).
func Import(target *Config, srcPath string, strategy string) (*ImportResult, error) {
	v := viper.New()
	v.SetConfigFile(srcPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading import file: %w", err)
	}

	var src Config
	if err := v.Unmarshal(&src); err != nil {
		return nil, fmt.Errorf("unmarshaling import file: %w", err)
	}

	result := &ImportResult{}

	if src.Aliases == nil {
		src.Aliases = make(map[string]Alias)
	}
	for name, alias := range src.Aliases {
		if _, exists := target.Aliases[name]; exists {
			if strategy == "replace" {
				target.Aliases[name] = alias
				result.Replaced = append(result.Replaced, name)
			} else {
				result.Skipped = append(result.Skipped, name)
			}
		} else {
			if target.Aliases == nil {
				target.Aliases = make(map[string]Alias)
			}
			target.Aliases[name] = alias
			result.Added = append(result.Added, name)
		}
	}

	if src.Workflows == nil {
		src.Workflows = make(map[string]Workflow)
	}
	for name, wf := range src.Workflows {
		if _, exists := target.Workflows[name]; exists {
			if strategy == "replace" {
				target.Workflows[name] = wf
				result.Replaced = append(result.Replaced, name)
			} else {
				result.Skipped = append(result.Skipped, name)
			}
		} else {
			if target.Workflows == nil {
				target.Workflows = make(map[string]Workflow)
			}
			target.Workflows[name] = wf
			result.Added = append(result.Added, name)
		}
	}

	return result, nil
}
