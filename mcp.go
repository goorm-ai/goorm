package goorm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// MCPServer implements a Model Context Protocol server for AI integration.
// It exposes GoORM operations as MCP tools that AI can discover and invoke.
//
// MCPServer 实现用于 AI 集成的模型上下文协议服务器。
// 它将 GoORM 操作公开为 AI 可以发现和调用的 MCP 工具。
type MCPServer struct {
	db      *DB
	name    string
	version string
	tools   map[string]*MCPTool
	mu      sync.RWMutex
	running bool
	input   io.Reader
	output  io.Writer
}

// MCPTool represents an MCP tool definition.
// MCPTool 表示 MCP 工具定义。
type MCPTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
	Handler     MCPToolHandler  `json:"-"`
}

// MCPToolHandler is the function signature for tool handlers.
// MCPToolHandler 是工具处理器的函数签名。
type MCPToolHandler func(ctx context.Context, params map[string]any) (any, error)

// MCPMessage represents an MCP protocol message.
// MCPMessage 表示 MCP 协议消息。
type MCPMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents an MCP error.
// MCPError 表示 MCP 错误。
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// NewMCPServer creates a new MCP server.
// NewMCPServer 创建新的 MCP 服务器。
func NewMCPServer(db *DB) *MCPServer {
	s := &MCPServer{
		db:      db,
		name:    "goorm-mcp",
		version: Version,
		tools:   make(map[string]*MCPTool),
		input:   os.Stdin,
		output:  os.Stdout,
	}

	// Register built-in tools
	// 注册内置工具
	s.registerBuiltinTools()

	return s
}

