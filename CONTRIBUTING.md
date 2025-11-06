# Contributing to PulsarDB

Thank you for your interest in contributing! 🎉

---

## Getting Started

1. **Fork the repository**
2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/pulsardb.git
   cd pulsardb
   ```
3. **Create a branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

---

## Development Setup

### Prerequisites
- Go 1.21 or higher
- Git
- Make (optional)

### Install Dependencies
```bash
go mod tidy
```

### Run Tests
```bash
go test ./...
```

### Run Server
```bash
go run cmd/pulsardb/main.go
```

---

## Code Guidelines

### Go Style
- Follow official [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- Use `gofmt` to format code
- Run `go vet` before committing

### Project Structure
```
cmd/        - Executable binaries
internal/   - Private code (not importable)
pkg/        - Public libraries
docs/       - Documentation
```

### Naming Conventions
- **Files:** lowercase with underscores (e.g., `memtable.go`)
- **Functions:** PascalCase for exported, camelCase for private
- **Variables:** camelCase
- **Constants:** PascalCase or SCREAMING_SNAKE_CASE

### Comments
- All exported functions must have comments
- Comments should explain "why", not "what"

Good:
```go
// Write appends a data point to the WAL before storing in memory
// to ensure durability in case of crashes.
func (e *Engine) Write(point *DataPoint) error {
```

Bad:
```go
// Write writes a point
func (e *Engine) Write(point *DataPoint) error {
```

---

## Testing

### Unit Tests
- Test files: `*_test.go`
- Table-driven tests preferred

Example:
```go
func TestMemTableInsert(t *testing.T) {
    tests := []struct {
        name    string
        point   *DataPoint
        wantErr bool
    }{
        {"valid point", &DataPoint{...}, false},
        {"nil point", nil, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

### Integration Tests
Coming soon

### Benchmarks
```go
func BenchmarkMemTableInsert(b *testing.B) {
    mt := NewMemTable(512)
    point := &DataPoint{...}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        mt.Insert(point)
    }
}
```

---

## Pull Request Process

1. **Create an issue first** (for non-trivial changes)
2. **Write tests** for new functionality
3. **Update documentation** if needed
4. **Ensure tests pass**
   ```bash
   go test ./...
   go vet ./...
   ```
5. **Commit with clear messages**
   ```
   Add WAL implementation for durability
   
   - Append-only log file
   - Crash recovery on startup
   - Fixes #123
   ```
6. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```
7. **Open a Pull Request**

### PR Title Format
```
[Type] Brief description

Types:
- feat: New feature
- fix: Bug fix
- docs: Documentation
- refactor: Code refactoring
- test: Tests
- chore: Maintenance
```

Examples:
- `[feat] Add WAL implementation`
- `[fix] Resolve race condition in MemTable`
- `[docs] Update API documentation`

---

## What to Contribute

### Good First Issues
- Documentation improvements
- Code comments
- Example code
- Bug fixes
- Test coverage

### Priority Areas
Check [roadmap.md](docs/roadmap.md) for current priorities:
- WAL implementation
- SSTable format
- Compression algorithms
- Client libraries

### Ideas Welcome
- Performance optimizations
- New features
- API improvements
- Developer tools

---

## Code Review

All submissions require review. We aim to review PRs within:
- Simple fixes: 1-2 days
- Features: 3-7 days
- Major changes: 1-2 weeks

### Review Criteria
- ✅ Code quality and style
- ✅ Test coverage
- ✅ Documentation
- ✅ Performance impact
- ✅ Backward compatibility

---

## Communication

- **GitHub Issues:** Bug reports, feature requests
- **GitHub Discussions:** Questions, ideas
- **Pull Requests:** Code contributions

Be respectful and constructive. We follow the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/).

---

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

## Questions?

Open an issue or discussion on GitHub. We're happy to help!

Thank you for contributing to PulsarDB! 🚀

