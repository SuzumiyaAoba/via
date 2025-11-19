package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Rule struct {
	Name       string   `yaml:"name,omitempty"`
	Extensions []string `yaml:"extensions,omitempty"`
	Regex      string   `yaml:"regex,omitempty"`
	Mime       string   `yaml:"mime,omitempty"`
	Scheme     string   `yaml:"scheme,omitempty"`
	OS         []string `yaml:"os,omitempty"`
	Background bool     `yaml:"background,omitempty"`
	Terminal   bool     `yaml:"terminal,omitempty"`
	Fallthrough bool    `yaml:"fallthrough,omitempty"`
	Command    string   `yaml:"command"`
}

type Config struct {
	Version        string            `yaml:"version"`
	DefaultCommand string            `yaml:"default_command,omitempty"`
	Default        string            `yaml:"default,omitempty"` // Shorter alias for DefaultCommand
	Aliases        map[string]string `yaml:"aliases,omitempty"`
	Rules          []Rule            `yaml:"rules"`
}

func LoadConfig(path string) (*Config, error) {
	configPath, err := GetConfigPath(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist, or maybe error?
			// For now, let's return an error so the user knows they need a config.
			return nil, fmt.Errorf("config file not found at %s", configPath)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// If 'default' is set, use it as DefaultCommand (unless DefaultCommand is already set)
	if cfg.Default != "" && cfg.DefaultCommand == "" {
		cfg.DefaultCommand = cfg.Default
	}

	return &cfg, nil
}

func GetConfigPath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}
	return filepath.Join(home, ".config", "entry", "config.yml"), nil
}

func SaveConfig(path string, cfg *Config) error {
	configPath, err := GetConfigPath(path)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
