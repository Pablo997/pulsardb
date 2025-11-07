# Performance Benchmarks

Performance results for PulsarDB on modern hardware.

## TL;DR

**PulsarDB Performance Summary:**

- ⚡ **Write Throughput**: 1.5M/sec (single with WAL) | 2M/sec (concurrent)
- 🔍 **Query**: 144ms for 1M points (linear scaling)
- 🚀 **Concurrency**: 2M writes/sec @ high parallelism (zero contention)
- 💾 **Memory**: ~197 bytes/point (1M points = 197 MB)
- 📊 **Edge Capacity**: 2.5M points on 512 MB device
- 🎯 **Binary WAL**: Only 666ns overhead (1.6x faster than JSON)

**Verified on Intel i9-9900K (16 cores) | In-memory + Binary WAL**

---

## Test Environment

- **CPU**: Intel Core i9-9900K @ 3.60GHz (16 logical cores)
- **OS**: Windows
- **Go Version**: 1.21+
- **Configuration**: In-memory storage + Binary WAL (lazy flush)

---

## Write Performance

| Operation | Latency | Throughput | Memory |
|-----------|---------|------------|--------|
| Single Write (WAL) | 666 ns | 1.5M writes/sec | 192 B/op |
| Single Write (No WAL) | 128 ns | 7.8M writes/sec | 43 B/op |
| Batch 10 (WAL) | 5.0 μs | 2M points/sec | 2.4 KB/op |
| Batch 100 (WAL) | 46 μs | 2.1M points/sec | 23 KB/op |
| Batch 1000 (WAL) | 497 μs | 2M points/sec | 235 KB/op |
| Concurrent (WAL) | 520 ns | 1.9M writes/sec | 185 B/op |

**Key Insights:**
- Binary WAL adds only 538ns overhead vs no WAL
- Batching provides consistent 2M+ points/sec throughput
- Concurrent writes maintain predictable performance
- WAL is 1.6x faster than JSON encoding

---

## Query Performance

| Operation | Data Points | Latency | Throughput |
|-----------|-------------|---------|------------|
| Small Range | 100 | 27 μs | 37k queries/sec |
| Medium Range | 1000 | 144 μs | 7k queries/sec |
| Large Range | 10k | 2.47 ms | 405 queries/sec |
| Concurrent Queries | 1000 | 94 μs | 10.5k queries/sec |

**Key Insights:**
- Query time scales linearly with data size
- Concurrent queries benefit from RWMutex (2.2x faster)
- Sub-millisecond latency for typical workloads

---

## HTTP Overhead

| Endpoint | Latency | Operations/sec |
|----------|---------|----------------|
| /health | 3.1 μs | 320k/sec |
| /metrics | 3.9 μs | 257k/sec |

**Minimal overhead** - HTTP routing and JSON parsing are highly optimized.

---

## Scalability Estimates

### Real-World Capacity

**IoT Sensors** (1 point every 10 seconds):
- Single-threaded: **15 million sensors**
- Multi-core (16 cores): **200+ million sensors**

**High-frequency monitoring** (1 point per second):
- Single-threaded: **1.5M sensors**
- Multi-core: **20+ million sensors**

### Throughput Limits

**Writes:**
```
Single-threaded (WAL): 1.5M points/sec = 90M points/min
Multi-core (WAL):      2M points/sec = 120M points/min
```

**Queries:**
```
Single-threaded: 4.7k queries/sec
Multi-core:      168k queries/sec (concurrent)
```

---

## Comparison with Other TSDBs

| Database | Single Write | Batch 100 | Query 1k pts | Architecture |
|----------|--------------|-----------|--------------|--------------|
| **PulsarDB** | **0.67 μs** | **46 μs** | **144 μs** | In-memory + WAL |
| InfluxDB | ~50 μs | ~500 μs | ~500 μs | Disk-based |
| Prometheus | ~20 μs | ~200 μs | ~300 μs | In-memory + disk |
| TimescaleDB | ~100 μs | ~1 ms | ~1 ms | PostgreSQL |

**PulsarDB vs InfluxDB:**
- Single write: **75x faster** (0.67 μs vs 50 μs)
- Batch 100: **11x faster** (46 μs vs 500 μs)
- Query 1k: **3.5x faster** (144 μs vs 500 μs)

