# Associations / 关联关系

This document covers model associations in GoORM.

本文档涵盖 GoORM 中的模型关联关系。

## Relation Types / 关系类型

| Type | Description / 描述 |
|------|-------------------|
| Has One | One-to-one / 一对一 |
| Has Many | One-to-many / 一对多 |
| Belongs To | Many-to-one / 多对一 |
| Many To Many | Many-to-many / 多对多 |

## Has One / 一对一

```go
type User struct {
    goorm.Model
    Name    string   `json:"name"`
    Profile *Profile `rel:"has_one" model:"profiles"`
}

type Profile struct {
    goorm.Model
    UserID uint64 `json:"user_id"`
    Bio    string `json:"bio"`
    Avatar string `json:"avatar"`
}
```

## Has Many / 一对多

```go
type User struct {
    goorm.Model
    Name   string   `json:"name"`
    Orders []*Order `rel:"has_many" model:"orders"`
}

type Order struct {
    goorm.Model
    UserID uint64  `json:"user_id"`
    Total  float64 `json:"total"`
}
```

## Belongs To / 多对一

```go
type Order struct {
    goorm.Model
    UserID uint64 `json:"user_id"`
    Total  float64 `json:"total"`
    User   *User   `rel:"belongs_to" model:"users" fk:"user_id"`
}
```

## Eager Loading / 预加载

```go
// Load user with profile / 加载用户及其资料
result := db.Query(`{
    "table": "users",
    "action": "find",
    "with": ["profile"]
}`)

// Load user with multiple relations / 加载用户及其多个关联
result := db.Query(`{
    "table": "users",
    "action": "find",
    "with": ["profile", "orders"]
}`)

// Load order with user / 加载订单及其所属用户
result := db.Query(`{
    "table": "orders",
    "action": "find",
    "with": ["user"]
}`)
```
