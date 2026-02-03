package goorm

import (
	"fmt"
	"strings"
)

// SQLBuilder builds SQL statements from JQL queries.
// It handles dialect-specific differences and parameter binding.
//
// SQLBuilder 从 JQL 查询构建 SQL 语句。
// 它处理方言特定的差异和参数绑定。
type SQLBuilder struct {
	dialect Dialect
	query   *Query
	params  []any
	paramN  int
}

// BuildResult contains the built SQL and parameters.
// BuildResult 包含构建的 SQL 和参数。
type BuildResult struct {
	SQL    string
	Params []any
}

// NewSQLBuilder creates a new SQL builder.
// NewSQLBuilder 创建新的 SQL 构建器。
func NewSQLBuilder(dialect Dialect, query *Query) *SQLBuilder {
	return &SQLBuilder{
		dialect: dialect,
		query:   query,
		params:  make([]any, 0),
		paramN:  0,
	}
}

// Build builds the SQL statement based on the query action.
// Build 根据查询操作构建 SQL 语句。
func (b *SQLBuilder) Build() (*BuildResult, error) {
	var sql string
	var err error

	switch b.query.Action {
	case ActionFind:
		sql, err = b.buildSelect()
	case ActionCreate:
		sql, err = b.buildInsert()
	case ActionCreateBatch:
		sql, err = b.buildInsertBatch()
	case ActionUpdate:
		sql, err = b.buildUpdate()
	case ActionDelete:
		sql, err = b.buildDelete()
	case ActionCount:
		sql, err = b.buildCount()
	case ActionAggregate:
		sql, err = b.buildAggregate()
	default:
		return nil, fmt.Errorf("unsupported action for SQL building: %s", b.query.Action)
	}

	if err != nil {
		return nil, err
	}

	return &BuildResult{
		SQL:    sql,
		Params: b.params,
	}, nil
}

// buildSelect builds a SELECT statement.
// buildSelect 构建 SELECT 语句。
func (b *SQLBuilder) buildSelect() (string, error) {
	var sb strings.Builder

	// SELECT clause
	// SELECT 子句
	sb.WriteString("SELECT ")
	if len(b.query.Select) > 0 {
		sb.WriteString(b.buildSelectColumns())
	} else {
		sb.WriteString("*")
	}

	// FROM clause
	// FROM 子句
	sb.WriteString(" FROM ")
	sb.WriteString(b.dialect.Quote(b.query.Table))

	// JOIN clause
	// JOIN 子句
	if len(b.query.Join) > 0 {
		joinSQL, err := b.buildJoins()
		if err != nil {
			return "", err
		}
		sb.WriteString(joinSQL)
	}

	// WHERE clause
	// WHERE 子句
	if len(b.query.Where) > 0 {
		whereSQL, err := b.buildWhere()
		if err != nil {
			return "", err
		}
		sb.WriteString(" WHERE ")
		sb.WriteString(whereSQL)
	}

	// GROUP BY clause
	// GROUP BY 子句
	if len(b.query.GroupBy) > 0 {
		sb.WriteString(" GROUP BY ")
		sb.WriteString(b.buildGroupBy())
	}

	// HAVING clause
	// HAVING 子句
	if len(b.query.Having) > 0 {
		havingSQL, err := b.buildHaving()
		if err != nil {
			return "", err
		}
		sb.WriteString(" HAVING ")
		sb.WriteString(havingSQL)
	}

	// ORDER BY clause
	// ORDER BY 子句
	if len(b.query.OrderBy) > 0 {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(b.buildOrderBy())
	}

	// LIMIT clause
	// LIMIT 子句
	if b.query.Limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", b.query.Limit))
	}

	// OFFSET clause
	// OFFSET 子句
	if b.query.Offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", b.query.Offset))
	}

	return sb.String(), nil
}

