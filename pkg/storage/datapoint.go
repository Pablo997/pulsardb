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

