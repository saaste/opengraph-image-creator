package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Site            string        `yaml:"site"`
	Secret          string        `yaml:"secret"`
	DateFormat      string        `yaml:"date_format"`
	LineBreakChars  []string      `yaml:"line_break_characters"`
	MaxCache        time.Duration `yaml:"max_cache"`
	CacheDir        string        `yaml:"cache_dir"`
	JpegCompression int           `yaml:"jpeg_compression"`
}

func Load() (*AppConfig, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config.yaml: %v", err)
	}

	var appSettings AppConfig
	if err := yaml.Unmarshal(data, &appSettings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config.yaml: %v", err)
	}

	return &appSettings, nil
}
