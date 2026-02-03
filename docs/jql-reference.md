# JQL Reference / JQL 参考

JQL (JSON Query Language) is GoORM's core protocol for database operations.

JQL（JSON 查询语言）是 GoORM 进行数据库操作的核心协议。

## Actions / 操作类型

| Action | Description / 描述 |
|--------|-------------------|
| `find` | Query records / 查询记录 |
| `create` | Insert single record / 插入单条记录 |
| `create_batch` | Insert multiple records / 批量插入记录 |
| `update` | Update records / 更新记录 |
| `delete` | Delete records / 删除记录 |
| `count` | Count records / 统计记录数 |
| `aggregate` | Aggregation (SUM, AVG, etc.) / 聚合运算 |
| `transaction` | Atomic operations / 原子操作 |

## Operators / 操作符

| Operator | Description / 描述 |
|----------|-------------------|
| `=`, `!=` | Equal, Not equal / 等于、不等于 |
| `>`, `>=`, `<`, `<=` | Comparison / 比较运算 |
| `in`, `not_in` | Array membership / 数组包含 |
| `like`, `ilike` | Pattern match / 模式匹配 |
| `between` | Range / 范围查询 |
| `null`, `not_null` | Null check / 空值检查 |

## Query Structure / 查询结构

```json
{
    "table": "users",
    "action": "find",
    "select": ["id", "name", "email"],
    "where": [
        {"field": "age", "op": ">", "value": 18},
        {"field": "status", "op": "=", "value": "active"}
    ],
    "order_by": [{"field": "created_at", "desc": true}],
    "limit": 10,
    "offset": 0,
    "with": ["orders", "profile"]
}
```

## Examples / 示例

### Find / 查询

```json
{
    "table": "users",
    "action": "find",
    "where": [{"field": "age", "op": ">", "value": 18}],
    "order_by": [{"field": "created_at", "desc": true}],
    "limit": 10
}
```

### Create / 创建

```json
{
    "table": "users",
    "action": "create",
    "data": {"name": "张三", "email": "zhangsan@example.com", "age": 25}
}
```

### Update / 更新

```json
{
    "table": "users",
    "action": "update",
    "where": [{"field": "id", "op": "=", "value": 1}],
    "data": {"name": "李四", "age": 26}
}
```

### Delete / 删除

```json
{
    "table": "users",
    "action": "delete",
    "where": [{"field": "id", "op": "=", "value": 1}]
}
```

### Transaction / 事务

```json
{
    "action": "transaction",
    "operations": [
        {"table": "users", "action": "create", "data": {"name": "小明"}, "ref": "new_user"},
        {"table": "orders", "action": "create", "data": {"user_id": "$new_user.id", "total": 100}}
    ]
}
```

### Aggregation / 聚合

```json
{
    "table": "orders",
    "action": "aggregate",
    "select": [{"sum": "total"}, {"count": "*"}],
    "group_by": ["status"],
    "having": [{"field": {"sum": "total"}, "op": ">", "value": 1000}]
}
```

### Complex Conditions / 复杂条件

```json
{
    "table": "users",
    "action": "find",
    "where": [
        {"field": "status", "op": "=", "value": "active"},
        {"or": [
            {"field": "role", "op": "=", "value": "admin"},
            {"field": "age", "op": ">=", "value": 21}
        ]}
    ]
}
```

### Eager Loading / 预加载

```json
{
    "table": "users",
    "action": "find",
    "with": ["profile", "orders"],
    "limit": 5
}
```
