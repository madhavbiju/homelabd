package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Docker   DockerConfig   `yaml:"docker"`
	Security SecurityConfig `yaml:"security"`
	Storage  StorageConfig  `yaml:"storage"`
}

type ServerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	LogLevel string `yaml:"log_level"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type DockerConfig struct {
	Enabled bool   `yaml:"enabled"`
	Socket  string `yaml:"socket"`
}

type SecurityConfig struct {
	AuthEnabled bool `yaml:"auth_enabled"`
}

type StorageConfig struct {
	Paths []string `yaml:"paths"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:     "0.0.0.0",
			Port:     8080,
			LogLevel: "info",
		},
		Database: DatabaseConfig{
			Path: "homelabd.db",
		},
		Docker: DockerConfig{
			Enabled: true,
			Socket:  "/var/run/docker.sock",
		},
		Security: SecurityConfig{
			AuthEnabled: true,
		},
		Storage: StorageConfig{
			Paths: []string{"/"},
		},
	}
}

func Load(path string) (*Config, error) {
	config := DefaultConfig()

	if path == "" {
		return config, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}
