package goorm

import (
	"regexp"
	"strconv"
	"strings"
)

// NLParser parses natural language queries into JQL.
// NLParser 将自然语言查询解析为 JQL。
type NLParser struct {
	db       *DB
	patterns []nlPattern
}

// nlPattern represents a natural language pattern.
// nlPattern 表示自然语言模式。
type nlPattern struct {
	regex   *regexp.Regexp
	action  Action
	handler func(matches []string, tables []string) *Query
}

// NewNLParser creates a new natural language parser.
// NewNLParser 创建新的自然语言解析器。
func NewNLParser(db *DB) *NLParser {
	parser := &NLParser{db: db}
	parser.initPatterns()
	return parser
}

// Parse converts a natural language query to JQL Query.
// Parse 将自然语言查询转换为 JQL Query。
func (p *NLParser) Parse(query string) (*Query, error) {
	query = strings.TrimSpace(query)
	queryLower := strings.ToLower(query)

	// Get available tables for matching
	// 获取可用的表用于匹配
	tables := p.db.registry.ListTables()
	tableNames := make([]string, len(tables))
	for i, t := range tables {
		tableNames[i] = t.Name
	}

	// Try to match patterns
	// 尝试匹配模式
	for _, pattern := range p.patterns {
		if matches := pattern.regex.FindStringSubmatch(queryLower); matches != nil {
			return pattern.handler(matches, tableNames), nil
		}
	}

	// Fallback: try to detect intent from keywords
	// 回退：尝试从关键词检测意图
	return p.parseByKeywords(query, queryLower, tableNames)
}

// initPatterns initializes the NL patterns.
// initPatterns 初始化自然语言模式。
func (p *NLParser) initPatterns() {
	p.patterns = []nlPattern{
		// English patterns
		// Find/Get/Query patterns
		{
			regex:  regexp.MustCompile(`(?i)^(?:find|get|query|show|list|select)\s+(?:all\s+)?(\w+)(?:\s+where\s+(\w+)\s*(=|>|<|>=|<=|!=|is|equals?)\s*(.+))?$`),
			action: ActionFind,
			handler: func(matches []string, tables []string) *Query {
				table := p.matchTable(matches[1], tables)
				q := &Query{Table: table, Action: ActionFind}
				if len(matches) > 2 && matches[2] != "" {
					q.Where = []Condition{{
						Field: matches[2],
						Op:    p.parseOperator(matches[3]),
						Value: p.parseValue(matches[4]),
					}}
				}
				return q
			},
		},
		// Count patterns
		{
			regex:  regexp.MustCompile(`(?i)^(?:count|how many)\s+(\w+)(?:\s+where\s+(.+))?$`),
			action: ActionCount,
			handler: func(matches []string, tables []string) *Query {
				table := p.matchTable(matches[1], tables)
				q := &Query{Table: table, Action: ActionCount}
				if len(matches) > 2 && matches[2] != "" {
					q.Where = p.parseWhereClause(matches[2])
				}
				return q
			},
		},
		// Find with age/number conditions
		{
			regex:  regexp.MustCompile(`(?i)^(?:find|get|query)\s+(?:all\s+)?(\w+)\s+(?:older|greater|more)\s+than\s+(\d+)$`),
			action: ActionFind,
			handler: func(matches []string, tables []string) *Query {
				table := p.matchTable(matches[1], tables)
				value, _ := strconv.Atoi(matches[2])
				return &Query{
					Table:  table,
					Action: ActionFind,
					Where: []Condition{{
						Field: "age",
						Op:    OpGreater,
						Value: value,
					}},
				}
			},
		},
		// Find with younger/less conditions
		{
			regex:  regexp.MustCompile(`(?i)^(?:find|get|query)\s+(?:all\s+)?(\w+)\s+(?:younger|less|fewer)\s+than\s+(\d+)$`),
			action: ActionFind,
			handler: func(matches []string, tables []string) *Query {
				table := p.matchTable(matches[1], tables)
				value, _ := strconv.Atoi(matches[2])
				return &Query{
					Table:  table,
					Action: ActionFind,
					Where: []Condition{{
						Field: "age",
						Op:    OpLess,
						Value: value,
					}},
				}
			},
		},
		// Delete patterns
		{
			regex:  regexp.MustCompile(`(?i)^(?:delete|remove)\s+(\w+)\s+where\s+(\w+)\s*(=|>|<|>=|<=|!=)\s*(.+)$`),
			action: ActionDelete,
			handler: func(matches []string, tables []string) *Query {
				table := p.matchTable(matches[1], tables)
				return &Query{
					Table:  table,
					Action: ActionDelete,
					Where: []Condition{{
						Field: matches[2],
						Op:    p.parseOperator(matches[3]),
						Value: p.parseValue(matches[4]),
					}},
				}
			},
		},
		// Chinese patterns / 中文模式
		// 查找/查询模式
		{
			regex:  regexp.MustCompile(`^(?:查找|查询|获取|显示|列出)(?:所有)?(\w+)$`),
			action: ActionFind,
			handler: func(matches []string, tables []string) *Query {
				table := p.matchTableChinese(matches[1], tables)
				return &Query{Table: table, Action: ActionFind}
			},
		},
		// 年龄大于
		{
			regex:  regexp.MustCompile(`^(?:查找|查询|获取)(?:所有)?(?:年龄)?(?:大于|超过|高于)(\d+)(?:岁)?的?(\w+)?$`),
			action: ActionFind,
			handler: func(matches []string, tables []string) *Query {
				value, _ := strconv.Atoi(matches[1])
				table := "users"
				if len(matches) > 2 && matches[2] != "" {
					table = p.matchTableChinese(matches[2], tables)
				}
				return &Query{
					Table:  table,
					Action: ActionFind,
					Where: []Condition{{
						Field: "age",
						Op:    OpGreater,
						Value: value,
					}},
				}
			},
		},
		// 年龄小于
		{
			regex:  regexp.MustCompile(`^(?:查找|查询|获取)(?:所有)?(?:年龄)?(?:小于|低于|不超过)(\d+)(?:岁)?的?(\w+)?$`),
			action: ActionFind,
			handler: func(matches []string, tables []string) *Query {
				value, _ := strconv.Atoi(matches[1])
				table := "users"
				if len(matches) > 2 && matches[2] != "" {
					table = p.matchTableChinese(matches[2], tables)
				}
				return &Query{
					Table:  table,
					Action: ActionFind,
					Where: []Condition{{
						Field: "age",
						Op:    OpLess,
						Value: value,
					}},
				}
			},
		},
		// 统计数量
		{
			regex:  regexp.MustCompile(`^(?:统计|计算|数一下|有多少)(\w+)(?:的数量)?$`),
			action: ActionCount,
			handler: func(matches []string, tables []string) *Query {
				table := p.matchTableChinese(matches[1], tables)
				return &Query{Table: table, Action: ActionCount}
			},
		},
		// 删除
		{
			regex:  regexp.MustCompile(`^(?:删除|移除)(\w+)\s*(?:其中|where)?\s*(\w+)\s*(=|>|<|等于|大于|小于)\s*(.+)$`),
			action: ActionDelete,
			handler: func(matches []string, tables []string) *Query {
				table := p.matchTableChinese(matches[1], tables)
				return &Query{
					Table:  table,
					Action: ActionDelete,
					Where: []Condition{{
						Field: matches[2],
						Op:    p.parseOperator(matches[3]),
						Value: p.parseValue(matches[4]),
					}},
				}
			},
		},
	}
}

