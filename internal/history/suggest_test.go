package history

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- isNoise ---

func TestIsNoise(t *testing.T) {
	tests := []struct {
		cmd   string
		noise bool
	}{
		{"ls", true},          // too short (< 4)
		{"cd ~", false},       // 4 chars exactly
		{"# comment", true},   // starts with #
		{"git status \\", true}, // ends with backslash (continuation)
		{"}{\\", true},        // all punctuation
		{"git status -sb", false},
		{"npm run build", false},
		{"   ", true}, // short after we check len(cmd) < 4 (spaces still < 4 meaningful but len("   ")==3)
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			assert.Equal(t, tt.noise, isNoise(tt.cmd))
		})
	}
}

// --- isArgLike ---

func TestIsArgLike(t *testing.T) {
	// Should be arg-like (true)
	assert.True(t, isArgLike("github.com/user/repo"))    // contains dot
	assert.True(t, isArgLike("./src/main.go"))           // contains /
	assert.True(t, isArgLike("user@host.com"))           // contains @
	assert.True(t, isArgLike("my-long-package-name"))    // len > 10
	assert.True(t, isArgLike("MyPackage"))               // uppercase
	assert.True(t, isArgLike("node123"))                 // has digit
	assert.True(t, isArgLike("http://example.com"))      // URL

	// Should NOT be arg-like (subcommand-like)
	assert.False(t, isArgLike("add"))
	assert.False(t, isArgLike("remove"))
	assert.False(t, isArgLike("list"))
	assert.False(t, isArgLike("status"))
	assert.False(t, isArgLike("push"))
}

// --- hasUnmatchedQuote ---

func TestHasUnmatchedQuote(t *testing.T) {
	assert.False(t, hasUnmatchedQuote(`git commit -m "fix bug"`))
	assert.False(t, hasUnmatchedQuote(`echo 'hello'`))
	assert.False(t, hasUnmatchedQuote("git status"))
	assert.True(t, hasUnmatchedQuote(`echo "unclosed`))
	assert.True(t, hasUnmatchedQuote(`echo 'unclosed`))
	assert.True(t, hasUnmatchedQuote(`echo "a" "b`))
}

// --- isDestructive ---

func TestIsDestructive(t *testing.T) {
	assert.True(t, isDestructive("rm -rf /tmp/foo"))
	assert.True(t, isDestructive("kill -9 1234"))
	assert.True(t, isDestructive("git reset --hard HEAD~1"))
	assert.True(t, isDestructive("git clean -fd"))
	assert.True(t, isDestructive("git push --force origin main"))
	assert.True(t, isDestructive("pkill node"))
	assert.False(t, isDestructive("git status"))
	assert.False(t, isDestructive("ls -la"))
	assert.False(t, isDestructive("npm run build"))
}

// --- safeAliasName ---

func TestSafeAliasName(t *testing.T) {
	used := map[string]bool{}

	// Basic: initials (2 tokens → 2-char name, not in universalAliases)
	assert.Equal(t, "nr", safeAliasName("npm run", used))

	// Avoids universal aliases
	// "gc" is universally claimed
	result := safeAliasName("git commit -m", used)
	assert.NotEqual(t, "gc", result)
	assert.NotEmpty(t, result)

	// Avoids already-used names
	used2 := map[string]bool{"nr": true}
	result2 := safeAliasName("npm run", used2)
	assert.NotEqual(t, "nr", result2)
}

func TestSafeAliasName_ReturnsEmptyWhenAllTaken(t *testing.T) {
	// All candidates for "git clone" are taken
	used := map[string]bool{
		"gc": true, "gcl": true, "gclo": true,
	}
	// "gc" is also in universalAliases, so candidates are: gc (universal, skip), gcl, gclo
	result := safeAliasName("git clone", used)
	assert.Empty(t, result)
}

// --- paramNameForPrefix ---

func TestParamNameForPrefix(t *testing.T) {
	tests := []struct {
		prefix string
		want   string
	}{
		{"git clone", "repo"},
		{"git checkout", "branch"},
		{"git push origin", "branch"},
		{"npm run", "script"},
		{"npm i -g", "package"},
		{"brew install", "package"},
		{"brew install --cask", "cask"},
		{"rm -rf", "path"},
		{"mkdir -p", "dir"},
		{"docker run", "image"},
		{"docker exec", "container"},
		{"kubectl get", "resource"},
		{"kubectl apply", "file"},
		{"ssh", "host"},
		{"curl", "url"},
		{"unknown prefix xyz", "arg"}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			assert.Equal(t, tt.want, paramNameForPrefix(tt.prefix))
		})
	}
}

// --- detectParamPatterns quality gates ---

func TestDetectParamPatterns_Gate1_SubcommandRejected(t *testing.T) {
	// git remote add, git remote remove, git remote list — tokens AFTER varying one
	entries := entriesFromCmds(
		"git remote add origin", "git remote add upstream", "git remote add backup",
		"git remote remove origin", "git remote remove upstream", "git remote remove backup",
		"git remote list",
	)
	// Multiply entries to meet minCount
	all := repeatEntries(entries, 5)
	sug := detectParamPatterns(all, map[string]string{}, 3)

	// Should NOT suggest "git remote" as a param pattern because tokens follow the varying one
	for _, s := range sug {
		assert.False(t, strings.Contains(s.Command, "git remote {"), "got unexpected: %s", s.Command)
	}
}

