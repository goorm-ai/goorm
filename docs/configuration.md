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

## Naming Convention / 命名规范

```go
// Table prefix / 表前缀
config.Naming.TablePrefix = "app_"

// Use singular table names / 使用单数表名
config.Naming.SingularTable = false

// Snake case for columns / 列名使用蛇形命名
config.Naming.SnakeCase = true
```

## Migration / 迁移设置

```go
// Auto migrate on startup / 启动时自动迁移
config.Migration.AutoMigrate = true

// Aggressive mode (drop columns) / 激进模式（删除列）
config.Migration.Aggressive = false

// Auto backup before migration / 迁移前自动备份
config.Migration.AutoBackup = true
```

## Health Check / 健康检查

```go
// Enable background health check / 启用后台健康检查
config.Health.Enabled = true

// Check interval / 检查间隔
config.Health.Interval = 30 * time.Second

// Timeout for health check / 健康检查超时
config.Health.Timeout = 5 * time.Second
```

## Logging / 日志

```go
// Log level / 日志级别
config.Log.Level = "info"  // debug, info, warn, error

// Log slow queries / 记录慢查询
config.Log.SlowThreshold = 200 * time.Millisecond
```

## Full Example / 完整示例

```go
config := goorm.DefaultConfig()

// Connection pool / 连接池
config.MaxOpenConns = 100
config.MaxIdleConns = 10
config.ConnMaxLifetime = time.Hour

// Naming / 命名
config.Naming.TablePrefix = "app_"
config.Naming.SingularTable = false

// Migration / 迁移
config.Migration.AutoMigrate = true
config.Migration.Aggressive = false
config.Migration.AutoBackup = true

// Health / 健康检查
config.Health.Enabled = true
config.Health.Interval = 30 * time.Second

// Logging / 日志
config.Log.Level = "info"
config.Log.SlowThreshold = 200 * time.Millisecond

db, err := goorm.ConnectWithConfig(dsn, config)
```
