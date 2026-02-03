package goorm

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel represents the logging level.
// LogLevel 表示日志级别。
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelSilent
)

// String returns the string representation of the log level.
// String 返回日志级别的字符串表示。
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "SILENT"
	}
}

// DefaultLogger is the default logger implementation.
// DefaultLogger 是默认的日志记录器实现。
type DefaultLogger struct {
	mu       sync.Mutex
	level    LogLevel
	output   io.Writer
	prefix   string
	showTime bool
	showSQL  bool
}

// NewDefaultLogger creates a new default logger.
// NewDefaultLogger 创建新的默认日志记录器。
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		level:    LogLevelInfo,
		output:   os.Stdout,
		showTime: true,
		showSQL:  false,
	}
}

// SetLevel sets the log level.
// SetLevel 设置日志级别。
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput sets the output writer.
// SetOutput 设置输出写入器。
func (l *DefaultLogger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// SetPrefix sets the log prefix.
// SetPrefix 设置日志前缀。
func (l *DefaultLogger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// ShowSQL enables/disables SQL logging.
// ShowSQL 启用/禁用 SQL 日志。
func (l *DefaultLogger) ShowSQL(show bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.showSQL = show
}

// Debug logs a debug message.
// Debug 记录调试消息。
func (l *DefaultLogger) Debug(msg string, args ...any) {
	l.log(LogLevelDebug, msg, args...)
}

// Info logs an info message.
// Info 记录信息消息。
func (l *DefaultLogger) Info(msg string, args ...any) {
	l.log(LogLevelInfo, msg, args...)
}

// Warn logs a warning message.
// Warn 记录警告消息。
func (l *DefaultLogger) Warn(msg string, args ...any) {
	l.log(LogLevelWarn, msg, args...)
}

// Error logs an error message.
// Error 记录错误消息。
func (l *DefaultLogger) Error(msg string, args ...any) {
	l.log(LogLevelError, msg, args...)
}

func (l *DefaultLogger) log(level LogLevel, msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	var buf []byte

	if l.showTime {
		buf = append(buf, time.Now().Format("2006-01-02 15:04:05.000")...)
		buf = append(buf, ' ')
	}

	buf = append(buf, '[')
	buf = append(buf, level.String()...)
	buf = append(buf, ']', ' ')

	if l.prefix != "" {
		buf = append(buf, l.prefix...)
		buf = append(buf, ' ')
	}

	buf = append(buf, msg...)

	// Format key-value pairs
	// 格式化键值对
	for i := 0; i < len(args); i += 2 {
		buf = append(buf, ' ')
		if i < len(args) {
			buf = append(buf, fmt.Sprintf("%v", args[i])...)
			buf = append(buf, '=')
		}
		if i+1 < len(args) {
			buf = append(buf, fmt.Sprintf("%v", args[i+1])...)
		}
	}

	buf = append(buf, '\n')
	l.output.Write(buf)
}

// QueryLogger logs query execution.
// QueryLogger 记录查询执行。
type QueryLogger struct {
	logger        Logger
	slowThreshold time.Duration
	logAll        bool
}

// NewQueryLogger creates a new query logger.
// NewQueryLogger 创建新的查询日志记录器。
func NewQueryLogger(logger Logger) *QueryLogger {
	return &QueryLogger{
		logger:        logger,
		slowThreshold: 200 * time.Millisecond,
		logAll:        false,
	}
}

// SetSlowThreshold sets the slow query threshold.
// SetSlowThreshold 设置慢查询阈值。
func (l *QueryLogger) SetSlowThreshold(d time.Duration) {
	l.slowThreshold = d
}

// LogAll enables logging of all queries.
// LogAll 启用记录所有查询。
func (l *QueryLogger) LogAll(enable bool) {
	l.logAll = enable
}

// LogQuery logs a query execution.
// LogQuery 记录查询执行。
func (l *QueryLogger) LogQuery(sql string, params []any, duration time.Duration, err error) {
	if !l.logAll && err == nil && duration < l.slowThreshold {
		return
	}

	if err != nil {
		l.logger.Error("query failed",
			"sql", sql,
			"params", fmt.Sprintf("%v", params),
			"duration_ms", duration.Milliseconds(),
			"error", err.Error(),
		)
	} else if duration >= l.slowThreshold {
		l.logger.Warn("slow query",
			"sql", sql,
			"params", fmt.Sprintf("%v", params),
			"duration_ms", duration.Milliseconds(),
		)
	} else if l.logAll {
		l.logger.Debug("query",
			"sql", sql,
			"params", fmt.Sprintf("%v", params),
			"duration_ms", duration.Milliseconds(),
		)
	}
}

// MetricsCollector collects query metrics.
// MetricsCollector 收集查询指标。
type MetricsCollector struct {
	mu            sync.RWMutex
	totalQueries  int64
	totalDuration time.Duration
	slowQueries   int64
	errorQueries  int64
	byAction      map[Action]*ActionMetrics
}

// ActionMetrics contains metrics for a specific action.
// ActionMetrics 包含特定操作的指标。
type ActionMetrics struct {
	Count    int64
	Duration time.Duration
	Errors   int64
}

// NewMetricsCollector creates a new metrics collector.
// NewMetricsCollector 创建新的指标收集器。
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		byAction: make(map[Action]*ActionMetrics),
	}
}

// RecordQuery records a query execution.
// RecordQuery 记录查询执行。
func (m *MetricsCollector) RecordQuery(action Action, duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalQueries++
	m.totalDuration += duration

	if duration >= 200*time.Millisecond {
		m.slowQueries++
	}

	if err != nil {
		m.errorQueries++
	}

	if m.byAction[action] == nil {
		m.byAction[action] = &ActionMetrics{}
	}

	m.byAction[action].Count++
	m.byAction[action].Duration += duration
	if err != nil {
		m.byAction[action].Errors++
	}
}

// GetStats returns the current metrics.
// GetStats 返回当前指标。
func (m *MetricsCollector) GetStats() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgDuration := time.Duration(0)
	if m.totalQueries > 0 {
		avgDuration = m.totalDuration / time.Duration(m.totalQueries)
	}

	byAction := make(map[string]any)
	for action, metrics := range m.byAction {
		avgActionDuration := time.Duration(0)
		if metrics.Count > 0 {
			avgActionDuration = metrics.Duration / time.Duration(metrics.Count)
		}
		byAction[string(action)] = map[string]any{
			"count":           metrics.Count,
			"total_duration":  metrics.Duration.String(),
			"avg_duration_ms": avgActionDuration.Milliseconds(),
			"errors":          metrics.Errors,
		}
	}

	return map[string]any{
		"total_queries":   m.totalQueries,
		"total_duration":  m.totalDuration.String(),
		"avg_duration_ms": avgDuration.Milliseconds(),
		"slow_queries":    m.slowQueries,
		"error_queries":   m.errorQueries,
		"by_action":       byAction,
	}
}

// Reset resets all metrics.
// Reset 重置所有指标。
func (m *MetricsCollector) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalQueries = 0
	m.totalDuration = 0
	m.slowQueries = 0
	m.errorQueries = 0
	m.byAction = make(map[Action]*ActionMetrics)
}