func TestDetectParamPatterns_Gate2_ArgLikeRequired(t *testing.T) {
	// Commands where varying token is subcommand-like (short, plain, no path signals)
	entries := entriesFromCmds(
		"docker network create", "docker network inspect", "docker network remove",
		"docker network list", "docker network connect", "docker network disconnect",
	)
	all := repeatEntries(entries, 5)
	sug := detectParamPatterns(all, map[string]string{}, 3)
	for _, s := range sug {
		assert.False(t, strings.Contains(s.Command, "docker network {"))
	}
}

func TestDetectParamPatterns_ValidPattern(t *testing.T) {
	// git clone with real repo paths — should produce a suggestion
	entries := entriesFromCmds(
		"git clone https://github.com/user/repo-a.git",
		"git clone https://github.com/user/repo-b.git",
		"git clone https://github.com/other/project.git",
		"git clone git@github.com:foo/bar.git",
	)
	all := repeatEntries(entries, 4) // each appears 4 times
	sug := detectParamPatterns(all, map[string]string{}, 3)

	found := false
	for _, s := range sug {
		if strings.HasPrefix(s.Command, "git clone") {
			found = true
			assert.Equal(t, "repo", s.Params[0])
		}
	}
	assert.True(t, found, "expected git clone param suggestion")
}

func TestDetectParamPatterns_SkipsExisting(t *testing.T) {
	entries := entriesFromCmds(
		"git clone https://github.com/a/b.git",
		"git clone https://github.com/c/d.git",
		"git clone https://github.com/e/f.git",
		"git clone https://github.com/g/h.git",
	)
	all := repeatEntries(entries, 4)
	existing := map[string]string{"gcl": "git clone {repo}"}
	sug := detectParamPatterns(all, existing, 3)
	for _, s := range sug {
		assert.NotEqual(t, "gcl", s.Name)
	}
}

// --- Suggest integration ---

func TestSuggest_NoiseFiltered(t *testing.T) {
	entries := entriesFromCmds("\\", "}{", "#comment", "ls")
	for i := 0; i < 10; i++ {
		entries = append(entries, entries...)
	}
	opts := SuggestOptions{MinFrequency: 3, MinSequence: 3, MaxSuggestions: 20}
	sug := Suggest(entries, map[string]string{}, opts)
	for _, s := range sug {
		assert.False(t, isNoise(s.Command), "noise command in suggestions: %s", s.Command)
	}
}

func TestSuggest_SortedByImpact(t *testing.T) {
	// high-keystroke command appears 5x, low-keystroke appears 10x
	var entries []Entry
	for i := 0; i < 5; i++ {
		entries = append(entries, Entry{Command: "a-very-long-command-name"})
	}
	for i := 0; i < 10; i++ {
		entries = append(entries, Entry{Command: "short-cmd"})
	}
	opts := SuggestOptions{MinFrequency: 3, MinSequence: 3, MaxSuggestions: 20}
	sug := Suggest(entries, map[string]string{}, opts)
	require.GreaterOrEqual(t, len(sug), 2)
	assert.GreaterOrEqual(t, sug[0].SavedKeystrokes, sug[1].SavedKeystrokes)
}

func TestSuggest_DestructiveGetsConfirm(t *testing.T) {
	var entries []Entry
	for i := 0; i < 10; i++ {
		entries = append(entries, Entry{Command: "rm -rf /tmp/build"})
	}
	opts := SuggestOptions{MinFrequency: 5, MinSequence: 3, MaxSuggestions: 20}
	sug := Suggest(entries, map[string]string{}, opts)
	require.NotEmpty(t, sug)
	assert.True(t, sug[0].Confirm)
}

func TestSuggest_UniversalAliasNotSuggested(t *testing.T) {
	// "git status" would naively map to "gs" — a universally claimed name
	var entries []Entry
	for i := 0; i < 20; i++ {
		entries = append(entries, Entry{Command: "git status"})
	}
	opts := SuggestOptions{MinFrequency: 5, MinSequence: 3, MaxSuggestions: 20}
	sug := Suggest(entries, map[string]string{}, opts)
	for _, s := range sug {
		assert.NotEqual(t, "gs", s.Name, "should not suggest universal alias 'gs'")
		assert.NotEqual(t, "gc", s.Name)
	}
}

// helpers

func entriesFromCmds(cmds ...string) []Entry {
	entries := make([]Entry, len(cmds))
	for i, cmd := range cmds {
		entries[i] = Entry{Command: cmd}
	}
	return entries
}

func repeatEntries(entries []Entry, n int) []Entry {
	out := make([]Entry, 0, len(entries)*n)
	for i := 0; i < n; i++ {
		out = append(out, entries...)
	}
	return out
}
