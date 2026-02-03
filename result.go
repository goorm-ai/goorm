package goorm

import (
	"encoding/json"
)

// Result represents the response from a JQL execution.
// All GoORM operations return this standardized structure.
//
// Result 表示 JQL 执行的响应。
// 所有 GoORM 操作都返回这个标准化结构。
type Result struct {
	// Success indicates whether the operation succeeded.
	// Success 表示操作是否成功。
	Success bool `json:"success"`

	// Data contains the query results (for find operations).
	// Data 包含查询结果（用于 find 操作）。
	Data []map[string]any `json:"data,omitempty"`

	// Count is the number of matching records (for find/count operations).
	// Count 是匹配记录的数量（用于 find/count 操作）。
	Count int64 `json:"count,omitempty"`

	// Affected is the number of affected rows (for update/delete operations).
	// Affected 是受影响的行数（用于 update/delete 操作）。
	Affected int64 `json:"affected,omitempty"`

	// ID is the last inserted ID (for create operations).
	// ID 是最后插入的 ID（用于 create 操作）。
	ID uint64 `json:"id,omitempty"`

	// IDs contains all inserted IDs (for batch create operations).
	// IDs 包含所有插入的 ID（用于批量 create 操作）。
	IDs []uint64 `json:"ids,omitempty"`

	// Meta contains additional metadata about the query.
	// Meta 包含查询的附加元数据。
	Meta *ResultMeta `json:"meta,omitempty"`

	// Error contains error information if the operation failed.
	// Error 包含操作失败时的错误信息。
	Error *ResultError `json:"error,omitempty"`

	// Status indicates special states like "pending_confirm".
	// Status 表示特殊状态，如 "pending_confirm"。
	Status string `json:"status,omitempty"`

	// ConfirmToken is used for confirming destructive operations.
	// ConfirmToken 用于确认破坏性操作。
	ConfirmToken string `json:"confirm_token,omitempty"`

	// Preview contains sample data for pending operations.
	// Preview 包含待处理操作的示例数据。
	Preview []map[string]any `json:"preview,omitempty"`

	// TransactionID is the ID of the transaction (for transaction operations).
	// TransactionID 是事务的 ID（用于事务操作）。
	TransactionID string `json:"transaction_id,omitempty"`

	// Results contains sub-results for transaction operations.
	// Results 包含事务操作的子结果。
	Results []Result `json:"results,omitempty"`

	// Tables contains table information (for list_tables operation).
	// Tables 包含表信息（用于 list_tables 操作）。
	Tables []TableInfo `json:"tables,omitempty"`

	// Schema contains table schema (for describe operation).
	// Schema 包含表 Schema（用于 describe 操作）。
	Schema *TableSchema `json:"schema,omitempty"`

	// Explain contains query explanation (for explain operation).
	// Explain 包含查询解释（用于 explain 操作）。
	Explain *ExplainResult `json:"explain,omitempty"`

	// NLQuery contains the original natural language query (for NL operations).
	// NLQuery 包含原始自然语言查询（用于 NL 操作）。
	NLQuery string `json:"nl_query,omitempty"`

	// ParsedJQL contains the JQL parsed from natural language (for NL operations).
	// ParsedJQL 包含从自然语言解析的 JQL（用于 NL 操作）。
	ParsedJQL string `json:"parsed_jql,omitempty"`
}

// ResultMeta contains metadata about the query execution.
// ResultMeta 包含查询执行的元数据。
type ResultMeta struct {
	// SQL is the generated SQL statement (when debug mode is enabled).
	// SQL 是生成的 SQL 语句（启用调试模式时）。
	SQL string `json:"sql,omitempty"`

	// Params contains the SQL parameters.
	// Params 包含 SQL 参数。
	Params []any `json:"params,omitempty"`

	// DurationMs is the execution time in milliseconds.
	// DurationMs 是执行时间（毫秒）。
	DurationMs float64 `json:"duration_ms,omitempty"`

	// RowsScanned is the number of rows scanned by the database.
	// RowsScanned 是数据库扫描的行数。
	RowsScanned int64 `json:"rows_scanned,omitempty"`

	// RowsReturned is the number of rows returned.
	// RowsReturned 是返回的行数。
	RowsReturned int64 `json:"rows_returned,omitempty"`
}

