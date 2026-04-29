package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// executeCmd sets args on RootCmd, captures output, and executes.
// Resets flags to defaults before each call to avoid cross-test pollution.
func executeCmd(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs(args)
	// Reset flags to avoid state leaking between tests
	_ = runCmd.Flags().Set("dry-run", "false")
	_ = runCmd.Flags().Set("yes", "false")
	_ = listCmd.Flags().Set("tag", "")
	_ = listCmd.Flags().Set("scope", "")
	_ = suggestCmd.Flags().Set("apply", "false")
	_ = suggestCmd.Flags().Set("min-frequency", "5")
	_ = suggestCmd.Flags().Set("min-sequence", "3")
	_ = configConvertCmd.Flags().Set("to", "")
	_ = exportCmd.Flags().Set("output", "")
	_ = exportCmd.Flags().Set("format", "toml")
	_ = importCmd.Flags().Set("replace", "false")
	_ = addCmd.Flags().Set("global", "false")
	_ = editCmd.Flags().Set("global", "false")
	_ = initCmd.Flags().Set("format", "")
	_ = initCmd.Flags().Set("project", "false")
	_ = shellInitCmd.Flags().Set("shell", "")
	err := RootCmd.Execute()
	return buf.String(), err
}

// setupTestHome sets HOME to a temp directory and returns its path.
func setupTestHome(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", tmpDir)
	}
	return tmpDir
}

// --- init ---

func TestInitCommand(t *testing.T) {
	home := setupTestHome(t)
	t.Setenv("SHELL", "/bin/zsh")

	out, err := executeCmd("init", "--format", "toml")
	require.NoError(t, err)
	assert.Contains(t, out, "Detected shell: zsh")
	assert.Contains(t, out, "Created toml config")
	assert.FileExists(t, filepath.Join(home, ".config", "ganbatte", "config.toml"))
}

func TestInitYaml(t *testing.T) {
	home := setupTestHome(t)

	out, err := executeCmd("init", "--format", "yaml")
	require.NoError(t, err)
	assert.Contains(t, out, "Created yaml config")
	assert.FileExists(t, filepath.Join(home, ".config", "ganbatte", "config.yaml"))
}

