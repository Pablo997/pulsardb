# PulsarDB

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Tests](https://img.shields.io/badge/Tests-Passing-success)](https://github.com/Pablo997/pulsardb)
[![Coverage](https://img.shields.io/badge/Coverage-90%25-brightgreen)](https://github.com/Pablo997/pulsardb)

A lightweight, high-performance time-series database designed for edge computing and IoT applications.

## Features

- 🚀 High-throughput writes and low-latency queries
- 💾 Memory-first architecture with smart caching
- 🔌 Simple HTTP REST API
- ⚡ Minimal resource footprint for edge devices
- 🔒 Thread-safe concurrent operations
- 📊 Real-time metrics monitoring

## Quick Start

```bash
# Clone repository
git clone https://github.com/Pablo997/pulsardb.git
cd pulsardb

# Install dependencies
go mod tidy

# Run server
go run cmd/pulsardb/main.go
```

Server starts on `http://localhost:8080`

## Basic Usage

### Write Data
```bash
curl -X POST http://localhost:8080/write \
  -H "Content-Type: application/json" \
  -d '{
    "metric": "temperature",
    "timestamp": 1699267200000,
    "value": 23.5,
    "tags": {"sensor": "sensor1"}
  }'
```

### Query Data
```bash
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{
    "metric": "temperature",
    "start": 1699267000000,
    "end": 1699267300000
  }'
```

## Documentation

- **[API Reference](docs/API.md)** - Complete API documentation
- **[Examples](docs/examples.md)** - Usage examples (curl, PowerShell, code)
- **[Architecture](docs/architecture.md)** - System design and internals
- **[Benchmarks](docs/benchmarks.md)** - Performance metrics and comparisons

## Configuration

Default config at startup. Custom config:

```bash
go run cmd/pulsardb/main.go -config config.json
```

See [config.dev.json](config.dev.json) for example.

## Development Status

**v0.2.0** - Core features + WAL implemented, well-tested (90% coverage).

### ✅ Implemented
- Write/Query HTTP endpoints (1.5M+ writes/sec)
- **Binary WAL for durability (~650ns overhead)**
- In-memory storage engine (verified up to 1M points)
- Real-time metrics tracking (atomic operations)
- Thread-safe concurrent operations (zero contention)
- Comprehensive test suite (90%+ coverage)
- Performance optimizations (binary encoding, lazy flush, pre-allocation)

### 🚧 Next
- Persistent storage (SSTables)
- Compression algorithms (Gorilla)
- Tag filtering
- Aggregation functions
- Compaction strategies

[Full roadmap](docs/roadmap.md)

## Building & Testing

```bash
# Build binary
go build -o pulsardb cmd/pulsardb/main.go

# Run tests
go test ./... -v

# Test coverage
go test ./... -cover

# Production build
go build -ldflags "-s -w" -o pulsardb cmd/pulsardb/main.go
```

## Contributing

Contributions are welcome! Please check [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - See [LICENSE](LICENSE) for details.

---

**Note:** This project is under active development. APIs may change.

