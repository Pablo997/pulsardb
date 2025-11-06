package storage

import (
	"sync"
)

// MemTable is an in-memory buffer for recent writes
type MemTable struct {
	maxSizeMB int
	mu        sync.RWMutex
	data      map[string][]*DataPoint // metric -> []DataPoint
	size      int64                   // approximate size in bytes
}

// NewMemTable creates a new memtable
func NewMemTable(maxSizeMB int) *MemTable {
	return &MemTable{
		maxSizeMB: maxSizeMB,
		data:      make(map[string][]*DataPoint),
	}
}

// Insert inserts a data point into the memtable
func (mt *MemTable) Insert(point *DataPoint) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	metric := point.Metric
	mt.data[metric] = append(mt.data[metric], point)
	
	// Rough size estimation (will improve later)
	mt.size += 64 // approximate bytes per point

	return nil
}

// Query queries the memtable for data points
func (mt *MemTable) Query(metric string, start, end int64) []*DataPoint {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	points, exists := mt.data[metric]
	if !exists {
		return nil
	}

	// Filter by time range
	result := make([]*DataPoint, 0)
	for _, point := range points {
		if point.Timestamp >= start && point.Timestamp <= end {
			result = append(result, point)
		}
	}

	return result
}

// IsFull returns true if the memtable should be flushed
func (mt *MemTable) IsFull() bool {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	
	return mt.size >= int64(mt.maxSizeMB*1024*1024)
}

// Clear clears the memtable
func (mt *MemTable) Clear() {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	
	mt.data = make(map[string][]*DataPoint)
	mt.size = 0
}

