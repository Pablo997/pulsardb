package config

import (
	"encoding/json"
	"os"
)

// Config holds all configuration for PulsarDB
type Config struct {
	HTTP    HTTPConfig    `json:"http"`
	Storage StorageConfig `json:"storage"`
}

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// StorageConfig holds storage engine configuration
type StorageConfig struct {
	DataDir        string `json:"data_dir"`
	MaxMemoryMB    int    `json:"max_memory_mb"`
	FlushInterval  int    `json:"flush_interval_seconds"`
	RetentionDays  int    `json:"retention_days"`
	CompressionOn  bool   `json:"compression_enabled"`
	WALEnabled     bool   `json:"wal_enabled"`
	WALPath        string `json:"wal_path"`
}

// Load loads configuration from file or returns defaults
func Load(path string) (*Config, error) {
	cfg := defaultConfig()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		HTTP: HTTPConfig{
			Address: "0.0.0.0",
			Port:    8080,
		},
		Storage: StorageConfig{
			DataDir:        "./data",
			MaxMemoryMB:    512,
			FlushInterval:  60,
			RetentionDays:  7,
			CompressionOn:  true,
			WALEnabled:     true,
			WALPath:        "./data/wal.log",
		},
	}
}

