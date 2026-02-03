# Hooks / 钩子函数

This document covers hooks in GoORM.

本文档涵盖 GoORM 中的钩子函数。

## Hook Types / 钩子类型

| Hook | Timing / 时机 |
|------|--------------|
| `HookBeforeCreate` | Before insert / 插入前 |
| `HookAfterCreate` | After insert / 插入后 |
| `HookBeforeUpdate` | Before update / 更新前 |
| `HookAfterUpdate` | After update / 更新后 |
| `HookBeforeDelete` | Before delete / 删除前 |
| `HookAfterDelete` | After delete / 删除后 |
| `HookBeforeFind` | Before query / 查询前 |
| `HookAfterFind` | After query / 查询后 |

## Registering Hooks / 注册钩子

```go
// Before create hook / 创建前钩子
db.Hook("users", goorm.HookBeforeCreate, func(ctx *goorm.HookContext) error {
    // Add default values / 添加默认值
    ctx.Data["created_by"] = "system"
    ctx.Data["status"] = "pending"
    return nil
})

// After create hook / 创建后钩子
db.Hook("users", goorm.HookAfterCreate, func(ctx *goorm.HookContext) error {
    // Log the creation / 记录创建日志
    log.Printf("User created: %v", ctx.Result.ID)
    return nil
})

// Before update hook / 更新前钩子
db.Hook("users", goorm.HookBeforeUpdate, func(ctx *goorm.HookContext) error {
    // Auto-update timestamp / 自动更新时间戳
    ctx.Data["updated_at"] = time.Now()
    return nil
})

// Before delete hook / 删除前钩子
db.Hook("users", goorm.HookBeforeDelete, func(ctx *goorm.HookContext) error {
    // Check if can delete / 检查是否可以删除
    if ctx.Where["id"] == 1 {
        return errors.New("cannot delete admin user")
    }
    return nil
})
```

## Hook Context / 钩子上下文

```go
type HookContext struct {
    Table   string                 // Table name / 表名
    Action  string                 // Action type / 操作类型
    Data    map[string]any         // Data for create/update / 创建/更新数据
    Where   map[string]any         // Conditions / 条件
    Result  *Result                // Query result (after hooks) / 查询结果
    Context context.Context        // Request context / 请求上下文
}
```

## Soft Delete / 软删除

```go
// Enable soft delete for all tables / 为所有表启用软删除
db.EnableSoftDeleteGlobal("")

// Enable for specific table / 为特定表启用
db.EnableSoftDelete("users")

// Custom deleted_at column / 自定义 deleted_at 列
db.EnableSoftDeleteGlobal("removed_at")
```

When soft delete is enabled, `delete` action sets `deleted_at` instead of removing the record.

启用软删除后，`delete` 操作会设置 `deleted_at` 而不是删除记录。
