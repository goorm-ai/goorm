package goorm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Transaction represents a database transaction.
// Transaction 表示数据库事务。
type Transaction struct {
	db      *DB
	tx      *sql.Tx
	results map[string]*Result
}

// executeTransactionInternal executes a JQL transaction.
// executeTransactionInternal 执行 JQL 事务。
func (db *DB) executeTransactionInternal(ctx context.Context, query *Query) *Result {
	if len(query.Operations) == 0 {
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "NO_OPERATIONS",
				Message: "transaction requires at least one operation",
			},
		}
	}

	// Start transaction
	// 开始事务
	tx, err := db.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "TX_BEGIN_ERROR",
				Message: err.Error(),
			},
		}
	}

	transaction := &Transaction{
		db:      db,
		tx:      tx,
		results: make(map[string]*Result),
	}

	// Execute each operation
	// 执行每个操作
	results := make([]Result, 0, len(query.Operations))
	transactionID := generateTransactionID()

	for i, op := range query.Operations {
		// Check for "check" action (conditional rollback)
		// 检查 "check" 操作（条件回滚）
		if op.Action == "check" {
			// Currently not implemented - just skip
			// TODO: Implement condition checking
			continue
		}

		// Resolve references from previous operations
		// 解析来自之前操作的引用
		resolvedOp := transaction.resolveReferences(op)

		// Execute the operation
		// 执行操作
		result := transaction.executeOperation(ctx, resolvedOp)

		if !result.Success {
			// Rollback on error
			// 出错时回滚
			tx.Rollback()
			return &Result{
				Success: false,
				Error: &ResultError{
					Code:    "TX_OPERATION_ERROR",
					Message: fmt.Sprintf("operation %d failed: %s", i+1, result.Error.Message),
					Details: map[string]any{
						"operation_index": i,
						"operation":       op,
					},
				},
			}
		}

		// Store result with alias if provided
		// 如果提供了别名则存储结果
		if op.As != "" {
			transaction.results[op.As] = result
		}

		stepResult := *result
		if op.As != "" {
			stepResult.Meta = &ResultMeta{SQL: op.As}
		}
		results = append(results, stepResult)
	}

	// Commit transaction
	// 提交事务
	if err := tx.Commit(); err != nil {
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "TX_COMMIT_ERROR",
				Message: err.Error(),
			},
		}
	}

	return &Result{
		Success:       true,
		TransactionID: transactionID,
		Results:       results,
	}
}

// executeOperation executes a single operation within a transaction.
// executeOperation 在事务中执行单个操作。
func (t *Transaction) executeOperation(ctx context.Context, query *Query) *Result {
	builder := NewSQLBuilder(t.db.dialect, query)
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

	switch query.Action {
	case ActionCreate:
		return t.executeCreate(ctx, buildResult)
	case ActionUpdate, ActionDelete:
		return t.executeWrite(ctx, buildResult)
	case ActionFind:
		return t.executeFind(ctx, buildResult)
	default:
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "UNSUPPORTED_TX_ACTION",
				Message: fmt.Sprintf("action %s not supported in transaction", query.Action),
			},
		}
	}
}

// executeCreate executes a create operation in transaction.
// executeCreate 在事务中执行创建操作。
func (t *Transaction) executeCreate(ctx context.Context, build *BuildResult) *Result {
	var lastID uint64

	if t.db.dialect.SupportsReturning() {
		err := t.tx.QueryRowContext(ctx, build.SQL, build.Params...).Scan(&lastID)
		if err != nil && err != sql.ErrNoRows {
			return &Result{
				Success: false,
				Error: &ResultError{
					Code:    "CREATE_ERROR",
					Message: err.Error(),
				},
			}
		}
	} else {
		result, err := t.tx.ExecContext(ctx, build.SQL, build.Params...)
		if err != nil {
			return &Result{
				Success: false,
				Error: &ResultError{
					Code:    "CREATE_ERROR",
					Message: err.Error(),
				},
			}
		}
		id, _ := result.LastInsertId()
		lastID = uint64(id)
	}

	return &Result{
		Success:  true,
		ID:       lastID,
		Affected: 1,
	}
}

