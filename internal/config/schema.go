package config

import "fmt"

// SchemaVersion is the current config schema version.
// Bump this when the config format changes in a breaking way.
const SchemaVersion = "1.0.0"

// SupportedVersions lists all schema versions that can be loaded.
var SupportedVersions = []string{"0.1.0", "1.0.0"}

// ValidateVersion checks if a config version is supported.
func ValidateVersion(version string) error {
	for _, v := range SupportedVersions {
		if v == version {
			return nil
		}
	}
	return fmt.Errorf("unsupported config version '%s' (supported: %v)", version, SupportedVersions)
}

// NeedsMigration returns true if the config version is older than current.
func NeedsMigration(version string) bool {
	return version != "" && version != SchemaVersion
}

// Migrate updates a config from an older version to the current version.
// Currently just bumps the version field since schema is compatible.
func Migrate(cfg *Config) *Config {
	cfg.Version = SchemaVersion
	return cfg
}
