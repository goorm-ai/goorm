package goorm

import (
	"fmt"
	"sync"
)

// Dialect is the interface for database-specific operations.
// Each supported database must implement this interface.
//
// Dialect 是数据库特定操作的接口。
// 每个支持的数据库必须实现此接口。
type Dialect interface {
	// Name returns the dialect name (e.g., "postgres", "mysql").
	// Name 返回方言名称（如 "postgres"、"mysql"）。
	Name() string

	// DriverName returns the database/sql driver name.
	// DriverName 返回 database/sql 驱动名称。
	DriverName() string

	// Quote quotes an identifier (table name, column name).
	// Quote 引用标识符（表名、列名）。
	Quote(identifier string) string

	// Placeholder returns the placeholder for the nth parameter.
	// Placeholder 返回第 n 个参数的占位符。
	Placeholder(n int) string

	// GoTypeToSQL converts a Go type to SQL type.
	// GoTypeToSQL 将 Go 类型转换为 SQL 类型。
	GoTypeToSQL(goType string, tags map[string]string) string

	// SupportsReturning indicates if the dialect supports RETURNING clause.
	// SupportsReturning 表示方言是否支持 RETURNING 子句。
	SupportsReturning() bool

	// SupportsUpsert indicates if the dialect supports UPSERT.
	// SupportsUpsert 表示方言是否支持 UPSERT。
	SupportsUpsert() bool

	// AutoIncrementClause returns the auto-increment clause.
	// AutoIncrementClause 返回自动递增子句。
	AutoIncrementClause() string

	// CurrentTimestamp returns the SQL for current timestamp.
	// CurrentTimestamp 返回当前时间戳的 SQL。
	CurrentTimestamp() string
}

// dialectRegistry holds all registered dialects.
// dialectRegistry 保存所有已注册的方言。
var (
	dialects   = make(map[string]Dialect)
	dialectsMu sync.RWMutex
)

// RegisterDialect registers a dialect by name.
// RegisterDialect 按名称注册方言。
func RegisterDialect(name string, dialect Dialect) {
	dialectsMu.Lock()
	defer dialectsMu.Unlock()
	dialects[name] = dialect
}

// GetDialect returns a dialect by name.
// GetDialect 按名称返回方言。
func GetDialect(name string) (Dialect, error) {
	dialectsMu.RLock()
	defer dialectsMu.RUnlock()

	dialect, ok := dialects[name]
	if !ok {
		return nil, fmt.Errorf("dialect %q not registered", name)
	}
	return dialect, nil
}

// --- PostgreSQL Dialect ---
// --- PostgreSQL 方言 ---

// PostgresDialect implements the Dialect interface for PostgreSQL.
// PostgresDialect 为 PostgreSQL 实现 Dialect 接口。
type PostgresDialect struct{}

// Name returns "postgres".
// Name 返回 "postgres"。
func (d *PostgresDialect) Name() string {
	return "postgres"
}

// DriverName returns "postgres" (pgx driver).
// DriverName 返回 "postgres"（pgx 驱动）。
func (d *PostgresDialect) DriverName() string {
	return "postgres"
}

// Quote quotes an identifier with double quotes.
// Quote 使用双引号引用标识符。
func (d *PostgresDialect) Quote(identifier string) string {
	return `"` + identifier + `"`
}

// Placeholder returns $N style placeholder.
// Placeholder 返回 $N 风格的占位符。
func (d *PostgresDialect) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

