package storage

import (
	"testing"
)

func TestNewMemTable(t *testing.T) {
	mt := NewMemTable(512)
	
	if mt == nil {
		t.Fatal("NewMemTable returned nil")
	}
	
	if mt.maxSizeMB != 512 {
		t.Errorf("expected maxSizeMB=512, got %d", mt.maxSizeMB)
	}
	
	if mt.data == nil {
		t.Error("data map not initialized")
	}
	
	if mt.size != 0 {
		t.Errorf("expected initial size=0, got %d", mt.size)
	}
}

func TestMemTableInsert(t *testing.T) {
	mt := NewMemTable(512)
	
	point := &DataPoint{
		Metric:    "temperature",
		Timestamp: 1699267200000,
		Value:     23.5,
		Tags:      map[string]string{"sensor": "sensor1"},
	}
	
	err := mt.Insert(point)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	
	// Check if data was inserted
	points := mt.data["temperature"]
	if len(points) != 1 {
		t.Errorf("expected 1 point, got %d", len(points))
	}
	
	if points[0].Value != 23.5 {
		t.Errorf("expected value=23.5, got %f", points[0].Value)
	}
}

func TestMemTableInsertMultiple(t *testing.T) {
	mt := NewMemTable(512)
	
	points := []*DataPoint{
		{Metric: "cpu", Timestamp: 1000, Value: 10.0},
		{Metric: "cpu", Timestamp: 2000, Value: 20.0},
		{Metric: "cpu", Timestamp: 3000, Value: 30.0},
		{Metric: "memory", Timestamp: 1000, Value: 50.0},
	}
	
	for _, p := range points {
		if err := mt.Insert(p); err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}
	
	// Check cpu metric
	cpuPoints := mt.data["cpu"]
	if len(cpuPoints) != 3 {
		t.Errorf("expected 3 cpu points, got %d", len(cpuPoints))
	}
	
	// Check memory metric
	memPoints := mt.data["memory"]
	if len(memPoints) != 1 {
		t.Errorf("expected 1 memory point, got %d", len(memPoints))
	}
}

func TestMemTableQuery(t *testing.T) {
	mt := NewMemTable(512)
	
	// Insert test data
	testPoints := []*DataPoint{
		{Metric: "temp", Timestamp: 1000, Value: 10.0},
		{Metric: "temp", Timestamp: 2000, Value: 20.0},
		{Metric: "temp", Timestamp: 3000, Value: 30.0},
		{Metric: "temp", Timestamp: 4000, Value: 40.0},
	}
	
	for _, p := range testPoints {
		mt.Insert(p)
	}
	
	tests := []struct {
		name   string
		metric string
		start  int64
		end    int64
		want   int
	}{
		{"all points", "temp", 0, 5000, 4},
		{"middle range", "temp", 1500, 3500, 2},
		{"single point", "temp", 2000, 2000, 1},
		{"no points", "temp", 5000, 6000, 0},
		{"non-existent metric", "cpu", 0, 5000, 0},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := mt.Query(tt.metric, tt.start, tt.end)
			if len(results) != tt.want {
				t.Errorf("Query(%s, %d, %d) returned %d points, want %d",
					tt.metric, tt.start, tt.end, len(results), tt.want)
			}
		})
	}
}

func TestMemTableQueryTimeRange(t *testing.T) {
	mt := NewMemTable(512)
	
	mt.Insert(&DataPoint{Metric: "test", Timestamp: 1000, Value: 1.0})
	mt.Insert(&DataPoint{Metric: "test", Timestamp: 2000, Value: 2.0})
	mt.Insert(&DataPoint{Metric: "test", Timestamp: 3000, Value: 3.0})
	
	// Query: start=1500, end=2500 (should get only timestamp=2000)
	results := mt.Query("test", 1500, 2500)
	
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	
	if results[0].Timestamp != 2000 {
		t.Errorf("expected timestamp=2000, got %d", results[0].Timestamp)
	}
	
	if results[0].Value != 2.0 {
		t.Errorf("expected value=2.0, got %f", results[0].Value)
	}
}

func TestMemTableIsFull(t *testing.T) {
	mt := NewMemTable(1) // 1 MB
	
	if mt.IsFull() {
		t.Error("empty memtable should not be full")
	}
	
	// Simulate filling up
	mt.size = 2 * 1024 * 1024 // 2 MB
	
	if !mt.IsFull() {
		t.Error("memtable should be full")
	}
}

func TestMemTableClear(t *testing.T) {
	mt := NewMemTable(512)
	
	// Add some data
	mt.Insert(&DataPoint{Metric: "test", Timestamp: 1000, Value: 1.0})
	mt.Insert(&DataPoint{Metric: "test", Timestamp: 2000, Value: 2.0})
	
	if len(mt.data) == 0 {
		t.Fatal("data should not be empty before clear")
	}
	
	// Clear
	mt.Clear()
	
	if len(mt.data) != 0 {
		t.Errorf("expected empty data after clear, got %d entries", len(mt.data))
	}
	
	if mt.size != 0 {
		t.Errorf("expected size=0 after clear, got %d", mt.size)
	}
}

func TestMemTableConcurrency(t *testing.T) {
	mt := NewMemTable(512)
	
	// Concurrent writes
	done := make(chan bool)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				point := &DataPoint{
					Metric:    "concurrent",
					Timestamp: int64(id*1000 + j),
					Value:     float64(id),
				}
				mt.Insert(point)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Should have 1000 points total
	results := mt.Query("concurrent", 0, 100000)
	if len(results) != 1000 {
		t.Errorf("expected 1000 points from concurrent writes, got %d", len(results))
	}
}

func BenchmarkMemTableInsert(b *testing.B) {
	mt := NewMemTable(512)
	point := &DataPoint{
		Metric:    "benchmark",
		Timestamp: 1699267200000,
		Value:     42.0,
		Tags:      map[string]string{"host": "server1"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mt.Insert(point)
	}
}

func BenchmarkMemTableQuery(b *testing.B) {
	mt := NewMemTable(512)
	
	// Populate with 1000 points
	for i := 0; i < 1000; i++ {
		mt.Insert(&DataPoint{
			Metric:    "benchmark",
			Timestamp: int64(i * 1000),
			Value:     float64(i),
		})
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mt.Query("benchmark", 0, 1000000)
	}
}

