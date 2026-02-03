# MCP Server / MCP 服务器

This document covers GoORM's built-in MCP (Model Context Protocol) server for AI integration.

本文档涵盖 GoORM 内置的 MCP（模型上下文协议）服务器，用于 AI 集成。

## Starting the Server / 启动服务器

```go
server := goorm.NewMCPServer(db)
server.Start(ctx)
```

## Available Tools / 可用工具

| Tool | Description / 描述 |
|------|-------------------|
| `execute_query` | Execute JQL query / 执行 JQL 查询 |
| `list_tables` | List all tables / 列出所有表 |
| `describe_table` | Get table schema / 获取表结构 |
| `find_records` | Query records / 查询记录 |
| `create_record` | Create record / 创建记录 |
| `update_records` | Update records / 更新记录 |
| `delete_records` | Delete records / 删除记录 |
| `count_records` | Count records / 统计记录数 |
| `execute_transaction` | Atomic operations / 原子操作 |
| `explain_query` | SQL preview / SQL 预览 |
| `aggregate` | Aggregations / 聚合查询 |
| `sync_schema` | Schema migration / 模式迁移 |
| `get_stats` | Database stats / 数据库统计 |
| `natural_language` | NL to JQL query / 自然语言查询 |

## Tool Examples / 工具示例

### execute_query

```json
{
    "tool": "execute_query",
    "arguments": {
        "query": "{\"table\": \"users\", \"action\": \"find\"}"
    }
}
```

### list_tables

```json
{
    "tool": "list_tables",
    "arguments": {}
}
```

### describe_table

```json
{
    "tool": "describe_table",
    "arguments": {
        "table": "users"
    }
}
```

### find_records

```json
{
    "tool": "find_records",
    "arguments": {
        "table": "users",
        "where": [{"field": "age", "op": ">", "value": 18}],
        "limit": 10
    }
}
```

### natural_language

```json
{
    "tool": "natural_language",
    "arguments": {
        "query": "Find all users older than 18 sorted by name"
    }
}
```

## Integration / 集成

The MCP server enables AI assistants to directly interact with your database using natural language or JQL.

MCP 服务器使 AI 助手能够使用自然语言或 JQL 直接与数据库交互。
