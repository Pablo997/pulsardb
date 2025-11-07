package storage

// DataPoint represents a single time-series data point
type DataPoint struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"` // Unix timestamp in milliseconds
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// Key returns a unique key for this data point
func (dp *DataPoint) Key() string {
	// TODO: Implement proper key generation with tags
	return dp.Metric
}

// ApproximateSize calculates the approximate memory size of this data point
func (dp *DataPoint) ApproximateSize() int64 {
	size := int64(len(dp.Metric)) // metric string
	size += 8                      // timestamp (int64)
	size += 8                      // value (float64)
	
	// Tags
	for k, v := range dp.Tags {
		size += int64(len(k) + len(v))
	}
	
	// Overhead: struct fields, pointers, map header
	size += 48
	
	return size
}