// registerBuiltinTools registers the 14 standard MCP tools.
// registerBuiltinTools 注册 14 个标准 MCP 工具。
func (s *MCPServer) registerBuiltinTools() {
	// 1. execute_query - Execute JQL query
	// 1. execute_query - 执行 JQL 查询
	s.RegisterTool(&MCPTool{
		Name:        "execute_query",
		Description: "Execute a JQL (JSON Query Language) query against the database. Supports find, create, update, delete, count, and aggregate operations. / 对数据库执行 JQL（JSON 查询语言）查询。支持 find、create、update、delete、count 和 aggregate 操作。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {
					"type": "object",
					"description": "The JQL query object"
				}
			},
			"required": ["query"]
		}`),
		Handler: s.handleExecuteQuery,
	})

	// 2. list_tables - List all tables
	// 2. list_tables - 列出所有表
	s.RegisterTool(&MCPTool{
		Name:        "list_tables",
		Description: "List all registered tables in the database with their basic information. / 列出数据库中所有已注册的表及其基本信息。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
		Handler: s.handleListTables,
	})

	// 3. describe_table - Describe table schema
	// 3. describe_table - 描述表结构
	s.RegisterTool(&MCPTool{
		Name:        "describe_table",
		Description: "Get detailed schema information for a specific table including columns, types, and constraints. / 获取特定表的详细 Schema 信息，包括列、类型和约束。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"table": {
					"type": "string",
					"description": "The table name to describe"
				}
			},
			"required": ["table"]
		}`),
		Handler: s.handleDescribeTable,
	})

	// 4. find_records - Find records with conditions
	// 4. find_records - 按条件查找记录
	s.RegisterTool(&MCPTool{
		Name:        "find_records",
		Description: "Find records in a table with optional filtering, sorting, and pagination. / 在表中查找记录，支持可选的过滤、排序和分页。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"table": {"type": "string", "description": "Table name"},
				"where": {"type": "array", "description": "Filter conditions"},
				"order_by": {"type": "array", "description": "Sort order"},
				"limit": {"type": "integer", "description": "Max records to return"},
				"offset": {"type": "integer", "description": "Records to skip"}
			},
			"required": ["table"]
		}`),
		Handler: s.handleFindRecords,
	})

	// 5. create_record - Create a new record
	// 5. create_record - 创建新记录
	s.RegisterTool(&MCPTool{
		Name:        "create_record",
		Description: "Create a new record in the specified table. / 在指定表中创建新记录。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"table": {"type": "string", "description": "Table name"},
				"data": {"type": "object", "description": "Record data"}
			},
			"required": ["table", "data"]
		}`),
		Handler: s.handleCreateRecord,
	})

	// 6. update_records - Update existing records
	// 6. update_records - 更新现有记录
	s.RegisterTool(&MCPTool{
		Name:        "update_records",
		Description: "Update records matching the specified conditions. / 更新匹配指定条件的记录。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"table": {"type": "string", "description": "Table name"},
				"where": {"type": "array", "description": "Filter conditions"},
				"data": {"type": "object", "description": "Data to update"}
			},
			"required": ["table", "where", "data"]
		}`),
		Handler: s.handleUpdateRecords,
	})

	// 7. delete_records - Delete records
	// 7. delete_records - 删除记录
	s.RegisterTool(&MCPTool{
		Name:        "delete_records",
		Description: "Delete records matching the specified conditions. / 删除匹配指定条件的记录。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"table": {"type": "string", "description": "Table name"},
				"where": {"type": "array", "description": "Filter conditions"}
			},
			"required": ["table", "where"]
		}`),
		Handler: s.handleDeleteRecords,
	})

	// 8. count_records - Count records
	// 8. count_records - 统计记录数
	s.RegisterTool(&MCPTool{
		Name:        "count_records",
		Description: "Count records matching the specified conditions. / 统计匹配指定条件的记录数。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"table": {"type": "string", "description": "Table name"},
				"where": {"type": "array", "description": "Optional filter conditions"}
			},
			"required": ["table"]
		}`),
		Handler: s.handleCountRecords,
	})

	// 9. execute_transaction - Execute transaction
	// 9. execute_transaction - 执行事务
	s.RegisterTool(&MCPTool{
		Name:        "execute_transaction",
		Description: "Execute multiple operations as an atomic transaction. / 将多个操作作为原子事务执行。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"operations": {
					"type": "array",
					"description": "Array of JQL operations to execute"
				}
			},
			"required": ["operations"]
		}`),
		Handler: s.handleExecuteTransaction,
	})

	// 10. explain_query - Explain query plan
	// 10. explain_query - 解释查询计划
	s.RegisterTool(&MCPTool{
		Name:        "explain_query",
		Description: "Get the SQL that would be generated for a JQL query without executing it. / 获取将为 JQL 查询生成的 SQL，而不执行它。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {"type": "object", "description": "The JQL query to explain"}
			},
			"required": ["query"]
		}`),
		Handler: s.handleExplainQuery,
	})

	// 11. aggregate - Perform aggregation
	// 11. aggregate - 执行聚合
	s.RegisterTool(&MCPTool{
		Name:        "aggregate",
		Description: "Perform aggregation operations like SUM, AVG, COUNT, MAX, MIN with optional grouping. / 执行聚合操作如 SUM、AVG、COUNT、MAX、MIN，支持可选分组。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"table": {"type": "string", "description": "Table name"},
				"select": {"type": "array", "description": "Aggregation functions"},
				"group_by": {"type": "array", "description": "Grouping columns"},
				"having": {"type": "array", "description": "Having conditions"},
				"where": {"type": "array", "description": "Filter conditions"}
			},
			"required": ["table", "select"]
		}`),
		Handler: s.handleAggregate,
	})

	// 12. sync_schema - Sync database schema
	// 12. sync_schema - 同步数据库 Schema
	s.RegisterTool(&MCPTool{
		Name:        "sync_schema",
		Description: "Synchronize the database schema with registered models. Creates or updates tables as needed. / 将数据库 Schema 与已注册的模型同步。根据需要创建或更新表。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"preview": {"type": "boolean", "description": "If true, only show changes without applying"}
			}
		}`),
		Handler: s.handleSyncSchema,
	})

	// 13. get_stats - Get database statistics
	// 13. get_stats - 获取数据库统计信息
	s.RegisterTool(&MCPTool{
		Name:        "get_stats",
		Description: "Get database statistics including connection pool status and query metrics. / 获取数据库统计信息，包括连接池状态和查询指标。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
		Handler: s.handleGetStats,
	})

	// 14. natural_language - Process natural language query
	// 14. natural_language - 处理自然语言查询
	s.RegisterTool(&MCPTool{
		Name:        "natural_language",
		Description: "Execute a natural language query. The system will convert it to JQL and execute. / 执行自然语言查询。系统将转换为 JQL 并执行。",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {"type": "string", "description": "Natural language query"}
			},
			"required": ["query"]
		}`),
		Handler: s.handleNaturalLanguage,
	})
}

// RegisterTool registers a custom MCP tool.
// RegisterTool 注册自定义 MCP 工具。
func (s *MCPServer) RegisterTool(tool *MCPTool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
}

// Start starts the MCP server.
// Start 启动 MCP 服务器。
func (s *MCPServer) Start(ctx context.Context) error {
	s.running = true

	for s.running {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := s.processMessage(); err != nil {
				if err == io.EOF {
					return nil
				}
				// Log error but continue
				continue
			}
		}
	}

	return nil
}

