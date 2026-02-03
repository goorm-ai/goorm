# Model Definition / 模型定义

This document covers model definition in GoORM.

本文档涵盖 GoORM 中的模型定义。

## Base Model / 基础模型

GoORM provides a base `Model` struct with common fields:

GoORM 提供了包含常用字段的基础 `Model` 结构体：

```go
type Model struct {
    ID        uint64     `json:"id"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
```

## Defining Models / 定义模型

```go
type User struct {
    goorm.Model
    Name     string  `json:"name"`
    Email    string  `json:"email" goorm:"unique"`
    Age      int     `json:"age"`
    Balance  float64 `json:"balance" goorm:"default:0"`
    IsActive bool    `json:"is_active" goorm:"default:true"`
}
```

## Tags / 标签

| Tag | Description / 描述 |
|-----|-------------------|
| `json:"name"` | JSON field name / JSON 字段名 |
| `goorm:"unique"` | Unique constraint / 唯一约束 |
| `goorm:"not_null"` | Not null constraint / 非空约束 |
| `goorm:"default:value"` | Default value / 默认值 |
| `goorm:"size:100"` | Field size / 字段大小 |
| `goorm:"index"` | Create index / 创建索引 |
| `goorm:"primary_key"` | Primary key / 主键 |
| `rel:"has_one"` | Has one relation / 一对一关系 |
| `rel:"has_many"` | Has many relation / 一对多关系 |
| `rel:"belongs_to"` | Belongs to relation / 多对一关系 |

## Field Types / 字段类型

| Go Type | Database Type |
|---------|---------------|
| `string` | VARCHAR/TEXT |
| `int`, `int64` | INTEGER/BIGINT |
| `float64` | DECIMAL/DOUBLE |
| `bool` | BOOLEAN |
| `time.Time` | TIMESTAMP |
| `[]byte` | BLOB/BYTEA |

## Registering Models / 注册模型

```go
// Register single model / 注册单个模型
db.Register(&User{})

// Register multiple models / 注册多个模型
db.Register(&User{}, &Order{}, &Product{})
```

## Auto Migration / 自动迁移

```go
// Sync all registered models / 同步所有已注册模型
db.AutoSync()
```
