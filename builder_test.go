package goorm

import "testing"

// TestSQLBuilderSelect tests SELECT statement building.
// TestSQLBuilderSelect 测试 SELECT 语句构建。
func TestSQLBuilderSelect(t *testing.T) {
	dialect := &PostgresDialect{}

	tests := []struct {
		name       string
		query      *Query
		wantSQL    string
		wantParams []any
	}{
		{
			name: "simple select all",
			query: &Query{
				Table:  "users",
				Action: ActionFind,
			},
			wantSQL:    `SELECT * FROM "users"`,
			wantParams: []any{},
		},
		{
			name: "select with columns",
			query: &Query{
				Table:  "users",
				Action: ActionFind,
				Select: []any{"name", "email"},
			},
			wantSQL:    `SELECT "name", "email" FROM "users"`,
			wantParams: []any{},
		},
		{
			name: "select with where",
			query: &Query{
				Table:  "users",
				Action: ActionFind,
				Where: []Condition{
					{Field: "age", Op: OpGreater, Value: 18},
				},
			},
			wantSQL:    `SELECT * FROM "users" WHERE "age" > $1`,
			wantParams: []any{18},
		},
		{
			name: "select with multiple conditions",
			query: &Query{
				Table:  "users",
				Action: ActionFind,
				Where: []Condition{
					{Field: "age", Op: OpGreater, Value: 18},
					{Field: "status", Op: OpEqual, Value: "active"},
				},
			},
			wantSQL:    `SELECT * FROM "users" WHERE "age" > $1 AND "status" = $2`,
			wantParams: []any{18, "active"},
		},
		{
			name: "select with OR condition",
			query: &Query{
				Table:  "users",
				Action: ActionFind,
				Where: []Condition{
					{Field: "role", Op: OpEqual, Value: "admin"},
					{Field: "role", Op: OpEqual, Value: "superadmin", Or: true},
				},
			},
			wantSQL:    `SELECT * FROM "users" WHERE "role" = $1 OR "role" = $2`,
			wantParams: []any{"admin", "superadmin"},
		},
		{
			name: "select with order and limit",
			query: &Query{
				Table:  "users",
				Action: ActionFind,
				OrderBy: []Order{
					{Field: "created_at", Desc: true},
				},
				Limit:  10,
				Offset: 20,
			},
			wantSQL:    `SELECT * FROM "users" ORDER BY "created_at" DESC LIMIT 10 OFFSET 20`,
			wantParams: []any{},
		},
		{
			name: "select with IN",
			query: &Query{
				Table:  "users",
				Action: ActionFind,
				Where: []Condition{
					{Field: "id", Op: OpIn, Value: []any{1, 2, 3}},
				},
			},
			wantSQL:    `SELECT * FROM "users" WHERE "id" IN ($1, $2, $3)`,
			wantParams: []any{1, 2, 3},
		},
		{
			name: "select with BETWEEN",
			query: &Query{
				Table:  "users",
				Action: ActionFind,
				Where: []Condition{
					{Field: "age", Op: OpBetween, Value: []any{18, 30}},
				},
			},
			wantSQL:    `SELECT * FROM "users" WHERE "age" BETWEEN $1 AND $2`,
			wantParams: []any{18, 30},
		},
		{
			name: "select with NULL check",
			query: &Query{
				Table:  "users",
				Action: ActionFind,
				Where: []Condition{
					{Field: "deleted_at", Op: OpNull},
				},
			},
			wantSQL:    `SELECT * FROM "users" WHERE "deleted_at" IS NULL`,
			wantParams: []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSQLBuilder(dialect, tt.query)
			result, err := builder.Build()

			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			if result.SQL != tt.wantSQL {
				t.Errorf("Build() SQL = %q, want %q", result.SQL, tt.wantSQL)
			}

			if len(result.Params) != len(tt.wantParams) {
				t.Errorf("Build() Params count = %d, want %d", len(result.Params), len(tt.wantParams))
			}
		})
	}
}

// TestSQLBuilderInsert tests INSERT statement building.
// TestSQLBuilderInsert 测试 INSERT 语句构建。
func TestSQLBuilderInsert(t *testing.T) {
	dialect := &PostgresDialect{}

	query := &Query{
		Table:  "users",
		Action: ActionCreate,
		Data: map[string]any{
			"name":  "张三",
			"email": "test@example.com",
		},
	}

	builder := NewSQLBuilder(dialect, query)
	result, err := builder.Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Check SQL structure (column order may vary)
	if !containsAll(result.SQL, []string{"INSERT INTO", `"users"`, "VALUES", "RETURNING id"}) {
		t.Errorf("Build() SQL = %q, missing expected parts", result.SQL)
	}

	if len(result.Params) != 2 {
		t.Errorf("Build() Params count = %d, want 2", len(result.Params))
	}
}

