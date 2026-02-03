package goorm

import (
	"encoding/json"
	"fmt"
)

// Action represents the type of database operation in JQL.
// Action 表示 JQL 中的数据库操作类型。
type Action string

// Supported JQL actions.
// 支持的 JQL 操作。
const (
	// ActionFind retrieves records from the database.
	// ActionFind 从数据库检索记录。
	ActionFind Action = "find"

	// ActionCreate inserts new records into the database.
	// ActionCreate 向数据库插入新记录。
	ActionCreate Action = "create"

	// ActionCreateBatch inserts multiple records at once.
	// ActionCreateBatch 批量插入多条记录。
	ActionCreateBatch Action = "create_batch"

	// ActionUpdate modifies existing records.
	// ActionUpdate 修改现有记录。
	ActionUpdate Action = "update"

	// ActionDelete removes records from the database.
	// ActionDelete 从数据库删除记录。
	ActionDelete Action = "delete"

	// ActionCount returns the number of matching records.
	// ActionCount 返回匹配记录的数量。
	ActionCount Action = "count"

	// ActionAggregate performs aggregation operations (SUM, AVG, etc.).
	// ActionAggregate 执行聚合操作（SUM、AVG 等）。
	ActionAggregate Action = "aggregate"

	// ActionTransaction executes multiple operations atomically.
	// ActionTransaction 原子地执行多个操作。
	ActionTransaction Action = "transaction"

	// ActionExplain returns the SQL that would be generated without executing.
	// ActionExplain 返回将生成的 SQL 而不执行。
	ActionExplain Action = "explain"

	// ActionValidate checks if the query is valid without executing.
	// ActionValidate 检查查询是否有效而不执行。
	ActionValidate Action = "validate"

	// ActionListTables returns all registered tables.
	// ActionListTables 返回所有已注册的表。
	ActionListTables Action = "list_tables"

	// ActionDescribe returns the schema of a table.
	// ActionDescribe 返回表的 Schema。
	ActionDescribe Action = "describe"
)

// Operator represents a comparison operator in JQL conditions.
// Operator 表示 JQL 条件中的比较运算符。
type Operator string

// Supported JQL operators.
// 支持的 JQL 运算符。
const (
	OpEqual       Operator = "="        // Equal / 等于
	OpNotEqual    Operator = "!="       // Not equal / 不等于
	OpGreater     Operator = ">"        // Greater than / 大于
	OpGreaterOrEq Operator = ">="       // Greater than or equal / 大于等于
	OpLess        Operator = "<"        // Less than / 小于
	OpLessOrEq    Operator = "<="       // Less than or equal / 小于等于
	OpIn          Operator = "in"       // In array / 在数组中
	OpNotIn       Operator = "not_in"   // Not in array / 不在数组中
	OpLike        Operator = "like"     // Pattern match / 模式匹配
	OpILike       Operator = "ilike"    // Case-insensitive pattern match / 不区分大小写模式匹配
	OpNotLike     Operator = "not_like" // Not pattern match / 不匹配模式
	OpBetween     Operator = "between"  // Between range / 在范围内
	OpNull        Operator = "null"     // Is null / 为空
	OpNotNull     Operator = "not_null" // Is not null / 不为空
	OpExists      Operator = "exists"   // Exists subquery / 存在子查询
)

