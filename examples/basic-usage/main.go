package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/goorm-ai/goorm"
	"modernc.org/sqlite"
)

// 将 modernc.org/sqlite 注册为 sqlite3 驱动名
func init() {
	sql.Register("sqlite3", &sqlite.Driver{})
}

// User 定义用户模型（简化版，移除约束便于测试）
type User struct {
	goorm.Model
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func main() {
	fmt.Println("=== GoORM 基本使用示例 ===")
	fmt.Println()

	// 使用 SQLite 内存数据库进行测试
	db, err := goorm.Connect("sqlite://:memory:")
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	fmt.Println("✓ 数据库连接成功")

	// 注册模型
	if err := db.Register(&User{}); err != nil {
		log.Fatalf("注册模型失败: %v", err)
	}
	fmt.Println("✓ 模型注册成功")

	// 自动同步表结构
	if err := db.AutoSync(); err != nil {
		log.Fatalf("同步表结构失败: %v", err)
	}
	fmt.Println("✓ 表结构同步成功")

	// 使用 JQL 创建用户
	createResult := db.Execute(`{
		"table": "users",
		"action": "create",
		"data": {
			"name": "张三",
			"email": "zhangsan@example.com",
			"age": 25
		}
	}`)
	if createResult.Error != nil {
		fmt.Printf("✗ 创建用户失败: %+v\n", createResult.Error)
	} else {
		fmt.Println("✓ 用户创建成功")
	}

	// 使用 JQL 查询用户
	findResult := db.Execute(`{
		"table": "users",
		"action": "find"
	}`)
	if findResult.Error != nil {
		fmt.Printf("✗ 查询用户失败: %+v\n", findResult.Error)
	} else {
		fmt.Printf("✓ 查询到 %d 个用户\n", findResult.Count)
	}

	// 列出所有表
	listResult := db.Execute(`{
		"action": "list_tables"
	}`)
	if listResult.Error != nil {
		fmt.Printf("✗ 列出表失败: %+v\n", listResult.Error)
	} else {
		fmt.Printf("✓ 数据库表: %v\n", listResult.Tables)
	}

	// 使用自然语言查询（注意：这个功能还未实现）
	nlResult := db.NL("查找所有用户")
	if nlResult.Error != nil {
		fmt.Printf("⚠ 自然语言查询暂未实现: %s\n", nlResult.Error.Message)
	} else {
		fmt.Printf("✓ 自然语言查询成功，结果: %d 条记录\n", nlResult.Count)
	}

	fmt.Println()
	fmt.Println("=== GoORM 引用测试完成 ===")
	fmt.Println("✓ 成功从 GitHub 引用 github.com/goorm-ai/goorm")
}
