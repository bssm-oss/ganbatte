package track

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogPath_DefaultsToHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	path, err := LogPath()
	require.NoError(t, err)
	home, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(home, ".local", "share", "ganbatte", "track.log"), path)
}

func TestLogPath_RespectsXDG(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/custom/data")
	path, err := LogPath()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/custom/data", "ganbatte", "track.log"), path)
}

func TestParse_ValidLog(t *testing.T) {
	f := writeLog(t, ""+
		"1700000000\t0\tgit status -sb\n"+
		"1700000060\t0\tgit log --oneline -10\n"+
		"1700000120\t1\tnpm run build\n",
	)
	entries, err := Parse(f)
	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, "git status -sb", entries[0].Command)
	assert.Equal(t, time.Unix(1700000000, 0), entries[0].Timestamp)
	assert.Equal(t, "npm run build", entries[2].Command)
}

func TestParse_SkipsMalformed(t *testing.T) {
	f := writeLog(t, ""+
		"notanumber\t0\tshould skip\n"+
		"1700000000\t0\tvalid cmd\n"+
		"missing-tab\n"+
		"\t\t\n",
	)
	entries, err := Parse(f)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "valid cmd", entries[0].Command)
}

func TestParse_SkipsEmptyCommands(t *testing.T) {
	f := writeLog(t, ""+
		"1700000000\t0\t\n"+
		"1700000001\t0\t   \n"+
		"1700000002\t0\tactual cmd\n",
	)
	entries, err := Parse(f)
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func TestParse_Nonexistent(t *testing.T) {
	entries, err := Parse("/nonexistent/path/track.log")
	assert.NoError(t, err)
	assert.Nil(t, entries)
}

func TestParse_TabInCommand(t *testing.T) {
	// SplitN(line, "\t", 3) means command can contain tabs
	f := writeLog(t, "1700000000\t0\techo hello\tworld\n")
	entries, err := Parse(f)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "echo hello\tworld", entries[0].Command)
}

func TestCount_Basic(t *testing.T) {
	f := writeLog(t, ""+
		"1700000000\t0\tcmd1\n"+
		"1700000001\t0\tcmd2\n"+
		"1700000002\t0\tcmd3\n",
	)
	n, err := Count(f)
	require.NoError(t, err)
	assert.Equal(t, 3, n)
}

func TestCount_SkipsEmptyLines(t *testing.T) {
	f := writeLog(t, "1700000000\t0\tcmd1\n\n1700000001\t0\tcmd2\n")
	n, err := Count(f)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestCount_Nonexistent(t *testing.T) {
	n, err := Count("/nonexistent/path/track.log")
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}

func writeLog(t *testing.T, content string) string {
	t.Helper()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "track.log")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}
