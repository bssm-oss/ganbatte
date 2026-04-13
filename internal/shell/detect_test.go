package shell

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name  string
		shell string
		want  string
	}{
		{"zsh", "/bin/zsh", "zsh"},
		{"bash", "/bin/bash", "bash"},
		{"fish", "/usr/bin/fish", "fish"},
		{"empty", "", "unknown"},
		{"sh", "/bin/sh", "sh"},
		{"nested path", "/usr/local/bin/zsh", "zsh"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SHELL", tt.shell)
			assert.Equal(t, tt.want, Detect())
		})
	}
}

func TestHistoryPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
	}
	t.Setenv("HISTFILE", "")
	t.Setenv("XDG_DATA_HOME", "")

	t.Run("zsh default", func(t *testing.T) {
		t.Setenv("HISTFILE", "")
		path := HistoryPath("zsh")
		assert.Equal(t, filepath.Join(home, ".zsh_history"), path)
	})

	t.Run("bash default", func(t *testing.T) {
		t.Setenv("HISTFILE", "")
		path := HistoryPath("bash")
		assert.Equal(t, filepath.Join(home, ".bash_history"), path)
	})

	t.Run("fish default", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "")
		path := HistoryPath("fish")
		assert.Equal(t, filepath.Join(home, ".local", "share", "fish", "fish_history"), path)
	})

	t.Run("fish with XDG_DATA_HOME", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/custom/data")
		path := HistoryPath("fish")
		assert.Equal(t, filepath.Join("/custom/data", "fish", "fish_history"), path)
	})

	t.Run("HISTFILE override", func(t *testing.T) {
		t.Setenv("HISTFILE", "/custom/history")
		assert.Equal(t, "/custom/history", HistoryPath("zsh"))
		assert.Equal(t, "/custom/history", HistoryPath("bash"))
	})

	t.Run("unknown shell", func(t *testing.T) {
		assert.Empty(t, HistoryPath("unknown"))
		assert.Empty(t, HistoryPath(""))
	})
}
