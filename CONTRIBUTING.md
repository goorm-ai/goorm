# Contributing to GoORM

Thank you for your interest in contributing to GoORM! This document provides guidelines for contributing.

感谢您有兴趣为 GoORM 做贡献！本文档提供贡献指南。

## How to Contribute | 如何贡献

### Reporting Issues | 报告问题

- Check if the issue already exists | 检查问题是否已存在
- Include clear description and reproduction steps | 包含清晰的描述和复现步骤
- Provide Go version and database information | 提供 Go 版本和数据库信息

### Submitting Changes | 提交更改

1. Fork the repository | Fork 仓库
2. Create a feature branch | 创建功能分支
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. Make your changes | 进行更改
4. Add tests for new functionality | 为新功能添加测试
5. Ensure all tests pass | 确保所有测试通过
   ```bash
   go test ./...
   ```
6. Commit your changes | 提交更改
   ```bash
   git commit -m 'Add amazing feature'
   ```
7. Push to the branch | 推送到分支
   ```bash
   git push origin feature/amazing-feature
   ```
8. Open a Pull Request | 打开拉取请求

## Code Style | 代码风格

- Follow standard Go conventions | 遵循标准 Go 约定
- Use bilingual comments (English/Chinese) | 使用双语注释
- Add tests for new functionality | 为新功能添加测试
- Keep functions focused and small | 保持函数专注且简洁

### Comment Style | 注释风格

```go
// FunctionName does something important.
// FunctionName 做一些重要的事情。
func FunctionName() {
    // Implementation detail
    // 实现细节
}
```

## Development Setup | 开发设置

```bash
# Clone the repository | 克隆仓库
git clone https://github.com/goorm-ai/goorm.git
cd goorm

# Run tests | 运行测试
go test ./...

# Run with verbose output | 详细输出运行
go test -v ./...

# Run specific tests | 运行特定测试
go test -v -run TestName ./...
```

## Project Structure | 项目结构

```
goorm/
├── goorm.go          # Package documentation | 包文档
├── query.go          # JQL query types | JQL 查询类型
├── builder.go        # SQL builder | SQL 构建器
├── executor.go       # Query execution | 查询执行
├── db.go             # Core DB instance | 核心数据库实例
├── dialect.go        # Database dialects | 数据库方言
├── migration.go      # Schema migration | 模式迁移
├── relation.go       # Relations | 关联关系
├── hooks.go          # Lifecycle hooks | 生命周期钩子
├── cache.go          # Query cache | 查询缓存
├── mcp.go            # MCP server | MCP 服务器
├── *_test.go         # Tests | 测试
└── examples/         # Usage examples | 使用示例
```

## Questions? | 有问题？

Feel free to open an issue for any questions or suggestions.

如有任何问题或建议，请随时提交 issue。

---

Thank you for contributing! | 感谢您的贡献！
