# Configuration / 配置选项

This document covers GoORM configuration options.

本文档涵盖 GoORM 的配置选项。

## Default Configuration / 默认配置

```go
config := goorm.DefaultConfig()
db, err := goorm.ConnectWithConfig(dsn, config)
```

## Connection Pool / 连接池

```go
config := goorm.DefaultConfig()

// Max open connections / 最大打开连接数
config.MaxOpenConns = 100

// Max idle connections / 最大空闲连接数
config.MaxIdleConns = 10

// Connection max lifetime / 连接最大生命周期
config.ConnMaxLifetime = time.Hour

// Connection max idle time / 连接最大空闲时间
config.ConnMaxIdleTime = 30 * time.Minute
```

## Timeout / 超时设置

```go
// Default timeout for all operations / 所有操作的默认超时
config.DefaultTimeout = 30 * time.Second

// Query timeout / 查询超时
config.QueryTimeout = 10 * time.Second

// Write timeout / 写操作超时
config.WriteTimeout = 30 * time.Second
```

## Naming Convention / 命名规范

```go
// Table prefix / 表前缀
config.Naming.TablePrefix = "app_"

// Custom table naming function / 自定义表名函数
config.Naming.TableNamer = goorm.SnakeCasePlural

// Custom column naming function / 自定义列名函数
config.Naming.ColumnNamer = goorm.SnakeCase

// Field names / 字段名称
config.Naming.PrimaryKey = "id"
config.Naming.CreatedAtField = "created_at"
config.Naming.UpdatedAtField = "updated_at"
config.Naming.DeletedAtField = "deleted_at"
```

## Migration / 迁移设置

```go
// Auto migrate on startup / 启动时自动迁移
config.Migration.AutoMigrate = true

// Aggressive mode (drop columns) / 激进模式（删除列）
config.Migration.Aggressive = false

// Auto backup before migration / 迁移前自动备份
config.Migration.AutoBackup = true

// Backup before delete / 删除前备份
config.Migration.BackupBeforeDelete = true

// Backup retention period / 备份保留时间
config.Migration.BackupRetention = 30 * 24 * time.Hour
```

## Security / 安全设置

```go
// Confirm destructive operations / 确认破坏性操作
config.Security.ConfirmDestructive = true

// Minimum affected rows to trigger confirmation / 触发确认的最小影响行数
config.Security.ConfirmThreshold = 10

// Enable audit logging / 启用审计日志
config.Security.AuditEnabled = true

// Mask sensitive fields / 脱敏敏感字段
config.Security.MaskSensitive = true
```

## Debug / 调试

```go
// Enable debug mode / 启用调试模式
config.Debug = true

// Custom logger / 自定义日志器
config.Logger = myCustomLogger
```

## Full Example / 完整示例

```go
config := goorm.DefaultConfig()

// Connection pool / 连接池
config.MaxOpenConns = 100
config.MaxIdleConns = 10
config.ConnMaxLifetime = time.Hour
config.ConnMaxIdleTime = 30 * time.Minute

// Timeout / 超时
config.DefaultTimeout = 30 * time.Second
config.QueryTimeout = 10 * time.Second

// Naming / 命名
config.Naming.TablePrefix = "app_"
config.Naming.TableNamer = goorm.SnakeCasePlural
config.Naming.ColumnNamer = goorm.SnakeCase

// Migration / 迁移
config.Migration.AutoMigrate = true
config.Migration.Aggressive = false
config.Migration.AutoBackup = true

// Security / 安全
config.Security.ConfirmDestructive = true
config.Security.ConfirmThreshold = 10
config.Security.AuditEnabled = true

// Debug / 调试
config.Debug = false

db, err := goorm.ConnectWithConfig(dsn, config)
```