// Query represents a JQL query structure.
// This is the core data structure that AI generates and GoORM executes.
//
// Query 表示 JQL 查询结构。
// 这是 AI 生成、GoORM 执行的核心数据结构。
type Query struct {
	// Table is the target table name.
	// Table 是目标表名。
	Table string `json:"table,omitempty"`

	// Action is the operation type (find, create, update, delete, etc.).
	// Action 是操作类型（find、create、update、delete 等）。
	Action Action `json:"action"`

	// Where contains the query conditions.
	// Where 包含查询条件。
	Where []Condition `json:"where,omitempty"`

	// Data contains the data for create/update operations.
	// Data 包含 create/update 操作的数据。
	Data map[string]any `json:"data,omitempty"`

	// DataBatch contains multiple records for batch create.
	// DataBatch 包含批量创建的多条记录。
	DataBatch []map[string]any `json:"data_batch,omitempty"`

	// Select specifies which columns to return.
	// Select 指定返回哪些列。
	Select []any `json:"select,omitempty"`

	// OrderBy specifies the sort order.
	// OrderBy 指定排序顺序。
	OrderBy []Order `json:"order_by,omitempty"`

	// GroupBy specifies the grouping columns.
	// GroupBy 指定分组列。
	GroupBy []string `json:"group_by,omitempty"`

	// Having specifies conditions for grouped results.
	// Having 指定分组结果的条件。
	Having []HavingCondition `json:"having,omitempty"`

	// Limit restricts the number of results.
	// Limit 限制结果数量。
	Limit int `json:"limit,omitempty"`

	// Offset skips the first N results.
	// Offset 跳过前 N 条结果。
	Offset int `json:"offset,omitempty"`

	// With specifies relations to preload.
	// With 指定要预加载的关联。
	With []any `json:"with,omitempty"`

	// Join specifies explicit join operations.
	// Join 指定显式的 JOIN 操作。
	Join []JoinClause `json:"join,omitempty"`

	// Has filters records that have the specified relation.
	// Has 过滤具有指定关联的记录。
	Has any `json:"has,omitempty"`

	// Operations contains sub-operations for transactions.
	// Operations 包含事务的子操作。
	Operations []Query `json:"operations,omitempty"`

	// As is an alias for the result (used in transactions).
	// As 是结果的别名（用于事务）。
	As string `json:"as,omitempty"`

	// Timeout specifies the query timeout duration.
	// Timeout 指定查询超时时间。
	Timeout string `json:"timeout,omitempty"`

	// Debug enables debug mode to return SQL in response.
	// Debug 启用调试模式，在响应中返回 SQL。
	Debug bool `json:"debug,omitempty"`

	// QueryToExplain is the query to explain (for ActionExplain).
	// QueryToExplain 是要解释的查询（用于 ActionExplain）。
	QueryToExplain *Query `json:"query,omitempty"`
}

// Condition represents a WHERE condition in JQL.
// Condition 表示 JQL 中的 WHERE 条件。
type Condition struct {
	// Field is the column name.
	// Field 是列名。
	Field string `json:"field,omitempty"`

	// Op is the comparison operator.
	// Op 是比较运算符。
	Op Operator `json:"op"`

	// Value is the comparison value.
	// Value 是比较值。
	Value any `json:"value,omitempty"`

	// Or indicates this condition should use OR instead of AND.
	// Or 表示此条件应使用 OR 而不是 AND。
	Or bool `json:"or,omitempty"`

	// Ref is a reference to another table's column (for correlated subqueries).
	// Ref 是对另一个表的列的引用（用于相关子查询）。
	Ref string `json:"ref,omitempty"`

	// Subquery is a nested query (for IN subqueries).
	// Subquery 是嵌套查询（用于 IN 子查询）。
	Subquery *Query `json:"subquery,omitempty"`

	// And contains nested AND conditions.
	// And 包含嵌套的 AND 条件。
	And []Condition `json:"and,omitempty"`

	// OrGroup contains nested OR conditions.
	// OrGroup 包含嵌套的 OR 条件。
	OrGroup []Condition `json:"or_group,omitempty"`
}

// Order represents ORDER BY clause in JQL.
// Order 表示 JQL 中的 ORDER BY 子句。
type Order struct {
	// Field is the column to sort by.
	// Field 是排序的列。
	Field string `json:"field"`

	// Desc indicates descending order.
	// Desc 表示降序。
	Desc bool `json:"desc,omitempty"`
}

