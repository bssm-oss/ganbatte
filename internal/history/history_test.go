package history

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixturesDir() string {
	// Walk up from internal/history to find testdata
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "testdata", "fixtures")
}

// --- Zsh Parser ---

func TestZshParser(t *testing.T) {
	p := &ZshParser{}
	entries, err := p.Parse(filepath.Join(fixturesDir(), "zsh_history"))
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	// First entry
	assert.Equal(t, "git status -sb", entries[0].Command)
	assert.False(t, entries[0].Timestamp.IsZero())

	// Count git status entries
	count := 0
	for _, e := range entries {
		if e.Command == "git status -sb" {
			count++
		}
	}
	assert.Equal(t, 4, count)
}

func TestZshParser_PlainFormat(t *testing.T) {
	// Test with plain commands (no timestamps)
	tmp := t.TempDir()
	histFile := filepath.Join(tmp, "history")
	err := os.WriteFile(histFile, []byte("echo hello\nls -la\n\n"), 0644)
	require.NoError(t, err)

	p := &ZshParser{}
	entries, err := p.Parse(histFile)
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "echo hello", entries[0].Command)
}

func TestZshParser_FileNotFound(t *testing.T) {
	p := &ZshParser{}
	_, err := p.Parse("/nonexistent/file")
	assert.Error(t, err)
}

// --- Bash Parser ---

func TestBashParser(t *testing.T) {
	p := &BashParser{}
	entries, err := p.Parse(filepath.Join(fixturesDir(), "bash_history"))
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	assert.Equal(t, "git status -sb", entries[0].Command)

	// Count npm test entries
	count := 0
	for _, e := range entries {
		if e.Command == "npm test" {
			count++
		}
	}
	assert.Equal(t, 7, count)
}

func TestBashParser_SkipsComments(t *testing.T) {
	tmp := t.TempDir()
	histFile := filepath.Join(tmp, "history")
	err := os.WriteFile(histFile, []byte("#1700000001\necho hello\n#comment\n"), 0644)
	require.NoError(t, err)

	p := &BashParser{}
	entries, err := p.Parse(histFile)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "echo hello", entries[0].Command)
}

// --- Fish Parser ---

func TestFishParser(t *testing.T) {
	p := &FishParser{}
	entries, err := p.Parse(filepath.Join(fixturesDir(), "fish_history"))
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	assert.Equal(t, "git status -sb", entries[0].Command)
	assert.False(t, entries[0].Timestamp.IsZero())

	// Count docker ps entries
	count := 0
	for _, e := range entries {
		if e.Command == "docker ps" {
			count++
		}
	}
	assert.Equal(t, 5, count)
}

// --- Suggest Engine ---

func TestSuggestAliases(t *testing.T) {
	entries := makeEntries(
		"git status -sb", "git status -sb", "git status -sb",
		"git status -sb", "git status -sb", "git status -sb",
		"ls -la", "echo hello",
	)

	suggestions := Suggest(entries, map[string]string{}, SuggestOptions{
		MinFrequency:   5,
		MaxSuggestions: 10,
	})

	// Should suggest "git status -sb" as alias (6 times >= 5)
	var found bool
	for _, s := range suggestions {
		if s.Type == "alias" && s.Command == "git status -sb" {
			found = true
			assert.Contains(t, s.Reason, "6 times")
		}
	}
	assert.True(t, found, "expected alias suggestion for 'git status -sb'")
}

func TestSuggestSkipsExisting(t *testing.T) {
	entries := makeEntries(
		"git status -sb", "git status -sb", "git status -sb",
		"git status -sb", "git status -sb", "git status -sb",
	)

	existing := map[string]string{"gs": "git status -sb"}
	suggestions := Suggest(entries, existing, SuggestOptions{MinFrequency: 5})

	for _, s := range suggestions {
		if s.Type == "alias" {
			assert.NotEqual(t, "git status -sb", s.Command, "should skip existing alias")
		}
	}
}

func TestSuggestWorkflows(t *testing.T) {
	// Repeat a sequence 3+ times
	entries := makeEntries(
		"git add .", "git commit -m x", "git push",
		"git add .", "git commit -m x", "git push",
		"git add .", "git commit -m x", "git push",
		"git add .", "git commit -m x", "git push",
	)

	suggestions := Suggest(entries, map[string]string{}, SuggestOptions{
		MinFrequency: 100, // high threshold to suppress alias suggestions
		MinSequence:  3,
	})

	var found bool
	for _, s := range suggestions {
		if s.Type == "workflow" && len(s.Steps) >= 2 {
			found = true
		}
	}
	assert.True(t, found, "expected workflow suggestion from repeated sequence")
}

func TestSuggestEmpty(t *testing.T) {
	suggestions := Suggest(nil, nil, DefaultSuggestOptions())
	assert.Empty(t, suggestions)
}

func TestGenerateAliasName(t *testing.T) {
	assert.Equal(t, "gss", generateAliasName("git status -sb"))
	assert.Equal(t, "ga", generateAliasName("git add ."))
	assert.Equal(t, "e", generateAliasName("echo"))
	assert.Equal(t, "cmd", generateAliasName(""))
}

// helper to create entries from command strings
func makeEntries(cmds ...string) []Entry {
	entries := make([]Entry, len(cmds))
	for i, cmd := range cmds {
		entries[i] = Entry{Command: cmd}
	}
	return entries
}
