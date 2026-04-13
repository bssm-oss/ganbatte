package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExport_All(t *testing.T) {
	cfg := &Config{
		Version: "0.1.0",
		Aliases: map[string]Alias{
			"gs": {Cmd: "git status -sb"},
			"ll": {Cmd: "ls -la"},
		},
		Workflows: map[string]Workflow{},
	}

	data, err := Export(cfg, ExportOptions{Format: "toml"})
	require.NoError(t, err)
	assert.Contains(t, string(data), "gs")
	assert.Contains(t, string(data), "git status")
}

func TestExport_Subset(t *testing.T) {
	cfg := &Config{
		Version: "0.1.0",
		Aliases: map[string]Alias{
			"gs": {Cmd: "git status"},
			"ll": {Cmd: "ls -la"},
		},
		Workflows: map[string]Workflow{},
	}

	data, err := Export(cfg, ExportOptions{
		Names:  []string{"gs"},
		Format: "yaml",
	})
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "gs")
	// ll should not be in the export since we only asked for gs
}

func TestImport_Merge(t *testing.T) {
	// Create source file
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "import.toml")
	err := os.WriteFile(srcFile, []byte(`
version = "0.1.0"
[alias.newone]
cmd = "echo new"
[alias.existing]
cmd = "echo imported"
`), 0644)
	require.NoError(t, err)

	target := &Config{
		Version: "0.1.0",
		Aliases: map[string]Alias{
			"existing": {Cmd: "echo original"},
		},
		Workflows: make(map[string]Workflow),
	}

	result, err := Import(target, srcFile, "merge")
	require.NoError(t, err)
	assert.Contains(t, result.Added, "newone")
	assert.Contains(t, result.Skipped, "existing")
	assert.Equal(t, "echo original", target.Aliases["existing"].Cmd) // not overwritten
	assert.Equal(t, "echo new", target.Aliases["newone"].Cmd)
}

func TestImport_Replace(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "import.toml")
	err := os.WriteFile(srcFile, []byte(`
version = "0.1.0"
[alias.existing]
cmd = "echo replaced"
`), 0644)
	require.NoError(t, err)

	target := &Config{
		Version:   "0.1.0",
		Aliases:   map[string]Alias{"existing": {Cmd: "echo original"}},
		Workflows: make(map[string]Workflow),
	}

	result, err := Import(target, srcFile, "replace")
	require.NoError(t, err)
	assert.Contains(t, result.Replaced, "existing")
	assert.Equal(t, "echo replaced", target.Aliases["existing"].Cmd)
}

func TestImport_FileNotFound(t *testing.T) {
	target := &Config{
		Aliases:   make(map[string]Alias),
		Workflows: make(map[string]Workflow),
	}
	_, err := Import(target, "/nonexistent/file.toml", "merge")
	assert.Error(t, err)
}

func TestLoadWithMeta(t *testing.T) {
	tmpDir := t.TempDir()
	setTestHome(t, tmpDir)

	// No config → meta is nil
	cfg, meta, err := LoadWithMeta()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Nil(t, meta)

	// Create config
	configDir := filepath.Join(tmpDir, ".config", "ganbatte")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	configFile := filepath.Join(configDir, "config.toml")
	require.NoError(t, os.WriteFile(configFile, []byte(`version = "0.1.0"`), 0644))

	cfg, meta, err = LoadWithMeta()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, meta)
	assert.Equal(t, configFile, meta.FilePath)
	assert.Equal(t, "toml", meta.Format)
}
