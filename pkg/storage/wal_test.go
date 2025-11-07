package storage

import (
	"path/filepath"
	"testing"
)

func TestNewWAL(t *testing.T) {
	tmpDir := t.TempDir()
	walPath := filepath.Join(tmpDir, "test.wal")

	wal, err := NewWAL(walPath)
	if err != nil {
		t.Fatalf("NewWAL failed: %v", err)
	}
	defer wal.Close()

	if wal.path != walPath {
		t.Errorf("Expected path %s, got %s", walPath, wal.path)
	}
}

func TestWALWrite(t *testing.T) {
	tmpDir := t.TempDir()
	walPath := filepath.Join(tmpDir, "test.wal")

	wal, err := NewWAL(walPath)
	if err != nil {
		t.Fatalf("NewWAL failed: %v", err)
	}
	defer wal.Close()

	point := &DataPoint{
		Metric:    "test.metric",
		Timestamp: 1234567890,
		Value:     42.0,
		Tags:      map[string]string{"host": "server1"},
	}

	if err := wal.Write(point); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := wal.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
}

func TestWALRecover(t *testing.T) {
	tmpDir := t.TempDir()
	walPath := filepath.Join(tmpDir, "test.wal")

	// Write some data points
	wal, err := NewWAL(walPath)
	if err != nil {
		t.Fatalf("NewWAL failed: %v", err)
	}

	points := []*DataPoint{
		{Metric: "cpu", Timestamp: 1000, Value: 10.0, Tags: map[string]string{"host": "server1"}},
		{Metric: "cpu", Timestamp: 2000, Value: 20.0, Tags: map[string]string{"host": "server1"}},
		{Metric: "mem", Timestamp: 3000, Value: 30.0, Tags: map[string]string{"host": "server2"}},
	}

	for _, point := range points {
		if err := wal.Write(point); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}

	if err := wal.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	wal.Close()

	// Recover data
	recovered, err := Recover(walPath)
	if err != nil {
		t.Fatalf("Recover failed: %v", err)
	}

	if len(recovered) != len(points) {
		t.Fatalf("Expected %d points, got %d", len(points), len(recovered))
	}

	for i, point := range recovered {
		if point.Metric != points[i].Metric {
			t.Errorf("Point %d: expected metric %s, got %s", i, points[i].Metric, point.Metric)
		}
		if point.Timestamp != points[i].Timestamp {
			t.Errorf("Point %d: expected timestamp %d, got %d", i, points[i].Timestamp, point.Timestamp)
		}
		if point.Value != points[i].Value {
			t.Errorf("Point %d: expected value %f, got %f", i, points[i].Value, point.Value)
		}
	}
}

func TestWALRecoverEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	walPath := filepath.Join(tmpDir, "nonexistent.wal")

	// Recover from non-existent file
	recovered, err := Recover(walPath)
	if err != nil {
		t.Fatalf("Recover should not fail on missing file: %v", err)
	}

	if len(recovered) != 0 {
		t.Errorf("Expected 0 points, got %d", len(recovered))
	}
}

func TestWALTruncate(t *testing.T) {
	tmpDir := t.TempDir()
	walPath := filepath.Join(tmpDir, "test.wal")

	wal, err := NewWAL(walPath)
	if err != nil {
		t.Fatalf("NewWAL failed: %v", err)
	}
	defer wal.Close()

	// Write some data
	point := &DataPoint{
		Metric:    "test.metric",
		Timestamp: 1234567890,
		Value:     42.0,
	}

	if err := wal.Write(point); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := wal.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Truncate
	if err := wal.Truncate(); err != nil {
		t.Fatalf("Truncate failed: %v", err)
	}

	// Close and recover
	wal.Close()

	recovered, err := Recover(walPath)
	if err != nil {
		t.Fatalf("Recover failed: %v", err)
	}

	if len(recovered) != 0 {
		t.Errorf("Expected 0 points after truncate, got %d", len(recovered))
	}
}

func TestWALConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	walPath := filepath.Join(tmpDir, "test.wal")

	wal, err := NewWAL(walPath)
	if err != nil {
		t.Fatalf("NewWAL failed: %v", err)
	}
	defer wal.Close()

	// Write concurrently
	const numGoroutines = 100
	const pointsPerGoroutine = 10

	done := make(chan bool)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < pointsPerGoroutine; j++ {
				point := &DataPoint{
					Metric:    "test.metric",
					Timestamp: int64(id*pointsPerGoroutine + j),
					Value:     float64(id),
				}
				if err := wal.Write(point); err != nil {
					t.Errorf("Write failed: %v", err)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	if err := wal.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	wal.Close()

	// Recover and check count
	recovered, err := Recover(walPath)
	if err != nil {
		t.Fatalf("Recover failed: %v", err)
	}

	expected := numGoroutines * pointsPerGoroutine
	if len(recovered) != expected {
		t.Errorf("Expected %d points, got %d", expected, len(recovered))
	}
}

func BenchmarkWALWrite(b *testing.B) {
	tmpDir := b.TempDir()
	walPath := filepath.Join(tmpDir, "bench.wal")

	wal, err := NewWAL(walPath)
	if err != nil {
		b.Fatalf("NewWAL failed: %v", err)
	}
	defer wal.Close()

	point := &DataPoint{
		Metric:    "test.metric",
		Timestamp: 1234567890,
		Value:     42.0,
		Tags:      map[string]string{"host": "server1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := wal.Write(point); err != nil {
			b.Fatalf("Write failed: %v", err)
		}
	}
}

func BenchmarkWALFlush(b *testing.B) {
	tmpDir := b.TempDir()
	walPath := filepath.Join(tmpDir, "bench.wal")

	wal, err := NewWAL(walPath)
	if err != nil {
		b.Fatalf("NewWAL failed: %v", err)
	}
	defer wal.Close()

	point := &DataPoint{
		Metric:    "test.metric",
		Timestamp: 1234567890,
		Value:     42.0,
	}

	// Write 100 points per flush
	for i := 0; i < 100; i++ {
		wal.Write(point)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := wal.Flush(); err != nil {
			b.Fatalf("Flush failed: %v", err)
		}
	}
}

func BenchmarkWALRecover(b *testing.B) {
	tmpDir := b.TempDir()
	walPath := filepath.Join(tmpDir, "bench.wal")

	// Create WAL with 1000 points
	wal, err := NewWAL(walPath)
	if err != nil {
		b.Fatalf("NewWAL failed: %v", err)
	}

	for i := 0; i < 1000; i++ {
		point := &DataPoint{
			Metric:    "test.metric",
			Timestamp: int64(i),
			Value:     float64(i),
		}
		wal.Write(point)
	}
	wal.Flush()
	wal.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Recover(walPath)
		if err != nil {
			b.Fatalf("Recover failed: %v", err)
		}
	}
}

