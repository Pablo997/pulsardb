package server

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/Pablo997/pulsardb/internal/config"
)

func BenchmarkHandleWriteSinglePoint(b *testing.B) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_data",
			MaxMemoryMB: 512,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	point := map[string]interface{}{
		"metric":    "benchmark",
		"timestamp": float64(1699267200000),
		"value":     42.0,
		"tags": map[string]interface{}{
			"host": "server1",
		},
	}

	body, _ := json.Marshal(point)

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

func BenchmarkHandleWriteBatch10(b *testing.B) {
	benchmarkHandleWriteBatch(b, 10)
}

func BenchmarkHandleWriteBatch100(b *testing.B) {
	benchmarkHandleWriteBatch(b, 100)
}

func BenchmarkHandleWriteBatch1000(b *testing.B) {
	benchmarkHandleWriteBatch(b, 1000)
}

func BenchmarkHandleWriteBatch10000(b *testing.B) {
	benchmarkHandleWriteBatch(b, 10000)
}

func BenchmarkHandleWriteBatch100000(b *testing.B) {
	benchmarkHandleWriteBatch(b, 100000)
}

func benchmarkHandleWriteBatch(b *testing.B, batchSize int) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_data_batch",
			MaxMemoryMB: 1024,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	// Create batch
	points := make([]map[string]interface{}, batchSize)
	for i := 0; i < batchSize; i++ {
		points[i] = map[string]interface{}{
			"metric":    "benchmark",
			"timestamp": float64(1699267200000 + int64(i)*1000),
			"value":     float64(i),
		}
	}

	body, _ := json.Marshal(points)

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

func BenchmarkHandleQuery(b *testing.B) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_data_query",
			MaxMemoryMB: 512,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	// Populate with test data
	for i := 0; i < 1000; i++ {
		point := map[string]interface{}{
			"metric":    "benchmark",
			"timestamp": float64(1699267200000 + int64(i)*1000),
			"value":     float64(i),
		}
		body, _ := json.Marshal(point)
		req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.handleWrite(w, req)
	}

	// Query payload
	query := map[string]interface{}{
		"metric": "benchmark",
		"start":  float64(1699267200000),
		"end":    float64(1699267200000 + 1000000),
	}
	queryBody, _ := json.Marshal(query)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/query", bytes.NewBuffer(queryBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		srv.handleQuery(w, req)

		if w.Code != 200 {
			b.Fatalf("unexpected status code: %d", w.Code)
		}
	}
}

func BenchmarkHandleQuerySmallRange(b *testing.B) {
	benchmarkHandleQueryRange(b, 100)
}

func BenchmarkHandleQueryMediumRange(b *testing.B) {
	benchmarkHandleQueryRange(b, 1000)
}

func BenchmarkHandleQueryLargeRange(b *testing.B) {
	benchmarkHandleQueryRange(b, 10000)
}

func BenchmarkHandleQueryXLargeRange(b *testing.B) {
	benchmarkHandleQueryRange(b, 100000)
}

func BenchmarkHandleQueryMassiveRange(b *testing.B) {
	benchmarkHandleQueryRange(b, 1000000)
}

func benchmarkHandleQueryRange(b *testing.B, dataPoints int) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_data_range",
			MaxMemoryMB: 1024,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	// Populate
	for i := 0; i < dataPoints; i++ {
		point := map[string]interface{}{
			"metric":    "benchmark",
			"timestamp": float64(1699267200000 + int64(i)*1000),
			"value":     float64(i),
		}
		body, _ := json.Marshal(point)
		req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.handleWrite(w, req)
	}

	query := map[string]interface{}{
		"metric": "benchmark",
		"start":  float64(1699267200000),
		"end":    float64(1699267200000 + int64(dataPoints)*1000),
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

func BenchmarkHandleHealth(b *testing.B) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_data_health",
			MaxMemoryMB: 128,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		srv.handleHealth(w, req)
	}
}

func BenchmarkHandleMetrics(b *testing.B) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_data_metrics",
			MaxMemoryMB: 128,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		srv.handleMetrics(w, req)
	}
}

func BenchmarkConcurrentWrites(b *testing.B) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_data_concurrent",
			MaxMemoryMB: 1024,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	point := map[string]interface{}{
		"metric":    "benchmark",
		"timestamp": float64(1699267200000),
		"value":     42.0,
	}
	body, _ := json.Marshal(point)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			srv.handleWrite(w, req)
		}
	})
}

func BenchmarkConcurrentQueries(b *testing.B) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./bench_data_concurrent_query",
			MaxMemoryMB: 512,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Stop()

	// Populate
	for i := 0; i < 1000; i++ {
		point := map[string]interface{}{
			"metric":    "benchmark",
			"timestamp": float64(1699267200000 + int64(i)*1000),
			"value":     float64(i),
		}
		body, _ := json.Marshal(point)
		req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.handleWrite(w, req)
	}

	query := map[string]interface{}{
		"metric": "benchmark",
		"start":  float64(1699267200000),
		"end":    float64(1699267200000 + 1000000),
	}
	queryBody, _ := json.Marshal(query)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/query", bytes.NewBuffer(queryBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			srv.handleQuery(w, req)
		}
	})
}

