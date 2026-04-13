package shell

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseShellAliases_Zsh(t *testing.T) {
	content := `# My aliases
alias gs='git status -sb'
alias gl="git log --oneline -20"
alias ll=ls\ -la
# alias commented='should not appear'

export PATH="/usr/local/bin:$PATH"
alias gco='git checkout'
`
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".zshrc")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	aliases, err := ParseShellAliases(path, "zsh")
	require.NoError(t, err)

	assert.Len(t, aliases, 4)
	assert.Equal(t, "gs", aliases[0].Name)
	assert.Equal(t, "git status -sb", aliases[0].Command)
	assert.Equal(t, "gl", aliases[1].Name)
	assert.Equal(t, "git log --oneline -20", aliases[1].Command)
	assert.Equal(t, "ll", aliases[2].Name)
	assert.Equal(t, "gco", aliases[3].Name)
	assert.Equal(t, "git checkout", aliases[3].Command)
}

func TestParseShellAliases_Bash(t *testing.T) {
	content := `alias gs='git status'
alias ll='ls -la'
`
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".bashrc")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	aliases, err := ParseShellAliases(path, "bash")
	require.NoError(t, err)

	assert.Len(t, aliases, 2)
	assert.Equal(t, "gs", aliases[0].Name)
	assert.Equal(t, "git status", aliases[0].Command)
}

func TestParseShellAliases_Fish(t *testing.T) {
	content := `# Fish aliases
alias gs 'git status -sb'
alias ll "ls -la"
`
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.fish")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	aliases, err := ParseShellAliases(path, "fish")
	require.NoError(t, err)

	assert.Len(t, aliases, 2)
	assert.Equal(t, "gs", aliases[0].Name)
	assert.Equal(t, "git status -sb", aliases[0].Command)
}

func TestParseShellAliases_Empty(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".zshrc")
	require.NoError(t, os.WriteFile(path, []byte("# no aliases\nexport FOO=bar\n"), 0o644))

	aliases, err := ParseShellAliases(path, "zsh")
	require.NoError(t, err)
	assert.Empty(t, aliases)
}

func TestParseShellAliases_FileNotFound(t *testing.T) {
	_, err := ParseShellAliases("/nonexistent/path", "zsh")
	assert.Error(t, err)
}

func TestFindShellConfigs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
	}

	// Create some config files
	require.NoError(t, os.WriteFile(filepath.Join(home, ".zshrc"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".aliases"), []byte(""), 0o644))

	configs := FindShellConfigs("zsh")
	assert.Contains(t, configs, filepath.Join(home, ".zshrc"))
	assert.Contains(t, configs, filepath.Join(home, ".aliases"))
}