// ResultError contains error information.
// ResultError 包含错误信息。
type ResultError struct {
	// Code is the error code (e.g., "INVALID_FIELD", "FK_VIOLATION").
	// Code 是错误代码（如 "INVALID_FIELD"、"FK_VIOLATION"）。
	Code string `json:"code"`

	// Message is a human-readable error message.
	// Message 是人类可读的错误消息。
	Message string `json:"message"`

	// Details contains additional error details.
	// Details 包含额外的错误详情。
	Details map[string]any `json:"details,omitempty"`

	// Suggestion is a suggested fix for the error.
	// Suggestion 是对错误的修复建议。
	Suggestion string `json:"suggestion,omitempty"`

	// ValidFields lists valid field names (for INVALID_FIELD errors).
	// ValidFields 列出有效的字段名（用于 INVALID_FIELD 错误）。
	ValidFields []string `json:"valid_fields,omitempty"`

	// AutoFix contains a query that could fix the issue.
	// AutoFix 包含可以修复问题的查询。
	AutoFix *Query `json:"auto_fix,omitempty"`
}

// TableInfo contains information about a registered table.
// TableInfo 包含已注册表的信息。
type TableInfo struct {
	// Name is the table name in the database.
	// Name 是数据库中的表名。
	Name string `json:"name"`

	// Model is the Go model name.
	// Model 是 Go 模型名。
	Model string `json:"model"`

	// Description is the table description.
	// Description 是表描述。
	Description string `json:"description,omitempty"`

	// Columns lists the column names.
	// Columns 列出列名。
	Columns []string `json:"columns"`

	// PrimaryKey is the primary key column name.
	// PrimaryKey 是主键列名。
	PrimaryKey string `json:"primary_key"`
}

// TableSchema contains detailed schema information for a table.
// TableSchema 包含表的详细 Schema 信息。
type TableSchema struct {
	// Table is the table name.
	// Table 是表名。
	Table string `json:"table"`

	// Model is the Go model name.
	// Model 是 Go 模型名。
	Model string `json:"model"`

	// Description is the table description.
	// Description 是表描述。
	Description string `json:"description,omitempty"`

	// Columns contains detailed column information.
	// Columns 包含详细的列信息。
	Columns []ColumnSchema `json:"columns"`

	// Indexes contains index information.
	// Indexes 包含索引信息。
	Indexes []IndexSchema `json:"indexes,omitempty"`

	// Relations contains relation information.
	// Relations 包含关联信息。
	Relations []RelationSchema `json:"relations,omitempty"`
}

// ColumnSchema contains schema information for a column.
// ColumnSchema 包含列的 Schema 信息。
type ColumnSchema struct {
	// Name is the column name.
	// Name 是列名。
	Name string `json:"name"`

	// Type is the SQL type.
	// Type 是 SQL 类型。
	Type string `json:"type"`

	// GoType is the Go type.
	// GoType 是 Go 类型。
	GoType string `json:"go_type,omitempty"`

	// Nullable indicates if the column can be NULL.
	// Nullable 表示列是否可为 NULL。
	Nullable bool `json:"nullable,omitempty"`

	// Primary indicates if this is the primary key.
	// Primary 表示是否为主键。
	Primary bool `json:"primary,omitempty"`

	// Unique indicates if the column has a unique constraint.
	// Unique 表示列是否有唯一约束。
	Unique bool `json:"unique,omitempty"`

	// Default is the default value.
	// Default 是默认值。
	Default string `json:"default,omitempty"`

	// Description is the column description.
	// Description 是列描述。
	Description string `json:"desc,omitempty"`

	// Sensitive indicates if this is a sensitive field.
	// Sensitive 表示是否为敏感字段。
	Sensitive bool `json:"sensitive,omitempty"`

	// Mask is the masking strategy for sensitive fields.
	// Mask 是敏感字段的脱敏策略。
	Mask string `json:"mask,omitempty"`
}

// IndexSchema contains schema information for an index.
// IndexSchema 包含索引的 Schema 信息。
type IndexSchema struct {
	// Name is the index name.
	// Name 是索引名。
	Name string `json:"name"`

	// Columns lists the indexed columns.
	// Columns 列出索引的列。
	Columns []string `json:"columns"`

	// Unique indicates if this is a unique index.
	// Unique 表示是否为唯一索引。
	Unique bool `json:"unique,omitempty"`
}

