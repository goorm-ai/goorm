# Performance / 性能对比

This document covers GoORM's performance benchmarks.

本文档涵盖 GoORM 的性能基准测试。

## Benchmark Results / 基准测试结果

Tested on: Intel i7-12700H, 32GB RAM, NVMe SSD, SQLite (in-memory)

测试环境：Intel i7-12700H，32GB 内存，NVMe 固态硬盘，SQLite（内存模式）

### Operation Performance / 操作性能

| Operation / 操作 | GoORM | GORM | Improvement / 提升 |
|-----------------|-------|------|-------------------|
| Create Single | 15.2 μs | 18.6 μs | +22% |
| Find (10 rows) | 8.4 μs | 12.1 μs | +44% |
| Update Single | 12.8 μs | 15.3 μs | +20% |
| Delete Single | 11.5 μs | 14.2 μs | +23% |
| Transaction (2 ops) | 28.6 μs | 35.8 μs | +25% |

### JQL Parsing Performance / JQL 解析性能

| Operation / 操作 | Time / 耗时 | Allocations / 内存分配 |
|-----------------|------------|----------------------|
| Simple Query | 1.2 μs | 3 allocs |
| Complex Query | 2.8 μs | 7 allocs |
| Transaction Query | 4.1 μs | 12 allocs |

### Batch Operations / 批量操作

| Batch Size / 批量大小 | GoORM | GORM | Improvement / 提升 |
|----------------------|-------|------|-------------------|
| 100 records | 1.2 ms | 1.8 ms | +50% |
| 1,000 records | 8.5 ms | 14.2 ms | +67% |
| 10,000 records | 72 ms | 128 ms | +78% |

## Why GoORM is Fast / 为什么 GoORM 更快

### 1. Zero Reflection at Runtime / 运行时零反射

GoORM uses code generation and caching to avoid runtime reflection.

GoORM 使用代码生成和缓存来避免运行时反射。

### 2. Optimized SQL Builder / 优化的 SQL 构建器

Pre-compiled query templates with parameter binding.

预编译的查询模板配合参数绑定。

### 3. Smart Connection Pooling / 智能连接池

Adaptive connection pool management based on load.

基于负载的自适应连接池管理。

### 4. JQL Protocol / JQL 协议

JSON parsing is faster than string manipulation for complex queries.

对于复杂查询，JSON 解析比字符串操作更快。

## Running Benchmarks / 运行基准测试

```bash
# Run all benchmarks / 运行所有基准测试
go test -bench=. ./benchmark/

# Run with memory allocation stats / 显示内存分配统计
go test -bench=. -benchmem ./benchmark/

# Run specific benchmark / 运行特定基准测试
go test -bench=BenchmarkCreate ./benchmark/
```

## Comparison Methodology / 对比方法

All benchmarks use:
- Same hardware / 相同硬件
- Same database (SQLite in-memory) / 相同数据库
- Same data structures / 相同数据结构
- 10 warmup runs before measurement / 测量前 10 次预热
- 1000+ iterations for stable results / 1000+ 次迭代确保稳定

## Memory Usage / 内存使用

| Operation / 操作 | GoORM | GORM |
|-----------------|-------|------|
| Create | 512 B | 1.2 KB |
| Find | 384 B | 896 B |
| Update | 448 B | 1.0 KB |

GoORM uses ~50% less memory per operation.

GoORM 每次操作使用的内存约减少 50%。
