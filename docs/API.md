# API Reference

Complete API documentation for PulsarDB.

## Base URL

```
http://localhost:8080
```

---

## Endpoints

### Health Check

Check if the server is running.

**Request:**
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy"
}
```

---

### Write Data

Write single or multiple data points.

#### Write Single Point

**Request:**
```http
POST /write
Content-Type: application/json

{
  "metric": "temperature",
  "timestamp": 1699267200000,
  "value": 23.5,
  "tags": {
    "sensor": "sensor1",
    "location": "room1"
  }
}
```

**Response:**
```json
{
  "written": 1
}
```

#### Write Multiple Points

**Request:**
```http
POST /write
Content-Type: application/json

[
  {
    "metric": "temperature",
    "timestamp": 1699267200000,
    "value": 23.5,
    "tags": {"sensor": "sensor1"}
  },
  {
    "metric": "temperature",
    "timestamp": 1699267260000,
    "value": 24.1,
    "tags": {"sensor": "sensor1"}
  }
]
```

**Response:**
```json
{
  "written": 2
}
```

#### Error Response

```json
{
  "written": 1,
  "errors": ["missing or invalid metric", "missing or invalid timestamp"]
}
```

**Fields:**
- `metric` (string, required): Metric name
- `timestamp` (int64, required): Unix timestamp in milliseconds
- `value` (float64, required): Numeric value
- `tags` (object, optional): Key-value pairs for metadata

---

### Query Data

Query time-series data within a time range.

**Request:**
```http
POST /query
Content-Type: application/json

{
  "metric": "temperature",
  "start": 1699267200000,
  "end": 1699353600000
}
```

**Response:**
```json
{
  "metric": "temperature",
  "start": 1699267200000,
  "end": 1699353600000,
  "points": [
    {
      "metric": "temperature",
      "timestamp": 1699267200000,
      "value": 23.5,
      "tags": {"sensor": "sensor1"}
    },
    {
      "metric": "temperature",
      "timestamp": 1699267260000,
      "value": 24.1,
      "tags": {"sensor": "sensor1"}
    }
  ],
  "count": 2
}
```

**Fields:**
- `metric` (string, required): Metric name to query
- `start` (int64, required): Start timestamp (inclusive)
- `end` (int64, required): End timestamp (inclusive)
- `tags` (object, optional): Filter by tags (not yet implemented)

**Error Responses:**

Missing metric:
```json
{
  "error": "missing or invalid metric"
}
```

Invalid time range:
```json
{
  "error": "start timestamp must be before end timestamp"
}
```

---

### Metrics

Get database metrics and statistics.

**Request:**
```http
GET /metrics
```

**Response:**
```json
{
  "points_written": 150,
  "queries_served": 42,
  "uptime_seconds": 3600
}
```

**Fields:**
- `points_written`: Total number of data points written
- `queries_served`: Total number of queries executed
- `uptime_seconds`: Server uptime in seconds

---

## HTTP Status Codes

- `200 OK`: Request successful
- `206 Partial Content`: Some points written, some failed
- `400 Bad Request`: Invalid request format or parameters
- `500 Internal Server Error`: Server error

---

## Data Types

### DataPoint

```json
{
  "metric": "string",
  "timestamp": 1234567890000,
  "value": 42.5,
  "tags": {
    "key1": "value1",
    "key2": "value2"
  }
}
```

- `metric`: Identifies the measurement (e.g., "cpu_usage", "temperature")
- `timestamp`: Unix time in milliseconds
- `value`: Floating-point number
- `tags`: Optional metadata for filtering and grouping

---

## Rate Limits

Currently no rate limits enforced. Will be added in future versions.

---

## Authentication

Currently no authentication. Will be added in future versions.

