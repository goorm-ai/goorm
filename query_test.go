package goorm

import (
	"encoding/json"
	"testing"
)

// TestParseQuery tests JQL parsing.
// TestParseQuery 测试 JQL 解析。
func TestParseQuery(t *testing.T) {
	tests := []struct {
		name    string
		jql     string
		wantErr bool
	}{
		{
			name: "simple find query",
			jql: `{
				"table": "users",
				"action": "find"
			}`,
			wantErr: false,
		},
		{
			name: "find with where",
			jql: `{
				"table": "users",
				"action": "find",
				"where": [{"field": "age", "op": ">", "value": 18}]
			}`,
			wantErr: false,
		},
		{
			name: "create query",
			jql: `{
				"table": "users",
				"action": "create",
				"data": {"name": "张三", "email": "test@example.com"}
			}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			jql:     `{"table": "users"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := ParseQuery(tt.jql)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && query == nil {
				t.Error("ParseQuery() returned nil query without error")
			}
		})
	}
}

// TestQueryValidate tests Query validation.
// TestQueryValidate 测试 Query 验证。
func TestQueryValidate(t *testing.T) {
	tests := []struct {
		name    string
		query   Query
		wantErr bool
	}{
		{
			name: "valid find query",
			query: Query{
				Table:  "users",
				Action: ActionFind,
			},
			wantErr: false,
		},
		{
			name: "missing action",
			query: Query{
				Table: "users",
			},
			wantErr: true,
		},
		{
			name: "find without table",
			query: Query{
				Action: ActionFind,
			},
			wantErr: true,
		},
		{
			name: "create without data",
			query: Query{
				Table:  "users",
				Action: ActionCreate,
			},
			wantErr: true,
		},
		{
			name: "valid create",
			query: Query{
				Table:  "users",
				Action: ActionCreate,
				Data:   map[string]any{"name": "test"},
			},
			wantErr: false,
		},
		{
			name: "valid list_tables",
			query: Query{
				Action: ActionListTables,
			},
			wantErr: false,
		},
		{
			name: "transaction without operations",
			query: Query{
				Action: ActionTransaction,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Query.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestQueryString tests Query JSON serialization.
// TestQueryString 测试 Query JSON 序列化。
func TestQueryString(t *testing.T) {
	query := &Query{
		Table:  "users",
		Action: ActionFind,
		Where: []Condition{
			{Field: "age", Op: OpGreater, Value: 18},
		},
		Limit: 10,
	}

	jsonStr := query.String()

	// Verify it's valid JSON
	// 验证是有效的 JSON
	var parsed Query
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Errorf("Query.String() returned invalid JSON: %v", err)
	}

	if parsed.Table != query.Table {
		t.Errorf("Query.String() round-trip: Table = %v, want %v", parsed.Table, query.Table)
	}
}