// buildInsert builds an INSERT statement.
// buildInsert 构建 INSERT 语句。
func (b *SQLBuilder) buildInsert() (string, error) {
	if len(b.query.Data) == 0 {
		return "", fmt.Errorf("no data provided for insert")
	}

	var sb strings.Builder

	// Collect columns and values
	// 收集列和值
	columns := make([]string, 0, len(b.query.Data))
	placeholders := make([]string, 0, len(b.query.Data))

	for col, val := range b.query.Data {
		columns = append(columns, b.dialect.Quote(col))
		placeholders = append(placeholders, b.addParam(val))
	}

	sb.WriteString("INSERT INTO ")
	sb.WriteString(b.dialect.Quote(b.query.Table))
	sb.WriteString(" (")
	sb.WriteString(strings.Join(columns, ", "))
	sb.WriteString(") VALUES (")
	sb.WriteString(strings.Join(placeholders, ", "))
	sb.WriteString(")")

	// Add RETURNING for PostgreSQL
	// 为 PostgreSQL 添加 RETURNING
	if b.dialect.SupportsReturning() {
		sb.WriteString(" RETURNING id")
	}

	return sb.String(), nil
}

// buildInsertBatch builds a batch INSERT statement.
// buildInsertBatch 构建批量 INSERT 语句。
func (b *SQLBuilder) buildInsertBatch() (string, error) {
	if len(b.query.DataBatch) == 0 {
		return "", fmt.Errorf("no data provided for batch insert")
	}

	var sb strings.Builder

	// Get columns from first record
	// 从第一条记录获取列
	firstRecord := b.query.DataBatch[0]
	columns := make([]string, 0, len(firstRecord))
	columnNames := make([]string, 0, len(firstRecord))

	for col := range firstRecord {
		columns = append(columns, b.dialect.Quote(col))
		columnNames = append(columnNames, col)
	}

	sb.WriteString("INSERT INTO ")
	sb.WriteString(b.dialect.Quote(b.query.Table))
	sb.WriteString(" (")
	sb.WriteString(strings.Join(columns, ", "))
	sb.WriteString(") VALUES ")

	// Build value rows
	// 构建值行
	valueRows := make([]string, 0, len(b.query.DataBatch))
	for _, record := range b.query.DataBatch {
		placeholders := make([]string, len(columnNames))
		for i, col := range columnNames {
			placeholders[i] = b.addParam(record[col])
		}
		valueRows = append(valueRows, "("+strings.Join(placeholders, ", ")+")")
	}

	sb.WriteString(strings.Join(valueRows, ", "))

	// Add RETURNING for PostgreSQL
	// 为 PostgreSQL 添加 RETURNING
	if b.dialect.SupportsReturning() {
		sb.WriteString(" RETURNING id")
	}

	return sb.String(), nil
}

// buildUpdate builds an UPDATE statement.
// buildUpdate 构建 UPDATE 语句。
func (b *SQLBuilder) buildUpdate() (string, error) {
	if len(b.query.Data) == 0 {
		return "", fmt.Errorf("no data provided for update")
	}

	var sb strings.Builder

	sb.WriteString("UPDATE ")
	sb.WriteString(b.dialect.Quote(b.query.Table))
	sb.WriteString(" SET ")

	// Build SET clause
	// 构建 SET 子句
	setParts := make([]string, 0, len(b.query.Data))
	for col, val := range b.query.Data {
		// Handle special operators like $incr, $decr
		// 处理特殊运算符如 $incr、$decr
		if m, ok := val.(map[string]any); ok {
			if incr, ok := m["$incr"]; ok {
				setParts = append(setParts, fmt.Sprintf("%s = %s + %s",
					b.dialect.Quote(col),
					b.dialect.Quote(col),
					b.addParam(incr)))
				continue
			}
			if decr, ok := m["$decr"]; ok {
				setParts = append(setParts, fmt.Sprintf("%s = %s - %s",
					b.dialect.Quote(col),
					b.dialect.Quote(col),
					b.addParam(decr)))
				continue
			}
		}
		setParts = append(setParts, fmt.Sprintf("%s = %s",
			b.dialect.Quote(col),
			b.addParam(val)))
	}

	sb.WriteString(strings.Join(setParts, ", "))

	// WHERE clause
	// WHERE 子句
	if len(b.query.Where) > 0 {
		whereSQL, err := b.buildWhere()
		if err != nil {
			return "", err
		}
		sb.WriteString(" WHERE ")
		sb.WriteString(whereSQL)
	}

	return sb.String(), nil
}

