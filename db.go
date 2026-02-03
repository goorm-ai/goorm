package goorm

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

// DB is the main GoORM database instance.
// It provides methods for executing JQL queries, managing models, and configuring behavior.
//
// DB 是 GoORM 的主数据库实例。
// 它提供执行 JQL 查询、管理模型和配置行为的方法。
type DB struct {
	// config holds the database configuration.
	// config 保存数据库配置。
	config Config

	// sqlDB is the underlying database connection.
	// sqlDB 是底层数据库连接。
	sqlDB *sql.DB

	// dialect is the database-specific dialect.
	// dialect 是数据库特定的方言。
	dialect Dialect

	// registry holds registered models and their metadata.
	// registry 保存已注册的模型及其元数据。
	registry *Registry

	// hooks is the hook manager for lifecycle events.
	// hooks 是生命周期事件的钩子管理器。
	hooks *HookManager

	// mu protects concurrent access.
	// mu 保护并发访问。
	mu sync.RWMutex

	// ctx is the default context for operations.
	// ctx 是操作的默认上下文。
	ctx context.Context

	// cancelFunc is the cancel function for the default context.
	// cancelFunc 是默认上下文的取消函数。
	cancelFunc context.CancelFunc
}

// Connect creates a new database connection with the given DSN.
// The DSN format should include the driver scheme:
//   - postgres://user:pass@host:port/dbname?sslmode=disable
//   - mysql://user:pass@host:port/dbname?charset=utf8mb4
//   - sqlite://path/to/database.db
//
// Connect 使用给定的 DSN 创建新的数据库连接。
// DSN 格式应包含驱动方案：
//   - postgres://user:pass@host:port/dbname?sslmode=disable
//   - mysql://user:pass@host:port/dbname?charset=utf8mb4
//   - sqlite://path/to/database.db
func Connect(dsn string) (*DB, error) {
	return ConnectWithConfig(dsn, DefaultConfig())
}

// ConnectWithConfig creates a new database connection with custom configuration.
// ConnectWithConfig 使用自定义配置创建新的数据库连接。
func ConnectWithConfig(dsn string, config Config) (*DB, error) {
	driver, cleanDSN, err := parseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	config.Driver = driver
	config.DSN = cleanDSN

	dialect, err := GetDialect(driver)
	if err != nil {
		return nil, fmt.Errorf("unsupported database driver %q: %w", driver, err)
	}

	sqlDB, err := sql.Open(dialect.DriverName(), cleanDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	// 配置连接池
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbCtx, dbCancel := context.WithCancel(context.Background())

	db := &DB{
		config:     config,
		sqlDB:      sqlDB,
		dialect:    dialect,
		registry:   NewRegistry(),
		hooks:      NewHookManager(),
		ctx:        dbCtx,
		cancelFunc: dbCancel,
	}

	// Register built-in hooks
	// 注册内置钩子
	db.hooks.RegisterGlobal(HookBeforeCreate, TimestampHook(config.Naming))
	db.hooks.RegisterGlobal(HookBeforeUpdate, TimestampHook(config.Naming))

	return db, nil
}

// parseDSN parses a DSN string and extracts the driver and clean DSN.
// parseDSN 解析 DSN 字符串并提取驱动程序和干净的 DSN。
func parseDSN(dsn string) (driver, cleanDSN string, err error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", "", err
	}

	driver = strings.ToLower(u.Scheme)
	if driver == "" {
		return "", "", fmt.Errorf("missing driver scheme in DSN")
	}

	// Convert URL format to driver-specific format
	// 将 URL 格式转换为驱动程序特定格式
	switch driver {
	case "postgres", "postgresql":
		driver = "postgres"
		// PostgreSQL accepts the URL format directly
		cleanDSN = dsn
	case "mysql":
		// MySQL requires: user:pass@tcp(host:port)/dbname
		password, _ := u.User.Password()
		host := u.Host
		if !strings.Contains(host, ":") {
			host += ":3306"
		}
		cleanDSN = fmt.Sprintf("%s:%s@tcp(%s)%s?%s",
			u.User.Username(),
			password,
			host,
			u.Path,
			u.RawQuery,
		)
	case "sqlite", "sqlite3":
		driver = "sqlite3"
		cleanDSN = strings.TrimPrefix(dsn, u.Scheme+"://")
	default:
		return "", "", fmt.Errorf("unsupported driver: %s", driver)
	}

	return driver, cleanDSN, nil
}

// Close closes the database connection and releases resources.
// Close 关闭数据库连接并释放资源。
func (db *DB) Close() error {
	db.cancelFunc()
	return db.sqlDB.Close()
}

// Execute executes a JQL query string and returns the result.
// This is the main method for AI-generated queries.
//
// Execute 执行 JQL 查询字符串并返回结果。
// 这是 AI 生成查询的主要方法。
//
// Example / 示例:
//
//	result := db.Execute(`{
//	    "table": "users",
//	    "action": "find",
//	    "where": [{"field": "age", "op": ">", "value": 18}]
//	}`)
func (db *DB) Execute(jql string) *Result {
	return db.ExecuteContext(db.ctx, jql)
}

