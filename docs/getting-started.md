# Getting Started / 快速入门

This guide will help you get started with GoORM.

本指南将帮助您快速上手 GoORM。

## Installation / 安装

```bash
go get -u github.com/goorm-ai/goorm
```

## Connecting to Database / 连接数据库

GoORM supports PostgreSQL, MySQL, and SQLite.

GoORM 支持 PostgreSQL、MySQL 和 SQLite。

```go
import "github.com/goorm-ai/goorm"

// PostgreSQL (requires: github.com/lib/pq)
db, err := goorm.Connect("postgres://user:pass@localhost:5432/dbname?sslmode=disable")

// MySQL (requires: github.com/go-sql-driver/mysql)
db, err := goorm.Connect("mysql://user:pass@localhost:3306/dbname")

// SQLite (requires: modernc.org/sqlite or github.com/mattn/go-sqlite3)
// SQLite（需要：modernc.org/sqlite 或 github.com/mattn/go-sqlite3）
import _ "modernc.org/sqlite" // Pure Go, recommended / 纯 Go 实现，推荐
db, err := goorm.Connect("sqlite://./data.db")
```

## Defining Models / 定义模型

```go
type User struct {
    goorm.Model
    Name   string   `json:"name"`
    Email  string   `json:"email" goorm:"unique"`
    Age    int      `json:"age"`
    Orders []*Order `rel:"has_many" model:"orders"`
}

type Order struct {
    goorm.Model
    UserID uint64  `json:"user_id"`
    Total  float64 `json:"total"`
    Status string  `json:"status"`
    User   *User   `rel:"belongs_to" model:"users" fk:"user_id"`
}
```

## Registering Models / 注册模型

```go
db.Register(&User{})
db.Register(&Order{})
```

## Auto Migration / 自动迁移

```go
// Sync all registered models to database
// 同步所有已注册的模型到数据库
db.AutoSync()
```

## Basic Query / 基础查询

```go
// Using JQL / 使用 JQL
result := db.Query(`{
    "table": "users",
    "action": "find",
    "where": [{"field": "age", "op": ">", "value": 18}],
    "limit": 10
}`)

// Using natural language / 使用自然语言
result := db.NL("Find all users older than 18")
```

## Next Steps / 下一步

- [JQL Reference](jql-reference.md) - Learn JQL syntax / 学习 JQL 语法
- [CRUD Operations](crud.md) - Detailed CRUD examples / 详细的增删改查示例
- [Associations](associations.md) - Working with relations / 使用关联关系