// Stop stops the MCP server.
// Stop 停止 MCP 服务器。
func (s *MCPServer) Stop() {
	s.running = false
}

// processMessage processes a single MCP message.
// processMessage 处理单个 MCP 消息。
func (s *MCPServer) processMessage() error {
	decoder := json.NewDecoder(s.input)
	var msg MCPMessage
	if err := decoder.Decode(&msg); err != nil {
		return err
	}

	response := s.handleMessage(&msg)
	if response != nil {
		encoder := json.NewEncoder(s.output)
		return encoder.Encode(response)
	}

	return nil
}

// handleMessage handles an MCP message and returns a response.
// handleMessage 处理 MCP 消息并返回响应。
func (s *MCPServer) handleMessage(msg *MCPMessage) *MCPMessage {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "tools/list":
		return s.handleToolsList(msg)
	case "tools/call":
		return s.handleToolsCall(msg)
	default:
		return &MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

// handleInitialize handles the initialize request.
// handleInitialize 处理初始化请求。
func (s *MCPServer) handleInitialize(msg *MCPMessage) *MCPMessage {
	return &MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]any{
				"name":    s.name,
				"version": s.version,
			},
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
		},
	}
}

// handleToolsList handles the tools/list request.
// handleToolsList 处理 tools/list 请求。
func (s *MCPServer) handleToolsList(msg *MCPMessage) *MCPMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]map[string]any, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}

	return &MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: map[string]any{
			"tools": tools,
		},
	}
}

// handleToolsCall handles the tools/call request.
// handleToolsCall 处理 tools/call 请求。
func (s *MCPServer) handleToolsCall(msg *MCPMessage) *MCPMessage {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}

	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return &MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	s.mu.RLock()
	tool, exists := s.tools[params.Name]
	s.mu.RUnlock()

	if !exists {
		return &MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: fmt.Sprintf("Tool not found: %s", params.Name),
			},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := tool.Handler(ctx, params.Arguments)
	if err != nil {
		return &MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Result: map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": fmt.Sprintf("Error: %s", err.Error()),
					},
				},
				"isError": true,
			},
		}
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return &MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": string(resultJSON),
				},
			},
		},
	}
}

// --- Tool Handlers ---
// --- 工具处理器 ---

func (s *MCPServer) handleExecuteQuery(ctx context.Context, params map[string]any) (any, error) {
	queryData, ok := params["query"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid query parameter")
	}

	queryJSON, _ := json.Marshal(queryData)
	query, err := ParseQuery(string(queryJSON))
	if err != nil {
		return nil, err
	}

	return s.db.ExecuteQuery(ctx, query), nil
}

func (s *MCPServer) handleListTables(ctx context.Context, params map[string]any) (any, error) {
	return s.db.ExecuteQuery(ctx, &Query{Action: ActionListTables}), nil
}

func (s *MCPServer) handleDescribeTable(ctx context.Context, params map[string]any) (any, error) {
	table, _ := params["table"].(string)
	if table == "" {
		return nil, fmt.Errorf("table is required")
	}
	return s.db.ExecuteQuery(ctx, &Query{Action: ActionDescribe, Table: table}), nil
}

func (s *MCPServer) handleFindRecords(ctx context.Context, params map[string]any) (any, error) {
	query := &Query{
		Table:  params["table"].(string),
		Action: ActionFind,
	}

	if where, ok := params["where"].([]any); ok {
		whereJSON, _ := json.Marshal(where)
		json.Unmarshal(whereJSON, &query.Where)
	}

	if orderBy, ok := params["order_by"].([]any); ok {
		orderJSON, _ := json.Marshal(orderBy)
		json.Unmarshal(orderJSON, &query.OrderBy)
	}

	if limit, ok := params["limit"].(float64); ok {
		query.Limit = int(limit)
	}

	if offset, ok := params["offset"].(float64); ok {
		query.Offset = int(offset)
	}

	return s.db.ExecuteQuery(ctx, query), nil
}

func (s *MCPServer) handleCreateRecord(ctx context.Context, params map[string]any) (any, error) {
	table, _ := params["table"].(string)
	data, _ := params["data"].(map[string]any)

	if table == "" || data == nil {
		return nil, fmt.Errorf("table and data are required")
	}

	return s.db.ExecuteQuery(ctx, &Query{
		Table:  table,
		Action: ActionCreate,
		Data:   data,
	}), nil
}

func (s *MCPServer) handleUpdateRecords(ctx context.Context, params map[string]any) (any, error) {
	query := &Query{
		Table:  params["table"].(string),
		Action: ActionUpdate,
	}

	if where, ok := params["where"].([]any); ok {
		whereJSON, _ := json.Marshal(where)
		json.Unmarshal(whereJSON, &query.Where)
	}

	if data, ok := params["data"].(map[string]any); ok {
		query.Data = data
	}

	return s.db.ExecuteQuery(ctx, query), nil
}

