# Contributing to GoORM

Thank you for your interest in contributing to GoORM! This document provides guidelines for contributing.

感谢您有兴趣�?GoORM 做出贡献！本文档提供了贡献指南�?

## How to Contribute | 如何贡献

### Reporting Issues | 报告问题

- Check if the issue already exists
- Include clear description and reproduction steps
- Provide Go version and database information

- 检查问题是否已存在
- 包含清晰的描述和复现步骤
- 提供 Go 版本和数据库信息

### Pull Requests | 拉取请求

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

1. Fork 仓库
2. 创建功能分支（`git checkout -b feature/amazing-feature`�?
3. 进行更改
4. 为新功能添加测试
5. 确保所有测试通过（`go test ./...`�?
6. 提交更改（`git commit -m 'Add amazing feature'`�?
7. 推送到分支（`git push origin feature/amazing-feature`�?
8. 打开拉取请求

## Code Style | 代码风格

- Follow standard Go conventions
- Use bilingual comments (English/Chinese)
- Add tests for new functionality
- Keep functions focused and small

- 遵循标准 Go 约定
- 使用双语注释（英�?中文�?
- 为新功能添加测试
- 保持函数专注且简�?

### Comment Style | 注释风格

```go
// FunctionName does something important.
// FunctionName 做一些重要的事情�?
func FunctionName() {
    // Implementation detail
    // 实现细节
}
```

## Development Setup | 开发设�?

```bash
# Clone the repository
git clone https://github.com/goorm-ai/goorm.git
cd goorm

# Run tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific tests
go test -v -run TestName ./...
```

## Project Structure | 项目结构

```
goorm/
├── goorm.go          # Package documentation
├── query.go          # JQL query types
├── builder.go        # SQL builder
├── executor.go       # Query execution
├── db.go             # Core DB instance
├── dialect.go        # Database dialects
├── migration.go      # Schema migration
├── relation.go       # Relations
├── hooks.go          # Lifecycle hooks
├── cache.go          # Query cache
├── mcp.go            # MCP server
├── *_test.go         # Tests
└── examples/         # Usage examples
```

## Questions? | 有问题？

Feel free to open an issue for any questions or suggestions.

如有任何问题或建议，请随时提�?issue�?

---

Thank you for contributing! 感谢您的贡献�?
