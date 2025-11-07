package storage

import (
	"fmt"
	"os"
	"sync"

	"github.com/Pablo997/pulsardb/internal/config"
)

// Engine is the main storage engine for time-series data
type Engine struct {
	config *config.StorageConfig
	mu     sync.RWMutex
	
	// In-memory buffer for recent data
	memTable *MemTable
	
	// Write-Ahead Log for durability (binary encoding)
	wal *WAL
	
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

	// Initialize WAL if enabled
	if cfg.WALEnabled {
		wal, err := NewWAL(cfg.WALPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create WAL: %w", err)
		}
		e.wal = wal

		// Recover data from WAL
		if err := e.recoverFromWAL(); err != nil {
			return nil, fmt.Errorf("failed to recover from WAL: %w", err)
		}
	}

	return e, nil
}

// recoverFromWAL replays WAL entries into memtable
func (e *Engine) recoverFromWAL() error {
	points, err := Recover(e.config.WALPath)
	if err != nil {
		return err
	}

	// Replay all points into memtable
	for _, point := range points {
		if err := e.memTable.Insert(point); err != nil {
			return fmt.Errorf("failed to insert point during recovery: %w", err)
		}
	}

	return nil
}

// Write writes a data point to the storage engine
func (e *Engine) Write(point *DataPoint) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Write to WAL first (if enabled) - binary encoding
	if e.wal != nil {
		if err := e.wal.Write(point); err != nil {
			return fmt.Errorf("WAL write failed: %w", err)
		}
	}

	// Write to memtable
	if err := e.memTable.Insert(point); err != nil {
		return err
	}

	// Flush if memtable is full (Lazy WAL strategy)
	if e.memTable.IsFull() {
		if err := e.flush(); err != nil {
			return fmt.Errorf("flush failed: %w", err)
		}
	}

	return nil
}

// flush persists memtable and truncates WAL
func (e *Engine) flush() error {
	// Flush WAL to disk
	if e.wal != nil {
		if err := e.wal.Flush(); err != nil {
			return err
		}
	}

	// TODO: Write memtable to SSTable
	
	// Clear memtable
	e.memTable.Clear()

	// Truncate WAL (data is now in SSTable)
	if e.wal != nil {
		if err := e.wal.Truncate(); err != nil {
			return err
		}
	}

	return nil
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

	// Flush remaining data
	if err := e.flush(); err != nil {
		return fmt.Errorf("final flush failed: %w", err)
	}

	// Close WAL
	if e.wal != nil {
		if err := e.wal.Close(); err != nil {
			return fmt.Errorf("WAL close failed: %w", err)
		}
	}

	// TODO: Close all SSTable file handles

	return nil
}