func (s *MCPServer) handleDeleteRecords(ctx context.Context, params map[string]any) (any, error) {
	query := &Query{
		Table:  params["table"].(string),
		Action: ActionDelete,
	}

	if where, ok := params["where"].([]any); ok {
		whereJSON, _ := json.Marshal(where)
		json.Unmarshal(whereJSON, &query.Where)
	}

	return s.db.ExecuteQuery(ctx, query), nil
}

func (s *MCPServer) handleCountRecords(ctx context.Context, params map[string]any) (any, error) {
	query := &Query{
		Table:  params["table"].(string),
		Action: ActionCount,
	}

	if where, ok := params["where"].([]any); ok {
		whereJSON, _ := json.Marshal(where)
		json.Unmarshal(whereJSON, &query.Where)
	}

	return s.db.ExecuteQuery(ctx, query), nil
}

func (s *MCPServer) handleExecuteTransaction(ctx context.Context, params map[string]any) (any, error) {
	operations, ok := params["operations"].([]any)
	if !ok {
		return nil, fmt.Errorf("operations is required")
	}

	ops := make([]Query, 0, len(operations))
	for _, op := range operations {
		opJSON, _ := json.Marshal(op)
		var q Query
		if err := json.Unmarshal(opJSON, &q); err != nil {
			return nil, err
		}
		ops = append(ops, q)
	}

	return s.db.ExecuteQuery(ctx, &Query{
		Action:     ActionTransaction,
		Operations: ops,
	}), nil
}

func (s *MCPServer) handleExplainQuery(ctx context.Context, params map[string]any) (any, error) {
	queryData, ok := params["query"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("query is required")
	}

	queryJSON, _ := json.Marshal(queryData)
	query, err := ParseQuery(string(queryJSON))
	if err != nil {
		return nil, err
	}

	return s.db.ExecuteQuery(ctx, &Query{
		Action:         ActionExplain,
		QueryToExplain: query,
	}), nil
}

func (s *MCPServer) handleAggregate(ctx context.Context, params map[string]any) (any, error) {
	query := &Query{
		Table:  params["table"].(string),
		Action: ActionAggregate,
	}

	if sel, ok := params["select"].([]any); ok {
		query.Select = sel
	}

	if groupBy, ok := params["group_by"].([]any); ok {
		for _, g := range groupBy {
			if gs, ok := g.(string); ok {
				query.GroupBy = append(query.GroupBy, gs)
			}
		}
	}

	if where, ok := params["where"].([]any); ok {
		whereJSON, _ := json.Marshal(where)
		json.Unmarshal(whereJSON, &query.Where)
	}

	if having, ok := params["having"].([]any); ok {
		havingJSON, _ := json.Marshal(having)
		json.Unmarshal(havingJSON, &query.Having)
	}

	return s.db.ExecuteQuery(ctx, query), nil
}

func (s *MCPServer) handleSyncSchema(ctx context.Context, params map[string]any) (any, error) {
	preview, _ := params["preview"].(bool)

	migrator := NewMigrator(s.db)
	plan, err := migrator.Plan(ctx)
	if err != nil {
		return nil, err
	}

	if preview {
		return map[string]any{
			"changes": plan.Changes,
			"preview": true,
		}, nil
	}

	if err := migrator.Execute(ctx, plan); err != nil {
		return nil, err
	}

	return map[string]any{
		"changes": plan.Changes,
		"applied": true,
	}, nil
}

func (s *MCPServer) handleGetStats(ctx context.Context, params map[string]any) (any, error) {
	stats := s.db.sqlDB.Stats()
	return map[string]any{
		"connection_pool": map[string]any{
			"open_connections":    stats.OpenConnections,
			"in_use":              stats.InUse,
			"idle":                stats.Idle,
			"max_open":            stats.MaxOpenConnections,
			"wait_count":          stats.WaitCount,
			"wait_duration_ms":    stats.WaitDuration.Milliseconds(),
			"max_idle_closed":     stats.MaxIdleClosed,
			"max_lifetime_closed": stats.MaxLifetimeClosed,
		},
		"tables": len(s.db.registry.ListTables()),
	}, nil
}

func (s *MCPServer) handleNaturalLanguage(ctx context.Context, params map[string]any) (any, error) {
	query, _ := params["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Use the NL method on DB
	// 使用 DB 上的 NL 方法
	return s.db.NL(query), nil
}