// ExecuteContext executes a JQL query with the given context.
// ExecuteContext 使用给定的上下文执行 JQL 查询。
func (db *DB) ExecuteContext(ctx context.Context, jql string) *Result {
	query, err := ParseQuery(jql)
	if err != nil {
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "PARSE_ERROR",
				Message: err.Error(),
			},
		}
	}

	return db.ExecuteQuery(ctx, query)
}

// ExecuteQuery executes a parsed Query struct.
// ExecuteQuery 执行解析后的 Query 结构体。
func (db *DB) ExecuteQuery(ctx context.Context, query *Query) *Result {
	// Validate the query
	// 验证查询
	if err := query.Validate(); err != nil {
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		}
	}

	// Apply timeout if specified
	// 如果指定了超时则应用
	if query.Timeout != "" {
		timeout, err := time.ParseDuration(query.Timeout)
		if err == nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
	}

	// Execute based on action
	// 根据操作执行
	switch query.Action {
	case ActionFind:
		return db.executeFind(ctx, query)
	case ActionCreate:
		return db.executeCreate(ctx, query)
	case ActionCreateBatch:
		return db.executeCreateBatch(ctx, query)
	case ActionUpdate:
		return db.executeUpdate(ctx, query)
	case ActionDelete:
		return db.executeDelete(ctx, query)
	case ActionCount:
		return db.executeCount(ctx, query)
	case ActionAggregate:
		return db.executeAggregate(ctx, query)
	case ActionTransaction:
		return db.executeTransaction(ctx, query)
	case ActionExplain:
		return db.executeExplain(ctx, query)
	case ActionValidate:
		return db.executeValidate(ctx, query)
	case ActionListTables:
		return db.executeListTables(ctx, query)
	case ActionDescribe:
		return db.executeDescribe(ctx, query)
	default:
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "UNKNOWN_ACTION",
				Message: fmt.Sprintf("unknown action: %s", query.Action),
			},
		}
	}
}

// NL executes a natural language query.
// The query is converted to JQL internally and then executed.
//
// NL 执行自然语言查询。
// 查询在内部转换为 JQL，然后执行。
//
// Example / 示例:
//
//	result := db.NL("查找所有18岁以上的用户")
//	result := db.NL("Find all users older than 18")
func (db *DB) NL(query string) *Result {
	return db.NLContext(db.ctx, query)
}

// NLContext executes a natural language query with the given context.
// NLContext 使用给定的上下文执行自然语言查询。
func (db *DB) NLContext(ctx context.Context, query string) *Result {
	// TODO: Implement natural language to JQL conversion
	// TODO: 实现自然语言到 JQL 的转换
	return &Result{
		Success: false,
		Error: &ResultError{
			Code:    "NOT_IMPLEMENTED",
			Message: "natural language queries are not yet implemented",
		},
	}
}

// Register registers one or more models with the database.
// Models should embed goorm.Model and define their fields.
//
// Register 向数据库注册一个或多个模型。
// 模型应嵌入 goorm.Model 并定义其字段。
//
// Example / 示例:
//
//	db.Register(&User{}, &Order{})
func (db *DB) Register(models ...any) error {
	for _, model := range models {
		if err := db.registry.Register(model, db.config.Naming); err != nil {
			return err
		}
	}
	return nil
}

// AutoSync synchronizes the database schema with registered models.
// In aggressive mode (default), it will delete columns/tables not in models.
//
// AutoSync 将数据库模式与已注册的模型同步。
// 在激进模式下（默认），它会删除模型中不存在的列/表。
func (db *DB) AutoSync() error {
	return db.AutoSyncContext(db.ctx)
}

// AutoSyncContext synchronizes the database schema with the given context.
// AutoSyncContext 使用给定的上下文同步数据库模式。
func (db *DB) AutoSyncContext(ctx context.Context) error {
	migrator := NewMigrator(db)
	return migrator.AutoSync(ctx)
}

// Configure updates the database configuration.
// Configure 更新数据库配置。
func (db *DB) Configure(config Config) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Update pool settings if they changed
	// 如果连接池设置已更改则更新
	if config.MaxOpenConns > 0 {
		db.config.MaxOpenConns = config.MaxOpenConns
		db.sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.config.MaxIdleConns = config.MaxIdleConns
		db.sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.ConnMaxLifetime > 0 {
		db.config.ConnMaxLifetime = config.ConnMaxLifetime
		db.sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	}

	// Update other settings
	// 更新其他设置
	if config.DefaultTimeout > 0 {
		db.config.DefaultTimeout = config.DefaultTimeout
	}
	if config.Naming.TableNamer != nil {
		db.config.Naming = config.Naming
	}
	if config.Debug {
		db.config.Debug = config.Debug
	}
}

