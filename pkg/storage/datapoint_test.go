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