// GoTypeToSQL converts Go type to PostgreSQL type.
// GoTypeToSQL 将 Go 类型转换为 PostgreSQL 类型。
func (d *PostgresDialect) GoTypeToSQL(goType string, tags map[string]string) string {
	if sqlType, ok := tags["type"]; ok {
		return sqlType
	}

	switch goType {
	case "int", "int32":
		return "INTEGER"
	case "int8":
		return "SMALLINT"
	case "int16":
		return "SMALLINT"
	case "int64":
		return "BIGINT"
	case "uint", "uint32":
		return "INTEGER"
	case "uint8":
		return "SMALLINT"
	case "uint16":
		return "SMALLINT"
	case "uint64":
		return "BIGINT"
	case "float32":
		return "REAL"
	case "float64":
		return "DOUBLE PRECISION"
	case "bool":
		return "BOOLEAN"
	case "string":
		if size, ok := tags["size"]; ok {
			return fmt.Sprintf("VARCHAR(%s)", size)
		}
		return "VARCHAR(255)"
	case "[]byte":
		return "BYTEA"
	case "time.Time":
		return "TIMESTAMP WITH TIME ZONE"
	case "*time.Time":
		return "TIMESTAMP WITH TIME ZONE"
	default:
		// Check for pointer types
		if len(goType) > 1 && goType[0] == '*' {
			return d.GoTypeToSQL(goType[1:], tags)
		}
		return "TEXT"
	}
}

// SupportsReturning returns true.
// SupportsReturning 返回 true。
func (d *PostgresDialect) SupportsReturning() bool {
	return true
}

// SupportsUpsert returns true.
// SupportsUpsert 返回 true。
func (d *PostgresDialect) SupportsUpsert() bool {
	return true
}

// AutoIncrementClause returns SERIAL-based clause.
// AutoIncrementClause 返回基于 SERIAL 的子句。
func (d *PostgresDialect) AutoIncrementClause() string {
	return "BIGSERIAL"
}

// CurrentTimestamp returns NOW().
// CurrentTimestamp 返回 NOW()。
func (d *PostgresDialect) CurrentTimestamp() string {
	return "NOW()"
}

// --- MySQL Dialect ---
// --- MySQL 方言 ---

// MySQLDialect implements the Dialect interface for MySQL.
// MySQLDialect 为 MySQL 实现 Dialect 接口。
type MySQLDialect struct{}

// Name returns "mysql".
// Name 返回 "mysql"。
func (d *MySQLDialect) Name() string {
	return "mysql"
}

// DriverName returns "mysql".
// DriverName 返回 "mysql"。
func (d *MySQLDialect) DriverName() string {
	return "mysql"
}

// Quote quotes an identifier with backticks.
// Quote 使用反引号引用标识符。
func (d *MySQLDialect) Quote(identifier string) string {
	return "`" + identifier + "`"
}

// Placeholder returns ? style placeholder.
// Placeholder 返回 ? 风格的占位符。
func (d *MySQLDialect) Placeholder(n int) string {
	return "?"
}

// GoTypeToSQL converts Go type to MySQL type.
// GoTypeToSQL 将 Go 类型转换为 MySQL 类型。
func (d *MySQLDialect) GoTypeToSQL(goType string, tags map[string]string) string {
	if sqlType, ok := tags["type"]; ok {
		return sqlType
	}

	switch goType {
	case "int", "int32":
		return "INT"
	case "int8":
		return "TINYINT"
	case "int16":
		return "SMALLINT"
	case "int64":
		return "BIGINT"
	case "uint", "uint32":
		return "INT UNSIGNED"
	case "uint8":
		return "TINYINT UNSIGNED"
	case "uint16":
		return "SMALLINT UNSIGNED"
	case "uint64":
		return "BIGINT UNSIGNED"
	case "float32":
		return "FLOAT"
	case "float64":
		return "DOUBLE"
	case "bool":
		return "TINYINT(1)"
	case "string":
		if size, ok := tags["size"]; ok {
			return fmt.Sprintf("VARCHAR(%s)", size)
		}
		return "VARCHAR(255)"
	case "[]byte":
		return "BLOB"
	case "time.Time":
		return "DATETIME"
	case "*time.Time":
		return "DATETIME"
	default:
		if len(goType) > 1 && goType[0] == '*' {
			return d.GoTypeToSQL(goType[1:], tags)
		}
		return "TEXT"
	}
}

