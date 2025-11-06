package storage

import (
	"os"
	"testing"
	
	"github.com/Pablo997/pulsardb/internal/config"
)

func TestNewEngine(t *testing.T) {
	cfg := &config.StorageConfig{
		DataDir:     "./test_data",
		MaxMemoryMB: 128,
	}
	
	// Clean up before and after
	defer os.RemoveAll(cfg.DataDir)
	
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("NewEngine failed: %v", err)
	}
	
	if engine == nil {
		t.Fatal("engine is nil")
	}
	
	if engine.memTable == nil {
		t.Error("memTable not initialized")
	}
	
	// Check data directory was created
	if _, err := os.Stat(cfg.DataDir); os.IsNotExist(err) {
		t.Error("data directory was not created")
	}
	
	engine.Close()
}

func TestEngineWrite(t *testing.T) {
	cfg := &config.StorageConfig{
		DataDir:     "./test_data_write",
		MaxMemoryMB: 128,
	}
	defer os.RemoveAll(cfg.DataDir)
	
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("NewEngine failed: %v", err)
	}
	defer engine.Close()
	
	point := &DataPoint{
		Metric:    "temperature",
		Timestamp: 1699267200000,
		Value:     23.5,
		Tags:      map[string]string{"sensor": "sensor1"},
	}
	
	err = engine.Write(point)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	
	// Verify data was written to memtable
	results, err := engine.Query("temperature", 0, 2000000000000)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("expected 1 point, got %d", len(results))
	}
	
	if results[0].Value != 23.5 {
		t.Errorf("expected value=23.5, got %f", results[0].Value)
	}
}

func TestEngineQuery(t *testing.T) {
	cfg := &config.StorageConfig{
		DataDir:     "./test_data_query",
		MaxMemoryMB: 128,
	}
	defer os.RemoveAll(cfg.DataDir)
	
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("NewEngine failed: %v", err)
	}
	defer engine.Close()
	
	// Write test data
	points := []*DataPoint{
		{Metric: "cpu", Timestamp: 1000, Value: 10.0},
		{Metric: "cpu", Timestamp: 2000, Value: 20.0},
		{Metric: "cpu", Timestamp: 3000, Value: 30.0},
		{Metric: "memory", Timestamp: 1000, Value: 50.0},
	}
	
	for _, p := range points {
		if err := engine.Write(p); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}
	
	// Query CPU metric
	results, err := engine.Query("cpu", 0, 5000)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	if len(results) != 3 {
		t.Errorf("expected 3 cpu points, got %d", len(results))
	}
	
	// Query memory metric
	results, err = engine.Query("memory", 0, 5000)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("expected 1 memory point, got %d", len(results))
	}
	
	// Query non-existent metric
	results, err = engine.Query("disk", 0, 5000)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	if len(results) != 0 {
		t.Errorf("expected 0 points for non-existent metric, got %d", len(results))
	}
}

func TestEngineConcurrentWrites(t *testing.T) {
	cfg := &config.StorageConfig{
		DataDir:     "./test_data_concurrent",
		MaxMemoryMB: 128,
	}
	defer os.RemoveAll(cfg.DataDir)
	
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("NewEngine failed: %v", err)
	}
	defer engine.Close()
	
	// Concurrent writes from multiple goroutines
	done := make(chan bool)
	numWriters := 10
	writesPerWriter := 100
	
	for i := 0; i < numWriters; i++ {
		go func(id int) {
			for j := 0; j < writesPerWriter; j++ {
				point := &DataPoint{
					Metric:    "concurrent_test",
					Timestamp: int64(id*10000 + j),
					Value:     float64(id),
				}
				if err := engine.Write(point); err != nil {
					t.Errorf("concurrent write failed: %v", err)
				}
			}
			done <- true
		}(i)
	}
	
	// Wait for all writers
	for i := 0; i < numWriters; i++ {
		<-done
	}
	
	// Verify all points were written
	results, err := engine.Query("concurrent_test", 0, 1000000)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	expected := numWriters * writesPerWriter
	if len(results) != expected {
		t.Errorf("expected %d points, got %d", expected, len(results))
	}
}

func TestEngineClose(t *testing.T) {
	cfg := &config.StorageConfig{
		DataDir:     "./test_data_close",
		MaxMemoryMB: 128,
	}
	defer os.RemoveAll(cfg.DataDir)
	
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("NewEngine failed: %v", err)
	}
	
	// Write some data
	point := &DataPoint{
		Metric:    "test",
		Timestamp: 1000,
		Value:     1.0,
	}
	engine.Write(point)
	
	// Close should not error
	err = engine.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func BenchmarkEngineWrite(b *testing.B) {
	cfg := &config.StorageConfig{
		DataDir:     "./bench_data",
		MaxMemoryMB: 512,
	}
	defer os.RemoveAll(cfg.DataDir)
	
	engine, _ := NewEngine(cfg)
	defer engine.Close()
	
	point := &DataPoint{
		Metric:    "benchmark",
		Timestamp: 1699267200000,
		Value:     42.0,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Write(point)
	}
}

func BenchmarkEngineQuery(b *testing.B) {
	cfg := &config.StorageConfig{
		DataDir:     "./bench_data_query",
		MaxMemoryMB: 512,
	}
	defer os.RemoveAll(cfg.DataDir)
	
	engine, _ := NewEngine(cfg)
	defer engine.Close()
	
	// Populate with data
	for i := 0; i < 10000; i++ {
		engine.Write(&DataPoint{
			Metric:    "benchmark",
			Timestamp: int64(i * 1000),
			Value:     float64(i),
		})
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Query("benchmark", 0, 10000000)
	}
}

