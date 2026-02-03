// Package main demonstrates GoORM usage with various features.
// 主包演示 GoORM 的各种功能用法。
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/goorm-ai/goorm"
)

// User model
// 用户模型
type User struct {
	goorm.Model
	Name   string   `json:"name"`
	Email  string   `json:"email" goorm:"unique"`
	Age    int      `json:"age"`
	Orders []*Order `rel:"has_many" model:"orders"`
}

// Order model
// 订单模型
type Order struct {
	goorm.Model
	UserID uint64  `json:"user_id"`
	Total  float64 `json:"total"`
	Status string  `json:"status"`
	User   *User   `rel:"belongs_to" model:"users" fk:"user_id"`
}

func main() {
	fmt.Println("=== GoORM Demo / GoORM 演示 ===")
	fmt.Println()

	// Try to connect to database
	// 尝试连接数据库
	db, err := goorm.Connect("postgres://user:pass@localhost:5432/testdb?sslmode=disable")
	if err != nil {
		log.Printf("Database not available / 数据库不可用: %v", err)
		log.Println("Showing example JQL queries instead... / 改为展示 JQL 查询示例...")
		fmt.Println()
		showExamples()
		return
	}
	defer db.Close()

	// Register models
	// 注册模型
	if err := db.Register(&User{}); err != nil {
		log.Fatalf("Failed to register User / 注册 User 失败: %v", err)
	}
	if err := db.Register(&Order{}); err != nil {
		log.Fatalf("Failed to register Order / 注册 Order 失败: %v", err)
	}

	// Sync database schema
	// 同步数据库结构
	if err := db.AutoSync(); err != nil {
		log.Fatalf("Failed to sync / 同步失败: %v", err)
	}

	// Enable soft delete for all tables
	// 为所有表启用软删除
	db.EnableSoftDeleteGlobal("")

	// Register a hook
	// 注册钩子
	db.Hook("users", goorm.HookBeforeCreate, func(ctx *goorm.HookContext) error {
		log.Println("Creating user / 正在创建用户:", ctx.Data["name"])
		return nil
	})

	ctx := context.Background()

	// Example: Create user
	// 示例：创建用户
	result := db.ExecuteQuery(ctx, &goorm.Query{
		Table:  "users",
		Action: goorm.ActionCreate,
		Data: map[string]any{
			"name":  "张三",
			"email": "zhangsan@example.com",
			"age":   25,
		},
	})
	printResult("Create User / 创建用户", result)

	// Example: Find users
	// 示例：查询用户
	result = db.ExecuteQuery(ctx, &goorm.Query{
		Table:  "users",
		Action: goorm.ActionFind,
		Where: []goorm.Condition{
			{Field: "age", Op: goorm.OpGreater, Value: 18},
		},
		Limit: 10,
	})
	printResult("Find Users / 查询用户", result)

	// Example: Count
	// 示例：统计数量
	result = db.ExecuteQuery(ctx, &goorm.Query{
		Table:  "users",
		Action: goorm.ActionCount,
	})
	printResult("Count Users / 统计用户", result)

	// Health check
	// 健康检查
	check := db.Health().Check(ctx)
	fmt.Printf("\nHealth Status / 健康状态: %s (Latency / 延迟: %v)\n", check.Status, check.Latency)
}

// printResult prints the query result
// printResult 打印查询结果
func printResult(title string, result *goorm.Result) {
	fmt.Printf("\n--- %s ---\n", title)
	if result.Success {
		fmt.Printf("Success / 成功: true\n")
		if result.Data != nil {
			fmt.Printf("Data / 数据: %d records / 条记录\n", len(result.Data))
		}
		if result.Count > 0 {
			fmt.Printf("Count / 数量: %d\n", result.Count)
		}
		if result.Affected > 0 {
			fmt.Printf("Affected / 影响行数: %d\n", result.Affected)
		}
		if result.ID > 0 {
			fmt.Printf("ID / 标识: %d\n", result.ID)
		}
	} else {
		fmt.Printf("Success / 成功: false\n")
		if result.Error != nil {
			fmt.Printf("Error / 错误: %s\n", result.Error.Message)
		}
	}
}

// showExamples displays JQL query examples when database is not available
// showExamples 在数据库不可用时展示 JQL 查询示例
func showExamples() {
	fmt.Println("=== JQL Examples / JQL 示例 ===")
	fmt.Println()

	examples := []struct {
		name  string
		query string
	}{
		{
			"Find users over 18 / 查找18岁以上的用户",
			`{
    "table": "users",
    "action": "find",
    "where": [{"field": "age", "op": ">", "value": 18}],
    "order_by": [{"field": "created_at", "desc": true}],
    "limit": 10
}`,
		},
		{
			"Create user / 创建用户",
			`{
    "table": "users",
    "action": "create",
    "data": {"name": "张三", "email": "zhangsan@example.com", "age": 25}
}`,
		},
		{
			"Update user / 更新用户",
			`{
    "table": "users",
    "action": "update",
    "where": [{"field": "id", "op": "=", "value": 1}],
    "data": {"name": "李四", "age": 26}
}`,
		},
		{
			"Transaction / 事务",
			`{
    "action": "transaction",
    "operations": [
        {"table": "users", "action": "create", "data": {"name": "小明"}, "ref": "new_user"},
        {"table": "orders", "action": "create", "data": {"user_id": "$new_user.id", "total": 100}}
    ]
}`,
		},
		{
			"Aggregation / 聚合查询",
			`{
    "table": "orders",
    "action": "aggregate",
    "select": [{"sum": "total"}, {"count": "*"}],
    "group_by": ["status"]
}`,
		},
		{
			"With relations (eager loading) / 关联查询（预加载）",
			`{
    "table": "users",
    "action": "find",
    "with": ["orders"],
    "limit": 5
}`,
		},
	}

	for _, ex := range examples {
		fmt.Printf("--- %s ---\n", ex.name)
		fmt.Println(ex.query)
		fmt.Println()

		// Validate the query
		// 验证查询语句
		query, err := goorm.ParseQuery(ex.query)
		if err != nil {
			fmt.Printf("Parse error / 解析错误: %v\n\n", err)
			continue
		}
		if err := query.Validate(); err != nil {
			fmt.Printf("Validation error / 验证错误: %v\n\n", err)
			continue
		}
		fmt.Println("Valid JQL / 有效的 JQL")
		fmt.Println()
	}

	// Show MCP tools
	// 展示 MCP 工具
	fmt.Println("=== MCP Tools / MCP 工具 ===")
	fmt.Println()
	tools := []string{
		"execute_query       - Execute JQL query / 执行 JQL 查询",
		"list_tables         - List all tables / 列出所有表",
		"describe_table      - Get table schema / 获取表结构",
		"find_records        - Query records / 查询记录",
		"create_record       - Create record / 创建记录",
		"update_records      - Update records / 更新记录",
		"delete_records      - Delete records / 删除记录",
		"count_records       - Count records / 统计记录数",
		"execute_transaction - Atomic operations / 原子操作",
		"explain_query       - SQL preview / SQL 预览",
		"aggregate           - Aggregations / 聚合查询",
		"sync_schema         - Schema migration / 模式迁移",
		"get_stats           - Database stats / 数据库统计",
		"natural_language    - NL query / 自然语言查询",
	}
	for _, tool := range tools {
		fmt.Printf("  - %s\n", tool)
	}

	// Suppress unused import warning
	// 抑制未使用导入警告
	_ = time.Now()
}
