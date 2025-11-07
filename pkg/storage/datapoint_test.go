package storage

import (
	"testing"
)

func TestDataPointKey(t *testing.T) {
	tests := []struct {
		name   string
		point  *DataPoint
		want   string
	}{
		{
			name: "simple metric",
			point: &DataPoint{
				Metric:    "temperature",
				Timestamp: 1699267200000,
				Value:     23.5,
			},
			want: "temperature",
		},
		{
			name: "with tags",
			point: &DataPoint{
				Metric:    "cpu_usage",
				Timestamp: 1699267200000,
				Value:     45.2,
				Tags:      map[string]string{"host": "server1"},
			},
			want: "cpu_usage",
		},
		{
			name: "empty metric",
			point: &DataPoint{
				Metric:    "",
				Timestamp: 1699267200000,
				Value:     0.0,
			},
			want: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.point.Key()
			if got != tt.want {
				t.Errorf("Key() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDataPointCreation(t *testing.T) {
	point := &DataPoint{
		Metric:    "test_metric",
		Timestamp: 1699267200000,
		Value:     42.0,
		Tags: map[string]string{
			"host":     "server1",
			"region":   "us-west",
			"datacenter": "dc1",
		},
	}
	
	if point.Metric != "test_metric" {
		t.Errorf("expected metric=test_metric, got %s", point.Metric)
	}
	
	if point.Timestamp != 1699267200000 {
		t.Errorf("expected timestamp=1699267200000, got %d", point.Timestamp)
	}
	
	if point.Value != 42.0 {
		t.Errorf("expected value=42.0, got %f", point.Value)
	}
	
	if len(point.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(point.Tags))
	}
	
	if point.Tags["host"] != "server1" {
		t.Errorf("expected host=server1, got %s", point.Tags["host"])
	}
}

func TestDataPointNilTags(t *testing.T) {
	point := &DataPoint{
		Metric:    "test",
		Timestamp: 1000,
		Value:     1.0,
		Tags:      nil, // nil tags should be ok
	}
	
	key := point.Key()
	if key != "test" {
		t.Errorf("Key() with nil tags = %q, want %q", key, "test")
	}
}

func TestDataPointEmptyTags(t *testing.T) {
	point := &DataPoint{
		Metric:    "test",
		Timestamp: 1000,
		Value:     1.0,
		Tags:      map[string]string{}, // empty map
	}
	
	if point.Tags == nil {
		t.Error("empty tags map should not be nil")
	}
	
	if len(point.Tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(point.Tags))
	}
}

func TestDataPointApproximateSize(t *testing.T) {
	tests := []struct {
		name  string
		point *DataPoint
		min   int64 // minimum expected size
	}{
		{
			name: "simple point",
			point: &DataPoint{
				Metric:    "test",
				Timestamp: 1000,
				Value:     1.0,
			},
			min: 60, // metric(4) + timestamp(8) + value(8) + overhead(48) = 68
		},
		{
			name: "point with tags",
			point: &DataPoint{
				Metric:    "temperature",
				Timestamp: 1000,
				Value:     23.5,
				Tags: map[string]string{
					"sensor":   "sensor1",
					"location": "room1",
				},
			},
			min: 90, // metric + timestamp + value + tags + overhead
		},
		{
			name: "point with long metric name",
			point: &DataPoint{
				Metric:    "very_long_metric_name_for_testing_purposes",
				Timestamp: 1000,
				Value:     1.0,
			},
			min: 100,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := tt.point.ApproximateSize()
			if size < tt.min {
				t.Errorf("ApproximateSize() = %d, want at least %d", size, tt.min)
			}
			
			// Size should be reasonable (not negative, not crazy large)
			if size <= 0 || size > 10000 {
				t.Errorf("ApproximateSize() = %d, unreasonable size", size)
			}
		})
	}
}

func TestDataPointApproximateSizeConsistency(t *testing.T) {
	point := &DataPoint{
		Metric:    "test",
		Timestamp: 1000,
		Value:     1.0,
		Tags: map[string]string{
			"key": "value",
		},
	}
	
	// Calling multiple times should return same result
	size1 := point.ApproximateSize()
	size2 := point.ApproximateSize()
	
	if size1 != size2 {
		t.Errorf("ApproximateSize() not consistent: %d vs %d", size1, size2)
	}
}

