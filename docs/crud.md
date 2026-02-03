# CRUD Operations / 增删改查

This document covers all CRUD operations in GoORM.

本文档涵盖 GoORM 中所有的增删改查操作。

## Create / 创建

### Single Record / 单条记录

```go
result := db.Query(`{
    "table": "users",
    "action": "create",
    "data": {"name": "张三", "email": "zhangsan@example.com", "age": 25}
}`)

// Or using Go API / 或使用 Go API
result := db.ExecuteQuery(ctx, &goorm.Query{
    Table:  "users",
    Action: goorm.ActionCreate,
    Data: map[string]any{
        "name":  "张三",
        "email": "zhangsan@example.com",
        "age":   25,
    },
})
```

### Batch Insert / 批量插入

```go
result := db.Query(`{
    "table": "users",
    "action": "create_batch",
    "data": [
        {"name": "用户1", "email": "user1@example.com"},
        {"name": "用户2", "email": "user2@example.com"},
        {"name": "用户3", "email": "user3@example.com"}
    ]
}`)
```

## Read / 查询

### Find All / 查询全部

```go
result := db.Query(`{
    "table": "users",
    "action": "find"
}`)
```

### Find with Conditions / 条件查询

```go
result := db.Query(`{
    "table": "users",
    "action": "find",
    "where": [
        {"field": "age", "op": ">", "value": 18},
        {"field": "status", "op": "=", "value": "active"}
    ]
}`)
```

### Select Columns / 选择列

```go
result := db.Query(`{
    "table": "users",
    "action": "find",
    "select": ["id", "name", "email"]
}`)
```

### Order and Limit / 排序和分页

```go
result := db.Query(`{
    "table": "users",
    "action": "find",
    "order_by": [{"field": "created_at", "desc": true}],
    "limit": 10,
    "offset": 20
}`)
```

### Count / 统计

```go
result := db.Query(`{
    "table": "users",
    "action": "count",
    "where": [{"field": "status", "op": "=", "value": "active"}]
}`)
```

## Update / 更新

### Update with Conditions / 条件更新

```go
result := db.Query(`{
    "table": "users",
    "action": "update",
    "where": [{"field": "id", "op": "=", "value": 1}],
    "data": {"name": "新名字", "age": 30}
}`)
```

### Batch Update / 批量更新

```go
result := db.Query(`{
    "table": "users",
    "action": "update",
    "where": [{"field": "status", "op": "=", "value": "pending"}],
    "data": {"status": "active"}
}`)
```

## Delete / 删除

### Delete with Conditions / 条件删除

```go
result := db.Query(`{
    "table": "users",
    "action": "delete",
    "where": [{"field": "id", "op": "=", "value": 1}]
}`)
```

### Soft Delete / 软删除

```go
// Enable soft delete globally / 全局启用软删除
db.EnableSoftDeleteGlobal("")

// Now delete will set deleted_at instead of removing
// 现在删除操作会设置 deleted_at 而不是真正删除
result := db.Query(`{
    "table": "users",
    "action": "delete",
    "where": [{"field": "id", "op": "=", "value": 1}]
}`)
```

## Aggregations / 聚合

```go
result := db.Query(`{
    "table": "orders",
    "action": "aggregate",
    "select": [
        {"sum": "total"},
        {"avg": "total"},
        {"count": "*"},
        {"min": "total"},
        {"max": "total"}
    ],
    "group_by": ["status"]
}`)
```
