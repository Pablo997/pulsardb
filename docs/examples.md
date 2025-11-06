# Usage Examples

Examples for interacting with PulsarDB using different tools and languages.

---

## Using curl (Linux/Mac/Windows)

### Write Single Point
```bash
curl -X POST http://localhost:8080/write \
  -H "Content-Type: application/json" \
  -d '{
    "metric": "cpu_usage",
    "timestamp": 1699267200000,
    "value": 45.2,
    "tags": {"host": "server1", "datacenter": "us-west"}
  }'
```

### Write Multiple Points
```bash
curl -X POST http://localhost:8080/write \
  -H "Content-Type: application/json" \
  -d '[
    {
      "metric": "memory_usage",
      "timestamp": 1699267200000,
      "value": 78.5,
      "tags": {"host": "server1"}
    },
    {
      "metric": "memory_usage",
      "timestamp": 1699267260000,
      "value": 79.2,
      "tags": {"host": "server1"}
    }
  ]'
```

### Query Data
```bash
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{
    "metric": "cpu_usage",
    "start": 1699267000000,
    "end": 1699267300000
  }'
```

### Get Metrics
```bash
curl http://localhost:8080/metrics
```

### Health Check
```bash
curl http://localhost:8080/health
```

---

## Using PowerShell (Windows)

### Write Data
```powershell
$body = @{
    metric = "cpu_usage"
    timestamp = 1699267200000
    value = 45.2
    tags = @{
        host = "server1"
        datacenter = "us-west"
    }
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/write" `
  -Method POST `
  -Body $body `
  -ContentType "application/json"
```

### Write Multiple Points
```powershell
$body = @(
    @{
        metric = "memory_usage"
        timestamp = 1699267200000
        value = 78.5
        tags = @{host = "server1"}
    },
    @{
        metric = "memory_usage"
        timestamp = 1699267260000
        value = 79.2
        tags = @{host = "server1"}
    }
) | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/write" `
  -Method POST `
  -Body $body `
  -ContentType "application/json"
```

### Query Data
```powershell
$query = @{
    metric = "cpu_usage"
    start = 1699267000000
    end = 1699267300000
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/query" `
  -Method POST `
  -Body $query `
  -ContentType "application/json"
```

### Get Metrics
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/metrics" -Method GET
```

---

## Using Go

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DataPoint struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
}

func main() {
	// Write data point
	point := DataPoint{
		Metric:    "temperature",
		Timestamp: time.Now().UnixMilli(),
		Value:     23.5,
		Tags: map[string]string{
			"sensor":   "sensor1",
			"location": "room1",
		},
	}

	data, _ := json.Marshal(point)
	resp, err := http.Post(
		"http://localhost:8080/write",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Write response:", resp.Status)

	// Query data
	query := map[string]interface{}{
		"metric": "temperature",
		"start":  time.Now().Add(-1 * time.Hour).UnixMilli(),
		"end":    time.Now().UnixMilli(),
	}

	queryData, _ := json.Marshal(query)
	resp, err = http.Post(
		"http://localhost:8080/query",
		"application/json",
		bytes.NewBuffer(queryData),
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Query result: %+v\n", result)
}
```

---

## Using Python

```python
import requests
import time

# Write single point
point = {
    "metric": "temperature",
    "timestamp": int(time.time() * 1000),
    "value": 23.5,
    "tags": {
        "sensor": "sensor1",
        "location": "room1"
    }
}

response = requests.post(
    "http://localhost:8080/write",
    json=point
)
print("Write response:", response.json())

# Write multiple points
points = [
    {
        "metric": "humidity",
        "timestamp": int(time.time() * 1000),
        "value": 65.2,
        "tags": {"sensor": "sensor1"}
    },
    {
        "metric": "humidity",
        "timestamp": int(time.time() * 1000) + 60000,
        "value": 66.1,
        "tags": {"sensor": "sensor1"}
    }
]

response = requests.post(
    "http://localhost:8080/write",
    json=points
)
print("Batch write:", response.json())

# Query data
query = {
    "metric": "temperature",
    "start": int((time.time() - 3600) * 1000),  # 1 hour ago
    "end": int(time.time() * 1000)
}

response = requests.post(
    "http://localhost:8080/query",
    json=query
)
result = response.json()
print(f"Found {result['count']} points")

# Get metrics
metrics = requests.get("http://localhost:8080/metrics").json()
print(f"Metrics: {metrics}")
```

---

## Using JavaScript/Node.js

```javascript
const axios = require('axios');

const BASE_URL = 'http://localhost:8080';

// Write single point
async function writePoint() {
  const point = {
    metric: 'temperature',
    timestamp: Date.now(),
    value: 23.5,
    tags: {
      sensor: 'sensor1',
      location: 'room1'
    }
  };

  const response = await axios.post(`${BASE_URL}/write`, point);
  console.log('Write response:', response.data);
}

// Write multiple points
async function writeBatch() {
  const points = [
    {
      metric: 'cpu_usage',
      timestamp: Date.now(),
      value: 45.2,
      tags: { host: 'server1' }
    },
    {
      metric: 'cpu_usage',
      timestamp: Date.now() + 60000,
      value: 46.8,
      tags: { host: 'server1' }
    }
  ];

  const response = await axios.post(`${BASE_URL}/write`, points);
  console.log('Batch write:', response.data);
}

// Query data
async function queryData() {
  const query = {
    metric: 'temperature',
    start: Date.now() - 3600000, // 1 hour ago
    end: Date.now()
  };

  const response = await axios.post(`${BASE_URL}/query`, query);
  console.log(`Found ${response.data.count} points`);
  console.log('Points:', response.data.points);
}

// Get metrics
async function getMetrics() {
  const response = await axios.get(`${BASE_URL}/metrics`);
  console.log('Metrics:', response.data);
}

// Run examples
(async () => {
  await writePoint();
  await writeBatch();
  await queryData();
  await getMetrics();
})();
```

---

## IoT Sensor Example

Simulating an IoT temperature sensor sending data every 10 seconds:

```python
import requests
import time
import random

BASE_URL = "http://localhost:8080"
SENSOR_ID = "sensor_001"
LOCATION = "warehouse_a"

def read_temperature():
    # Simulate sensor reading
    return round(random.uniform(18.0, 28.0), 2)

while True:
    temp = read_temperature()
    
    point = {
        "metric": "temperature",
        "timestamp": int(time.time() * 1000),
        "value": temp,
        "tags": {
            "sensor": SENSOR_ID,
            "location": LOCATION
        }
    }
    
    try:
        response = requests.post(f"{BASE_URL}/write", json=point)
        print(f"Sent {temp}°C - Response: {response.json()}")
    except Exception as e:
        print(f"Error: {e}")
    
    time.sleep(10)  # Wait 10 seconds
```

---

## Monitoring Script

Get metrics every 30 seconds:

```bash
#!/bin/bash

while true; do
  echo "=== PulsarDB Metrics $(date) ==="
  curl -s http://localhost:8080/metrics | jq '.'
  echo ""
  sleep 30
done
```

