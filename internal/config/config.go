package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Rule struct {
	Extensions []string `yaml:"extensions,omitempty"`
	Regex      string   `yaml:"regex,omitempty"`
	Mime       string   `yaml:"mime,omitempty"`
	OS         []string `yaml:"os,omitempty"`
	Command    string   `yaml:"command"`
}

type Config struct {
	Version        string            `yaml:"version"`
	DefaultCommand string            `yaml:"default_command,omitempty"`
	Aliases        map[string]string `yaml:"aliases,omitempty"`
	Rules          []Rule            `yaml:"rules"`
}

func LoadConfig(path string) (*Config, error) {
	var configPath string
	if path != "" {
		configPath = path
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home dir: %w", err)
		}
		configPath = filepath.Join(home, ".config", "entry", "config.yml")
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

	return &cfg, nil
}