// parseByKeywords parses query using keyword detection.
// parseByKeywords 使用关键词检测解析查询。
func (p *NLParser) parseByKeywords(query, queryLower string, tables []string) (*Query, error) {
	// Detect action
	// 检测操作类型
	action := ActionFind
	if containsAny(queryLower, []string{"count", "how many", "统计", "计算", "有多少"}) {
		action = ActionCount
	} else if containsAny(queryLower, []string{"delete", "remove", "删除", "移除"}) {
		action = ActionDelete
	} else if containsAny(queryLower, []string{"update", "change", "modify", "更新", "修改"}) {
		action = ActionUpdate
	} else if containsAny(queryLower, []string{"create", "add", "insert", "创建", "添加", "插入", "新增"}) {
		action = ActionCreate
	}

	// Find table name
	// 查找表名
	table := ""
	for _, t := range tables {
		if strings.Contains(queryLower, strings.ToLower(t)) {
			table = t
			break
		}
	}

	// Fallback to first table or common table names
	// 回退到第一个表或常见表名
	if table == "" {
		if containsAny(queryLower, []string{"user", "用户"}) {
			table = "users"
		} else if containsAny(queryLower, []string{"order", "订单"}) {
			table = "orders"
		} else if containsAny(queryLower, []string{"product", "商品", "产品"}) {
			table = "products"
		} else if len(tables) > 0 {
			table = tables[0]
		} else {
			table = "users" // Default table
		}
	}

	// Parse conditions
	// 解析条件
	var where []Condition

	// Look for age conditions
	// 查找年龄条件
	if ageMatch := regexp.MustCompile(`(?:older than|greater than|age\s*>\s*|大于|超过)\s*(\d+)`).FindStringSubmatch(queryLower); ageMatch != nil {
		value, _ := strconv.Atoi(ageMatch[1])
		where = append(where, Condition{Field: "age", Op: OpGreater, Value: value})
	} else if ageMatch := regexp.MustCompile(`(?:younger than|less than|age\s*<\s*|小于|低于)\s*(\d+)`).FindStringSubmatch(queryLower); ageMatch != nil {
		value, _ := strconv.Atoi(ageMatch[1])
		where = append(where, Condition{Field: "age", Op: OpLess, Value: value})
	}

	// Look for status conditions
	// 查找状态条件
	if containsAny(queryLower, []string{"active", "活跃", "激活"}) {
		where = append(where, Condition{Field: "status", Op: OpEqual, Value: "active"})
	} else if containsAny(queryLower, []string{"inactive", "禁用", "未激活"}) {
		where = append(where, Condition{Field: "status", Op: OpEqual, Value: "inactive"})
	}

	// Look for limit
	// 查找限制
	var limit int
	if limitMatch := regexp.MustCompile(`(?:top|first|limit|前|最多)\s*(\d+)`).FindStringSubmatch(queryLower); limitMatch != nil {
		limit, _ = strconv.Atoi(limitMatch[1])
	}

	// Look for order
	// 查找排序
	var orderBy []Order
	if containsAny(queryLower, []string{"newest", "latest", "recent", "最新", "最近"}) {
		orderBy = append(orderBy, Order{Field: "created_at", Desc: true})
	} else if containsAny(queryLower, []string{"oldest", "earliest", "最早", "最旧"}) {
		orderBy = append(orderBy, Order{Field: "created_at", Desc: false})
	}

	result := &Query{
		Table:  table,
		Action: action,
	}
	if len(where) > 0 {
		result.Where = where
	}
	if limit > 0 {
		result.Limit = limit
	}
	if len(orderBy) > 0 {
		result.OrderBy = orderBy
	}

	return result, nil
}

