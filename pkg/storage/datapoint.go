package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
)

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

// EncodeBinary encodes the DataPoint to binary format (3-5x faster than JSON)
// Format: [metric_len][metric][timestamp][value][num_tags][tag_key_len][tag_key][tag_val_len][tag_val]...
func (dp *DataPoint) EncodeBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write metric length and metric
	metricBytes := []byte(dp.Metric)
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(metricBytes))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(metricBytes); err != nil {
		return nil, err
	}

	// Write timestamp
	if err := binary.Write(buf, binary.LittleEndian, dp.Timestamp); err != nil {
		return nil, err
	}

	// Write value
	if err := binary.Write(buf, binary.LittleEndian, dp.Value); err != nil {
		return nil, err
	}

	// Write number of tags
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(dp.Tags))); err != nil {
		return nil, err
	}

	// Write tags (sorted for deterministic encoding)
	keys := make([]string, 0, len(dp.Tags))
	for k := range dp.Tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := dp.Tags[key]
		
		// Write key length and key
		keyBytes := []byte(key)
		if err := binary.Write(buf, binary.LittleEndian, uint32(len(keyBytes))); err != nil {
			return nil, err
		}
		if _, err := buf.Write(keyBytes); err != nil {
			return nil, err
		}

		// Write value length and value
		valueBytes := []byte(value)
		if err := binary.Write(buf, binary.LittleEndian, uint32(len(valueBytes))); err != nil {
			return nil, err
		}
		if _, err := buf.Write(valueBytes); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// DecodeDataPoint decodes a DataPoint from binary format
func DecodeDataPoint(data []byte) (*DataPoint, error) {
	buf := bytes.NewReader(data)
	dp := &DataPoint{Tags: make(map[string]string)}

	// Read metric length
	var metricLen uint32
	if err := binary.Read(buf, binary.LittleEndian, &metricLen); err != nil {
		return nil, fmt.Errorf("failed to read metric length: %w", err)
	}

	// Read metric
	metricBytes := make([]byte, metricLen)
	if _, err := buf.Read(metricBytes); err != nil {
		return nil, fmt.Errorf("failed to read metric: %w", err)
	}
	dp.Metric = string(metricBytes)

	// Read timestamp
	if err := binary.Read(buf, binary.LittleEndian, &dp.Timestamp); err != nil {
		return nil, fmt.Errorf("failed to read timestamp: %w", err)
	}

	// Read value
	if err := binary.Read(buf, binary.LittleEndian, &dp.Value); err != nil {
		return nil, fmt.Errorf("failed to read value: %w", err)
	}

	// Read number of tags
	var numTags uint32
	if err := binary.Read(buf, binary.LittleEndian, &numTags); err != nil {
		return nil, fmt.Errorf("failed to read tag count: %w", err)
	}

	// Read tags
	for i := uint32(0); i < numTags; i++ {
		// Read key length
		var keyLen uint32
		if err := binary.Read(buf, binary.LittleEndian, &keyLen); err != nil {
			return nil, fmt.Errorf("failed to read tag key length: %w", err)
		}

		// Read key
		keyBytes := make([]byte, keyLen)
		if _, err := buf.Read(keyBytes); err != nil {
			return nil, fmt.Errorf("failed to read tag key: %w", err)
		}

		// Read value length
		var valueLen uint32
		if err := binary.Read(buf, binary.LittleEndian, &valueLen); err != nil {
			return nil, fmt.Errorf("failed to read tag value length: %w", err)
		}

		// Read value
		valueBytes := make([]byte, valueLen)
		if _, err := buf.Read(valueBytes); err != nil {
			return nil, fmt.Errorf("failed to read tag value: %w", err)
		}

		dp.Tags[string(keyBytes)] = string(valueBytes)
	}

	return dp, nil
}

