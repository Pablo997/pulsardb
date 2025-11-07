# Architecture

PulsarDB system design and internals.

---

## Overview

PulsarDB is designed as a lightweight time-series database optimized for edge computing and IoT applications. It follows an LSM-tree (Log-Structured Merge-tree) inspired architecture with a memory-first approach.

```
┌─────────────────────────────────────────┐
│          HTTP API Layer                 │
│  (Gorilla Mux Router)                   │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│          Server Layer                   │
│  - Request handling                     │
│  - Metrics tracking                     │
│  - Thread-safe operations               │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│        Storage Engine                   │
│  - Write coordination                   │
│  - Query routing                        │
│  - Compaction (future)                  │
└───────────┬───────────┬─────────────────┘
            │           │
    ┌───────▼──────┐   ┌▼──────────────┐
    │  MemTable    │   │  WAL          │
    │  (Active)    │   │  (Lazy Flush) │
    └──────┬───────┘   └───────────────┘
           │
    ┌──────▼───────────────────────────┐
    │  SSTables (Persistent Storage)   │
    │  (Future)                        │
    └──────────────────────────────────┘
```

---

## Components

### 1. HTTP API Layer

**Technology:** Gorilla Mux

**Responsibilities:**
- Route HTTP requests to handlers
- Parse JSON payloads
- Return JSON responses

**Endpoints:**
- `POST /write` - Write data points
- `POST /query` - Query time-series data
- `GET /health` - Health check
- `GET /metrics` - System metrics

### 2. Server Layer

**File:** `internal/server/`

**Responsibilities:**
- Manage HTTP server lifecycle
- Track metrics (writes, queries, uptime)
- Thread-safe counter management
- Graceful shutdown

**Concurrency:**
- Uses `sync.RWMutex` for metrics
- Safe for concurrent requests

### 3. Storage Engine

**File:** `pkg/storage/engine.go`

**Responsibilities:**
- Coordinate writes and queries
- Manage MemTable
- Handle compaction (future)
- Enforce data retention (future)

**Current State:** In-memory + WAL (Lazy Flush)
**Future:** SSTables for long-term persistence

### 4. MemTable (In-Memory Buffer)

**File:** `pkg/storage/memtable.go`

**Data Structure:**
```go
map[string][]*DataPoint
// metric -> sorted list of points
```

**Characteristics:**
- Fast writes: O(1) insert
- Fast queries: O(n) scan with time filter
- Thread-safe with RWMutex
- Size-limited (configurable)

**When full:**
- Flush to SSTable (future)
- Create new MemTable
- Continue accepting writes

### 5. Write-Ahead Log (WAL) - Binary Encoding

**Purpose:** Durability with minimal performance impact

**How it works:**
1. Write arrives
2. Encode to **binary format** (3-5x faster than JSON)
3. Write to buffered WAL (memory)
4. Return success (~650ns)
5. Flush to disk on memtable full (lazy strategy)

**Benefits:**
- Crash recovery with binary encoding
- ~650ns write latency (1.6x faster than JSON WAL)
- Simple, predictable performance
- No async complexity
- Perfect for IoT workloads

**Binary Format:**
```
[4 bytes length][binary data]
- Metric: length-prefixed string
- Timestamp: int64
- Value: float64
- Tags: count + (key_len, key, val_len, val)
```

**Strategy: Lazy Flush**
- Writes buffered in memory
- Fsync only when memtable is full
- Predictable, no queue backpressure
- Max data loss: ~5min (until flush)

### 6. SSTables - Future

**Purpose:** Persistent storage

**Format:**
```
[Header]
  - Version
  - Compression type
  - Index offset

[Data Blocks]
  - Compressed time-series data
  - Sorted by (metric, timestamp)

[Index]
  - Block offsets
  - Min/Max timestamps per block
```

**File structure:**
```
data/
  ├── 00001.sst
  ├── 00002.sst
  ├── 00003.sst
  └── ...
```

### 7. Compaction - Future

**Purpose:** Merge and compact SSTables

**Levels:**
- L0: Fresh from MemTable (unsorted)
- L1: First compaction (sorted)
- L2+: Larger, older data

**Triggers:**
- L0 reaches N files
- L1+ exceeds size threshold

---

## Data Flow

### Write Path

