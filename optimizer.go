package goorm

import (
	"fmt"
	"strings"
)

// QueryOptimizer analyzes and optimizes JQL queries.
// QueryOptimizer 分析和优化 JQL 查询。
type QueryOptimizer struct {
	db      *DB
	enabled bool
	hints   []OptimizationHint
}

// OptimizationHint represents a query optimization suggestion.
// OptimizationHint 表示查询优化建议。
type OptimizationHint struct {
	// Type is the type of optimization.
	// Type 是优化类型。
	Type string `json:"type"`

	// Severity is the importance level (info, warning, error).
	// Severity 是重要性级别（info、warning、error）。
	Severity string `json:"severity"`

	// Message describes the optimization suggestion.
	// Message 描述优化建议。
	Message string `json:"message"`

	// Suggestion provides actionable advice.
	// Suggestion 提供可操作的建议。
	Suggestion string `json:"suggestion"`
}

// OptimizationResult contains optimization analysis results.
// OptimizationResult 包含优化分析结果。
type OptimizationResult struct {
	// Query is the original query.
	// Query 是原始查询。
	Query *Query `json:"query"`

	// Hints are optimization suggestions.
	// Hints 是优化建议。
	Hints []OptimizationHint `json:"hints"`

	// OptimizedQuery is the optimized query if changes were made.
	// OptimizedQuery 是优化后的查询（如果有更改）。
	OptimizedQuery *Query `json:"optimized_query,omitempty"`

	// Score is the optimization score (0-100).
	// Score 是优化评分（0-100）。
	Score int `json:"score"`
}

// NewQueryOptimizer creates a new query optimizer.
// NewQueryOptimizer 创建新的查询优化器。
func NewQueryOptimizer(db *DB) *QueryOptimizer {
	return &QueryOptimizer{
		db:      db,
		enabled: true,
		hints:   make([]OptimizationHint, 0),
	}
}

// Analyze analyzes a query and returns optimization suggestions.
// Analyze 分析查询并返回优化建议。
func (o *QueryOptimizer) Analyze(query *Query) *OptimizationResult {
	result := &OptimizationResult{
		Query: query,
		Hints: make([]OptimizationHint, 0),
		Score: 100,
	}

	// Check for missing indexes
	// 检查缺少的索引
	o.checkMissingIndexes(query, result)

	// Check for SELECT *
	// 检查 SELECT *
	o.checkSelectAll(query, result)

	// Check for missing LIMIT
	// 检查缺少 LIMIT
	o.checkMissingLimit(query, result)

	// Check for N+1 potential
	// 检查 N+1 问题潜在风险
	o.checkN1Potential(query, result)

	// Check for inefficient operators
	// 检查低效操作符
	o.checkInefficiententOperators(query, result)

	return result
}

// checkMissingIndexes checks for queries that may benefit from indexes.
// checkMissingIndexes 检查可能受益于索引的查询。
func (o *QueryOptimizer) checkMissingIndexes(query *Query, result *OptimizationResult) {
	if query.Table == "" || len(query.Where) == 0 {
		return
	}

	meta, ok := o.db.registry.Get(query.Table)
	if !ok {
		return
	}

	// Build index map
	// 构建索引映射
	indexedFields := make(map[string]bool)
	for _, field := range meta.Fields {
		if field.PrimaryKey || field.Unique {
			indexedFields[field.ColumnName] = true
		}
	}

	// Check each WHERE condition
	// 检查每个 WHERE 条件
	for _, cond := range query.Where {
		field := cond.Field
		if strings.Contains(field, ".") {
			parts := strings.Split(field, ".")
			field = parts[len(parts)-1]
		}

		if !indexedFields[field] {
			result.Hints = append(result.Hints, OptimizationHint{
				Type:       "missing_index",
				Severity:   "warning",
				Message:    fmt.Sprintf("Column '%s' in WHERE clause may not be indexed", field),
				Suggestion: fmt.Sprintf("Consider adding an index on '%s.%s'", query.Table, field),
			})
			result.Score -= 10
		}
	}
}

// checkSelectAll checks for SELECT * usage.
// checkSelectAll 检查 SELECT * 的使用。
func (o *QueryOptimizer) checkSelectAll(query *Query, result *OptimizationResult) {
	if query.Action != ActionFind {
		return
	}

	if len(query.Select) == 0 {
		result.Hints = append(result.Hints, OptimizationHint{
			Type:       "select_all",
			Severity:   "info",
			Message:    "Query selects all columns (SELECT *)",
			Suggestion: "Consider selecting only required columns to reduce data transfer",
		})
		result.Score -= 5
	}
}

