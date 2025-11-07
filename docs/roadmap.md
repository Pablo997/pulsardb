# Roadmap

PulsarDB development roadmap and future plans.

---

## ✅ Phase 1: MVP (Completed)

**Goal:** Basic functional time-series database

- [x] Project structure
- [x] HTTP server with routing
- [x] Configuration system
- [x] In-memory storage (MemTable)
- [x] Write endpoint (single & batch)
- [x] Query endpoint (time range)
- [x] Health check & metrics
- [x] Thread-safe operations
- [x] Graceful shutdown
- [x] Comprehensive test suite (90% coverage)
- [x] Performance optimizations (atomic operations, pre-allocation)
- [x] Stress benchmarks (verified up to 1M points)
- [x] Professional documentation

**Status:** ✅ Complete (November 2025)

---

## 🚧 Phase 2: Persistence (In Progress)

**Goal:** Don't lose data on restart

- [x] Write-Ahead Log (WAL)
  - Binary encoding (1.6x faster than JSON)
  - Lazy flush strategy (on memtable full)
  - ~650ns write latency
  - Crash recovery
  - Simple and predictable
- [ ] SSTable writer
  - Flush MemTable to disk
  - Sorted file format
  - Basic indexing
- [ ] SSTable reader
  - Query from disk
  - Merge memory + disk results

**Status:** WAL ✅ Complete (November 2025)  
**Target:** SSTables by Q1 2026

---

## 📋 Phase 3: Optimization

**Goal:** Production-ready performance

- [ ] Compression
  - Gorilla algorithm for time-series
  - Block-level compression
  - Configurable compression levels
- [ ] Tag indexing
  - Inverted index for tags
  - Fast tag filtering
  - Bitmap indexes
- [ ] Compaction
  - Multi-level compaction strategy
  - Background workers
  - Optimize read performance
- [ ] Query optimization
  - Index utilization
  - Query planning
  - Caching layer

**Target:** Q2 2025

---

## 🎯 Phase 4: Advanced Features

**Goal:** Feature completeness

- [ ] Aggregation functions
  - Sum, Avg, Min, Max, Count
  - Rate, Delta, Derivative
  - Time-window aggregations
- [ ] Downsampling
  - Automatic data rollup
  - Configurable retention per resolution
- [ ] Tag filtering in queries
  - Complex tag expressions
  - Regex support
- [ ] Data retention policies
  - Automatic old data deletion
  - Per-metric retention rules
- [ ] Continuous queries
  - Automatic aggregation
  - Materialized views

**Target:** Q3 2025

---

## 🔧 Phase 5: Operations

**Goal:** Easy to deploy and monitor

- [ ] Docker image
  - Official image on Docker Hub
  - Multi-arch support (amd64, arm64)
- [ ] Kubernetes deployment
  - Helm charts
  - StatefulSet examples
  - Persistent volume claims
- [ ] Observability
  - Prometheus metrics export
  - Structured logging
  - Tracing (OpenTelemetry)
- [ ] Admin API
  - Compaction triggers
  - Cache management
  - Configuration reload
- [ ] Backup/Restore
  - Snapshot mechanism
  - Incremental backups
  - Point-in-time recovery

**Target:** Q4 2025

---

## 🌐 Phase 6: Ecosystem

**Goal:** Easy integration

- [ ] Client libraries
  - Go client
  - Python client
  - JavaScript/TypeScript client
  - Java client
- [ ] Grafana plugin
  - Native datasource
  - Query builder UI
- [ ] Documentation
  - Complete API docs
  - Tutorials
  - Best practices guide
  - Performance tuning guide
- [ ] Examples
  - IoT sensor integration
  - Monitoring use cases
  - Edge deployment patterns

**Target:** 2026

---

## 🚀 Phase 7: Scale (Future)

**Goal:** Distributed time-series database

- [ ] Clustering
  - Multi-node setup
  - Raft consensus
  - Leader election
- [ ] Replication
  - Async replication
  - Configurable replication factor
- [ ] Sharding
  - Automatic data distribution
  - Shard rebalancing
- [ ] Federation
  - Cross-cluster queries
  - Global view

**Target:** TBD

---

## Community Feedback

We're open to community input! If you have feature requests or suggestions, please:

1. Open an issue on GitHub
2. Describe your use case
3. Explain why the feature is important

Popular requests will be prioritized.

---

## Performance Goals

### Phase 2 (Persistence)
- Write: 50K points/sec
- Query: 5K queries/sec
- Storage: 10GB/day compressed

### Phase 3 (Optimization)
- Write: 200K points/sec
- Query: 20K queries/sec
- Storage: 5GB/day compressed

### Phase 4 (Advanced)
- Write: 500K points/sec
- Query: 50K queries/sec
- Storage: 2GB/day compressed

### Phase 7 (Scale)
- Write: 5M points/sec (cluster)
- Query: 100K queries/sec (cluster)
- Storage: Petabyte scale

---

## Contributing

Want to help? Check out:

- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
- [Good First Issues](https://github.com/Pablo997/pulsardb/labels/good-first-issue) - Easy tasks for newcomers
- [Architecture](architecture.md) - System design

Join the development!

