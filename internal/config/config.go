package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
)

type Config struct {
	Port    string `json:"port"`
	Version string `json:"version"`
}

var cfg atomic.Value

func defaultConfig() *Config {
	return &Config{
		Port:    "8080",
		Version: "0.0.1",
	}
}

func Parse(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			defaults := defaultConfig()
			if writeErr := writeConfig(path, defaults); writeErr != nil {
				return nil, fmt.Errorf("ошибка создания файла конфига %s: %w", path, writeErr)
			}
			return defaults, nil
		}
		return nil, fmt.Errorf("ошибка чтения файла конфига %s: %w", path, err)
	}

	var c Config

	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON в файле %s: %w", path, err)
	}
	return &c, nil
}

func writeConfig(path string, c *Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Load(path string) error {
	data, err := Parse(path)
	if err != nil {
		return err
	}
	Set(data)
	return nil
}

func Set(c *Config) {
	cfg.Store(c)
}

func Get() *Config {
	return cfg.Load().(*Config)
}
