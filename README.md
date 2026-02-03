<p align="center">
  <h1 align="center">GoORM</h1>
  <p align="center">The AI-Native ORM library for Go</p>
  <p align="center">
    <a href="https://pkg.go.dev/github.com/goorm-ai/goorm"><img src="https://pkg.go.dev/badge/github.com/goorm-ai/goorm.svg" alt="Go Reference"></a>
    <a href="https://goreportcard.com/report/github.com/goorm-ai/goorm"><img src="https://goreportcard.com/badge/github.com/goorm-ai/goorm" alt="Go Report Card"></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
  </p>
  <p align="center">
    English | <a href="README_zh.md">中文</a>
  </p>
</p>

## Overview

GoORM is the first **pure AI-native** ORM library for Go. It uses **JQL (JSON Query Language)** as the core protocol, designed specifically for AI generation and understanding while providing all traditional ORM capabilities.

## Features

### Full-Featured ORM

- **CRUD Operations** - Create, Read, Update, Delete with JQL or Go API
- **Associations** - Has One, Has Many, Belongs To, Many To Many
- **Hooks** - Before/After Create, Update, Delete, Find
- **Eager Loading** - Preload related data with `with` clause
- **Transactions** - Full transaction support with atomic operations
- **Batch Operations** - Batch Insert, Batch Update, FindInBatches
- **SQL Builder** - Upsert, Locking, Aggregations, Subqueries
- **Auto Migrations** - Automatic schema sync with safe migration options
- **Soft Delete** - Global or per-table soft delete support

### AI-Native Design

- **JQL Protocol** - JSON Query Language designed for AI generation
- **Natural Language** - Built-in natural language to JQL conversion
- **MCP Server** - 14 built-in AI tools via Model Context Protocol
- **Query Optimizer** - Static analysis with optimization hints

### Smart Features

- **Health Monitoring** - Background health checks with status reporting
- **Metrics Collection** - Query performance and statistics tracking
- **Connection Pool** - Configurable connection pool management
- **Multi-Database** - PostgreSQL, MySQL, SQLite support

### Developer Friendly

- **Bilingual Docs** - Full English and Chinese documentation
- **Type Safe** - Strong typing with Go generics
- **Extensible** - Plugin architecture for custom extensions
- **Well Tested** - Comprehensive test coverage

## Installation

```bash
go get -u github.com/goorm-ai/goorm
```

## Quick Start

```go
import "github.com/goorm-ai/goorm"

// Connect to database
db, err := goorm.Connect("postgres://user:pass@localhost:5432/dbname?sslmode=disable")

// Define and register model
type User struct {
    goorm.Model
    Name  string `json:"name"`
    Email string `json:"email" goorm:"unique"`
}

db.Register(&User{})
db.AutoSync()

// Query with JQL
result := db.Query(`{"table": "users", "action": "find", "where": [{"field": "age", "op": ">", "value": 18}]}`)

// Or use natural language
result := db.NL("Find all active users")
```

## Documentation

- [Getting Started](docs/getting-started.md)
- [JQL Reference](docs/jql-reference.md)
- [Model Definition](docs/models.md)
- [CRUD Operations](docs/crud.md)
- [Associations](docs/associations.md)
- [Transactions](docs/transactions.md)
- [Hooks](docs/hooks.md)
- [MCP Server](docs/mcp-server.md)
- [Configuration](docs/configuration.md)
- [Performance](docs/performance.md)

## Database Support

| Database   | Status       |
|------------|--------------|
| PostgreSQL | Full Support |
| MySQL      | Full Support |
| SQLite     | Full Support |

## MCP Tools

GoORM provides 14 built-in AI tools via MCP:

| Tool | Description |
|------|-------------|
| `execute_query` | Execute JQL query |
| `list_tables` | List all tables |
| `describe_table` | Get table schema |
| `find_records` | Query records |
| `create_record` | Create record |
| `update_records` | Update records |
| `delete_records` | Delete records |
| `count_records` | Count records |
| `execute_transaction` | Atomic operations |
| `explain_query` | SQL preview |
| `aggregate` | Aggregations |
| `sync_schema` | Schema migration |
| `get_stats` | Database stats |
| `natural_language` | NL to JQL query |

## Performance

GoORM is designed for high performance with minimal overhead.

| Operation | GoORM | GORM | Improvement |
|-----------|-------|------|-------------|
| Create | 15.2 μs | 18.6 μs | **+22%** |
| Find (10 rows) | 8.4 μs | 12.1 μs | **+44%** |
| Update | 12.8 μs | 15.3 μs | **+20%** |
| Transaction | 28.6 μs | 35.8 μs | **+25%** |

- **~50% less memory** per operation
- **Zero reflection** at runtime
- **Pre-compiled** query templates

See [Performance Benchmarks](docs/performance.md) for detailed results.

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md).

## License

GoORM is released under the [MIT License](LICENSE).

---

<p align="center">Made with ❤️ for the AI era</p>
