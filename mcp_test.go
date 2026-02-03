package goorm

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"
)

// TestMCPServerToolRegistration tests tool registration.
// TestMCPServerToolRegistration 测试工具注册。
func TestMCPServerToolRegistration(t *testing.T) {
	// Note: This test doesn't need a real database connection
	// 注意：此测试不需要真正的数据库连接
	s := &MCPServer{
		name:    "test-server",
		version: "1.0.0",
		tools:   make(map[string]*MCPTool),
	}

	// Register a custom tool
	// 注册自定义工具
	s.RegisterTool(&MCPTool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: json.RawMessage(`{"type": "object"}`),
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{"result": "success"}, nil
		},
	})

	if len(s.tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(s.tools))
	}

	if s.tools["test_tool"] == nil {
		t.Error("test_tool should be registered")
	}
}

// TestMCPMessageHandling tests message handling.
// TestMCPMessageHandling 测试消息处理。
func TestMCPMessageHandling(t *testing.T) {
	s := &MCPServer{
		name:    "test-server",
		version: "1.0.0",
		tools:   make(map[string]*MCPTool),
	}

	// Test initialize
	// 测试初始化
	msg := &MCPMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	response := s.handleMessage(msg)

	if response.Error != nil {
		t.Errorf("unexpected error: %v", response.Error)
	}

	result, ok := response.Result.(map[string]any)
	if !ok {
		t.Fatal("result should be a map")
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("unexpected protocol version: %v", result["protocolVersion"])
	}
}

// TestMCPToolsList tests the tools/list method.
// TestMCPToolsList 测试 tools/list 方法。
func TestMCPToolsList(t *testing.T) {
	s := &MCPServer{
		name:    "test-server",
		version: "1.0.0",
		tools:   make(map[string]*MCPTool),
	}

	s.RegisterTool(&MCPTool{
		Name:        "tool1",
		Description: "Tool 1",
		InputSchema: json.RawMessage(`{}`),
	})

	s.RegisterTool(&MCPTool{
		Name:        "tool2",
		Description: "Tool 2",
		InputSchema: json.RawMessage(`{}`),
	})

	msg := &MCPMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	response := s.handleMessage(msg)

	result, ok := response.Result.(map[string]any)
	if !ok {
		t.Fatal("result should be a map")
	}

	tools, ok := result["tools"].([]map[string]any)
	if !ok {
		t.Fatal("tools should be an array")
	}

	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}
}

// TestMCPToolsCall tests the tools/call method.
// TestMCPToolsCall 测试 tools/call 方法。
func TestMCPToolsCall(t *testing.T) {
	s := &MCPServer{
		name:    "test-server",
		version: "1.0.0",
		tools:   make(map[string]*MCPTool),
	}

	s.RegisterTool(&MCPTool{
		Name:        "echo",
		Description: "Echo tool",
		InputSchema: json.RawMessage(`{}`),
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return params, nil
		},
	})

	params, _ := json.Marshal(map[string]any{
		"name":      "echo",
		"arguments": map[string]any{"message": "hello"},
	})

	msg := &MCPMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  params,
	}

	response := s.handleMessage(msg)

	if response.Error != nil {
		t.Errorf("unexpected error: %v", response.Error)
	}
}

// TestMCPMethodNotFound tests handling of unknown methods.
// TestMCPMethodNotFound 测试未知方法的处理。
func TestMCPMethodNotFound(t *testing.T) {
	s := &MCPServer{
		name:    "test-server",
		version: "1.0.0",
		tools:   make(map[string]*MCPTool),
	}

	msg := &MCPMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown/method",
	}

	response := s.handleMessage(msg)

	if response.Error == nil {
		t.Error("expected error for unknown method")
	}

	if response.Error.Code != -32601 {
		t.Errorf("expected error code -32601, got %d", response.Error.Code)
	}
}

// TestMCPProcessMessage tests processing a message via input/output.
// TestMCPProcessMessage 测试通过输入/输出处理消息。
func TestMCPProcessMessage(t *testing.T) {
	s := &MCPServer{
		name:    "test-server",
		version: "1.0.0",
		tools:   make(map[string]*MCPTool),
	}

	// Create input message
	// 创建输入消息
	msg := MCPMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}
	inputData, _ := json.Marshal(msg)

	input := bytes.NewReader(inputData)
	output := &bytes.Buffer{}

	s.input = input
	s.output = output

	// Process the message
	// 处理消息
	err := s.processMessage()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check output
	// 检查输出
	var response MCPMessage
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Errorf("failed to parse response: %v", err)
	}

	if response.Error != nil {
		t.Errorf("unexpected error in response: %v", response.Error)
	}
}

// TestDefaultLogger tests the default logger.
// TestDefaultLogger 测试默认日志记录器。
func TestDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger()
	output := &bytes.Buffer{}
	logger.SetOutput(output)
	logger.SetLevel(LogLevelDebug)

	logger.Debug("debug message", "key", "value")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	content := output.String()

	if !containsAll(content, []string{"DEBUG", "INFO", "WARN", "ERROR"}) {
		t.Errorf("expected all log levels in output: %s", content)
	}
}

// TestLoggerLevel tests log level filtering.
// TestLoggerLevel 测试日志级别过滤。
func TestLoggerLevel(t *testing.T) {
	logger := NewDefaultLogger()
	output := &bytes.Buffer{}
	logger.SetOutput(output)
	logger.SetLevel(LogLevelWarn)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	content := output.String()

	if containsHelper(content, "DEBUG") || containsHelper(content, "INFO") {
		t.Error("debug and info should be filtered out")
	}

	if !containsHelper(content, "WARN") || !containsHelper(content, "ERROR") {
		t.Error("warn and error should be present")
	}
}

// TestMetricsCollector tests the metrics collector.
// TestMetricsCollector 测试指标收集器。
func TestMetricsCollector(t *testing.T) {
	m := NewMetricsCollector()

	m.RecordQuery(ActionFind, 100*time.Millisecond, nil)
	m.RecordQuery(ActionFind, 300*time.Millisecond, nil) // slow
	m.RecordQuery(ActionCreate, 50*time.Millisecond, nil)

	stats := m.GetStats()

	if stats["total_queries"].(int64) != 3 {
		t.Errorf("expected 3 queries, got %d", stats["total_queries"])
	}

	if stats["slow_queries"].(int64) != 1 {
		t.Errorf("expected 1 slow query, got %d", stats["slow_queries"])
	}

	byAction := stats["by_action"].(map[string]any)
	findStats := byAction["find"].(map[string]any)
	if findStats["count"].(int64) != 2 {
		t.Errorf("expected 2 find queries, got %d", findStats["count"])
	}
}

// TestMetricsReset tests metrics reset.
// TestMetricsReset 测试指标重置。
func TestMetricsReset(t *testing.T) {
	m := NewMetricsCollector()

	m.RecordQuery(ActionFind, 100*time.Millisecond, nil)
	m.Reset()

	stats := m.GetStats()
	if stats["total_queries"].(int64) != 0 {
		t.Error("metrics should be reset")
	}
}