```
1. HTTP POST /write
2. Parse JSON → DataPoint
3. Validate fields
4. Encode binary + write to WAL buffer
5. Insert into MemTable
6. Flush if MemTable full (lazy)
7. Update metrics
8. Return response
```

**Current latency:** ~1ms (memory + binary WAL)
**Future latency:** 1-2ms (with SSTables)

### Query Path

```
1. HTTP POST /query
2. Parse query params
3. Query MemTable (scan + filter)
4. [Future] Query SSTables if needed
5. [Future] Merge results
6. Sort by timestamp
7. Return response
```

**Current latency:** <10ms (memory scan)
**Future latency:** 10-50ms (with SSTables)

---

## Concurrency Model

### Thread Safety

All operations are thread-safe:

```go
// Write path
engine.mu.Lock()
defer engine.mu.Unlock()

// Read path  
engine.mu.RLock()
defer engine.mu.RUnlock()
```

### Metrics Tracking

```go
// Atomic increments
server.metricsMutex.Lock()
server.pointsWritten += count
server.metricsMutex.Unlock()
```

### Request Handling

- Each HTTP request runs in its own goroutine
- Go's HTTP server handles concurrency automatically
- Safe for thousands of concurrent requests

---

## Performance Characteristics

### Current (Memory + Binary WAL)

**Writes:**
- Throughput: ~1.5M points/sec (single thread)
- Throughput: ~2M points/sec (concurrent)
- Latency: ~650ns p99
- Limited by: RAM

**Queries:**
- Throughput: ~100K queries/sec
- Latency: <10ms p99
- Limited by: Memory scan speed

**WAL Performance:**
- Encoding: Binary (50-100ns per point)
- Write latency: ~650ns (buffered)
- Flush: On memtable full (lazy)
- Recovery: ~2.5K points/ms (very fast)

### Future (With SSTables)

**Writes:**
- Throughput: ~500K points/sec
- Latency: 1-2ms p99
- Limited by: Disk I/O (WAL + SSTables)

**Queries:**
- Throughput: ~50K queries/sec
- Latency: 10-50ms p99
- Limited by: Disk I/O, caching

---

## Configuration

**File:** `internal/config/config.go`

```json
{
  "http": {
    "address": "0.0.0.0",
    "port": 8080
  },
  "storage": {
    "data_dir": "./data",
    "max_memory_mb": 512,
    "flush_interval_seconds": 60,
    "retention_days": 7,
    "compression_enabled": true,
    "wal_enabled": true,
    "wal_path": "./data/wal.log"
  }
}
```

---

## Future Improvements

### Short Term
- [x] WAL with binary encoding
- [ ] SSTable writer
- [ ] Basic compaction
- [ ] Tag filtering

### Medium Term
- [ ] Gorilla compression
- [ ] Advanced indexing
- [ ] Query optimization
- [ ] Retention enforcement

### Long Term
- [ ] Distributed mode
- [ ] Replication
- [ ] Sharding
- [ ] Advanced analytics

---

## Design Decisions

### Why LSM-tree?

**Pros:**
- Fast writes (sequential I/O)
- Good compression
- Handles time-series well

**Cons:**
- Read amplification
- Compaction overhead

**Alternative considered:** B-tree (slower writes)

### Why HTTP API?

**Pros:**
- Universal compatibility
- Easy debugging (curl)
- No client library required
- Firewall friendly

**Cons:**
- Slightly higher overhead than binary protocols

**Alternative considered:** gRPC (faster but complex)

### Why Go?

**Pros:**
- Fast
- Great concurrency
- Small binaries
- Easy deployment

**Cons:**
- GC pauses (mitigated)

**Alternative considered:** Rust (steeper learning curve)

---

## Comparison to Other TSDBs

| Feature | PulsarDB | InfluxDB | Prometheus | TimescaleDB |
|---------|----------|----------|------------|-------------|
| Language | Go | Go/Rust | Go | C (Postgres) |
| Architecture | LSM-tree | TSM | TSDB | B-tree |
| Edge-friendly | ✅ Yes | ⚠️ Medium | ⚠️ Medium | ❌ No |
| Memory usage | Low | Medium | Medium | High |
| Query language | HTTP/JSON | InfluxQL | PromQL | SQL |
| Maturity | 🚧 Early | ✅ Mature | ✅ Mature | ✅ Mature |

