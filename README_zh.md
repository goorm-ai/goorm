<p align="center">
  <h1 align="center">GoORM</h1>
  <p align="center">Go 语言 AI 原生 ORM 框架</p>
  <p align="center">
    <a href="https://pkg.go.dev/github.com/goorm-ai/goorm"><img src="https://pkg.go.dev/badge/github.com/goorm-ai/goorm.svg" alt="Go 文档"></a>
    <a href="https://goreportcard.com/report/github.com/goorm-ai/goorm"><img src="https://goreportcard.com/badge/github.com/goorm-ai/goorm" alt="代码质量"></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="许可证"></a>
  </p>
  <p align="center">
    <a href="README.md">English</a> | 中文
  </p>
</p>

## 概述

GoORM 是首个**纯 AI 原生**的 Go 语言 ORM 框架。它使用 **JQL（JSON 查询语言）**作为核心协议，专为 AI 生成和理解而设计，同时提供所有传统 ORM 功能。

## 功能特性

### 完整的 ORM 功能

- **增删改查** - 通过 JQL 或 Go API 进行 Create、Read、Update、Delete 操作
- **关联关系** - 一对一、一对多、多对一、多对多
- **钩子函数** - 创建、更新、删除、查询的前置/后置钩子
- **预加载** - 通过 `with` 子句预加载关联数据
- **事务支持** - 完整的事务支持，支持原子操作
- **批量操作** - 批量插入、批量更新、分批查询
- **SQL 构建器** - Upsert、锁定、聚合、子查询
- **自动迁移** - 自动同步表结构，提供安全迁移选项
- **软删除** - 全局或单表软删除支持

### AI 原生设计

- **JQL 协议** - 专为 AI 生成设计的 JSON 查询语言
- **自然语言** - 内置自然语言到 JQL 的转换功能
- **MCP 服务器** - 通过模型上下文协议提供 14 个内置 AI 工具
- **查询优化器** - 静态分析并提供优化建议

### 智能功能

- **健康监控** - 后台健康检查与状态报告
- **指标采集** - 查询性能和统计数据追踪
- **连接池** - 可配置的连接池管理
- **多数据库** - 支持 PostgreSQL、MySQL、SQLite

### 开发者友好

- **双语文档** - 完整的中英文文档
- **类型安全** - 使用 Go 泛型实现强类型
- **可扩展** - 插件架构支持自定义扩展
- **完善测试** - 全面的测试覆盖

## 安装

```bash
go get -u github.com/goorm-ai/goorm
```

## 快速开始

```go
import "github.com/goorm-ai/goorm"

// 连接数据库
db, err := goorm.Connect("postgres://user:pass@localhost:5432/dbname?sslmode=disable")

// 定义并注册模型
type User struct {
    goorm.Model
    Name  string `json:"name"`
    Email string `json:"email" goorm:"unique"`
}

db.Register(&User{})
db.AutoSync()

// 使用 JQL 查询
result := db.Query(`{"table": "users", "action": "find", "where": [{"field": "age", "op": ">", "value": 18}]}`)

// 或使用自然语言
result := db.NL("查找所有活跃用户")
```

## 文档

- [快速入门](docs/getting-started.md)
- [JQL 参考](docs/jql-reference.md)
- [模型定义](docs/models.md)
- [增删改查](docs/crud.md)
- [关联关系](docs/associations.md)
- [事务处理](docs/transactions.md)
- [钩子函数](docs/hooks.md)
- [MCP 服务器](docs/mcp-server.md)
- [配置选项](docs/configuration.md)
- [性能对比](docs/performance.md)

## 数据库支持

| 数据库 | 状态 |
|--------|------|
| PostgreSQL | 完全支持 |
| MySQL | 完全支持 |
| SQLite | 完全支持 |

## MCP 工具

GoORM 通过 MCP 提供 14 个内置 AI 工具：

| 工具 | 描述 |
|------|------|
| `execute_query` | 执行 JQL 查询 |
| `list_tables` | 列出所有表 |
| `describe_table` | 获取表结构 |
| `find_records` | 查询记录 |
| `create_record` | 创建记录 |
| `update_records` | 更新记录 |
| `delete_records` | 删除记录 |
| `count_records` | 统计记录数 |
| `execute_transaction` | 原子操作 |
| `explain_query` | SQL 预览 |
| `aggregate` | 聚合查询 |
| `sync_schema` | 模式迁移 |
| `get_stats` | 数据库统计 |
| `natural_language` | 自然语言查询 |

## 性能对比

GoORM 专为高性能设计，开销极小。

| 操作 | GoORM | GORM | 提升 |
|------|-------|------|------|
| 创建 | 15.2 μs | 18.6 μs | **+22%** |
| 查询 (10行) | 8.4 μs | 12.1 μs | **+44%** |
| 更新 | 12.8 μs | 15.3 μs | **+20%** |
| 事务 | 28.6 μs | 35.8 μs | **+25%** |

- **内存减少约 50%**
- **运行时零反射**
- **预编译查询模板**

详见 [性能基准测试](docs/performance.md)。

## 贡献

欢迎贡献代码！请先阅读我们的[贡献指南](CONTRIBUTING.md)。

## 许可证

GoORM 基于 [MIT 许可证](LICENSE) 发布。

---

<p align="center">为 AI 时代倾心打造 ❤️</p>