func TestInitInvalidFormat(t *testing.T) {
	setupTestHome(t)

	_, err := executeCmd("init", "--format", "xml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

// --- add ---

func TestAddAliasCommand(t *testing.T) {
	setupTestHome(t)
	_, err := executeCmd("init", "--format", "toml")
	require.NoError(t, err)

	out, err := executeCmd("add", "gs", "git status -sb")
	require.NoError(t, err)
	assert.Contains(t, out, "Added alias 'gs'")

	out, err = executeCmd("list")
	require.NoError(t, err)
	assert.Contains(t, out, "gs")
	assert.Contains(t, out, "git status -sb")
}

func TestAddDuplicateAlias(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")
	_, _ = executeCmd("add", "dup", "first command")

	_, err := executeCmd("add", "dup", "second command")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAddMissingArgs(t *testing.T) {
	setupTestHome(t)
	_, err := executeCmd("add")
	require.Error(t, err)
	_, err = executeCmd("add", "onlyname")
	require.Error(t, err)
}

// --- edit ---

func TestEditCommand(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")
	_, _ = executeCmd("add", "gs", "git status")

	out, err := executeCmd("edit", "gs", "git status -sb")
	require.NoError(t, err)
	assert.Contains(t, out, "Updated alias 'gs'")

	out, _ = executeCmd("list")
	assert.Contains(t, out, "git status -sb")
}

func TestEditNonexistent(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")

	_, err := executeCmd("edit", "nope", "something")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// --- remove ---

func TestRemoveCommand(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")
	_, _ = executeCmd("add", "rm_me", "echo bye")

	out, err := executeCmd("remove", "rm_me")
	require.NoError(t, err)
	assert.Contains(t, out, "Removed alias 'rm_me'")

	out, _ = executeCmd("list")
	assert.NotContains(t, out, "rm_me")
}

func TestRemoveNonexistent(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")

	_, err := executeCmd("remove", "ghost")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// --- run ---

func TestRunAliasDryRun(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")
	_, _ = executeCmd("add", "gs", "git status -sb")

	out, err := executeCmd("run", "gs", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "[dry-run]")
	assert.Contains(t, out, "git status -sb")
}

func TestRunAliasDestructiveDryRun(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")
	_, _ = executeCmd("add", "danger", "rm -rf /tmp/test")

	out, err := executeCmd("run", "danger", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "[DESTRUCTIVE]")
}

func TestRunNonexistent(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")

	_, err := executeCmd("run", "nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRunAliasActual(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")
	_, _ = executeCmd("add", "hi", "echo hello")

	out, err := executeCmd("run", "hi")
	require.NoError(t, err)
	assert.Contains(t, out, "Running: echo hello")
}

// --- list ---

func TestListEmpty(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")

	out, err := executeCmd("list")
	require.NoError(t, err)
	assert.Contains(t, out, "No aliases found")
	assert.Contains(t, out, "No workflows found")
}

func TestListTagFilter(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")
	_, _ = executeCmd("add", "gs", "git status")

	// Without tag filter: shows aliases
	out, _ := executeCmd("list")
	assert.Contains(t, out, "gs")

	// With tag filter: aliases hidden (no tags on aliases)
	out, _ = executeCmd("list", "--tag", "deploy")
	assert.NotContains(t, out, "gs")
}

func TestListTagFilterWorkflow(t *testing.T) {
	home := setupTestHome(t)

	// Write config with workflow + tags directly
	configDir := filepath.Join(home, ".config", "ganbatte")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(`
version = "0.1.0"
[alias.gs]
cmd = "git status"
[workflow.deploy]
description = "Deploy app"
tags = ["ci", "deploy"]
[[workflow.deploy.steps]]
run = "pnpm build"
`), 0o644))

	// Without filter: shows both
	out, err := executeCmd("list")
	require.NoError(t, err)
	assert.Contains(t, out, "gs")
	assert.Contains(t, out, "deploy")

	// With tag filter: only workflow with matching tag
	out, err = executeCmd("list", "--tag", "ci")
	require.NoError(t, err)
	assert.Contains(t, out, "deploy")
	assert.Contains(t, out, "No aliases found") // aliases hidden when tag filter active

	// With non-matching tag: no workflows
	out, err = executeCmd("list", "--tag", "nonexistent")
	require.NoError(t, err)
	assert.NotContains(t, out, "deploy")
}

func TestListScopeFilter(t *testing.T) {
	home := setupTestHome(t)
	projectDir := t.TempDir()
	t.Chdir(projectDir)

	configDir := filepath.Join(home, ".config", "ganbatte")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(`
version = "0.1.0"
[alias.global]
cmd = "echo global"
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, ".ganbatte.toml"), []byte(`
version = "0.1.0"
[alias.project]
cmd = "echo project"
`), 0o644))

	out, err := executeCmd("list", "--scope", "global")
	require.NoError(t, err)
	assert.Contains(t, out, "global")
	assert.NotContains(t, out, "project")

	out, err = executeCmd("list", "--scope", "project")
	require.NoError(t, err)
	assert.Contains(t, out, "project")
	assert.NotContains(t, out, "global")
}

func TestListScopeFilterConflictLabels(t *testing.T) {
	home := setupTestHome(t)
	projectDir := t.TempDir()
	t.Chdir(projectDir)

	configDir := filepath.Join(home, ".config", "ganbatte")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(`
version = "0.1.0"
[alias.shared]
cmd = "echo global"
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, ".ganbatte.toml"), []byte(`
version = "0.1.0"
[alias.shared]
cmd = "echo project"
`), 0o644))

	out, err := executeCmd("list", "--scope", "global")
	require.NoError(t, err)
	assert.Contains(t, out, "shared: echo global [global]")
	assert.NotContains(t, out, "[project]")

	out, err = executeCmd("list", "--scope", "project")
	require.NoError(t, err)
	assert.Contains(t, out, "shared: echo project [project]")
}

func TestProjectOverrides(t *testing.T) {
	scoped := &config.ScopedConfig{
		Global: &config.Config{
			Aliases:   map[string]config.Alias{"shared": {Cmd: "echo global"}},
			Workflows: map[string]config.Workflow{"deploy": {Description: "global"}},
		},
		Project: &config.Config{
			Aliases:   map[string]config.Alias{"shared": {Cmd: "echo project"}},
			Workflows: map[string]config.Workflow{"deploy": {Description: "project"}},
		},
	}

	assert.True(t, projectOverrides(scoped, "shared", "alias"))
	assert.True(t, projectOverrides(scoped, "deploy", "workflow"))
	assert.False(t, projectOverrides(scoped, "missing", "alias"))
}

func TestListInvalidScope(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")

	_, err := executeCmd("list", "--scope", "local")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid scope")
}

func TestShowWorkflow(t *testing.T) {
	home := setupTestHome(t)

	configDir := filepath.Join(home, ".config", "ganbatte")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(`
version = "0.1.0"
[workflow.deploy]
description = "Deploy app"
params = ["branch"]
tags = ["ci"]
[[workflow.deploy.steps]]
run = "pnpm build"
on_fail = "stop"
[[workflow.deploy.steps]]
run = "git push origin {branch}"
confirm = true
`), 0o644))

	out, err := executeCmd("show", "deploy")
	require.NoError(t, err)
	assert.Contains(t, out, "Workflow: deploy")
	assert.Contains(t, out, "Deploy app")
	assert.Contains(t, out, "pnpm build")
	assert.Contains(t, out, "on_fail: stop")
	assert.Contains(t, out, "confirm: true")
	assert.Contains(t, out, "ci")
}

