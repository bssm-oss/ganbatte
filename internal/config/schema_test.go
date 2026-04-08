package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateVersion(t *testing.T) {
	require.NoError(t, ValidateVersion("0.1.0"))
	require.NoError(t, ValidateVersion("1.0.0"))
	assert.Error(t, ValidateVersion("99.0.0"))
	assert.Error(t, ValidateVersion(""))
}

func TestNeedsMigration(t *testing.T) {
	assert.True(t, NeedsMigration("0.1.0"))
	assert.False(t, NeedsMigration(SchemaVersion))
	assert.False(t, NeedsMigration(""))
}

func TestMigrate(t *testing.T) {
	cfg := &Config{Version: "0.1.0"}
	migrated := Migrate(cfg)
	assert.Equal(t, SchemaVersion, migrated.Version)
}
