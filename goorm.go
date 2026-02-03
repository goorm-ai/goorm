// Package goorm provides an AI-First ORM framework for Go.
// It uses JSON Query Language (JQL) as the core API, making it easy for
// AI systems to generate and execute database operations.
//
// GoORM 是一个 AI 优先的 Go 语言 ORM 框架。
// 它使用 JSON 查询语言（JQL）作为核心 API，让 AI 系统能够轻松生成和执行数据库操作。
//
// # Quick Start / 快速开始
//
//	db := goorm.Connect("postgres://user:pass@localhost/mydb")
//	db.Register(&User{})
//	db.AutoSync()
//
//	// Execute JQL / 执行 JQL
//	result := db.Execute(`{
//	    "table": "users",
//	    "action": "find",
//	    "where": [{"field": "age", "op": ">", "value": 18}]
//	}`)
//
//	// Or use natural language / 或使用自然语言
//	result := db.NL("查找所有18岁以上的用户")
//
// # Features / 特性
//
//   - JSON Query Language (JQL) - AI can directly generate queries
//   - Natural Language Interface - Ultimate simplification
//   - Zero-value updates - Native support, no map/Select needed
//   - Aggressive migration - Database matches models exactly
//   - Auto backup - Data backed up before destructive operations
//   - MCP Protocol - 14 tools for AI Agents
//
// For more information, see https://github.com/goorm-ai/goorm
package goorm

// Version is the current version of GoORM.
// 当前 GoORM 版本号
const Version = "0.1.0"