// matchTable finds the best matching table name.
// matchTable 查找最佳匹配的表名。
func (p *NLParser) matchTable(input string, tables []string) string {
	inputLower := strings.ToLower(input)

	// Exact match
	// 精确匹配
	for _, t := range tables {
		if strings.ToLower(t) == inputLower {
			return t
		}
	}

	// Singular/Plural match
	// 单复数匹配
	for _, t := range tables {
		tLower := strings.ToLower(t)
		if tLower == inputLower+"s" || tLower+"s" == inputLower {
			return t
		}
	}

	// Partial match
	// 部分匹配
	for _, t := range tables {
		if strings.Contains(strings.ToLower(t), inputLower) || strings.Contains(inputLower, strings.ToLower(t)) {
			return t
		}
	}

	// Default: return input as table name (pluralized)
	// 默认：返回输入作为表名（复数形式）
	if !strings.HasSuffix(inputLower, "s") {
		return inputLower + "s"
	}
	return inputLower
}

// matchTableChinese matches Chinese table references.
// matchTableChinese 匹配中文表引用。
func (p *NLParser) matchTableChinese(input string, tables []string) string {
	// Chinese to English table mappings
	// 中英文表名映射
	chineseMap := map[string]string{
		"用户":  "users",
		"订单":  "orders",
		"产品":  "products",
		"商品":  "products",
		"文章":  "articles",
		"帖子":  "posts",
		"评论":  "comments",
		"分类":  "categories",
		"标签":  "tags",
		"日志":  "logs",
		"记录":  "records",
		"配置":  "configs",
		"设置":  "settings",
		"角色":  "roles",
		"权限":  "permissions",
		"部门":  "departments",
		"员工":  "employees",
		"客户":  "customers",
		"供应商": "suppliers",
		"账户":  "accounts",
		"交易":  "transactions",
		"支付":  "payments",
		"通知":  "notifications",
		"消息":  "messages",
		"任务":  "tasks",
		"项目":  "projects",
		"文件":  "files",
		"图片":  "images",
		"视频":  "videos",
	}

	if tableName, ok := chineseMap[input]; ok {
		// Check if this table exists
		// 检查表是否存在
		for _, t := range tables {
			if strings.ToLower(t) == tableName {
				return t
			}
		}
		return tableName
	}

	// Try English matching
	// 尝试英文匹配
	return p.matchTable(input, tables)
}

// parseOperator converts operator string to Operator.
// parseOperator 将操作符字符串转换为 Operator。
func (p *NLParser) parseOperator(op string) Operator {
	switch strings.ToLower(strings.TrimSpace(op)) {
	case "=", "is", "equal", "equals", "等于":
		return OpEqual
	case "!=", "not", "不等于":
		return OpNotEqual
	case ">", "greater", "大于":
		return OpGreater
	case ">=", "大于等于":
		return OpGreaterOrEq
	case "<", "less", "小于":
		return OpLess
	case "<=", "小于等于":
		return OpLessOrEq
	case "like", "包含", "含有":
		return OpLike
	default:
		return OpEqual
	}
}

// parseValue converts value string to appropriate type.
// parseValue 将值字符串转换为适当的类型。
func (p *NLParser) parseValue(value string) any {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)

	// Try integer
	// 尝试整数
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}

	// Try float
	// 尝试浮点数
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	// Try boolean
	// 尝试布尔值
	switch strings.ToLower(value) {
	case "true", "是", "yes":
		return true
	case "false", "否", "no":
		return false
	}

	return value
}

// parseWhereClause parses a simple where clause.
// parseWhereClause 解析简单的 where 子句。
func (p *NLParser) parseWhereClause(clause string) []Condition {
	var conditions []Condition

	// Simple pattern: field op value
	// 简单模式：字段 操作符 值
	re := regexp.MustCompile(`(\w+)\s*(=|!=|>|<|>=|<=)\s*(\S+)`)
	matches := re.FindAllStringSubmatch(clause, -1)

	for _, match := range matches {
		if len(match) >= 4 {
			conditions = append(conditions, Condition{
				Field: match[1],
				Op:    p.parseOperator(match[2]),
				Value: p.parseValue(match[3]),
			})
		}
	}

	return conditions
}
