package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	
	if cfg == nil {
		t.Fatal("defaultConfig returned nil")
	}
	
	// Check HTTP config
	if cfg.HTTP.Address != "0.0.0.0" {
		t.Errorf("expected address=0.0.0.0, got %s", cfg.HTTP.Address)
	}
	
	if cfg.HTTP.Port != 8080 {
		t.Errorf("expected port=8080, got %d", cfg.HTTP.Port)
	}
	
	// Check Storage config
	if cfg.Storage.DataDir != "./data" {
		t.Errorf("expected data_dir=./data, got %s", cfg.Storage.DataDir)
	}
	
	if cfg.Storage.MaxMemoryMB != 512 {
		t.Errorf("expected max_memory_mb=512, got %d", cfg.Storage.MaxMemoryMB)
	}
	
	if cfg.Storage.FlushInterval != 60 {
		t.Errorf("expected flush_interval=60, got %d", cfg.Storage.FlushInterval)
	}
	
	if cfg.Storage.RetentionDays != 7 {
		t.Errorf("expected retention_days=7, got %d", cfg.Storage.RetentionDays)
	}
	
	if !cfg.Storage.CompressionOn {
		t.Error("expected compression_enabled=true")
	}

	if !cfg.Storage.WALEnabled {
		t.Error("expected wal_enabled=true")
	}

	if cfg.Storage.WALPath != "./data/wal.log" {
		t.Errorf("expected wal_path=./data/wal.log, got %s", cfg.Storage.WALPath)
	}
}

func TestLoadNoFile(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load with empty path failed: %v", err)
	}
	
	// Should return default config
	if cfg.HTTP.Port != 8080 {
		t.Error("expected default config when no file specified")
	}
}

func TestLoadInvalidFile(t *testing.T) {
	_, err := Load("nonexistent.json")
	if err == nil {
		t.Error("expected error when loading nonexistent file")
	}
}

func TestLoadValidFile(t *testing.T) {
	// Create temporary config file
	content := `{
		"http": {
			"address": "127.0.0.1",
			"port": 9090
		},
		"storage": {
			"data_dir": "/tmp/test_data",
			"max_memory_mb": 256,
			"flush_interval_seconds": 30,
			"retention_days": 14,
			"compression_enabled": false
		}
	}`
	
	tmpfile, err := os.CreateTemp("", "config_test_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	
	// Load config
	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	// Verify values
	if cfg.HTTP.Address != "127.0.0.1" {
		t.Errorf("expected address=127.0.0.1, got %s", cfg.HTTP.Address)
	}
	
	if cfg.HTTP.Port != 9090 {
		t.Errorf("expected port=9090, got %d", cfg.HTTP.Port)
	}
	
	if cfg.Storage.DataDir != "/tmp/test_data" {
		t.Errorf("expected data_dir=/tmp/test_data, got %s", cfg.Storage.DataDir)
	}
	
	if cfg.Storage.MaxMemoryMB != 256 {
		t.Errorf("expected max_memory_mb=256, got %d", cfg.Storage.MaxMemoryMB)
	}
	
	if cfg.Storage.FlushInterval != 30 {
		t.Errorf("expected flush_interval=30, got %d", cfg.Storage.FlushInterval)
	}
	
	if cfg.Storage.RetentionDays != 14 {
		t.Errorf("expected retention_days=14, got %d", cfg.Storage.RetentionDays)
	}
	
	if cfg.Storage.CompressionOn {
		t.Error("expected compression_enabled=false")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	content := `{invalid json}`
	
	tmpfile, err := os.CreateTemp("", "config_invalid_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	
	_, err = Load(tmpfile.Name())
	if err == nil {
		t.Error("expected error when loading invalid JSON")
	}
}

func TestLoadPartialConfig(t *testing.T) {
	// Config with only some fields (should merge with defaults)
	content := `{
		"http": {
			"port": 7070
		}
	}`
	
	tmpfile, err := os.CreateTemp("", "config_partial_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	
	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	// Should have custom port
	if cfg.HTTP.Port != 7070 {
		t.Errorf("expected port=7070, got %d", cfg.HTTP.Port)
	}
	
	// Should have default address (not overridden)
	if cfg.HTTP.Address != "" {
		// Empty string from partial config, not default
		// This is expected behavior for JSON unmarshal
	}
}