// RelationSchema contains schema information for a relation.
// RelationSchema 包含关联的 Schema 信息。
type RelationSchema struct {
	// Name is the relation name (field name in Go struct).
	// Name 是关联名（Go 结构体中的字段名）。
	Name string `json:"name"`

	// Type is the relation type (has_one, has_many, belongs_to, many_to_many).
	// Type 是关联类型（has_one、has_many、belongs_to、many_to_many）。
	Type string `json:"type"`

	// Model is the related model/table name.
	// Model 是关联的模型/表名。
	Model string `json:"model"`

	// Target is the target table name (alias for Model).
	// Target 是目标表名（Model 的别名）。
	Target string `json:"target,omitempty"`

	// ForeignKey is the foreign key column.
	// ForeignKey 是外键列。
	ForeignKey string `json:"fk,omitempty"`

	// ReferenceKey is the referenced column (usually primary key).
	// ReferenceKey 是引用的列（通常是主键）。
	ReferenceKey string `json:"ref,omitempty"`

	// JoinTable is the join table name (for many_to_many).
	// JoinTable 是连接表名（用于 many_to_many）。
	JoinTable string `json:"join_table,omitempty"`

	// JoinForeignKey is the foreign key in join table for this model.
	// JoinForeignKey 是连接表中此模型的外键。
	JoinForeignKey string `json:"join_fk,omitempty"`

	// JoinReferenceKey is the foreign key in join table for related model.
	// JoinReferenceKey 是连接表中关联模型的外键。
	JoinReferenceKey string `json:"join_ref,omitempty"`
}

// ExplainResult contains the query explanation.
// ExplainResult 包含查询解释。
type ExplainResult struct {
	// SQL is the generated SQL statement.
	// SQL 是生成的 SQL 语句。
	SQL string `json:"sql"`

	// Params contains the SQL parameters.
	// Params 包含 SQL 参数。
	Params []any `json:"params"`

	// EstimatedRows is the estimated number of rows.
	// EstimatedRows 是预估的行数。
	EstimatedRows int64 `json:"estimated_rows,omitempty"`

	// EstimatedCost is the estimated query cost.
	// EstimatedCost 是预估的查询成本。
	EstimatedCost float64 `json:"estimated_cost,omitempty"`

	// IndexUsed lists the indexes used by the query.
	// IndexUsed 列出查询使用的索引。
	IndexUsed []string `json:"index_used,omitempty"`

	// IndexSuggested lists suggested indexes to improve performance.
	// IndexSuggested 列出建议的索引以提高性能。
	IndexSuggested []string `json:"index_suggested,omitempty"`

	// Warnings contains any warnings about the query.
	// Warnings 包含关于查询的警告。
	Warnings []string `json:"warnings,omitempty"`
}

// String returns the JSON representation of the Result.
// String 返回 Result 的 JSON 表示。
func (r *Result) String() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// IsSuccess returns true if the operation succeeded.
// IsSuccess 如果操作成功则返回 true。
func (r *Result) IsSuccess() bool {
	return r.Success
}

// IsPendingConfirm returns true if the operation requires confirmation.
// IsPendingConfirm 如果操作需要确认则返回 true。
func (r *Result) IsPendingConfirm() bool {
	return r.Status == "pending_confirm"
}

// Err returns an error if the operation failed, nil otherwise.
// Err 如果操作失败返回错误，否则返回 nil。
func (r *Result) Err() error {
	if r.Error != nil {
		return &QueryError{
			Code:       r.Error.Code,
			Message:    r.Error.Message,
			Suggestion: r.Error.Suggestion,
		}
	}
	return nil
}

// First returns the first record from the result, or nil if empty.
// First 返回结果中的第一条记录，如果为空则返回 nil。
func (r *Result) First() map[string]any {
	if len(r.Data) > 0 {
		return r.Data[0]
	}
	return nil
}

// QueryError represents an error from query execution.
// QueryError 表示查询执行的错误。
type QueryError struct {
	Code       string
	Message    string
	Suggestion string
}

// Error implements the error interface.
// Error 实现 error 接口。
func (e *QueryError) Error() string {
	if e.Suggestion != "" {
		return e.Message + " (suggestion: " + e.Suggestion + ")"
	}
	return e.Message
}
