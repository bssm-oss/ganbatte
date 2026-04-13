package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewViper(t *testing.T) {
	v := NewViper()
	assert.NotNil(t, v)
	assert.IsType(t, &viper.Viper{}, v)
}

func TestLoadDefaultConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Set HOME to temp directory
	oldHome := os.Getenv("HOME")
	err := os.Setenv("HOME", tmpDir)
	require.NoError(t, err)
	defer func() {
		if oldHome == "" {
			os.Unsetenv("HOME")
		} else {
			os.Setenv("HOME", oldHome)
		}
	}()

	// Load config (should return default since no file exists)
	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "0.1.0", cfg.Version)
	assert.True(t, cfg.Global)
	assert.Empty(t, cfg.Aliases)
	assert.Empty(t, cfg.Workflows)
}

func TestLoadTomlConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create config directory and file
	configDir := filepath.Join(tmpDir, ".config", "ganbatte")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configFile := filepath.Join(configDir, "config.toml")
	configContent := `
version = "0.2.0"
global_scope = false

[alias.test]
cmd = "echo hello"

[workflow.test]
description = "Test workflow"
steps = [
  { run = "echo test" }
]
`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set HOME to temp directory
	oldHome := os.Getenv("HOME")
	err = os.Setenv("HOME", tmpDir)
	require.NoError(t, err)
	defer func() {
		if oldHome == "" {
			os.Unsetenv("HOME")
		} else {
			os.Setenv("HOME", oldHome)
		}
	}()

	// Load config
	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "0.2.0", cfg.Version)
	assert.False(t, cfg.Global)

	// Check alias
	assert.Len(t, cfg.Aliases, 1)
	assert.Equal(t, Alias{Cmd: "echo hello"}, cfg.Aliases["test"])

	// Check workflow
	assert.Len(t, cfg.Workflows, 1)
	assert.Equal(t, "Test workflow", cfg.Workflows["test"].Description)
	assert.Len(t, cfg.Workflows["test"].Steps, 1)
	assert.Equal(t, "echo test", cfg.Workflows["test"].Steps[0].Run)
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create initial config
	configDir := filepath.Join(tmpDir, ".config", "ganbatte")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configFile := filepath.Join(configDir, "config.toml")
	initialContent := `version = "0.1.0"
global_scope = true
`
	err = os.WriteFile(configFile, []byte(initialContent), 0644)
	require.NoError(t, err)

	// Set HOME to temp directory
	oldHome := os.Getenv("HOME")
	err = os.Setenv("HOME", tmpDir)
	require.NoError(t, err)
	defer func() {
		if oldHome == "" {
			os.Unsetenv("HOME")
		} else {
			os.Setenv("HOME", oldHome)
		}
	}()

	// Load existing config
	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "0.1.0", cfg.Version)

	// Modify config
	cfg.Version = "0.3.0"
	cfg.Global = false
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]Alias)
	}
	cfg.Aliases["new-alias"] = Alias{Cmd: "new command"}

	// Save config
	err = cfg.Save()
	require.NoError(t, err)

	// Load again to verify
	cfg2, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg2)
	assert.Equal(t, "0.3.0", cfg2.Version)
	assert.False(t, cfg2.Global)
	assert.Len(t, cfg2.Aliases, 1)
	assert.Equal(t, "new command", cfg2.Aliases["new-alias"].Cmd)
}

func TestConfigFormatEquivalence(t *testing.T) {
	// Test that TOML, YAML, and JSON formats produce equivalent results
	testCases := []struct {
		name     string
		ext      string
		content  string
		expected map[string]Alias
	}{
		{
			name: "TOML",
			ext:  "toml",
			content: `[alias.test]
cmd = "test command"
`,
			expected: map[string]Alias{
				"test": {Cmd: "test command"},
			},
		},
		{
			name: "YAML",
			ext:  "yaml",
			content: `alias:
  test:
    cmd: "test command"
`,
			expected: map[string]Alias{
				"test": {Cmd: "test command"},
			},
		},
		{
			name: "JSON",
			ext:  "json",
			content: `{
  "alias": {
    "test": {
      "cmd": "test command"
    }
  }
}
`,
			expected: map[string]Alias{
				"test": {Cmd: "test command"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()

			// Create config directory and file
			configDir := filepath.Join(tmpDir, ".config", "ganbatte")
			err := os.MkdirAll(configDir, 0755)
			require.NoError(t, err)

			configFile := filepath.Join(configDir, "config."+tc.ext)
			err = os.WriteFile(configFile, []byte(tc.content), 0644)
			require.NoError(t, err)

			// Set HOME to temp directory
			oldHome := os.Getenv("HOME")
			err = os.Setenv("HOME", tmpDir)
			require.NoError(t, err)
			defer func() {
				if oldHome == "" {
					os.Unsetenv("HOME")
				} else {
					os.Setenv("HOME", oldHome)
				}
			}()

			// Load config WITH explicit config type
			v := viper.New()
			v.SetConfigFile(configFile)
			v.SetConfigType(tc.ext) // Set the config type explicitly

			// Set paths
			home, err := os.UserHomeDir()
			require.NoError(t, err)
			v.AddConfigPath(filepath.Join(home, ".config", "ganbatte"))
			v.AddConfigPath(".")

			// Read config
			err = v.ReadInConfig()
			require.NoError(t, err)

			// Unmarshal
			var cfg Config
			err = v.Unmarshal(&cfg)
			require.NoError(t, err)

			// Check that aliases match expected
			assert.Equal(t, tc.expected, cfg.Aliases, "Aliases should match for %s", tc.name)
		})
	}
}
