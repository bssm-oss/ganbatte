package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Version   string              `mapstructure:"version"`
	Global    bool                `mapstructure:"global_scope"`
	Aliases   map[string]Alias    `mapstructure:"alias"`
	Workflows map[string]Workflow `mapstructure:"workflow"`
}

// Alias represents a shell alias
type Alias struct {
	Cmd           string            `mapstructure:"cmd"`
	Params        []string          `mapstructure:"params"`
	DefaultParams map[string]string `mapstructure:"default_params"`
	Confirm       bool              `mapstructure:"confirm"`
	Tags          []string          `mapstructure:"tags"`
}

// Workflow represents a sequence of steps
type Workflow struct {
	Description string   `mapstructure:"description"`
	Params      []string `mapstructure:"params"`
	Steps       []Step   `mapstructure:"steps"`
	Tags        []string `mapstructure:"tags"`
}

// Step represents a single step in a workflow
type Step struct {
	Run     string `mapstructure:"run"`
	OnFail  string `mapstructure:"on_fail"` // stop, continue, prompt
	Confirm bool   `mapstructure:"confirm"`
}

// LoadMeta contains metadata about the loaded configuration.
type LoadMeta struct {
	FilePath string // absolute path to the loaded config file
	Format   string // "toml", "yaml", or "json"
}

// NewViper returns a new viper instance with default configuration
func NewViper() *viper.Viper {
	return viper.New()
}

// Load 설정 파일을 로드하고 기본값을 설정합니다.
// detectGlobalConfig 와 동일한 우선순위(toml > yaml > yml > json)로 파일을 선택하므로
// 여러 형식의 파일이 동시에 존재해도 일관된 파일을 읽습니다.
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	globalConfigPath := filepath.Join(home, ".config", "ganbatte")
	configFile, format := detectGlobalConfig(globalConfigPath)

	if _, err := os.Stat(configFile); err != nil {
		return &Config{
			Version:   "0.1.0",
			Global:    true,
			Aliases:   make(map[string]Alias),
			Workflows: make(map[string]Workflow),
		}, nil
	}

	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configFile)
	v.SetConfigType(format)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if cfg.Version == "" {
		cfg.Version = "0.1.0"
	}
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]Alias)
	}
	if cfg.Workflows == nil {
		cfg.Workflows = make(map[string]Workflow)
	}

	return &cfg, nil
}

// LoadWithMeta loads config and returns metadata about the loaded file.
func LoadWithMeta() (*Config, *LoadMeta, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, fmt.Errorf("getting home directory: %w", err)
	}

	globalConfigPath := filepath.Join(home, ".config", "ganbatte")
	configFile, format := detectGlobalConfig(globalConfigPath)

	if _, err := os.Stat(configFile); err != nil {
		return &Config{
			Version:   "0.1.0",
			Global:    true,
			Aliases:   make(map[string]Alias),
			Workflows: make(map[string]Workflow),
		}, nil, nil
	}

	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configFile)
	v.SetConfigType(format)

	if err := v.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if cfg.Version == "" {
		cfg.Version = "0.1.0"
	}
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]Alias)
	}
	if cfg.Workflows == nil {
		cfg.Workflows = make(map[string]Workflow)
	}

	return &cfg, &LoadMeta{FilePath: configFile, Format: format}, nil
}

// SaveGlobal 글로벌 설정 파일에 강제로 저장합니다.
// 기존 글로벌 설정 파일의 포맷을 감지하여 동일한 포맷으로 저장합니다.
func (c *Config) SaveGlobal() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "ganbatte")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// 기존 글로벌 설정 파일 포맷 감지
	configFile, format := detectGlobalConfig(configDir)

	v := viper.New()
	v.SetConfigType(format)
	v.SetConfigFile(configFile)
	v.Set("version", c.Version)
	v.Set("global_scope", c.Global)
	v.Set("alias", c.Aliases)
	v.Set("workflow", c.Workflows)

	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// detectGlobalConfig finds the existing global config file and its format.
// Falls back to config.toml if no existing file is found.
func detectGlobalConfig(configDir string) (filePath, format string) {
	for _, ext := range []string{"toml", "yaml", "yml", "json"} {
		path := filepath.Join(configDir, "config."+ext)
		if _, err := os.Stat(path); err == nil {
			f := ext
			if f == "yml" {
				f = "yaml"
			}
			return path, f
		}
	}
	return filepath.Join(configDir, "config.toml"), "toml"
}

// Save 설정 파일을 저장합니다.
func (c *Config) Save() error {
	v := viper.New()

	// 설정 파일 경로 결정 (우선순위: 현재 디렉토리 > 홈 디렉토리)
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	// 현재 디렉토리에 .ganbatte 파일이 있으면 그것을 사용
	// 확장자 순서대로 확인: toml, yaml, yml, json
	var configFile string
	if _, err := os.Stat(".ganbatte.toml"); err == nil {
		configFile = ".ganbatte.toml"
		v.SetConfigType("toml")
	} else if _, err := os.Stat(".ganbatte.yaml"); err == nil {
		configFile = ".ganbatte.yaml"
		v.SetConfigType("yaml")
	} else if _, err := os.Stat(".ganbatte.yml"); err == nil {
		configFile = ".ganbatte.yml"
		v.SetConfigType("yaml")
	} else if _, err := os.Stat(".ganbatte.json"); err == nil {
		configFile = ".ganbatte.json"
		v.SetConfigType("json")
	} else {
		// 홈 디렉토리의 설정 파일 사용 (기본값: toml)
		configDir := filepath.Join(home, ".config", "ganbatte")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return fmt.Errorf("creating config directory: %w", err)
		}

		configFile = filepath.Join(configDir, "config.toml")
		v.SetConfigType("toml")
	}

	v.SetConfigFile(configFile)

	// 설정 값 설정
	v.Set("version", c.Version)
	v.Set("global_scope", c.Global)
	v.Set("alias", c.Aliases)
	v.Set("workflow", c.Workflows)

	// 설정 파일 쓰기
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
