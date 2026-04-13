package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Convert reads a config file and writes it in the target format.
// Returns the path to the new file.
func Convert(srcPath, targetFormat string) (string, error) {
	targetFormat = strings.ToLower(targetFormat)
	switch targetFormat {
	case "toml", "yaml", "json":
		// valid
	default:
		return "", fmt.Errorf("unsupported target format '%s' (use toml, yaml, or json)", targetFormat)
	}

	v := viper.New()
	v.SetConfigFile(srcPath)

	if err := v.ReadInConfig(); err != nil {
		return "", fmt.Errorf("reading source config: %w", err)
	}

	// Determine output path: same directory, new extension
	dir := filepath.Dir(srcPath)
	base := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))

	ext := targetFormat
	if ext == "yaml" {
		ext = "yaml"
	}
	destPath := filepath.Join(dir, base+"."+ext)

	// Write in the target format
	v.SetConfigType(targetFormat)
	if err := v.WriteConfigAs(destPath); err != nil {
		return "", fmt.Errorf("writing converted config: %w", err)
	}

	return destPath, nil
}