// SupportsReturning returns false.
// SupportsReturning 返回 false。
func (d *MySQLDialect) SupportsReturning() bool {
	return false
}

// SupportsUpsert returns true (ON DUPLICATE KEY UPDATE).
// SupportsUpsert 返回 true（ON DUPLICATE KEY UPDATE）。
func (d *MySQLDialect) SupportsUpsert() bool {
	return true
}

// AutoIncrementClause returns AUTO_INCREMENT.
// AutoIncrementClause 返回 AUTO_INCREMENT。
func (d *MySQLDialect) AutoIncrementClause() string {
	return "AUTO_INCREMENT"
}

// CurrentTimestamp returns NOW().
// CurrentTimestamp 返回 NOW()。
func (d *MySQLDialect) CurrentTimestamp() string {
	return "NOW()"
}

// --- SQLite Dialect ---
// --- SQLite 方言 ---

// SQLiteDialect implements the Dialect interface for SQLite.
// SQLiteDialect 为 SQLite 实现 Dialect 接口。
type SQLiteDialect struct{}

// Name returns "sqlite".
// Name 返回 "sqlite"。
func (d *SQLiteDialect) Name() string {
	return "sqlite"
}

// DriverName returns "sqlite3".
// DriverName 返回 "sqlite3"。
func (d *SQLiteDialect) DriverName() string {
	return "sqlite3"
}

// Quote quotes an identifier with double quotes.
// Quote 使用双引号引用标识符。
func (d *SQLiteDialect) Quote(identifier string) string {
	return `"` + identifier + `"`
}

// Placeholder returns ? style placeholder.
// Placeholder 返回 ? 风格的占位符。
func (d *SQLiteDialect) Placeholder(n int) string {
	return "?"
}

// GoTypeToSQL converts Go type to SQLite type.
// GoTypeToSQL 将 Go 类型转换为 SQLite 类型。
func (d *SQLiteDialect) GoTypeToSQL(goType string, tags map[string]string) string {
	if sqlType, ok := tags["type"]; ok {
		return sqlType
	}

	switch goType {
	case "int", "int8", "int16", "int32", "int64":
		return "INTEGER"
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return "INTEGER"
	case "float32", "float64":
		return "REAL"
	case "bool":
		return "INTEGER"
	case "string":
		return "TEXT"
	case "[]byte":
		return "BLOB"
	case "time.Time", "*time.Time":
		return "DATETIME"
	default:
		if len(goType) > 1 && goType[0] == '*' {
			return d.GoTypeToSQL(goType[1:], tags)
		}
		return "TEXT"
	}
}

// SupportsReturning returns true (SQLite 3.35+).
// SupportsReturning 返回 true（SQLite 3.35+）。
func (d *SQLiteDialect) SupportsReturning() bool {
	return true
}

// SupportsUpsert returns true (INSERT OR REPLACE).
// SupportsUpsert 返回 true（INSERT OR REPLACE）。
func (d *SQLiteDialect) SupportsUpsert() bool {
	return true
}

// AutoIncrementClause returns AUTOINCREMENT.
// AutoIncrementClause 返回 AUTOINCREMENT。
func (d *SQLiteDialect) AutoIncrementClause() string {
	return "AUTOINCREMENT"
}

// CurrentTimestamp returns CURRENT_TIMESTAMP.
// CurrentTimestamp 返回 CURRENT_TIMESTAMP。
func (d *SQLiteDialect) CurrentTimestamp() string {
	return "CURRENT_TIMESTAMP"
}

// init registers the default dialects.
// init 注册默认方言。
func init() {
	RegisterDialect("postgres", &PostgresDialect{})
	RegisterDialect("postgresql", &PostgresDialect{})
	RegisterDialect("mysql", &MySQLDialect{})
	RegisterDialect("sqlite", &SQLiteDialect{})
	RegisterDialect("sqlite3", &SQLiteDialect{})
}
