package server

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Pablo997/pulsardb/internal/config"
)

// BenchmarkMemoryStress tests memory usage with large dataset
func BenchmarkMemoryStress(b *testing.B) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_memory_stress",
			MaxMemoryMB: 2048, // 2GB for stress test
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	// Pre-populate with 1M points
	b.Log("Populating with 1M data points...")
	for i := 0; i < 1000000; i++ {
		point := map[string]interface{}{
			"metric":    "stress_test",
			"timestamp": float64(1699267200000 + int64(i)*1000),
			"value":     float64(i % 100),
			"tags": map[string]interface{}{
				"host": "server1",
				"dc":   "us-west",
			},
		}
		body, _ := json.Marshal(point)
		req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.handleWrite(w, req)

		if i%100000 == 0 {
			b.Logf("Progress: %d/1000000", i)
		}
	}

	b.Log("Starting benchmark with 1M points in memory")

	// Query payload
	query := map[string]interface{}{
		"metric": "stress_test",
		"start":  float64(1699267200000),
		"end":    float64(1699267200000 + 1000000000), // All points
	}
	queryBody, _ := json.Marshal(query)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/query", bytes.NewBuffer(queryBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		srv.handleQuery(w, req)
	}
}

// BenchmarkHighThroughputWrites simulates sustained high write load
func BenchmarkHighThroughputWrites(b *testing.B) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_high_throughput",
			MaxMemoryMB: 2048,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	// Batch of 1000 points
	points := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		points[i] = map[string]interface{}{
			"metric":    "throughput_test",
			"timestamp": float64(1699267200000 + int64(i)*1000),
			"value":     float64(i),
			"tags": map[string]interface{}{
				"sensor": "sensor1",
			},
		}
	}
	body, _ := json.Marshal(points)

	b.ResetTimer()
	b.ReportAllocs()

	// Simulate sustained load
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		srv.handleWrite(w, req)

		if w.Code != 200 {
			b.Fatalf("unexpected status code: %d", w.Code)
		}
	}
}

// BenchmarkConcurrentStress tests extreme concurrency
func BenchmarkConcurrentStress(b *testing.B) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_concurrent_stress",
			MaxMemoryMB: 2048,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	point := map[string]interface{}{
		"metric":    "concurrent_stress",
		"timestamp": float64(1699267200000),
		"value":     42.0,
	}
	body, _ := json.Marshal(point)

	b.ResetTimer()
	b.ReportAllocs()

	// Maximum parallelism
	b.SetParallelism(1000) // 1000x parallelism

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			srv.handleWrite(w, req)
		}
	})
}

// BenchmarkMixedWorkload simulates realistic mixed read/write workload
func BenchmarkMixedWorkload(b *testing.B) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_mixed_workload",
			MaxMemoryMB: 1024,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	// Pre-populate with some data
	for i := 0; i < 10000; i++ {
		point := map[string]interface{}{
			"metric":    "mixed_test",
			"timestamp": float64(1699267200000 + int64(i)*1000),
			"value":     float64(i),
		}
		body, _ := json.Marshal(point)
		req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.handleWrite(w, req)
	}

	writePoint := map[string]interface{}{
		"metric":    "mixed_test",
		"timestamp": float64(1699267200000),
		"value":     42.0,
	}
	writeBody, _ := json.Marshal(writePoint)

	query := map[string]interface{}{
		"metric": "mixed_test",
		"start":  float64(1699267200000),
		"end":    float64(1699267200000 + 10000000),
	}
	queryBody, _ := json.Marshal(query)

	b.ResetTimer()
	b.ReportAllocs()

	var wg sync.WaitGroup
	
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			counter++
			
			// 80% reads, 20% writes (typical workload)
			if counter%5 == 0 {
				// Write
				req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(writeBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				srv.handleWrite(w, req)
			} else {
				// Query
				req := httptest.NewRequest("POST", "/query", bytes.NewBuffer(queryBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				srv.handleQuery(w, req)
			}
		}
	})
	
	wg.Wait()
}

// BenchmarkLargePayload tests handling of very large JSON payloads
func BenchmarkLargePayload(b *testing.B) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_large_payload",
			MaxMemoryMB: 2048,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	// Create a batch of 50k points (large JSON payload)
	points := make([]map[string]interface{}, 50000)
	for i := 0; i < 50000; i++ {
		tags := make(map[string]interface{})
		for j := 0; j < 10; j++ {
			tags["tag"+string(rune('a'+j))] = "value"
		}
		
		points[i] = map[string]interface{}{
			"metric":    "large_payload_test",
			"timestamp": float64(1699267200000 + int64(i)*1000),
			"value":     float64(i),
			"tags":      tags,
		}
	}

	body, _ := json.Marshal(points)
	b.Logf("Payload size: %.2f MB", float64(len(body))/(1024*1024))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		srv.handleWrite(w, req)

		if w.Code != 200 {
			b.Fatalf("unexpected status code: %d", w.Code)
		}
	}
}