// buildDelete builds a DELETE statement.
// buildDelete 构建 DELETE 语句。
func (b *SQLBuilder) buildDelete() (string, error) {
	var sb strings.Builder

	sb.WriteString("DELETE FROM ")
	sb.WriteString(b.dialect.Quote(b.query.Table))

	// WHERE clause
	// WHERE 子句
	if len(b.query.Where) > 0 {
		whereSQL, err := b.buildWhere()
		if err != nil {
			return "", err
		}
		sb.WriteString(" WHERE ")
		sb.WriteString(whereSQL)
	}

	return sb.String(), nil
}

// buildCount builds a COUNT statement.
// buildCount 构建 COUNT 语句。
func (b *SQLBuilder) buildCount() (string, error) {
	var sb strings.Builder

	sb.WriteString("SELECT COUNT(*) FROM ")
	sb.WriteString(b.dialect.Quote(b.query.Table))

	// WHERE clause
	// WHERE 子句
	if len(b.query.Where) > 0 {
		whereSQL, err := b.buildWhere()
		if err != nil {
			return "", err
		}
		sb.WriteString(" WHERE ")
		sb.WriteString(whereSQL)
	}

	return sb.String(), nil
}

// buildAggregate builds an aggregate query.
// buildAggregate 构建聚合查询。
func (b *SQLBuilder) buildAggregate() (string, error) {
	var sb strings.Builder

	sb.WriteString("SELECT ")
	sb.WriteString(b.buildSelectColumns())
	sb.WriteString(" FROM ")
	sb.WriteString(b.dialect.Quote(b.query.Table))

	// JOIN clause
	// JOIN 子句
	if len(b.query.Join) > 0 {
		joinSQL, err := b.buildJoins()
		if err != nil {
			return "", err
		}
		sb.WriteString(joinSQL)
	}

	// WHERE clause
	// WHERE 子句
	if len(b.query.Where) > 0 {
		whereSQL, err := b.buildWhere()
		if err != nil {
			return "", err
		}
		sb.WriteString(" WHERE ")
		sb.WriteString(whereSQL)
	}

	// GROUP BY clause
	// GROUP BY 子句
	if len(b.query.GroupBy) > 0 {
		sb.WriteString(" GROUP BY ")
		sb.WriteString(b.buildGroupBy())
	}

	// HAVING clause
	// HAVING 子句
	if len(b.query.Having) > 0 {
		havingSQL, err := b.buildHaving()
		if err != nil {
			return "", err
		}
		sb.WriteString(" HAVING ")
		sb.WriteString(havingSQL)
	}

	// ORDER BY clause
	// ORDER BY 子句
	if len(b.query.OrderBy) > 0 {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(b.buildOrderBy())
	}

	return sb.String(), nil
}

// buildSelectColumns builds the SELECT column list.
// buildSelectColumns 构建 SELECT 列列表。
func (b *SQLBuilder) buildSelectColumns() string {
	parts := make([]string, 0, len(b.query.Select))

	for _, sel := range b.query.Select {
		switch v := sel.(type) {
		case string:
			// Simple column name
			// 简单列名
			if v == "*" || strings.Contains(v, ".") {
				parts = append(parts, v)
			} else {
				parts = append(parts, b.dialect.Quote(v))
			}
		case map[string]any:
			// Aggregate function
			// 聚合函数
			fn, _ := v["fn"].(string)
			field, _ := v["field"].(string)
			as, _ := v["as"].(string)

			var expr string
			if field == "" || field == "*" {
				expr = fmt.Sprintf("%s(*)", strings.ToUpper(fn))
			} else {
				expr = fmt.Sprintf("%s(%s)", strings.ToUpper(fn), b.dialect.Quote(field))
			}

			if as != "" {
				expr = fmt.Sprintf("%s AS %s", expr, b.dialect.Quote(as))
			}
			parts = append(parts, expr)
		}
	}

	if len(parts) == 0 {
		return "*"
	}
	return strings.Join(parts, ", ")
}