// SqlDB returns the underlying *sql.DB connection.
// This is useful for advanced operations or integrating with other libraries.
//
// SqlDB 返回底层的 *sql.DB 连接。
// 这对于高级操作或与其他库集成很有用。
func (db *DB) SqlDB() *sql.DB {
	return db.sqlDB
}

// Dialect returns the database dialect.
// Dialect 返回数据库方言。
func (db *DB) Dialect() Dialect {
	return db.dialect
}

// --- Internal execution methods ---
// --- 内部执行方法 ---

func (db *DB) executeFind(ctx context.Context, query *Query) *Result {
	executor := NewExecutor(db)
	return executor.ExecuteFind(ctx, query)
}

func (db *DB) executeCreate(ctx context.Context, query *Query) *Result {
	executor := NewExecutor(db)
	return executor.ExecuteCreate(ctx, query)
}

func (db *DB) executeCreateBatch(ctx context.Context, query *Query) *Result {
	executor := NewExecutor(db)
	return executor.ExecuteCreateBatch(ctx, query)
}

func (db *DB) executeUpdate(ctx context.Context, query *Query) *Result {
	executor := NewExecutor(db)
	return executor.ExecuteUpdate(ctx, query)
}

func (db *DB) executeDelete(ctx context.Context, query *Query) *Result {
	executor := NewExecutor(db)
	return executor.ExecuteDelete(ctx, query)
}

func (db *DB) executeCount(ctx context.Context, query *Query) *Result {
	executor := NewExecutor(db)
	return executor.ExecuteCount(ctx, query)
}

func (db *DB) executeAggregate(ctx context.Context, query *Query) *Result {
	executor := NewExecutor(db)
	return executor.ExecuteAggregate(ctx, query)
}

func (db *DB) executeTransaction(ctx context.Context, query *Query) *Result {
	return db.executeTransactionInternal(ctx, query)
}

func (db *DB) executeExplain(ctx context.Context, query *Query) *Result {
	if query.QueryToExplain == nil {
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "MISSING_QUERY",
				Message: "query field is required for explain action",
			},
		}
	}

	builder := NewSQLBuilder(db.dialect, query.QueryToExplain)
	buildResult, err := builder.Build()
	if err != nil {
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "BUILD_ERROR",
				Message: err.Error(),
			},
		}
	}

	return &Result{
		Success: true,
		Explain: &ExplainResult{
			SQL:    buildResult.SQL,
			Params: buildResult.Params,
		},
	}
}

func (db *DB) executeValidate(ctx context.Context, query *Query) *Result {
	if query.QueryToExplain != nil {
		if err := query.QueryToExplain.Validate(); err != nil {
			return &Result{
				Success: false,
				Error: &ResultError{
					Code:    "VALIDATION_ERROR",
					Message: err.Error(),
				},
			}
		}
	}
	return &Result{Success: true}
}

func (db *DB) executeListTables(ctx context.Context, query *Query) *Result {
	tables := db.registry.ListTables()
	return &Result{
		Success: true,
		Tables:  tables,
	}
}

func (db *DB) executeDescribe(ctx context.Context, query *Query) *Result {
	schema, err := db.registry.GetSchema(query.Table)
	if err != nil {
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "TABLE_NOT_FOUND",
				Message: err.Error(),
			},
		}
	}
	return &Result{
		Success: true,
		Schema:  schema,
	}
}

// Hook registers a hook for a specific table.
// Hook 为特定表注册钩子。
func (db *DB) Hook(table string, hookType HookType, fn HookFunc) {
	db.hooks.Register(table, hookType, fn)
}

// HookGlobal registers a global hook for all tables.
// HookGlobal 为所有表注册全局钩子。
func (db *DB) HookGlobal(hookType HookType, fn HookFunc) {
	db.hooks.RegisterGlobal(hookType, fn)
}

// EnableSoftDelete enables soft delete for a table.
// EnableSoftDelete 为表启用软删除。
func (db *DB) EnableSoftDelete(table string, deletedAtField string) {
	if deletedAtField == "" {
		deletedAtField = db.config.Naming.DeletedAtField
	}
	if deletedAtField == "" {
		deletedAtField = "deleted_at"
	}
	db.hooks.Register(table, HookBeforeDelete, SoftDeleteHook(deletedAtField))
}

// EnableSoftDeleteGlobal enables soft delete for all tables.
// EnableSoftDeleteGlobal 为所有表启用软删除。
func (db *DB) EnableSoftDeleteGlobal(deletedAtField string) {
	if deletedAtField == "" {
		deletedAtField = db.config.Naming.DeletedAtField
	}
	if deletedAtField == "" {
		deletedAtField = "deleted_at"
	}
	db.hooks.RegisterGlobal(HookBeforeDelete, SoftDeleteHook(deletedAtField))
}

// Hooks returns the hook manager for advanced customization.
// Hooks 返回钩子管理器用于高级自定义。
func (db *DB) Hooks() *HookManager {
	return db.hooks
}