// HavingCondition represents a HAVING clause condition.
// HavingCondition 表示 HAVING 子句条件。
type HavingCondition struct {
	// Fn is the aggregate function (count, sum, avg, etc.).
	// Fn 是聚合函数（count、sum、avg 等）。
	Fn string `json:"fn"`

	// Field is the column to aggregate (optional for count).
	// Field 是要聚合的列（对于 count 可选）。
	Field string `json:"field,omitempty"`

	// Op is the comparison operator.
	// Op 是比较运算符。
	Op Operator `json:"op"`

	// Value is the comparison value.
	// Value 是比较值。
	Value any `json:"value"`
}

// JoinClause represents a JOIN clause in JQL.
// JoinClause 表示 JQL 中的 JOIN 子句。
type JoinClause struct {
	// Table is the table to join.
	// Table 是要连接的表。
	Table string `json:"table"`

	// Type is the join type (left, right, inner, full).
	// Type 是连接类型（left、right、inner、full）。
	Type string `json:"type,omitempty"`

	// On specifies the join conditions as column mappings.
	// On 指定连接条件，以列映射形式。
	On map[string]string `json:"on"`
}

// AggregateField represents an aggregate function in SELECT.
// AggregateField 表示 SELECT 中的聚合函数。
type AggregateField struct {
	// Fn is the aggregate function name (count, sum, avg, max, min).
	// Fn 是聚合函数名（count、sum、avg、max、min）。
	Fn string `json:"fn"`

	// Field is the column to aggregate (optional for count(*)).
	// Field 是要聚合的列（对于 count(*) 可选）。
	Field string `json:"field,omitempty"`

	// As is the alias for the result.
	// As 是结果的别名。
	As string `json:"as"`
}

// ParseQuery parses a JQL JSON string into a Query struct.
// Returns an error if the JSON is invalid or cannot be parsed.
//
// ParseQuery 将 JQL JSON 字符串解析为 Query 结构体。
// 如果 JSON 无效或无法解析，则返回错误。
func ParseQuery(jql string) (*Query, error) {
	var query Query
	if err := json.Unmarshal([]byte(jql), &query); err != nil {
		return nil, fmt.Errorf("failed to parse JQL: %w", err)
	}
	return &query, nil
}

// String returns the JSON representation of the Query.
// String 返回 Query 的 JSON 表示。
func (q *Query) String() string {
	data, _ := json.MarshalIndent(q, "", "  ")
	return string(data)
}

// Validate checks if the Query is valid.
// Returns an error describing the validation failure, or nil if valid.
//
// Validate 检查 Query 是否有效。
// 返回描述验证失败的错误，如果有效则返回 nil。
func (q *Query) Validate() error {
	if q.Action == "" {
		return fmt.Errorf("action is required")
	}

	switch q.Action {
	case ActionFind, ActionCount, ActionDelete, ActionUpdate, ActionAggregate:
		if q.Table == "" {
			return fmt.Errorf("table is required for action %q", q.Action)
		}
	case ActionCreate, ActionCreateBatch:
		if q.Table == "" {
			return fmt.Errorf("table is required for action %q", q.Action)
		}
		if q.Action == ActionCreate && len(q.Data) == 0 {
			return fmt.Errorf("data is required for action %q", q.Action)
		}
		if q.Action == ActionCreateBatch && len(q.DataBatch) == 0 {
			return fmt.Errorf("data_batch is required for action %q", q.Action)
		}
	case ActionTransaction:
		if len(q.Operations) == 0 {
			return fmt.Errorf("operations is required for action %q", q.Action)
		}
	case ActionListTables:
		// No additional validation needed
	case ActionDescribe:
		if q.Table == "" {
			return fmt.Errorf("table is required for action %q", q.Action)
		}
	case ActionExplain, ActionValidate:
		if q.QueryToExplain == nil {
			return fmt.Errorf("query is required for action %q", q.Action)
		}
	default:
		return fmt.Errorf("unknown action: %q", q.Action)
	}

	return nil
}
