package goorm

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Executor handles SQL execution and result mapping.
// It coordinates between the SQL builder and the database connection.
//
// Executor 处理 SQL 执行和结果映射。
// 它协调 SQL 构建器和数据库连接之间的关系。
type Executor struct {
	db      *DB
	dialect Dialect
}

// NewExecutor creates a new executor.
// NewExecutor 创建新的执行器。
func NewExecutor(db *DB) *Executor {
	return &Executor{
		db:      db,
		dialect: db.dialect,
	}
}

// ExecuteFind executes a find query and returns results.
// ExecuteFind 执行查找查询并返回结果。
func (e *Executor) ExecuteFind(ctx context.Context, query *Query) *Result {
	startTime := time.Now()

	builder := NewSQLBuilder(e.dialect, query)
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

	rows, err := e.db.sqlDB.QueryContext(ctx, buildResult.SQL, buildResult.Params...)
	if err != nil {
		return e.handleSQLError(err, buildResult)
	}
	defer rows.Close()

	// Get column names
	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "COLUMN_ERROR",
				Message: err.Error(),
			},
		}
	}

	// Scan rows
	// 扫描行
	data := make([]map[string]any, 0)
	for rows.Next() {
		// Create slice of interface{} to hold column values
		// 创建 interface{} 切片来保存列值
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return &Result{
				Success: false,
				Error: &ResultError{
					Code:    "SCAN_ERROR",
					Message: err.Error(),
				},
			}
		}

		// Convert to map
		// 转换为 map
		row := make(map[string]any, len(columns))
		for i, col := range columns {
			val := values[i]
			// Handle []byte conversion to string
			// 处理 []byte 到 string 的转换
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			row[col] = val
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "ROWS_ERROR",
				Message: err.Error(),
			},
		}
	}

	result := &Result{
		Success: true,
		Data:    data,
		Count:   int64(len(data)),
	}

	// Add debug info if requested
	// 如果请求则添加调试信息
	if query.Debug || e.db.config.Debug {
		result.Meta = &ResultMeta{
			SQL:          buildResult.SQL,
			Params:       buildResult.Params,
			DurationMs:   float64(time.Since(startTime).Microseconds()) / 1000,
			RowsReturned: int64(len(data)),
		}
	}

	return result
}

// ExecuteCreate executes a create query.
// ExecuteCreate 执行创建查询。
func (e *Executor) ExecuteCreate(ctx context.Context, query *Query) *Result {
	startTime := time.Now()

	builder := NewSQLBuilder(e.dialect, query)
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

	var lastID uint64

	if e.dialect.SupportsReturning() {
		// PostgreSQL: use RETURNING
		err = e.db.sqlDB.QueryRowContext(ctx, buildResult.SQL, buildResult.Params...).Scan(&lastID)
		if err != nil && err != sql.ErrNoRows {
			return e.handleSQLError(err, buildResult)
		}
	} else {
		// MySQL: use LastInsertId
		result, err := e.db.sqlDB.ExecContext(ctx, buildResult.SQL, buildResult.Params...)
		if err != nil {
			return e.handleSQLError(err, buildResult)
		}
		id, err := result.LastInsertId()
		if err == nil {
			lastID = uint64(id)
		}
	}

	r := &Result{
		Success:  true,
		ID:       lastID,
		Affected: 1,
	}

	if query.Debug || e.db.config.Debug {
		r.Meta = &ResultMeta{
			SQL:        buildResult.SQL,
			Params:     buildResult.Params,
			DurationMs: float64(time.Since(startTime).Microseconds()) / 1000,
		}
	}

	return r
}

