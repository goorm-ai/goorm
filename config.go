package goorm

import (
	"time"
)

// Config contains the configuration options for GoORM.
// 配置选项用于初始化和配置 GoORM 实例。
//
// Config 包含 GoORM 的配置选项。
type Config struct {
	// Driver is the database driver name (postgres, mysql, sqlite).
	// Driver 是数据库驱动名称（postgres、mysql、sqlite）。
	Driver string

	// DSN is the data source name (connection string).
	// DSN 是数据源名称（连接字符串）。
	DSN string

	// MaxOpenConns sets the maximum number of open connections.
	// MaxOpenConns 设置最大打开连接数。
	MaxOpenConns int

	// MaxIdleConns sets the maximum number of idle connections.
	// MaxIdleConns 设置最大空闲连接数。
	MaxIdleConns int

	// ConnMaxLifetime sets the maximum lifetime of a connection.
	// ConnMaxLifetime 设置连接的最大生命周期。
	ConnMaxLifetime time.Duration

	// ConnMaxIdleTime sets the maximum idle time of a connection.
	// ConnMaxIdleTime 设置连接的最大空闲时间。
	ConnMaxIdleTime time.Duration

	// DefaultTimeout is the default timeout for all operations.
	// DefaultTimeout 是所有操作的默认超时时间。
	DefaultTimeout time.Duration

	// QueryTimeout is the timeout for query operations.
	// QueryTimeout 是查询操作的超时时间。
	QueryTimeout time.Duration

	// WriteTimeout is the timeout for write operations.
	// WriteTimeout 是写操作的超时时间。
	WriteTimeout time.Duration

	// Naming contains naming convention configuration.
	// Naming 包含命名约定配置。
	Naming NamingConfig

	// Migration contains migration configuration.
	// Migration 包含迁移配置。
	Migration MigrationConfig

	// Security contains security configuration.
	// Security 包含安全配置。
	Security SecurityConfig

	// Debug enables debug mode for all queries.
	// Debug 为所有查询启用调试模式。
	Debug bool

	// Logger is the logger for GoORM.
	// Logger 是 GoORM 的日志记录器。
	Logger Logger
}

// NamingConfig contains naming convention configuration.
// NamingConfig 包含命名约定配置。
type NamingConfig struct {
	// TableNamer is the strategy for table names.
	// TableNamer 是表名的命名策略。
	TableNamer TableNamerFunc

	// ColumnNamer is the strategy for column names.
	// ColumnNamer 是列名的命名策略。
	ColumnNamer ColumnNamerFunc

	// TablePrefix is the prefix for all table names.
	// TablePrefix 是所有表名的前缀。
	TablePrefix string

	// PrimaryKey is the default primary key column name.
	// PrimaryKey 是默认的主键列名。
	PrimaryKey string

	// CreatedAtField is the name of the created_at field.
	// CreatedAtField 是 created_at 字段的名称。
	CreatedAtField string

	// UpdatedAtField is the name of the updated_at field.
	// UpdatedAtField 是 updated_at 字段的名称。
	UpdatedAtField string

	// DeletedAtField is the name of the deleted_at field.
	// DeletedAtField 是 deleted_at 字段的名称。
	DeletedAtField string
}

// TableNamerFunc is a function type for converting struct name to table name.
// TableNamerFunc 是将结构体名转换为表名的函数类型。
type TableNamerFunc func(structName string) string

// ColumnNamerFunc is a function type for converting field name to column name.
// ColumnNamerFunc 是将字段名转换为列名的函数类型。
type ColumnNamerFunc func(fieldName string) string

// MigrationConfig contains migration configuration.
// MigrationConfig 包含迁移配置。
type MigrationConfig struct {
	// AutoMigrate enables automatic migration on startup.
	// AutoMigrate 在启动时启用自动迁移。
	AutoMigrate bool

	// Aggressive enables aggressive migration mode (delete missing columns/tables).
	// Aggressive 启用激进迁移模式（删除缺失的列/表）。
	Aggressive bool

	// AutoBackup enables automatic backup before destructive operations.
	// AutoBackup 在破坏性操作之前启用自动备份。
	AutoBackup bool

	// BackupBeforeDelete enables backup before deleting columns/tables.
	// BackupBeforeDelete 在删除列/表之前启用备份。
	BackupBeforeDelete bool

	// BackupRetention is how long to keep backups.
	// BackupRetention 是备份保留时间。
	BackupRetention time.Duration
}

// SecurityConfig contains security configuration.
// SecurityConfig 包含安全配置。
type SecurityConfig struct {
	// ConfirmDestructive requires confirmation for destructive operations.
	// ConfirmDestructive 对破坏性操作要求确认。
	ConfirmDestructive bool

	// ConfirmThreshold is the minimum affected rows to trigger confirmation.
	// ConfirmThreshold 是触发确认的最小影响行数。
	ConfirmThreshold int

	// AuditEnabled enables audit logging.
	// AuditEnabled 启用审计日志。
	AuditEnabled bool

	// MaskSensitive enables automatic masking of sensitive fields.
	// MaskSensitive 启用敏感字段的自动脱敏。
	MaskSensitive bool
}

// Logger is the interface for logging.
// Logger 是日志记录的接口。
type Logger interface {
	// Debug logs a debug message.
	// Debug 记录调试消息。
	Debug(msg string, args ...any)

	// Info logs an info message.
	// Info 记录信息消息。
	Info(msg string, args ...any)

	// Warn logs a warning message.
	// Warn 记录警告消息。
	Warn(msg string, args ...any)

	// Error logs an error message.
	// Error 记录错误消息。
	Error(msg string, args ...any)
}

// DefaultConfig returns the default configuration.
// DefaultConfig 返回默认配置。
func DefaultConfig() Config {
	return Config{
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		DefaultTimeout:  30 * time.Second,
		QueryTimeout:    10 * time.Second,
		WriteTimeout:    30 * time.Second,
		Naming: NamingConfig{
			TableNamer:     SnakeCasePlural,
			ColumnNamer:    SnakeCase,
			PrimaryKey:     "id",
			CreatedAtField: "created_at",
			UpdatedAtField: "updated_at",
			DeletedAtField: "deleted_at",
		},
		Migration: MigrationConfig{
			AutoMigrate:        false,
			Aggressive:         true,
			AutoBackup:         true,
			BackupBeforeDelete: true,
			BackupRetention:    30 * 24 * time.Hour, // 30 days
		},
		Security: SecurityConfig{
			ConfirmDestructive: true,
			ConfirmThreshold:   10,
			AuditEnabled:       true,
			MaskSensitive:      true,
		},
	}
}
