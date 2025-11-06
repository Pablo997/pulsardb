package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pablo997/pulsardb/internal/config"
	"github.com/Pablo997/pulsardb/pkg/storage"
)

func setupTestServer(t *testing.T) *Server {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Address: "127.0.0.1",
			Port:    8080,
		},
		Storage: config.StorageConfig{
			DataDir:     "./test_data_handlers",
			MaxMemoryMB: 128,
		},
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	return srv
}

func TestHandleHealth(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	srv.handleHealth(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["status"] != "healthy" {
		t.Errorf("expected status=healthy, got %s", result["status"])
	}
}

func TestHandleWriteSinglePoint(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	point := map[string]interface{}{
		"metric":    "temperature",
		"timestamp": float64(1699267200000),
		"value":     23.5,
		"tags": map[string]interface{}{
			"sensor": "sensor1",
		},
	}

	body, _ := json.Marshal(point)
	req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleWrite(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	written := int(result["written"].(float64))
	if written != 1 {
		t.Errorf("expected written=1, got %d", written)
	}

	// Verify metrics were updated
	pointsWritten, _, _ := srv.getMetrics()
	if pointsWritten != 1 {
		t.Errorf("expected pointsWritten=1, got %d", pointsWritten)
	}
}

func TestHandleWriteMultiplePoints(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	points := []map[string]interface{}{
		{
			"metric":    "cpu",
			"timestamp": float64(1000),
			"value":     10.0,
		},
		{
			"metric":    "cpu",
			"timestamp": float64(2000),
			"value":     20.0,
		},
		{
			"metric":    "cpu",
			"timestamp": float64(3000),
			"value":     30.0,
		},
	}

	body, _ := json.Marshal(points)
	req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleWrite(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	written := int(result["written"].(float64))
	if written != 3 {
		t.Errorf("expected written=3, got %d", written)
	}
}

func TestHandleWriteInvalidJSON(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	req := httptest.NewRequest("POST", "/write", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleWrite(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleWriteMissingFields(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	tests := []struct {
		name  string
		point map[string]interface{}
	}{
		{
			"missing metric",
			map[string]interface{}{
				"timestamp": float64(1000),
				"value":     10.0,
			},
		},
		{
			"missing timestamp",
			map[string]interface{}{
				"metric": "test",
				"value":  10.0,
			},
		},
		{
			"missing value",
			map[string]interface{}{
				"metric":    "test",
				"timestamp": float64(1000),
			},
		},
		{
			"empty metric",
			map[string]interface{}{
				"metric":    "",
				"timestamp": float64(1000),
				"value":     10.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.point)
			req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			srv.handleWrite(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)

			// Should have errors
			if result["errors"] == nil {
				t.Error("expected errors in response")
			}

			written := int(result["written"].(float64))
			if written != 0 {
				t.Errorf("expected written=0, got %d", written)
			}
		})
	}
}

func TestHandleQuery(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	// Write test data first
	writeData := []map[string]interface{}{
		{"metric": "temp", "timestamp": float64(1000), "value": 10.0},
		{"metric": "temp", "timestamp": float64(2000), "value": 20.0},
		{"metric": "temp", "timestamp": float64(3000), "value": 30.0},
	}

	body, _ := json.Marshal(writeData)
	writeReq := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
	writeReq.Header.Set("Content-Type", "application/json")
	writeW := httptest.NewRecorder()
	srv.handleWrite(writeW, writeReq)

	// Now query
	query := map[string]interface{}{
		"metric": "temp",
		"start":  float64(0),
		"end":    float64(5000),
	}

	queryBody, _ := json.Marshal(query)
	req := httptest.NewRequest("POST", "/query", bytes.NewBuffer(queryBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleQuery(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	count := int(result["count"].(float64))
	if count != 3 {
		t.Errorf("expected count=3, got %d", count)
	}

	// Verify metrics were updated
	_, queriesServed, _ := srv.getMetrics()
	if queriesServed != 1 {
		t.Errorf("expected queriesServed=1, got %d", queriesServed)
	}
}

func TestHandleQueryTimeRange(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	// Write test data
	srv.storage.Write(&storage.DataPoint{Metric: "test", Timestamp: 1000, Value: 1.0})
	srv.storage.Write(&storage.DataPoint{Metric: "test", Timestamp: 2000, Value: 2.0})
	srv.storage.Write(&storage.DataPoint{Metric: "test", Timestamp: 3000, Value: 3.0})

	// Query middle range
	query := map[string]interface{}{
		"metric": "test",
		"start":  float64(1500),
		"end":    float64(2500),
	}

	body, _ := json.Marshal(query)
	req := httptest.NewRequest("POST", "/query", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleQuery(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	count := int(result["count"].(float64))
	if count != 1 {
		t.Errorf("expected count=1, got %d", count)
	}

	points := result["points"].([]interface{})
	if len(points) != 1 {
		t.Errorf("expected 1 point, got %d", len(points))
	}
}

func TestHandleQueryInvalidJSON(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleQuery(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleQueryMissingFields(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	tests := []struct {
		name  string
		query map[string]interface{}
	}{
		{
			"missing metric",
			map[string]interface{}{
				"start": float64(1000),
				"end":   float64(2000),
			},
		},
		{
			"missing start",
			map[string]interface{}{
				"metric": "test",
				"end":    float64(2000),
			},
		},
		{
			"missing end",
			map[string]interface{}{
				"metric": "test",
				"start":  float64(1000),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.query)
			req := httptest.NewRequest("POST", "/query", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			srv.handleQuery(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", resp.StatusCode)
			}
		})
	}
}

func TestHandleQueryInvalidTimeRange(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	query := map[string]interface{}{
		"metric": "test",
		"start":  float64(2000),
		"end":    float64(1000), // end before start
	}

	body, _ := json.Marshal(query)
	req := httptest.NewRequest("POST", "/query", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleQuery(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleMetrics(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	// Do some writes and queries to generate metrics
	point := map[string]interface{}{
		"metric":    "test",
		"timestamp": float64(1000),
		"value":     1.0,
	}
	body, _ := json.Marshal(point)
	
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/write", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.handleWrite(w, req)
	}

	// Wait a bit for uptime
	time.Sleep(100 * time.Millisecond)

	// Get metrics
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	srv.handleMetrics(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	pointsWritten := int(result["points_written"].(float64))
	if pointsWritten != 5 {
		t.Errorf("expected points_written=5, got %d", pointsWritten)
	}

	queriesServed := int(result["queries_served"].(float64))
	if queriesServed != 0 {
		t.Errorf("expected queries_served=0, got %d", queriesServed)
	}

	uptimeSeconds := int(result["uptime_seconds"].(float64))
	if uptimeSeconds < 0 {
		t.Errorf("uptime should be >= 0, got %d", uptimeSeconds)
	}
}

func TestHandleMetricsInitialState(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	srv.handleMetrics(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	// Initial state should be all zeros
	pointsWritten := int(result["points_written"].(float64))
	if pointsWritten != 0 {
		t.Errorf("expected initial points_written=0, got %d", pointsWritten)
	}

	queriesServed := int(result["queries_served"].(float64))
	if queriesServed != 0 {
		t.Errorf("expected initial queries_served=0, got %d", queriesServed)
	}
}

func TestHandleWriteQueryIntegration(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Stop()

	// Write data
	points := []map[string]interface{}{
		{"metric": "integration", "timestamp": float64(1000), "value": 1.0},
		{"metric": "integration", "timestamp": float64(2000), "value": 2.0},
		{"metric": "integration", "timestamp": float64(3000), "value": 3.0},
	}

	writeBody, _ := json.Marshal(points)
	writeReq := httptest.NewRequest("POST", "/write", bytes.NewBuffer(writeBody))
	writeReq.Header.Set("Content-Type", "application/json")
	writeW := httptest.NewRecorder()
	srv.handleWrite(writeW, writeReq)

	// Query back
	query := map[string]interface{}{
		"metric": "integration",
		"start":  float64(0),
		"end":    float64(5000),
	}

	queryBody, _ := json.Marshal(query)
	queryReq := httptest.NewRequest("POST", "/query", bytes.NewBuffer(queryBody))
	queryReq.Header.Set("Content-Type", "application/json")
	queryW := httptest.NewRecorder()
	srv.handleQuery(queryW, queryReq)

	// Check results
	var queryResult map[string]interface{}
	json.NewDecoder(queryW.Result().Body).Decode(&queryResult)

	count := int(queryResult["count"].(float64))
	if count != 3 {
		t.Errorf("expected count=3, got %d", count)
	}

	// Check metrics
	pointsWritten, queriesServed, _ := srv.getMetrics()
	if pointsWritten != 3 {
		t.Errorf("expected pointsWritten=3, got %d", pointsWritten)
	}
	if queriesServed != 1 {
		t.Errorf("expected queriesServed=1, got %d", queriesServed)
	}
}

