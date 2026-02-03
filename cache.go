package goorm

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// Cache is the interface for query result caching.
// Cache 是查询结果缓存的接口。
type Cache interface {
	// Get retrieves a cached result by key.
	// Get 通过键获取缓存的结果。
	Get(key string) (*Result, bool)

	// Set stores a result in the cache.
	// Set 将结果存储到缓存中。
	Set(key string, result *Result, ttl time.Duration)

	// Delete removes a key from the cache.
	// Delete 从缓存中删除键。
	Delete(key string)

	// Clear clears all cached entries for a table.
	// Clear 清除表的所有缓存条目。
	Clear(table string)

	// ClearAll clears all cached entries.
	// ClearAll 清除所有缓存条目。
	ClearAll()
}

// MemoryCache is an in-memory cache implementation.
// MemoryCache 是内存缓存的实现。
type MemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	tables  map[string]map[string]bool // table -> keys
}

type cacheEntry struct {
	result    *Result
	expiresAt time.Time
}

// NewMemoryCache creates a new memory cache.
// NewMemoryCache 创建新的内存缓存。
func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		entries: make(map[string]*cacheEntry),
		tables:  make(map[string]map[string]bool),
	}

	// Start cleanup goroutine
	// 启动清理 goroutine
	go c.cleanup()

	return c
}

// Get retrieves a cached result.
// Get 获取缓存的结果。
func (c *MemoryCache) Get(key string) (*Result, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.result, true
}

// Set stores a result in the cache.
// Set 将结果存储到缓存中。
func (c *MemoryCache) Set(key string, result *Result, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &cacheEntry{
		result:    result,
		expiresAt: time.Now().Add(ttl),
	}
}

// SetWithTable stores a result with table tracking.
// SetWithTable 存储带表跟踪的结果。
func (c *MemoryCache) SetWithTable(key, table string, result *Result, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &cacheEntry{
		result:    result,
		expiresAt: time.Now().Add(ttl),
	}

	if c.tables[table] == nil {
		c.tables[table] = make(map[string]bool)
	}
	c.tables[table][key] = true
}

// Delete removes a key from the cache.
// Delete 从缓存中删除键。
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear clears all cached entries for a table.
// Clear 清除表的所有缓存条目。
func (c *MemoryCache) Clear(table string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys, exists := c.tables[table]
	if !exists {
		return
	}

	for key := range keys {
		delete(c.entries, key)
	}
	delete(c.tables, table)
}

// ClearAll clears all cached entries.
// ClearAll 清除所有缓存条目。
func (c *MemoryCache) ClearAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
	c.tables = make(map[string]map[string]bool)
}

// Stats returns cache statistics.
// Stats 返回缓存统计信息。
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		Entries: len(c.entries),
		Tables:  len(c.tables),
	}
}

// cleanup periodically removes expired entries.
// cleanup 定期删除过期的条目。
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.expiresAt) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// CacheStats contains cache statistics.
// CacheStats 包含缓存统计信息。
type CacheStats struct {
	Entries int
	Tables  int
}

// CacheManager manages query caching.
// CacheManager 管理查询缓存。
type CacheManager struct {
	cache   Cache
	enabled bool
	ttl     time.Duration
	tables  map[string]bool // tables to cache
}

// NewCacheManager creates a new cache manager.
// NewCacheManager 创建新的缓存管理器。
func NewCacheManager() *CacheManager {
	return &CacheManager{
		cache:   NewMemoryCache(),
		enabled: false,
		ttl:     5 * time.Minute,
		tables:  make(map[string]bool),
	}
}

// Enable enables caching.
// Enable 启用缓存。
func (m *CacheManager) Enable() {
	m.enabled = true
}

// Disable disables caching.
// Disable 禁用缓存。
func (m *CacheManager) Disable() {
	m.enabled = false
}

// SetTTL sets the default TTL for cached entries.
// SetTTL 设置缓存条目的默认 TTL。
func (m *CacheManager) SetTTL(ttl time.Duration) {
	m.ttl = ttl
}

// EnableTable enables caching for a specific table.
// EnableTable 为特定表启用缓存。
func (m *CacheManager) EnableTable(table string) {
	m.tables[table] = true
}

// DisableTable disables caching for a specific table.
// DisableTable 为特定表禁用缓存。
func (m *CacheManager) DisableTable(table string) {
	delete(m.tables, table)
}

// ShouldCache checks if a query should be cached.
// ShouldCache 检查查询是否应该被缓存。
func (m *CacheManager) ShouldCache(query *Query) bool {
	if !m.enabled {
		return false
	}

	// Only cache read queries
	// 只缓存读取查询
	if query.Action != ActionFind && query.Action != ActionCount {
		return false
	}

	// Check if table is in the cache list
	// 检查表是否在缓存列表中
	if len(m.tables) > 0 {
		return m.tables[query.Table]
	}

	return true
}

// GenerateKey generates a cache key for a query.
// GenerateKey 为查询生成缓存键。
func (m *CacheManager) GenerateKey(query *Query) string {
	data, _ := json.Marshal(query)
	hash := md5.Sum(data)
	return query.Table + ":" + hex.EncodeToString(hash[:])
}

// Get retrieves a cached result.
// Get 获取缓存的结果。
func (m *CacheManager) Get(query *Query) (*Result, bool) {
	if !m.ShouldCache(query) {
		return nil, false
	}

	key := m.GenerateKey(query)
	return m.cache.Get(key)
}

// Set stores a result in the cache.
// Set 将结果存储到缓存中。
func (m *CacheManager) Set(query *Query, result *Result) {
	if !m.ShouldCache(query) {
		return
	}

	key := m.GenerateKey(query)
	if mc, ok := m.cache.(*MemoryCache); ok {
		mc.SetWithTable(key, query.Table, result, m.ttl)
	} else {
		m.cache.Set(key, result, m.ttl)
	}
}

// Invalidate invalidates cache for a table.
// Invalidate 使表的缓存失效。
func (m *CacheManager) Invalidate(table string) {
	m.cache.Clear(table)
}

// InvalidateAll invalidates all cache.
// InvalidateAll 使所有缓存失效。
func (m *CacheManager) InvalidateAll() {
	m.cache.ClearAll()
}

// SetCache sets a custom cache implementation.
// SetCache 设置自定义缓存实现。
func (m *CacheManager) SetCache(cache Cache) {
	m.cache = cache
}
