# Transactions / 事务处理

This document covers transaction handling in GoORM.

本文档涵盖 GoORM 中的事务处理。

## Basic Transaction / 基础事务

Use the `transaction` action to execute multiple operations atomically.

使用 `transaction` 操作来原子地执行多个操作。

```go
result := db.Query(`{
    "action": "transaction",
    "operations": [
        {"table": "users", "action": "create", "data": {"name": "张三"}},
        {"table": "orders", "action": "create", "data": {"total": 100}}
    ]
}`)
```

## Reference Results / 引用结果

Use `ref` to name an operation and `$ref.field` to reference its result.

使用 `ref` 命名操作，使用 `$ref.field` 引用其结果。

```go
result := db.Query(`{
    "action": "transaction",
    "operations": [
        {
            "table": "users",
            "action": "create",
            "data": {"name": "小明"},
            "ref": "new_user"
        },
        {
            "table": "orders",
            "action": "create",
            "data": {"user_id": "$new_user.id", "total": 100}
        }
    ]
}`)
```

## Complex Transaction / 复杂事务

```go
result := db.Query(`{
    "action": "transaction",
    "operations": [
        {
            "table": "accounts",
            "action": "update",
            "where": [{"field": "id", "op": "=", "value": 1}],
            "data": {"balance": {"$decrement": 100}}
        },
        {
            "table": "accounts",
            "action": "update",
            "where": [{"field": "id", "op": "=", "value": 2}],
            "data": {"balance": {"$increment": 100}}
        },
        {
            "table": "transfers",
            "action": "create",
            "data": {"from_id": 1, "to_id": 2, "amount": 100}
        }
    ]
}`)
```

## Transaction Behavior / 事务行为

- All operations succeed or all fail / 所有操作要么全部成功，要么全部失败
- Automatic rollback on error / 错误时自动回滚
- Operations execute in order / 操作按顺序执行
