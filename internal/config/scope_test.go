package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerge_BothNil(t *testing.T) {
	merged, conflicts := Merge(nil, nil)
	assert.NotNil(t, merged)
	assert.Empty(t, conflicts)
	assert.Empty(t, merged.Aliases)
}

func TestMerge_GlobalOnly(t *testing.T) {
	global := &Config{
		Aliases:   map[string]Alias{"gs": {Cmd: "git status"}},
		Workflows: map[string]Workflow{},
	}
	merged, conflicts := Merge(global, nil)
	assert.Equal(t, global, merged)
	assert.Empty(t, conflicts)
}

func TestMerge_ProjectOnly(t *testing.T) {
	project := &Config{
		Aliases:   map[string]Alias{"gs": {Cmd: "git status -sb"}},
		Workflows: map[string]Workflow{},
	}
	merged, conflicts := Merge(nil, project)
	assert.Equal(t, project, merged)
	assert.Empty(t, conflicts)
}

func TestMerge_NoConflict(t *testing.T) {
	global := &Config{
		Aliases:   map[string]Alias{"gs": {Cmd: "git status"}},
		Workflows: map[string]Workflow{},
	}
	project := &Config{
		Aliases:   map[string]Alias{"ll": {Cmd: "ls -la"}},
		Workflows: map[string]Workflow{},
	}

	merged, conflicts := Merge(global, project)
	assert.Empty(t, conflicts)
	assert.Len(t, merged.Aliases, 2)
	assert.Equal(t, "git status", merged.Aliases["gs"].Cmd)
	assert.Equal(t, "ls -la", merged.Aliases["ll"].Cmd)
}

func TestMerge_AliasConflict(t *testing.T) {
	global := &Config{
		Aliases:   map[string]Alias{"gs": {Cmd: "git status"}},
		Workflows: map[string]Workflow{},
	}
	project := &Config{
		Aliases:   map[string]Alias{"gs": {Cmd: "git status -sb"}},
		Workflows: map[string]Workflow{},
	}

	merged, conflicts := Merge(global, project)
	assert.Len(t, conflicts, 1)
	assert.Equal(t, "gs", conflicts[0].Name)
	assert.Equal(t, "alias", conflicts[0].Type)
	assert.Equal(t, "git status", conflicts[0].GlobalVal)
	assert.Equal(t, "git status -sb", conflicts[0].ProjectVal)

	// Project wins
	assert.Equal(t, "git status -sb", merged.Aliases["gs"].Cmd)
}

func TestMerge_WorkflowConflict(t *testing.T) {
	global := &Config{
		Aliases: map[string]Alias{},
		Workflows: map[string]Workflow{
			"deploy": {Description: "Global deploy"},
		},
	}
	project := &Config{
		Aliases: map[string]Alias{},
		Workflows: map[string]Workflow{
			"deploy": {Description: "Project deploy"},
		},
	}

	merged, conflicts := Merge(global, project)
	assert.Len(t, conflicts, 1)
	assert.Equal(t, "workflow", conflicts[0].Type)
	assert.Equal(t, "Project deploy", merged.Workflows["deploy"].Description)
}

func TestLoadScoped_GlobalOnly(t *testing.T) {
	tmpDir := t.TempDir()
	setTestHome(t, tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "ganbatte")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(`
version = "0.1.0"
[alias.gs]
cmd = "git status"
`), 0o644))

	scoped, err := LoadScoped()
	require.NoError(t, err)
	assert.NotNil(t, scoped.Global)
	assert.Nil(t, scoped.Project)
	assert.NotNil(t, scoped.Merged)
	assert.Equal(t, "git status", scoped.Merged.Aliases["gs"].Cmd)
	assert.Empty(t, scoped.Conflicts)
}

func TestLoadScoped_GlobalFormatPriority(t *testing.T) {
	tmpDir := t.TempDir()
	setTestHome(t, tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "ganbatte")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{
  "version": "0.1.0",
  "alias": {"json_alias": {"cmd": "echo json"}}
}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(`
version = "0.1.0"
[alias.toml_alias]
cmd = "echo toml"
`), 0o644))

	scoped, err := LoadScoped()
	require.NoError(t, err)
	assert.Contains(t, scoped.Global.Aliases, "toml_alias")
	assert.NotContains(t, scoped.Global.Aliases, "json_alias")
}

