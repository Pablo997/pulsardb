package server

import (
	"encoding/json"
	"net/http"

	"github.com/Pablo997/pulsardb/pkg/storage"
)

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// handleWrite handles data point writes
func (s *Server) handleWrite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	defer r.Body.Close()
	
	var body interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid JSON format",
		})
		return
	}

	// Handle both single point and array of points
	var points []map[string]interface{}
	switch v := body.(type) {
	case map[string]interface{}:
		points = []map[string]interface{}{v}
	case []interface{}:
		for _, item := range v {
			if point, ok := item.(map[string]interface{}); ok {
				points = append(points, point)
			}
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "expected object or array",
		})
		return
	}

	written := 0
	errors := []string{}

	for _, pointData := range points {
		// Extract fields
		metric, ok := pointData["metric"].(string)
		if !ok || metric == "" {
			errors = append(errors, "missing or invalid metric")
			continue
		}

		timestamp, ok := pointData["timestamp"].(float64)
		if !ok {
			errors = append(errors, "missing or invalid timestamp")
			continue
		}

		value, ok := pointData["value"].(float64)
		if !ok {
			errors = append(errors, "missing or invalid value")
			continue
		}

		// Extract tags (optional)
		tags := make(map[string]string)
		if tagsData, ok := pointData["tags"].(map[string]interface{}); ok {
			for k, v := range tagsData {
				if strVal, ok := v.(string); ok {
					tags[k] = strVal
				}
			}
		}

		// Create DataPoint
		dp := &storage.DataPoint{
			Metric:    metric,
			Timestamp: int64(timestamp),
			Value:     value,
			Tags:      tags,
		}

		// Write to storage
		if err := s.storage.Write(dp); err != nil {
			errors = append(errors, err.Error())
			continue
		}

		written++
	}

	// Update metrics
	if written > 0 {
		s.incrementPointsWritten(int64(written))
	}

	// Response
	response := map[string]interface{}{
		"written": written,
	}

	if len(errors) > 0 {
		response["errors"] = errors
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

// handleQuery handles time-series queries
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	// Parse query request
	var queryReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&queryReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid JSON format",
		})
		return
	}

	// Extract and validate metric
	metric, ok := queryReq["metric"].(string)
	if !ok || metric == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "missing or invalid metric",
		})
		return
	}

	// Extract and validate start timestamp
	start, ok := queryReq["start"].(float64)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "missing or invalid start timestamp",
		})
		return
	}

	// Extract and validate end timestamp
	end, ok := queryReq["end"].(float64)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "missing or invalid end timestamp",
		})
		return
	}

	// Validate time range
	if start > end {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "start timestamp must be before end timestamp",
		})
		return
	}

	// Query storage
	points, err := s.storage.Query(metric, int64(start), int64(end))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	// TODO: Filter by tags if provided
	// tags, _ := queryReq["tags"].(map[string]interface{})

	// Update metrics
	s.incrementQueriesServed()

	// Response
	response := map[string]interface{}{
		"metric": metric,
		"start":  int64(start),
		"end":    int64(end),
		"points": points,
		"count":  len(points),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleMetrics returns database metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Get current metrics
	pointsWritten, queriesServed, uptime := s.getMetrics()
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"points_written": pointsWritten,
		"queries_served": queriesServed,
		"uptime_seconds": uptime,
	})
}