// buildWhere builds the WHERE clause from conditions.
// buildWhere 从条件构建 WHERE 子句。
func (b *SQLBuilder) buildWhere() (string, error) {
	return b.buildConditions(b.query.Where)
}

// buildConditions builds SQL from a slice of conditions.
// buildConditions 从条件切片构建 SQL。
func (b *SQLBuilder) buildConditions(conditions []Condition) (string, error) {
	if len(conditions) == 0 {
		return "", nil
	}

	parts := make([]string, 0, len(conditions))

	for i, cond := range conditions {
		condSQL, err := b.buildCondition(cond)
		if err != nil {
			return "", err
		}

		if i == 0 {
			parts = append(parts, condSQL)
		} else if cond.Or {
			parts = append(parts, "OR "+condSQL)
		} else {
			parts = append(parts, "AND "+condSQL)
		}
	}

	return strings.Join(parts, " "), nil
}

// buildCondition builds a single condition.
// buildCondition 构建单个条件。
func (b *SQLBuilder) buildCondition(cond Condition) (string, error) {
	// Handle nested AND conditions
	// 处理嵌套的 AND 条件
	if len(cond.And) > 0 {
		andSQL, err := b.buildConditions(cond.And)
		if err != nil {
			return "", err
		}
		return "(" + andSQL + ")", nil
	}

	// Handle nested OR conditions
	// 处理嵌套的 OR 条件
	if len(cond.OrGroup) > 0 {
		orParts := make([]string, 0, len(cond.OrGroup))
		for _, orCond := range cond.OrGroup {
			orSQL, err := b.buildCondition(orCond)
			if err != nil {
				return "", err
			}
			orParts = append(orParts, orSQL)
		}
		return "(" + strings.Join(orParts, " OR ") + ")", nil
	}

	// Handle subquery
	// 处理子查询
	if cond.Subquery != nil {
		subBuilder := NewSQLBuilder(b.dialect, cond.Subquery)
		// Transfer current param count
		subBuilder.paramN = b.paramN
		subResult, err := subBuilder.buildSelect()
		if err != nil {
			return "", err
		}
		// Transfer params back
		b.params = append(b.params, subBuilder.params...)
		b.paramN = subBuilder.paramN

		return fmt.Sprintf("%s %s (%s)",
			b.dialect.Quote(cond.Field),
			b.opToSQL(cond.Op),
			subResult), nil
	}

	// Handle reference to another column
	// 处理对另一列的引用
	if cond.Ref != "" {
		return fmt.Sprintf("%s %s %s",
			b.dialect.Quote(cond.Field),
			b.opToSQL(cond.Op),
			cond.Ref), nil
	}

	// Handle different operators
	// 处理不同运算符
	field := b.dialect.Quote(cond.Field)
	if strings.Contains(cond.Field, ".") {
		// Table.Column format, don't quote
		field = cond.Field
	}

	switch cond.Op {
	case OpNull:
		return fmt.Sprintf("%s IS NULL", field), nil
	case OpNotNull:
		return fmt.Sprintf("%s IS NOT NULL", field), nil
	case OpIn, OpNotIn:
		if values, ok := cond.Value.([]any); ok {
			placeholders := make([]string, len(values))
			for i, v := range values {
				placeholders[i] = b.addParam(v)
			}
			op := "IN"
			if cond.Op == OpNotIn {
				op = "NOT IN"
			}
			return fmt.Sprintf("%s %s (%s)", field, op, strings.Join(placeholders, ", ")), nil
		}
		return "", fmt.Errorf("IN operator requires array value")
	case OpBetween:
		if values, ok := cond.Value.([]any); ok && len(values) == 2 {
			return fmt.Sprintf("%s BETWEEN %s AND %s",
				field,
				b.addParam(values[0]),
				b.addParam(values[1])), nil
		}
		return "", fmt.Errorf("BETWEEN operator requires array of two values")
	case OpLike, OpNotLike:
		op := "LIKE"
		if cond.Op == OpNotLike {
			op = "NOT LIKE"
		}
		return fmt.Sprintf("%s %s %s", field, op, b.addParam(cond.Value)), nil
	case OpExists:
		// EXISTS is handled with subquery
		return "", fmt.Errorf("EXISTS operator requires subquery")
	default:
		return fmt.Sprintf("%s %s %s", field, b.opToSQL(cond.Op), b.addParam(cond.Value)), nil
	}
}