func TestRunWorkflowDryRun(t *testing.T) {
	home := setupTestHome(t)

	configDir := filepath.Join(home, ".config", "ganbatte")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(`
version = "0.1.0"
[workflow.deploy]
description = "Deploy"
params = ["branch"]
[[workflow.deploy.steps]]
run = "echo building"
[[workflow.deploy.steps]]
run = "git push -f origin {branch}"
confirm = true
`), 0o644))

	out, err := executeCmd("run", "deploy", "main", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "dry-run")
	assert.Contains(t, out, "echo building")
	assert.Contains(t, out, "[DESTRUCTIVE]")
	assert.Contains(t, out, "git push -f origin main")
}

func TestRunWorkflowYesSkipsConfirm(t *testing.T) {
	home := setupTestHome(t)

	configDir := filepath.Join(home, ".config", "ganbatte")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(`
version = "0.1.0"
[workflow.deploy]
description = "Deploy"
[[workflow.deploy.steps]]
run = "printf deploy"
confirm = true
`), 0o644))

	out, err := executeCmd("run", "deploy", "--yes")
	require.NoError(t, err)
	assert.Contains(t, out, "Running workflow: Deploy")
	assert.Contains(t, out, "Step 1/1: printf deploy")
	assert.NotContains(t, out, "Run 'printf deploy'?")
}

func TestRunProjectAliasFromParentDir(t *testing.T) {
	setupTestHome(t)
	projectRoot := t.TempDir()
	nestedDir := filepath.Join(projectRoot, "sub", "dir")
	require.NoError(t, os.MkdirAll(nestedDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(projectRoot, ".ganbatte.toml"), []byte(`
version = "0.1.0"
[alias.project]
cmd = "printf project"
`), 0o644))
	t.Chdir(nestedDir)

	out, err := executeCmd("run", "project")
	require.NoError(t, err)
	assert.Contains(t, out, "Running: printf project")
}

func TestShowProjectAliasFromParentDir(t *testing.T) {
	setupTestHome(t)
	projectRoot := t.TempDir()
	nestedDir := filepath.Join(projectRoot, "sub", "dir")
	require.NoError(t, os.MkdirAll(nestedDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(projectRoot, ".ganbatte.toml"), []byte(`
version = "0.1.0"
[alias.project]
cmd = "echo project"
`), 0o644))
	t.Chdir(nestedDir)

	out, err := executeCmd("show", "project")
	require.NoError(t, err)
	assert.Contains(t, out, "Alias: project")
	assert.Contains(t, out, "Command: echo project")
}

func TestAddGlobalFlag(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")

	out, err := executeCmd("add", "--global", "gs", "git status")
	require.NoError(t, err)
	assert.Contains(t, out, "(global)")
}

