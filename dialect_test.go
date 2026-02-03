package goorm

import (
	"testing"
)

// TestSQLiteDialect tests the SQLite dialect.
// TestSQLiteDialect 测试 SQLite 方言。
func TestSQLiteDialect(t *testing.T) {
	d := &SQLiteDialect{}

	if d.Name() != "sqlite" {
		t.Errorf("expected 'sqlite', got %s", d.Name())
	}

	if d.DriverName() != "sqlite3" {
		t.Errorf("expected 'sqlite3', got %s", d.DriverName())
	}

	if d.Quote("users") != `"users"` {
		t.Errorf("expected '\"users\"', got %s", d.Quote("users"))
	}

	if d.Placeholder(1) != "?" {
		t.Errorf("expected '?', got %s", d.Placeholder(1))
	}
}

// TestSQLiteTypeMapping tests SQLite type mapping.
// TestSQLiteTypeMapping 测试 SQLite 类型映射。
func TestSQLiteTypeMapping(t *testing.T) {
	d := &SQLiteDialect{}

	tests := []struct {
		goType   string
		tags     map[string]string
		expected string
	}{
		{"int", nil, "INTEGER"},
		{"int64", nil, "INTEGER"},
		{"float64", nil, "REAL"},
		{"bool", nil, "INTEGER"},
		{"string", nil, "TEXT"},
		{"[]byte", nil, "BLOB"},
		{"time.Time", nil, "DATETIME"},
	}

	for _, tt := range tests {
		result := d.GoTypeToSQL(tt.goType, tt.tags)
		if result != tt.expected {
			t.Errorf("GoTypeToSQL(%s) = %s, want %s", tt.goType, result, tt.expected)
		}
	}
}

// TestDialectRegistration tests dialect registration.
// TestDialectRegistration 测试方言注册。
func TestDialectRegistration(t *testing.T) {
	dialects := []string{"postgres", "postgresql", "mysql", "sqlite", "sqlite3"}

	for _, name := range dialects {
		d, err := GetDialect(name)
		if err != nil {
			t.Errorf("dialect %s should be registered: %v", name, err)
		}
		if d == nil {
			t.Errorf("dialect %s should not be nil", name)
		}
	}

	_, err := GetDialect("unknown")
	if err == nil {
		t.Error("unknown dialect should return error")
	}
}

// TestQueryOptimizer tests the query optimizer.
// TestQueryOptimizer 测试查询优化器。
func TestQueryOptimizer(t *testing.T) {
	// Create a mock DB with registry
	// 创建一个带注册表的模拟 DB
	db := &DB{
		registry: NewRegistry(),
	}

	// Register a test model
	// 注册测试模型
	type TestUser struct {
		Model
		Name string
	}
	db.registry.Register(&TestUser{}, NamingConfig{})

	optimizer := NewQueryOptimizer(db)

	// Test SELECT * detection
	// 测试 SELECT * 检测
	query := &Query{
		Table:  "test_users",
		Action: ActionFind,
	}

	result := optimizer.Analyze(query)

	// Should have hints for SELECT * and missing LIMIT
	// 应该有 SELECT * 和缺少 LIMIT 的提示
	if len(result.Hints) < 2 {
		t.Errorf("expected at least 2 hints, got %d", len(result.Hints))
	}

	// Score should be reduced
	// 分数应该降低
	if result.Score >= 100 {
		t.Error("score should be reduced for non-optimal query")
	}
}

// TestQueryOptimizerWithLimit tests optimization with LIMIT.
// TestQueryOptimizerWithLimit 测试带 LIMIT 的优化。
func TestQueryOptimizerWithLimit(t *testing.T) {
	db := &DB{registry: NewRegistry()}
	optimizer := NewQueryOptimizer(db)

	query := &Query{
		Table:  "users",
		Action: ActionFind,
		Limit:  10,
		Select: []any{"id", "name"},
	}

	result := optimizer.Analyze(query)

	// Should have fewer hints
	// 应该有更少的提示
	hasLimitHint := false
	hasSelectAllHint := false
	for _, hint := range result.Hints {
		if hint.Type == "missing_limit" {
			hasLimitHint = true
		}
		if hint.Type == "select_all" {
			hasSelectAllHint = true
		}
	}

	if hasLimitHint {
		t.Error("should not have missing_limit hint")
	}
	if hasSelectAllHint {
		t.Error("should not have select_all hint")
	}
}

// TestQueryOptimizerOptimize tests the Optimize method.
// TestQueryOptimizerOptimize 测试 Optimize 方法。
func TestQueryOptimizerOptimize(t *testing.T) {
	db := &DB{registry: NewRegistry()}
	optimizer := NewQueryOptimizer(db)

	query := &Query{
		Table:  "users",
		Action: ActionFind,
	}

	optimized := optimizer.Optimize(query)

	// Should add a default limit
	// 应该添加默认限制
	if optimized.Limit != 1000 {
		t.Errorf("expected limit 1000, got %d", optimized.Limit)
	}
}

// TestHealthThresholds tests default health thresholds.
// TestHealthThresholds 测试默认健康阈值。
func TestHealthThresholds(t *testing.T) {
	thresholds := DefaultHealthThresholds()

	if thresholds.MaxLatency == 0 {
		t.Error("MaxLatency should be set")
	}

	if thresholds.MaxOpenConnectionsPercent == 0 {
		t.Error("MaxOpenConnectionsPercent should be set")
	}

	if thresholds.MaxWaitCount == 0 {
		t.Error("MaxWaitCount should be set")
	}
}
