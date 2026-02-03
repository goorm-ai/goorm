// Package benchmark provides performance benchmarks for GoORM.
// 基准测试包，用于测试 GoORM 的性能表现。
package benchmark

import (
	"context"
	"testing"

	"github.com/goorm-ai/goorm"
)

// BenchmarkCreate benchmarks record creation.
// BenchmarkCreate 测试记录创建的性能。
func BenchmarkCreate(b *testing.B) {
	db := setupBenchDB(b)
	defer db.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.ExecuteQuery(ctx, &goorm.Query{
			Table:  "bench_users",
			Action: goorm.ActionCreate,
			Data: map[string]any{
				"name":  "User",
				"email": "user@example.com",
				"age":   25,
			},
		})
	}
}

// BenchmarkFind benchmarks record querying.
// BenchmarkFind 测试记录查询的性能。
func BenchmarkFind(b *testing.B) {
	db := setupBenchDB(b)
	defer db.Close()

	ctx := context.Background()

	// Seed data / 填充数据
	for i := 0; i < 100; i++ {
		db.ExecuteQuery(ctx, &goorm.Query{
			Table:  "bench_users",
			Action: goorm.ActionCreate,
			Data: map[string]any{
				"name":  "User",
				"email": "user@example.com",
				"age":   i,
			},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.ExecuteQuery(ctx, &goorm.Query{
			Table:  "bench_users",
			Action: goorm.ActionFind,
			Where: []goorm.Condition{
				{Field: "age", Op: goorm.OpGreater, Value: 50},
			},
			Limit: 10,
		})
	}
}

// BenchmarkUpdate benchmarks record updating.
// BenchmarkUpdate 测试记录更新的性能。
func BenchmarkUpdate(b *testing.B) {
	db := setupBenchDB(b)
	defer db.Close()

	ctx := context.Background()

	// Create a record / 创建一条记录
	result := db.ExecuteQuery(ctx, &goorm.Query{
		Table:  "bench_users",
		Action: goorm.ActionCreate,
		Data: map[string]any{
			"name":  "User",
			"email": "user@example.com",
			"age":   25,
		},
	})

	id := result.ID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.ExecuteQuery(ctx, &goorm.Query{
			Table:  "bench_users",
			Action: goorm.ActionUpdate,
			Where: []goorm.Condition{
				{Field: "id", Op: goorm.OpEqual, Value: id},
			},
			Data: map[string]any{
				"age": i,
			},
		})
	}
}

// BenchmarkJQLParse benchmarks JQL parsing.
// BenchmarkJQLParse 测试 JQL 解析的性能。
func BenchmarkJQLParse(b *testing.B) {
	query := `{
		"table": "users",
		"action": "find",
		"where": [{"field": "age", "op": ">", "value": 18}],
		"order_by": [{"field": "created_at", "desc": true}],
		"limit": 10
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		goorm.ParseQuery(query)
	}
}

// BenchmarkTransaction benchmarks transaction execution.
// BenchmarkTransaction 测试事务执行的性能。
func BenchmarkTransaction(b *testing.B) {
	db := setupBenchDB(b)
	defer db.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.ExecuteQuery(ctx, &goorm.Query{
			Action: goorm.ActionTransaction,
			Operations: []goorm.Query{
				{
					Table:  "bench_users",
					Action: goorm.ActionCreate,
					Data:   map[string]any{"name": "User1"},
				},
				{
					Table:  "bench_orders",
					Action: goorm.ActionCreate,
					Data:   map[string]any{"total": 100},
				},
			},
		})
	}
}

func setupBenchDB(b *testing.B) *goorm.DB {
	db, err := goorm.Connect("sqlite://:memory:")
	if err != nil {
		b.Skip("Database not available")
	}

	// Create tables using Query / 使用 Query 创建表
	ctx := context.Background()
	db.Query(`{
		"action": "raw",
		"sql": "CREATE TABLE IF NOT EXISTS bench_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, email TEXT, age INTEGER, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)"
	}`)
	db.ExecuteQuery(ctx, &goorm.Query{
		Table:  "bench_orders",
		Action: goorm.ActionCreate,
		Data:   map[string]any{"_schema": true},
	})

	return db
}