// ExecuteCreateBatch executes a batch create query.
// ExecuteCreateBatch 执行批量创建查询。
func (e *Executor) ExecuteCreateBatch(ctx context.Context, query *Query) *Result {
	startTime := time.Now()

	builder := NewSQLBuilder(e.dialect, query)
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

	var ids []uint64
	affected := int64(len(query.DataBatch))

	if e.dialect.SupportsReturning() {
		// PostgreSQL: use RETURNING
		rows, err := e.db.sqlDB.QueryContext(ctx, buildResult.SQL, buildResult.Params...)
		if err != nil {
			return e.handleSQLError(err, buildResult)
		}
		defer rows.Close()

		for rows.Next() {
			var id uint64
			if err := rows.Scan(&id); err == nil {
				ids = append(ids, id)
			}
		}
	} else {
		// MySQL: execute and get last insert ID
		result, err := e.db.sqlDB.ExecContext(ctx, buildResult.SQL, buildResult.Params...)
		if err != nil {
			return e.handleSQLError(err, buildResult)
		}
		lastID, _ := result.LastInsertId()
		// MySQL auto-increment IDs are sequential
		for i := int64(0); i < affected; i++ {
			ids = append(ids, uint64(lastID+i))
		}
	}

	r := &Result{
		Success:  true,
		IDs:      ids,
		Affected: affected,
	}

	if query.Debug || e.db.config.Debug {
		r.Meta = &ResultMeta{
			SQL:        buildResult.SQL,
			Params:     buildResult.Params,
			DurationMs: float64(time.Since(startTime).Microseconds()) / 1000,
		}
	}

	return r
}

// ExecuteUpdate executes an update query.
// ExecuteUpdate 执行更新查询。
func (e *Executor) ExecuteUpdate(ctx context.Context, query *Query) *Result {
	// Check for destructive operation confirmation
	// 检查破坏性操作确认
	if e.db.config.Security.ConfirmDestructive && len(query.Where) == 0 {
		// UPDATE without WHERE - dangerous!
		// 没有 WHERE 的 UPDATE - 危险！
		count, err := e.getAffectedCount(ctx, query)
		if err == nil && count > int64(e.db.config.Security.ConfirmThreshold) {
			return &Result{
				Success:      false,
				Status:       "pending_confirm",
				ConfirmToken: generateConfirmToken(),
				Error: &ResultError{
					Code:    "CONFIRM_REQUIRED",
					Message: fmt.Sprintf("此操作将更新 %d 条记录，需要确认 / This will update %d records, confirmation required", count, count),
				},
			}
		}
	}

	return e.executeWriteQuery(ctx, query)
}

// ExecuteDelete executes a delete query.
// ExecuteDelete 执行删除查询。
func (e *Executor) ExecuteDelete(ctx context.Context, query *Query) *Result {
	// Check for destructive operation confirmation
	// 检查破坏性操作确认
	if e.db.config.Security.ConfirmDestructive && len(query.Where) == 0 {
		// DELETE without WHERE - very dangerous!
		// 没有 WHERE 的 DELETE - 非常危险！
		count, err := e.getAffectedCount(ctx, query)
		if err == nil && count > int64(e.db.config.Security.ConfirmThreshold) {
			return &Result{
				Success:      false,
				Status:       "pending_confirm",
				ConfirmToken: generateConfirmToken(),
				Error: &ResultError{
					Code:    "CONFIRM_REQUIRED",
					Message: fmt.Sprintf("此操作将删除 %d 条记录，需要确认 / This will delete %d records, confirmation required", count, count),
				},
			}
		}
	}

	return e.executeWriteQuery(ctx, query)
}

// executeWriteQuery executes an update or delete query.
// executeWriteQuery 执行更新或删除查询。
func (e *Executor) executeWriteQuery(ctx context.Context, query *Query) *Result {
	startTime := time.Now()

	builder := NewSQLBuilder(e.dialect, query)
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

	result, err := e.db.sqlDB.ExecContext(ctx, buildResult.SQL, buildResult.Params...)
	if err != nil {
		return e.handleSQLError(err, buildResult)
	}

	affected, _ := result.RowsAffected()

	r := &Result{
		Success:  true,
		Affected: affected,
	}

	if query.Debug || e.db.config.Debug {
		r.Meta = &ResultMeta{
			SQL:        buildResult.SQL,
			Params:     buildResult.Params,
			DurationMs: float64(time.Since(startTime).Microseconds()) / 1000,
		}
	}

	return r
}