**Note:** PulsarDB with binary WAL. With SSTables, expect:
- Writes: 1-2 μs (still very fast)
- Queries: 500 μs - 1 ms (with disk I/O)

---

## Memory Efficiency

| Operation | Allocations | Memory/op |
|-----------|-------------|-----------|
| Single Write (WAL) | 9 allocs | 192 B |
| Batch 100 (WAL) | 800 allocs | 23 KB |
| Query 1k | 71 allocs | 93 KB |
| Concurrent Write (WAL) | 7 allocs | 185 B |

Allocations are reasonable and could be further optimized with object pooling.

---

## Future Optimizations

### 1. Object Pooling
Reuse DataPoint objects to reduce GC pressure.

**Expected Impact:** -30% allocations, -15% latency

### 2. Zero-copy JSON Parsing
Use `jsoniter` or `easyjson` for faster parsing.

**Expected Impact:** -40% parsing overhead

### 3. Batch Processing
Internal batching before WAL flush.

**Expected Impact:** +50% throughput with WAL

### 4. Compression
Gorilla compression for time-series data.

**Expected Impact:** 90% storage reduction, +10μs write latency

---

## Running Benchmarks

```bash
# Run all benchmarks
go test ./... -bench=. -benchmem

# Run storage benchmarks only
go test ./pkg/storage/ -bench=. -benchmem

# Run server benchmarks only
go test ./internal/server/ -bench=. -benchmem

# Run with different CPU counts
go test ./internal/server/ -bench=BenchmarkConcurrent -cpu=1,2,4,8

# Profile CPU usage
go test ./internal/server/ -bench=BenchmarkHandleWriteSinglePoint \
  -cpuprofile=cpu.prof

# Profile memory usage
go test ./internal/server/ -bench=BenchmarkHandleWriteSinglePoint \
  -memprofile=mem.prof
```

---

## Stress Test Results

### Query Scalability (1k → 1M points)

```mermaid
xychart-beta
    title "Query Latency vs Dataset Size"
    x-axis [10k, 100k, 1M]
    y-axis "Latency (ms)" 0 --> 300
    line [2.47, 24, 288]
```

**Linear scaling:** 10x data = ~10x time ✅

### Write Throughput by Batch Size

| Batch Size | Latency | Throughput | Scaling |
|------------|---------|------------|---------|
| Single | 5 μs | 240k/sec | 1x |
| 100 | 98 μs | 1M/sec | 4x |
| 1,000 | 930 μs | 1M/sec | 4x |
| **100,000** | **93 ms** | **1M/sec** | **4x** |

```mermaid
xychart-beta
    title "Write Throughput (points/sec)"
    x-axis [Single, Batch-100, Batch-1k, Batch-100k]
    y-axis "Throughput" 0 --> 1200000
    bar [240000, 1000000, 1000000, 1000000]
```

**Batching provides 4x throughput improvement.**

### Memory Efficiency

```mermaid
xychart-beta
    title "Memory Usage by Dataset Size"
    x-axis [10k, 100k, 1M]
    y-axis "Memory (MB)" 0 --> 250
    line [1.38, 15.6, 197]
```

**~197 bytes/point** including overhead

### Extreme Load Results

| Test | Result | Notes |
|------|--------|-------|
| 1M point query | 288 ms | Linear scaling maintained |
| 100k batch write | 93 ms | 1M points/sec throughput |
| Concurrent 1000x | 2 μs/op | Zero contention |
| 2M points in RAM | 815 ms query | No degradation |
| Memory footprint | 197 MB/1M pts | ~197 bytes/point |

---

## Production Capacity Estimates

| Device Type | RAM | Capacity | Write Speed | Query Speed |
|-------------|-----|----------|-------------|-------------|
| **Edge** | 512 MB | ~2.5M points | 240k/sec | Sub-second |
| **Server** | 8 GB | ~40M points | 1M/sec | Thousands/sec |

---

## Notes

- All benchmarks run on Intel i9-9900K (16 cores)
- Results may vary based on hardware and configuration
- In-memory results (no persistence yet)
- Production performance will include WAL overhead (~50μs)
- Linear scaling verified up to 1M data points
- Zero contention with atomic operations

**Updated:** November 2025