// buildJoins builds JOIN clauses.
// buildJoins 构建 JOIN 子句。
func (b *SQLBuilder) buildJoins() (string, error) {
	var sb strings.Builder

	for _, join := range b.query.Join {
		joinType := strings.ToUpper(join.Type)
		if joinType == "" {
			joinType = "INNER"
		}

		sb.WriteString(fmt.Sprintf(" %s JOIN %s ON ", joinType, b.dialect.Quote(join.Table)))

		onParts := make([]string, 0, len(join.On))
		for left, right := range join.On {
			onParts = append(onParts, fmt.Sprintf("%s = %s", left, right))
		}
		sb.WriteString(strings.Join(onParts, " AND "))
	}

	return sb.String(), nil
}

// buildGroupBy builds GROUP BY clause.
// buildGroupBy 构建 GROUP BY 子句。
func (b *SQLBuilder) buildGroupBy() string {
	cols := make([]string, len(b.query.GroupBy))
	for i, col := range b.query.GroupBy {
		if strings.Contains(col, ".") {
			cols[i] = col
		} else {
			cols[i] = b.dialect.Quote(col)
		}
	}
	return strings.Join(cols, ", ")
}

// buildHaving builds HAVING clause.
// buildHaving 构建 HAVING 子句。
func (b *SQLBuilder) buildHaving() (string, error) {
	parts := make([]string, 0, len(b.query.Having))

	for _, h := range b.query.Having {
		var expr string
		if h.Field == "" || h.Field == "*" {
			expr = fmt.Sprintf("%s(*)", strings.ToUpper(h.Fn))
		} else {
			expr = fmt.Sprintf("%s(%s)", strings.ToUpper(h.Fn), b.dialect.Quote(h.Field))
		}

		parts = append(parts, fmt.Sprintf("%s %s %s",
			expr,
			b.opToSQL(h.Op),
			b.addParam(h.Value)))
	}

	return strings.Join(parts, " AND "), nil
}

// buildOrderBy builds ORDER BY clause.
// buildOrderBy 构建 ORDER BY 子句。
func (b *SQLBuilder) buildOrderBy() string {
	parts := make([]string, len(b.query.OrderBy))
	for i, o := range b.query.OrderBy {
		field := b.dialect.Quote(o.Field)
		if strings.Contains(o.Field, ".") {
			field = o.Field
		}
		if o.Desc {
			parts[i] = field + " DESC"
		} else {
			parts[i] = field + " ASC"
		}
	}
	return strings.Join(parts, ", ")
}

// addParam adds a parameter and returns the placeholder.
// addParam 添加参数并返回占位符。
func (b *SQLBuilder) addParam(value any) string {
	b.paramN++
	b.params = append(b.params, value)
	return b.dialect.Placeholder(b.paramN)
}

// opToSQL converts an Operator to SQL string.
// opToSQL 将 Operator 转换为 SQL 字符串。
func (b *SQLBuilder) opToSQL(op Operator) string {
	switch op {
	case OpEqual:
		return "="
	case OpNotEqual:
		return "!="
	case OpGreater:
		return ">"
	case OpGreaterOrEq:
		return ">="
	case OpLess:
		return "<"
	case OpLessOrEq:
		return "<="
	case OpIn:
		return "IN"
	case OpNotIn:
		return "NOT IN"
	case OpLike:
		return "LIKE"
	case OpNotLike:
		return "NOT LIKE"
	default:
		return string(op)
	}
}