// checkMissingLimit checks for queries without LIMIT.
// checkMissingLimit 检查没有 LIMIT 的查询。
func (o *QueryOptimizer) checkMissingLimit(query *Query, result *OptimizationResult) {
	if query.Action != ActionFind {
		return
	}

	if query.Limit == 0 {
		result.Hints = append(result.Hints, OptimizationHint{
			Type:       "missing_limit",
			Severity:   "warning",
			Message:    "Query has no LIMIT clause",
			Suggestion: "Add a LIMIT clause to prevent returning excessive rows",
		})
		result.Score -= 15
	}
}

// checkN1Potential checks for potential N+1 query issues.
// checkN1Potential 检查潜在的 N+1 查询问题。
func (o *QueryOptimizer) checkN1Potential(query *Query, result *OptimizationResult) {
	if query.Action != ActionFind {
		return
	}

	// If there are relations but no With clause
	// 如果有关联但没有 With 子句
	if query.Table != "" && len(query.With) == 0 {
		meta, ok := o.db.registry.Get(query.Table)
		if !ok {
			return
		}

		if len(meta.Relations) > 0 {
			result.Hints = append(result.Hints, OptimizationHint{
				Type:       "n1_potential",
				Severity:   "info",
				Message:    "Table has relations but query does not use eager loading",
				Suggestion: "Consider using 'with' clause for eager loading if you need related data",
			})
		}
	}
}

// checkInefficiententOperators checks for inefficient operators.
// checkInefficiententOperators 检查低效操作符。
func (o *QueryOptimizer) checkInefficiententOperators(query *Query, result *OptimizationResult) {
	for _, cond := range query.Where {
		// Check for LIKE with leading wildcard
		// 检查带有前导通配符的 LIKE
		if cond.Op == OpLike || cond.Op == OpILike {
			if strVal, ok := cond.Value.(string); ok {
				if strings.HasPrefix(strVal, "%") {
					result.Hints = append(result.Hints, OptimizationHint{
						Type:       "leading_wildcard",
						Severity:   "warning",
						Message:    fmt.Sprintf("LIKE pattern on '%s' starts with wildcard", cond.Field),
						Suggestion: "Leading wildcards prevent index usage. Consider full-text search for better performance",
					})
					result.Score -= 15
				}
			}
		}

		// Check for NOT IN with large lists
		// 检查带有大列表的 NOT IN
		if cond.Op == OpNotIn {
			if arr, ok := cond.Value.([]any); ok && len(arr) > 10 {
				result.Hints = append(result.Hints, OptimizationHint{
					Type:       "large_not_in",
					Severity:   "warning",
					Message:    fmt.Sprintf("NOT IN clause on '%s' has %d values", cond.Field, len(arr)),
					Suggestion: "Large NOT IN clauses can be slow. Consider using a subquery or temporary table",
				})
				result.Score -= 10
			}
		}
	}
}

// Optimize applies automatic optimizations to a query.
// Optimize 对查询应用自动优化。
func (o *QueryOptimizer) Optimize(query *Query) *Query {
	// Create a copy
	// 创建副本
	optimized := *query

	// Apply automatic optimizations
	// 应用自动优化

	// Add default limit if missing for find queries
	// 如果查找查询缺少限制，添加默认限制
	if optimized.Action == ActionFind && optimized.Limit == 0 {
		optimized.Limit = 1000 // Default reasonable limit
	}

	return &optimized
}

// --- DB Query Optimization Methods ---
// --- DB 查询优化方法 ---

// AnalyzeQuery analyzes a query and returns optimization suggestions.
// AnalyzeQuery 分析查询并返回优化建议。
func (db *DB) AnalyzeQuery(query *Query) *OptimizationResult {
	optimizer := NewQueryOptimizer(db)
	return optimizer.Analyze(query)
}

// OptimizeQuery applies automatic optimizations to a query.
// OptimizeQuery 对查询应用自动优化。
func (db *DB) OptimizeQuery(query *Query) *Query {
	optimizer := NewQueryOptimizer(db)
	return optimizer.Optimize(query)
}