// TestSQLBuilderUpdate tests UPDATE statement building.
// TestSQLBuilderUpdate 测试 UPDATE 语句构建。
func TestSQLBuilderUpdate(t *testing.T) {
	dialect := &PostgresDialect{}

	query := &Query{
		Table:  "users",
		Action: ActionUpdate,
		Where: []Condition{
			{Field: "id", Op: OpEqual, Value: 1},
		},
		Data: map[string]any{
			"name": "新名字",
		},
	}

	builder := NewSQLBuilder(dialect, query)
	result, err := builder.Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !containsAll(result.SQL, []string{"UPDATE", `"users"`, "SET", "WHERE"}) {
		t.Errorf("Build() SQL = %q, missing expected parts", result.SQL)
	}
}

// TestSQLBuilderUpdateIncr tests UPDATE with $incr operator.
// TestSQLBuilderUpdateIncr 测试带 $incr 运算符的 UPDATE。
func TestSQLBuilderUpdateIncr(t *testing.T) {
	dialect := &PostgresDialect{}

	query := &Query{
		Table:  "users",
		Action: ActionUpdate,
		Where: []Condition{
			{Field: "id", Op: OpEqual, Value: 1},
		},
		Data: map[string]any{
			"login_count": map[string]any{"$incr": 1},
		},
	}

	builder := NewSQLBuilder(dialect, query)
	result, err := builder.Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	expected := `UPDATE "users" SET "login_count" = "login_count" + $1 WHERE "id" = $2`
	if result.SQL != expected {
		t.Errorf("Build() SQL = %q, want %q", result.SQL, expected)
	}
}

// TestSQLBuilderDelete tests DELETE statement building.
// TestSQLBuilderDelete 测试 DELETE 语句构建。
func TestSQLBuilderDelete(t *testing.T) {
	dialect := &PostgresDialect{}

	query := &Query{
		Table:  "users",
		Action: ActionDelete,
		Where: []Condition{
			{Field: "id", Op: OpEqual, Value: 1},
		},
	}

	builder := NewSQLBuilder(dialect, query)
	result, err := builder.Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	expected := `DELETE FROM "users" WHERE "id" = $1`
	if result.SQL != expected {
		t.Errorf("Build() SQL = %q, want %q", result.SQL, expected)
	}
}

// TestSQLBuilderCount tests COUNT statement building.
// TestSQLBuilderCount 测试 COUNT 语句构建。
func TestSQLBuilderCount(t *testing.T) {
	dialect := &PostgresDialect{}

	query := &Query{
		Table:  "users",
		Action: ActionCount,
		Where: []Condition{
			{Field: "status", Op: OpEqual, Value: "active"},
		},
	}

	builder := NewSQLBuilder(dialect, query)
	result, err := builder.Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	expected := `SELECT COUNT(*) FROM "users" WHERE "status" = $1`
	if result.SQL != expected {
		t.Errorf("Build() SQL = %q, want %q", result.SQL, expected)
	}
}

// TestSQLBuilderAggregate tests aggregate statement building.
// TestSQLBuilderAggregate 测试聚合语句构建。
func TestSQLBuilderAggregate(t *testing.T) {
	dialect := &PostgresDialect{}

	query := &Query{
		Table:  "orders",
		Action: ActionAggregate,
		Select: []any{
			"status",
			map[string]any{"fn": "count", "as": "count"},
			map[string]any{"fn": "sum", "field": "amount", "as": "total"},
		},
		GroupBy: []string{"status"},
		Having: []HavingCondition{
			{Fn: "count", Op: OpGreater, Value: 5},
		},
	}

	builder := NewSQLBuilder(dialect, query)
	result, err := builder.Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !containsAll(result.SQL, []string{
		"SELECT", `"status"`, "COUNT(*)", "SUM(", "GROUP BY", "HAVING",
	}) {
		t.Errorf("Build() SQL = %q, missing expected parts", result.SQL)
	}
}

// TestSQLBuilderMySQL tests MySQL dialect.
// TestSQLBuilderMySQL 测试 MySQL 方言。
func TestSQLBuilderMySQL(t *testing.T) {
	dialect := &MySQLDialect{}

	query := &Query{
		Table:  "users",
		Action: ActionFind,
		Where: []Condition{
			{Field: "age", Op: OpGreater, Value: 18},
		},
	}

	builder := NewSQLBuilder(dialect, query)
	result, err := builder.Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// MySQL uses backticks and ?
	expected := "SELECT * FROM `users` WHERE `age` > ?"
	if result.SQL != expected {
		t.Errorf("Build() SQL = %q, want %q", result.SQL, expected)
	}
}

// containsAll checks if a string contains all substrings.
// containsAll 检查字符串是否包含所有子字符串。
func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

// contains checks if a string contains a substring.
// contains 检查字符串是否包含子字符串。
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsHelper(s, sub))
}

func containsHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