// executeWrite executes an update/delete operation in transaction.
// executeWrite 在事务中执行更新/删除操作。
func (t *Transaction) executeWrite(ctx context.Context, build *BuildResult) *Result {
	result, err := t.tx.ExecContext(ctx, build.SQL, build.Params...)
	if err != nil {
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "WRITE_ERROR",
				Message: err.Error(),
			},
		}
	}

	affected, _ := result.RowsAffected()
	return &Result{
		Success:  true,
		Affected: affected,
	}
}

// executeFind executes a find operation in transaction.
// executeFind 在事务中执行查询操作。
func (t *Transaction) executeFind(ctx context.Context, build *BuildResult) *Result {
	rows, err := t.tx.QueryContext(ctx, build.SQL, build.Params...)
	if err != nil {
		return &Result{
			Success: false,
			Error: &ResultError{
				Code:    "QUERY_ERROR",
				Message: err.Error(),
			},
		}
	}
	defer rows.Close()

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

	data := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]any, len(columns))
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			row[col] = val
		}
		data = append(data, row)
	}

	return &Result{
		Success: true,
		Data:    data,
		Count:   int64(len(data)),
	}
}

// resolveReferences replaces $reference.field with actual values.
// resolveReferences 将 $reference.field 替换为实际值。
func (t *Transaction) resolveReferences(query Query) *Query {
	result := query

	// Check data for references
	// 检查 data 中的引用
	if result.Data != nil {
		newData := make(map[string]any, len(result.Data))
		for k, v := range result.Data {
			newData[k] = t.resolveValue(v)
		}
		result.Data = newData
	}

	// Check where conditions for references
	// 检查 where 条件中的引用
	if result.Where != nil {
		newWhere := make([]Condition, len(result.Where))
		for i, cond := range result.Where {
			newCond := cond
			newCond.Value = t.resolveValue(cond.Value)
			newWhere[i] = newCond
		}
		result.Where = newWhere
	}

	return &result
}

// resolveValue resolves a reference value like "$new_user.id".
// resolveValue 解析引用值如 "$new_user.id"。
func (t *Transaction) resolveValue(v any) any {
	str, ok := v.(string)
	if !ok {
		return v
	}

	if !strings.HasPrefix(str, "$") {
		return v
	}

	// Parse $alias.field
	// 解析 $alias.field
	parts := strings.SplitN(str[1:], ".", 2)
	if len(parts) != 2 {
		return v
	}

	alias := parts[0]
	field := parts[1]

	result, exists := t.results[alias]
	if !exists {
		return v
	}

	// Get field value from result
	// 从结果获取字段值
	switch field {
	case "id":
		return result.ID
	case "affected":
		return result.Affected
	case "count":
		return result.Count
	default:
		// Try to get from first data row
		// 尝试从第一个数据行获取
		if len(result.Data) > 0 {
			if val, ok := result.Data[0][field]; ok {
				return val
			}
		}
		return v
	}
}

// generateTransactionID generates a unique transaction ID.
// generateTransactionID 生成唯一的事务 ID。
func generateTransactionID() string {
	return fmt.Sprintf("tx_%d", time.Now().UnixNano())
}

// Begin starts a manual transaction.
// Begin 开始一个手动事务。
func (db *DB) Begin() (*Transaction, error) {
	return db.BeginContext(db.ctx)
}

// BeginContext starts a manual transaction with context.
// BeginContext 使用上下文开始一个手动事务。
func (db *DB) BeginContext(ctx context.Context) (*Transaction, error) {
	tx, err := db.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &Transaction{
		db:      db,
		tx:      tx,
		results: make(map[string]*Result),
	}, nil
}

// Commit commits the transaction.
// Commit 提交事务。
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction.
// Rollback 回滚事务。
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

// Execute executes a JQL query within the transaction.
// Execute 在事务中执行 JQL 查询。
func (t *Transaction) Execute(jql string) *Result {
	return t.ExecuteContext(context.Background(), jql)
}

// ExecuteContext executes a JQL query with context within the transaction.
// ExecuteContext 使用上下文在事务中执行 JQL 查询。
func (t *Transaction) ExecuteContext(ctx context.Context, jql string) *Result {
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

	return t.executeOperation(ctx, query)
}
