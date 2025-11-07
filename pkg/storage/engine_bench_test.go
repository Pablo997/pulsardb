package storage

import (
	"fmt"
	"testing"

	"github.com/Pablo997/pulsardb/internal/config"
)

// Benchmark Engine writes with WAL enabled (binary encoding)
func BenchmarkEngineWriteWithWAL(b *testing.B) {
	cfg := &config.StorageConfig{
		DataDir:     b.TempDir(),
		MaxMemoryMB: 512,
		WALEnabled:  true,
		WALPath:     b.TempDir() + "/wal.log",
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		b.Fatalf("NewEngine failed: %v", err)
	}
	defer engine.Close()

	point := &DataPoint{
		Metric:    "test.metric",
		Timestamp: 1234567890,
		Value:     42.0,
		Tags:      map[string]string{"host": "server1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := engine.Write(point); err != nil {
			b.Fatalf("Write failed: %v", err)
		}
	}
}

// Benchmark Engine writes without WAL
func BenchmarkEngineWriteNoWAL(b *testing.B) {
	cfg := &config.StorageConfig{
		DataDir:     b.TempDir(),
		MaxMemoryMB: 512,
		WALEnabled:  false,
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		b.Fatalf("NewEngine failed: %v", err)
	}
	defer engine.Close()

	point := &DataPoint{
		Metric:    "test.metric",
		Timestamp: 1234567890,
		Value:     42.0,
		Tags:      map[string]string{"host": "server1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := engine.Write(point); err != nil {
			b.Fatalf("Write failed: %v", err)
		}
	}
}

// Benchmark batch writes with WAL (binary encoding)
func BenchmarkEngineWriteBatchWithWAL(b *testing.B) {
	batchSizes := []int{10, 100, 1000, 10000}

	for _, size := range batchSizes {
		b.Run(fmt.Sprintf("Batch%d", size), func(b *testing.B) {
			cfg := &config.StorageConfig{
				DataDir:     b.TempDir(),
				MaxMemoryMB: 512,
				WALEnabled:  true,
				WALPath:     b.TempDir() + "/wal.log",
			}

			engine, err := NewEngine(cfg)
			if err != nil {
				b.Fatalf("NewEngine failed: %v", err)
			}
			defer engine.Close()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < size; j++ {
					point := &DataPoint{
						Metric:    "test.metric",
						Timestamp: int64(i*size + j),
						Value:     float64(j),
					}
					if err := engine.Write(point); err != nil {
						b.Fatalf("Write failed: %v", err)
					}
				}
			}
		})
	}
}

// Benchmark recovery from WAL
func BenchmarkEngineRecovery(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dPoints", size), func(b *testing.B) {
			tmpDir := b.TempDir()
			cfg := &config.StorageConfig{
				DataDir:     tmpDir,
				MaxMemoryMB: 512,
				WALEnabled:  true,
				WALPath:     tmpDir + "/wal.log",
			}

			// Create engine and write data
			engine, err := NewEngine(cfg)
			if err != nil {
				b.Fatalf("NewEngine failed: %v", err)
			}

			for i := 0; i < size; i++ {
				point := &DataPoint{
					Metric:    "test.metric",
					Timestamp: int64(i),
					Value:     float64(i),
				}
				engine.Write(point)
			}
			
			// Flush WAL to disk
			if engine.wal != nil {
				engine.wal.Flush()
			}
			engine.Close()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Create new engine (triggers recovery)
				engine, err := NewEngine(cfg)
				if err != nil {
					b.Fatalf("NewEngine failed: %v", err)
				}
				engine.Close()
			}
		})
	}
}

// Benchmark concurrent writes with WAL (binary encoding)
func BenchmarkEngineConcurrentWritesWithWAL(b *testing.B) {
	cfg := &config.StorageConfig{
		DataDir:     b.TempDir(),
		MaxMemoryMB: 512,
		WALEnabled:  true,
		WALPath:     b.TempDir() + "/wal.log",
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		b.Fatalf("NewEngine failed: %v", err)
	}
	defer engine.Close()

	b.RunParallel(func(pb *testing.PB) {
		point := &DataPoint{
			Metric:    "test.metric",
			Timestamp: 1234567890,
			Value:     42.0,
		}

		for pb.Next() {
			if err := engine.Write(point); err != nil {
				b.Fatalf("Write failed: %v", err)
			}
		}
	})
}

// Benchmark concurrent writes without WAL
func BenchmarkEngineConcurrentWritesNoWAL(b *testing.B) {
	cfg := &config.StorageConfig{
		DataDir:     b.TempDir(),
		MaxMemoryMB: 512,
		WALEnabled:  false,
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		b.Fatalf("NewEngine failed: %v", err)
	}
	defer engine.Close()

	b.RunParallel(func(pb *testing.PB) {
		point := &DataPoint{
			Metric:    "test.metric",
			Timestamp: 1234567890,
			Value:     42.0,
		}

		for pb.Next() {
			if err := engine.Write(point); err != nil {
				b.Fatalf("Write failed: %v", err)
			}
		}
	})
}

// Benchmark memtable flush with WAL (binary encoding)
func BenchmarkEngineFlushWithWAL(b *testing.B) {
	cfg := &config.StorageConfig{
		DataDir:     b.TempDir(),
		MaxMemoryMB: 1, // Small memtable to trigger flushes
		WALEnabled:  true,
		WALPath:     b.TempDir() + "/wal.log",
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		b.Fatalf("NewEngine failed: %v", err)
	}
	defer engine.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Write until flush
		for j := 0; j < 1000; j++ {
			point := &DataPoint{
				Metric:    "test.metric",
				Timestamp: int64(i*1000 + j),
				Value:     float64(j),
			}
			engine.Write(point)
		}
	}
}