func TestLoadScoped_ProjectParentDiscovery(t *testing.T) {
	tmpDir := t.TempDir()
	setTestHome(t, tmpDir)

	projectRoot := filepath.Join(tmpDir, "repo")
	nestedDir := filepath.Join(projectRoot, "apps", "cli")
	require.NoError(t, os.MkdirAll(nestedDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(projectRoot, ".ganbatte.toml"), []byte(`
version = "0.1.0"
[alias.project]
cmd = "echo project"
`), 0o644))
	t.Chdir(nestedDir)

	scoped, err := LoadScoped()
	require.NoError(t, err)
	require.NotNil(t, scoped.Project)
	assert.Equal(t, "echo project", scoped.Project.Aliases["project"].Cmd)
	assert.Equal(t, "echo project", scoped.Merged.Aliases["project"].Cmd)
}

func TestLoadScoped_SkipsUnsafeProjectConfig(t *testing.T) {
	tmpDir := t.TempDir()
	setTestHome(t, tmpDir)

	projectRoot := filepath.Join(tmpDir, "repo")
	nestedDir := filepath.Join(projectRoot, "sub")
	require.NoError(t, os.MkdirAll(nestedDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "unsafe.toml"), []byte(`
version = "0.1.0"
[alias.unsafe]
cmd = "echo unsafe"
`), 0o644))
	require.NoError(t, os.Symlink(filepath.Join(tmpDir, "unsafe.toml"), filepath.Join(projectRoot, ".ganbatte.toml")))
	t.Chdir(nestedDir)

	scoped, err := LoadScoped()
	require.NoError(t, err)
	assert.Nil(t, scoped.Project)
}

func TestLoadScoped_SkipsWritableProjectConfig(t *testing.T) {
	tmpDir := t.TempDir()
	setTestHome(t, tmpDir)

	projectRoot := filepath.Join(tmpDir, "repo")
	nestedDir := filepath.Join(projectRoot, "sub")
	require.NoError(t, os.MkdirAll(nestedDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, ".git"), 0o755))
	configPath := filepath.Join(projectRoot, ".ganbatte.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
version = "0.1.0"
[alias.unsafe]
cmd = "echo unsafe"
`), 0o666))
	require.NoError(t, os.Chmod(configPath, 0o666))
	t.Chdir(nestedDir)

	scoped, err := LoadScoped()
	require.NoError(t, err)
	assert.Nil(t, scoped.Project)
}

func TestLoadScoped_NoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	setTestHome(t, tmpDir)

	scoped, err := LoadScoped()
	require.NoError(t, err)
	assert.NotNil(t, scoped.Global)
	assert.Nil(t, scoped.Project)
	assert.Empty(t, scoped.Merged.Aliases)
}

func TestSaveGlobal(t *testing.T) {
	tmpDir := t.TempDir()
	setTestHome(t, tmpDir)

	cfg := &Config{
		Version:   "0.1.0",
		Aliases:   map[string]Alias{"gs": {Cmd: "git status -sb"}},
		Workflows: map[string]Workflow{},
	}
	require.NoError(t, cfg.SaveGlobal())

	configFile := filepath.Join(tmpDir, ".config", "ganbatte", "config.toml")
	assert.FileExists(t, configFile)

	// Reload and verify
	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "git status -sb", loaded.Aliases["gs"].Cmd)
}

func TestSave_DefaultPath(t *testing.T) {
	tmpDir := t.TempDir()
	setTestHome(t, tmpDir)

	// Create config dir first
	configDir := filepath.Join(tmpDir, ".config", "ganbatte")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	cfg := &Config{
		Version:   "0.1.0",
		Aliases:   map[string]Alias{"ll": {Cmd: "ls -la"}},
		Workflows: map[string]Workflow{},
	}
	require.NoError(t, cfg.Save())

	configFile := filepath.Join(configDir, "config.toml")
	assert.FileExists(t, configFile)
}

func TestLoad_NoConfigReturnsDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	setTestHome(t, tmpDir)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "0.1.0", cfg.Version)
	assert.Empty(t, cfg.Aliases)
	assert.Empty(t, cfg.Workflows)
}

func TestLoad_WithTomlFixture(t *testing.T) {
	tmpDir := t.TempDir()
	setTestHome(t, tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "ganbatte")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	src, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", "config.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.toml"), src, 0o644))

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Aliases)
	assert.NotEmpty(t, cfg.Workflows)
}

func TestMerge_Mixed(t *testing.T) {
	global := &Config{
		Aliases:   map[string]Alias{"gs": {Cmd: "git status"}, "global_only": {Cmd: "echo global"}},
		Workflows: map[string]Workflow{"deploy": {Description: "Global deploy"}},
	}
	project := &Config{
		Aliases:   map[string]Alias{"gs": {Cmd: "git status -sb"}, "proj_only": {Cmd: "echo project"}},
		Workflows: map[string]Workflow{"test": {Description: "Project test"}},
	}

	merged, conflicts := Merge(global, project)
	assert.Len(t, conflicts, 1) // gs conflict only
	assert.Len(t, merged.Aliases, 3)
	assert.Len(t, merged.Workflows, 2)
}