func TestEditGlobalFlag(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")
	_, _ = executeCmd("add", "gs", "git status")

	out, err := executeCmd("edit", "--global", "gs", "git status -sb")
	require.NoError(t, err)
	assert.Contains(t, out, "(global)")
}

// --- show ---

func TestShowAlias(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")
	_, _ = executeCmd("add", "gs", "git status -sb")

	out, err := executeCmd("show", "gs")
	require.NoError(t, err)
	assert.Contains(t, out, "Alias: gs")
	assert.Contains(t, out, "Command: git status -sb")
}

func TestShowNonexistent(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")

	_, err := executeCmd("show", "nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// --- doctor ---

func TestDoctorCommand(t *testing.T) {
	setupTestHome(t)
	t.Setenv("SHELL", "/bin/zsh")

	// Before init: warns about missing config
	out, err := executeCmd("doctor")
	require.NoError(t, err)
	assert.Contains(t, out, "[OK] Shell: zsh")
	assert.Contains(t, out, "[WARN] No global config found")

	// After init: config OK
	_, _ = executeCmd("init", "--format", "toml")
	out, err = executeCmd("doctor")
	require.NoError(t, err)
	assert.Contains(t, out, "[OK] Global config:")
}

// --- config path ---

func TestConfigPath(t *testing.T) {
	home := setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")

	out, err := executeCmd("config", "path")
	require.NoError(t, err)
	assert.Contains(t, out, home)
	assert.Contains(t, out, "config.toml")
}

func TestConfigPathNoConfig(t *testing.T) {
	setupTestHome(t)

	out, err := executeCmd("config", "path")
	require.NoError(t, err)
	assert.Contains(t, out, "No config file found")
}

// --- config convert ---

func TestConfigConvert(t *testing.T) {
	home := setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")

	out, err := executeCmd("config", "convert", "--to", "yaml")
	require.NoError(t, err)
	assert.Contains(t, out, "Converted")
	assert.Contains(t, out, "yaml")

	// Verify yaml file was created
	yamlPath := filepath.Join(home, ".config", "ganbatte", "config.yaml")
	assert.FileExists(t, yamlPath)
}

func TestConfigConvertMissingFlag(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")

	_, err := executeCmd("config", "convert")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--to flag is required")
}

// --- export / import ---

func TestExportImport(t *testing.T) {
	home := setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")
	_, _ = executeCmd("add", "gs", "git status -sb")
	_, _ = executeCmd("add", "ll", "ls -la")

	exportPath := filepath.Join(home, "backup.toml")
	out, err := executeCmd("export", "--output", exportPath)
	require.NoError(t, err)
	assert.Contains(t, out, "Exported")
	assert.FileExists(t, exportPath)

	// Remove aliases, then import
	_, _ = executeCmd("remove", "gs")
	_, _ = executeCmd("remove", "ll")

	out, err = executeCmd("import", exportPath)
	require.NoError(t, err)
	assert.Contains(t, out, "Added")
	assert.Contains(t, out, "Config saved")

	// Verify imported
	out, _ = executeCmd("list")
	assert.Contains(t, out, "gs")
	assert.Contains(t, out, "ll")
}

func TestExportMissingOutput(t *testing.T) {
	setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")

	_, err := executeCmd("export")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--output flag is required")
}

func TestImportMergeSkipsExisting(t *testing.T) {
	home := setupTestHome(t)
	_, _ = executeCmd("init", "--format", "toml")
	_, _ = executeCmd("add", "gs", "git status")

	// Export
	exportPath := filepath.Join(home, "export.toml")
	_, _ = executeCmd("export", "--output", exportPath)

	// Import into same config (gs already exists)
	out, err := executeCmd("import", exportPath)
	require.NoError(t, err)
	assert.Contains(t, out, "Skipped")
}

// --- root ---

func TestRootCommand(t *testing.T) {
	setupTestHome(t)

	// Without config: shows empty state message
	out, err := executeCmd()
	require.NoError(t, err)
	assert.Contains(t, out, "No aliases or workflows configured")

	// With config + items: would launch TUI (can't test interactive TUI here)
}
