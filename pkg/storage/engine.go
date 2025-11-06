package storage

import (
	"fmt"
	"os"
	"sync"

	"github.com/yourusername/pulsardb/internal/config"
)

// Engine is the main storage engine for time-series data
type Engine struct {
	config *config.StorageConfig
	mu     sync.RWMutex
	
	// In-memory buffer for recent data
	memTable *MemTable
	
	// TODO: Add WAL (Write-Ahead Log)
	// TODO: Add SSTable management
	// TODO: Add compaction
}

// NewEngine creates a new storage engine
func NewEngine(cfg *config.StorageConfig) (*Engine, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	e := &Engine{
		config:   cfg,
		memTable: NewMemTable(cfg.MaxMemoryMB),
	}

	return e, nil
}

// Write writes a data point to the storage engine
func (e *Engine) Write(point *DataPoint) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// TODO: Implement WAL write
	// TODO: Check if memtable is full and needs flushing
	
	return e.memTable.Insert(point)
}

// Query queries data points within a time range
func (e *Engine) Query(metric string, start, end int64) ([]*DataPoint, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// TODO: Query memtable and SSTables
	// TODO: Merge and sort results
	
	return e.memTable.Query(metric, start, end), nil
}

// Close closes the storage engine
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// TODO: Flush memtable to disk
	// TODO: Close WAL
	// TODO: Close all file handles

	return nil
}

