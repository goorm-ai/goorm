package goorm

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

// HealthStatus represents the health status of the database connection.
// HealthStatus 表示数据库连接的健康状态。
type HealthStatus string

const (
	// HealthStatusHealthy indicates the database is healthy.
	// HealthStatusHealthy 表示数据库健康。
	HealthStatusHealthy HealthStatus = "healthy"

	// HealthStatusDegraded indicates the database is experiencing issues.
	// HealthStatusDegraded 表示数据库正在经历问题。
	HealthStatusDegraded HealthStatus = "degraded"

	// HealthStatusUnhealthy indicates the database is not accessible.
	// HealthStatusUnhealthy 表示数据库无法访问。
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck contains the result of a health check.
// HealthCheck 包含健康检查的结果。
type HealthCheck struct {
	// Status is the overall health status.
	// Status 是整体健康状态。
	Status HealthStatus `json:"status"`

	// Latency is the ping latency.
	// Latency 是 ping 延迟。
	Latency time.Duration `json:"latency"`

	// LastCheck is when the last check was performed.
	// LastCheck 是上次检查的时间。
	LastCheck time.Time `json:"last_check"`

	// Error is the error message if unhealthy.
	// Error 是不健康时的错误消息。
	Error string `json:"error,omitempty"`

	// Pool contains connection pool statistics.
	// Pool 包含连接池统计信息。
	Pool *PoolStats `json:"pool"`

	// Details contains additional health check details.
	// Details 包含额外的健康检查详情。
	Details map[string]any `json:"details,omitempty"`
}

// PoolStats contains connection pool statistics.
// PoolStats 包含连接池统计信息。
type PoolStats struct {
	// OpenConnections is the number of open connections.
	// OpenConnections 是打开的连接数。
	OpenConnections int `json:"open_connections"`

	// InUse is the number of connections currently in use.
	// InUse 是当前正在使用的连接数。
	InUse int `json:"in_use"`

	// Idle is the number of idle connections.
	// Idle 是空闲连接数。
	Idle int `json:"idle"`

	// MaxOpen is the maximum number of open connections allowed.
	// MaxOpen 是允许的最大打开连接数。
	MaxOpen int `json:"max_open"`

	// WaitCount is the total number of connections waited for.
	// WaitCount 是等待连接的总次数。
	WaitCount int64 `json:"wait_count"`

	// WaitDuration is the total time blocked waiting for connections.
	// WaitDuration 是等待连接的总阻塞时间。
	WaitDuration time.Duration `json:"wait_duration"`

	// MaxIdleClosed is the total number of connections closed due to max idle.
	// MaxIdleClosed 是因最大空闲而关闭的连接总数。
	MaxIdleClosed int64 `json:"max_idle_closed"`

	// MaxLifetimeClosed is the total number of connections closed due to max lifetime.
	// MaxLifetimeClosed 是因最大生命周期而关闭的连接总数。
	MaxLifetimeClosed int64 `json:"max_lifetime_closed"`
}

// HealthChecker manages health checks for the database.
// HealthChecker 管理数据库的健康检查。
type HealthChecker struct {
	db           *DB
	mu           sync.RWMutex
	lastCheck    *HealthCheck
	checkTimeout time.Duration
	interval     time.Duration
	thresholds   HealthThresholds
	running      bool
	stopCh       chan struct{}
}

// HealthThresholds defines thresholds for health status.
// HealthThresholds 定义健康状态的阈值。
type HealthThresholds struct {
	// MaxLatency is the maximum acceptable ping latency.
	// MaxLatency 是最大可接受的 ping 延迟。
	MaxLatency time.Duration

	// MaxOpenConnectionsPercent is the max percentage of open connections.
	// MaxOpenConnectionsPercent 是打开连接的最大百分比。
	MaxOpenConnectionsPercent float64

	// MaxWaitCount is the maximum wait count before degraded status.
	// MaxWaitCount 是降级前的最大等待次数。
	MaxWaitCount int64
}

// DefaultHealthThresholds returns default health thresholds.
// DefaultHealthThresholds 返回默认的健康阈值。
func DefaultHealthThresholds() HealthThresholds {
	return HealthThresholds{
		MaxLatency:                500 * time.Millisecond,
		MaxOpenConnectionsPercent: 80.0,
		MaxWaitCount:              100,
	}
}

// NewHealthChecker creates a new health checker.
// NewHealthChecker 创建新的健康检查器。
func NewHealthChecker(db *DB) *HealthChecker {
	return &HealthChecker{
		db:           db,
		checkTimeout: 5 * time.Second,
		interval:     30 * time.Second,
		thresholds:   DefaultHealthThresholds(),
		stopCh:       make(chan struct{}),
	}
}

// SetTimeout sets the check timeout.
// SetTimeout 设置检查超时。
func (h *HealthChecker) SetTimeout(timeout time.Duration) {
	h.checkTimeout = timeout
}

// SetInterval sets the background check interval.
// SetInterval 设置后台检查间隔。
func (h *HealthChecker) SetInterval(interval time.Duration) {
	h.interval = interval
}

// SetThresholds sets the health thresholds.
// SetThresholds 设置健康阈值。
func (h *HealthChecker) SetThresholds(thresholds HealthThresholds) {
	h.thresholds = thresholds
}

// Check performs a health check and returns the result.
// Check 执行健康检查并返回结果。
func (h *HealthChecker) Check(ctx context.Context) *HealthCheck {
	check := &HealthCheck{
		Status:    HealthStatusHealthy,
		LastCheck: time.Now(),
		Details:   make(map[string]any),
	}

	// Ping database
	// Ping 数据库
	start := time.Now()
	pingCtx, cancel := context.WithTimeout(ctx, h.checkTimeout)
	defer cancel()

	if err := h.db.sqlDB.PingContext(pingCtx); err != nil {
		check.Status = HealthStatusUnhealthy
		check.Error = err.Error()
		check.Latency = time.Since(start)
		h.updateLastCheck(check)
		return check
	}
	check.Latency = time.Since(start)

	// Get pool stats
	// 获取连接池统计
	check.Pool = h.getPoolStats()

	// Evaluate health based on thresholds
	// 根据阈值评估健康状态
	h.evaluateHealth(check)

	// Add additional details
	// 添加额外详情
	check.Details["dialect"] = h.db.dialect.Name()
	check.Details["tables"] = len(h.db.registry.ListTables())

	h.updateLastCheck(check)
	return check
}

// getPoolStats returns connection pool statistics.
// getPoolStats 返回连接池统计信息。
func (h *HealthChecker) getPoolStats() *PoolStats {
	stats := h.db.sqlDB.Stats()
	return &PoolStats{
		OpenConnections:   stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		MaxOpen:           stats.MaxOpenConnections,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}
}

// evaluateHealth evaluates health based on thresholds.
// evaluateHealth 根据阈值评估健康状态。
func (h *HealthChecker) evaluateHealth(check *HealthCheck) {
	// Check latency
	// 检查延迟
	if check.Latency > h.thresholds.MaxLatency {
		check.Status = HealthStatusDegraded
		check.Details["latency_warning"] = "Ping latency exceeds threshold"
	}

	// Check connection pool usage
	// 检查连接池使用率
	if check.Pool.MaxOpen > 0 {
		usagePercent := float64(check.Pool.OpenConnections) / float64(check.Pool.MaxOpen) * 100
		check.Details["pool_usage_percent"] = usagePercent

		if usagePercent > h.thresholds.MaxOpenConnectionsPercent {
			if check.Status == HealthStatusHealthy {
				check.Status = HealthStatusDegraded
			}
			check.Details["pool_warning"] = "Connection pool usage exceeds threshold"
		}
	}

	// Check wait count
	// 检查等待次数
	if check.Pool.WaitCount > h.thresholds.MaxWaitCount {
		if check.Status == HealthStatusHealthy {
			check.Status = HealthStatusDegraded
		}
		check.Details["wait_warning"] = "Connection wait count exceeds threshold"
	}
}

// updateLastCheck updates the cached last check result.
// updateLastCheck 更新缓存的最后检查结果。
func (h *HealthChecker) updateLastCheck(check *HealthCheck) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastCheck = check
}

