package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvert_TomlToYaml(t *testing.T) {
	src := filepath.Join("..", "..", "testdata", "fixtures", "config.toml")
	tmpDir := t.TempDir()

	// Copy source to temp dir
	data, err := os.ReadFile(src)
	require.NoError(t, err)
	srcCopy := filepath.Join(tmpDir, "config.toml")
	require.NoError(t, os.WriteFile(srcCopy, data, 0o644))

	dest, err := Convert(srcCopy, "yaml")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "config.yaml"), dest)
	assert.FileExists(t, dest)
}

func TestConvert_TomlToJson(t *testing.T) {
	src := filepath.Join("..", "..", "testdata", "fixtures", "config.toml")
	tmpDir := t.TempDir()

	data, err := os.ReadFile(src)
	require.NoError(t, err)
	srcCopy := filepath.Join(tmpDir, "config.toml")
	require.NoError(t, os.WriteFile(srcCopy, data, 0o644))

	dest, err := Convert(srcCopy, "json")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "config.json"), dest)
	assert.FileExists(t, dest)
}

func TestConvert_InvalidFormat(t *testing.T) {
	_, err := Convert("/tmp/dummy.toml", "xml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported target format")
}

func TestConvert_MissingSource(t *testing.T) {
	_, err := Convert("/nonexistent/file.toml", "yaml")
	assert.Error(t, err)
}