// ExecuteCount executes a count query.
// ExecuteCount 执行计数查询。
func (e *Executor) ExecuteCount(ctx context.Context, query *Query) *Result {
	startTime := time.Now()

	builder := NewSQLBuilder(e.dialect, query)
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

	var count int64
	err = e.db.sqlDB.QueryRowContext(ctx, buildResult.SQL, buildResult.Params...).Scan(&count)
	if err != nil {
		return e.handleSQLError(err, buildResult)
	}

	r := &Result{
		Success: true,
		Count:   count,
	}

	if query.Debug || e.db.config.Debug {
		r.Meta = &ResultMeta{
			SQL:        buildResult.SQL,
			Params:     buildResult.Params,
			DurationMs: float64(time.Since(startTime).Microseconds()) / 1000,
		}
	}

	return r
}

// ExecuteAggregate executes an aggregate query.
// ExecuteAggregate 执行聚合查询。
func (e *Executor) ExecuteAggregate(ctx context.Context, query *Query) *Result {
	// Aggregate queries are similar to find queries
	// 聚合查询类似于查找查询
	return e.ExecuteFind(ctx, query)
}

// getAffectedCount gets the count of rows that would be affected.
// getAffectedCount 获取将受影响的行数。
func (e *Executor) getAffectedCount(ctx context.Context, query *Query) (int64, error) {
	countQuery := &Query{
		Table:  query.Table,
		Action: ActionCount,
		Where:  query.Where,
	}

	builder := NewSQLBuilder(e.dialect, countQuery)
	buildResult, err := builder.Build()
	if err != nil {
		return 0, err
	}

	var count int64
	err = e.db.sqlDB.QueryRowContext(ctx, buildResult.SQL, buildResult.Params...).Scan(&count)
	return count, err
}

// handleSQLError converts SQL errors to Result errors.
// handleSQLError 将 SQL 错误转换为 Result 错误。
func (e *Executor) handleSQLError(err error, buildResult *BuildResult) *Result {
	code := "SQL_ERROR"
	message := err.Error()
	suggestion := ""

	// Detect specific error types
	// 检测特定错误类型
	errStr := err.Error()

	switch {
	case containsAny(errStr, []string{"duplicate key", "Duplicate entry", "unique constraint"}):
		code = "DUPLICATE_KEY"
		suggestion = "使用 UPDATE 而不是 INSERT，或检查唯一约束 / Use UPDATE instead of INSERT, or check unique constraint"

	case containsAny(errStr, []string{"foreign key", "FOREIGN KEY", "a]"}):
		code = "FK_VIOLATION"
		suggestion = "确保引用的记录存在 / Ensure the referenced record exists"

	case containsAny(errStr, []string{"column", "does not exist", "Unknown column"}):
		code = "INVALID_COLUMN"
		suggestion = "检查列名拼写 / Check column name spelling"

	case containsAny(errStr, []string{"table", "does not exist", "doesn't exist"}):
		code = "TABLE_NOT_FOUND"
		suggestion = "运行 db.AutoSync() 创建表 / Run db.AutoSync() to create table"

	case containsAny(errStr, []string{"syntax error", "near"}):
		code = "SYNTAX_ERROR"
		suggestion = "检查 JQL 语法 / Check JQL syntax"

	case containsAny(errStr, []string{"timeout", "deadline exceeded"}):
		code = "TIMEOUT"
		suggestion = "增加超时时间或优化查询 / Increase timeout or optimize query"

	case containsAny(errStr, []string{"connection refused", "no connection"}):
		code = "CONNECTION_ERROR"
		suggestion = "检查数据库连接 / Check database connection"
	}

	return &Result{
		Success: false,
		Error: &ResultError{
			Code:       code,
			Message:    message,
			Suggestion: suggestion,
			Details: map[string]any{
				"sql":    buildResult.SQL,
				"params": buildResult.Params,
			},
		},
	}
}

// containsAny checks if s contains any of the substrings.
// containsAny 检查 s 是否包含任何子字符串。
func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// generateConfirmToken generates a confirmation token.
// generateConfirmToken 生成确认令牌。
func generateConfirmToken() string {
	// Simple implementation - in production use crypto/rand
	// 简单实现 - 生产环境使用 crypto/rand
	return fmt.Sprintf("confirm_%d", time.Now().UnixNano())
}