// LastCheck returns the last health check result.
// LastCheck 返回上次健康检查结果。
func (h *HealthChecker) LastCheck() *HealthCheck {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastCheck
}

// Start starts background health checking.
// Start 启动后台健康检查。
func (h *HealthChecker) Start(ctx context.Context) {
	h.running = true

	go func() {
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()

		// Initial check
		// 初始检查
		h.Check(ctx)

		for {
			select {
			case <-ctx.Done():
				h.running = false
				return
			case <-h.stopCh:
				h.running = false
				return
			case <-ticker.C:
				h.Check(ctx)
			}
		}
	}()
}

// Stop stops background health checking.
// Stop 停止后台健康检查。
func (h *HealthChecker) Stop() {
	if h.running {
		close(h.stopCh)
		h.running = false
	}
}

// IsRunning returns whether background checking is running.
// IsRunning 返回后台检查是否正在运行。
func (h *HealthChecker) IsRunning() bool {
	return h.running
}

// --- DB Health Methods ---
// --- DB 健康方法 ---

// Health returns a health checker for the database.
// Health 返回数据库的健康检查器。
func (db *DB) Health() *HealthChecker {
	return NewHealthChecker(db)
}

// Ping pings the database with the given context.
// Ping 使用给定的上下文 ping 数据库。
func (db *DB) PingContext(ctx context.Context) error {
	return db.sqlDB.PingContext(ctx)
}

// Ping pings the database with the default context.
// Ping 使用默认上下文 ping 数据库。
func (db *DB) Ping() error {
	return db.PingContext(db.ctx)
}

// Stats returns the database connection pool statistics.
// Stats 返回数据库连接池统计信息。
func (db *DB) Stats() sql.DBStats {
	return db.sqlDB.Stats()
}

// IsHealthy performs a quick health check and returns true if healthy.
// IsHealthy 执行快速健康检查，如果健康则返回 true。
func (db *DB) IsHealthy(ctx context.Context) bool {
	checker := NewHealthChecker(db)
	checker.SetTimeout(2 * time.Second)
	check := checker.Check(ctx)
	return check.Status == HealthStatusHealthy
}
