# PulsarDB

A lightweight, high-performance time-series database designed for edge computing and IoT applications.

## Features

- 🚀 **Fast**: Optimized for high-throughput writes and low-latency queries
- 💾 **Efficient**: Memory-first architecture with automatic compression
- 🔌 **Simple**: Easy-to-use HTTP API
- 📦 **Embedded**: Can run as standalone server or embedded library
- ⚡ **Edge-Optimized**: Minimal resource footprint for IoT and edge devices

## Quick Start

### Installation

```bash
go get github.com/Pablo997/pulsardb
```

### Running the Server

```bash
# Using default configuration
go run cmd/pulsardb/main.go

# With custom config
go run cmd/pulsardb/main.go -config config.json
```

### Configuration

Create a `config.json` file:

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
    "compression_enabled": true
  }
}
```

## API Endpoints

### Write Data Point

```bash
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

### Query Data

```bash
POST /query
Content-Type: application/json

{
  "metric": "temperature",
  "start": 1699267200000,
  "end": 1699353600000,
  "tags": {
    "sensor": "sensor1"
  }
}
```

### Health Check

```bash
GET /health
```

### Metrics

```bash
GET /metrics
```

## Architecture

PulsarDB uses a LSM-tree (Log-Structured Merge-tree) inspired architecture:

1. **MemTable**: In-memory buffer for recent writes
2. **WAL**: Write-Ahead Log for durability
3. **SSTables**: Sorted String Tables for persistent storage
4. **Compaction**: Background merge and compression

## Development Status

🚧 **Early Development** - This is a work in progress.

### Implemented
- [x] Basic project structure
- [x] HTTP server with routing
- [x] Configuration management
- [x] MemTable (in-memory storage)

### TODO
- [ ] Write-Ahead Log (WAL)
- [ ] SSTable implementation
- [ ] Compaction strategy
- [ ] Query engine
- [ ] Compression algorithms
- [ ] Tag indexing
- [ ] Aggregation functions
- [ ] Client libraries
- [ ] Benchmarks
- [ ] Documentation

## Building

```bash
# Build binary
go build -o pulsardb cmd/pulsardb/main.go

# Run tests
go test ./...

# Build for production
go build -ldflags "-s -w" -o pulsardb cmd/pulsardb/main.go
```

## License

MIT License - See LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

