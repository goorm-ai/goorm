package goorm

import (
	"testing"
	"time"
)

// TestMemoryCache tests the memory cache.
// TestMemoryCache 测试内存缓存。
func TestMemoryCache(t *testing.T) {
	cache := NewMemoryCache()

	// Test set and get
	// 测试设置和获取
	result := &Result{Success: true, Count: 10}
	cache.Set("test_key", result, time.Hour)

	got, found := cache.Get("test_key")
	if !found {
		t.Error("expected to find cached entry")
	}
	if got.Count != 10 {
		t.Errorf("expected count 10, got %d", got.Count)
	}

	// Test non-existent key
	// 测试不存在的键
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("should not find non-existent key")
	}
}

// TestMemoryCacheExpiry tests cache expiry.
// TestMemoryCacheExpiry 测试缓存过期。
func TestMemoryCacheExpiry(t *testing.T) {
	cache := NewMemoryCache()

	result := &Result{Success: true}
	cache.Set("expire_key", result, 10*time.Millisecond)

	// Should find immediately
	// 应该立即找到
	_, found := cache.Get("expire_key")
	if !found {
		t.Error("expected to find entry immediately")
	}

	// Wait for expiry
	// 等待过期
	time.Sleep(20 * time.Millisecond)

	_, found = cache.Get("expire_key")
	if found {
		t.Error("entry should have expired")
	}
}

// TestMemoryCacheDelete tests cache deletion.
// TestMemoryCacheDelete 测试缓存删除。
func TestMemoryCacheDelete(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("delete_key", &Result{Success: true}, time.Hour)
	cache.Delete("delete_key")

	_, found := cache.Get("delete_key")
	if found {
		t.Error("entry should have been deleted")
	}
}

// TestMemoryCacheClear tests clearing cache by table.
// TestMemoryCacheClear 测试按表清除缓存。
func TestMemoryCacheClear(t *testing.T) {
	cache := NewMemoryCache()

	cache.SetWithTable("users:1", "users", &Result{Success: true}, time.Hour)
	cache.SetWithTable("users:2", "users", &Result{Success: true}, time.Hour)
	cache.SetWithTable("orders:1", "orders", &Result{Success: true}, time.Hour)

	cache.Clear("users")

	_, found := cache.Get("users:1")
	if found {
		t.Error("users entries should be cleared")
	}

	_, found = cache.Get("orders:1")
	if !found {
		t.Error("orders entries should still exist")
	}
}

// TestMemoryCacheClearAll tests clearing all cache.
// TestMemoryCacheClearAll 测试清除所有缓存。
func TestMemoryCacheClearAll(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("key1", &Result{Success: true}, time.Hour)
	cache.Set("key2", &Result{Success: true}, time.Hour)

	cache.ClearAll()

	stats := cache.Stats()
	if stats.Entries != 0 {
		t.Errorf("expected 0 entries, got %d", stats.Entries)
	}
}

// TestCacheManager tests the cache manager.
// TestCacheManager 测试缓存管理器。
func TestCacheManager(t *testing.T) {
	m := NewCacheManager()

	// Should not cache when disabled
	// 禁用时不应该缓存
	query := &Query{Table: "users", Action: ActionFind}
	if m.ShouldCache(query) {
		t.Error("should not cache when disabled")
	}

	m.Enable()

	// Should cache find queries
	// 应该缓存查找查询
	if !m.ShouldCache(query) {
		t.Error("should cache find queries")
	}

	// Should not cache write queries
	// 不应该缓存写入查询
	writeQuery := &Query{Table: "users", Action: ActionCreate}
	if m.ShouldCache(writeQuery) {
		t.Error("should not cache write queries")
	}
}

// TestCacheManagerWithTables tests table-specific caching.
// TestCacheManagerWithTables 测试表特定的缓存。
func TestCacheManagerWithTables(t *testing.T) {
	m := NewCacheManager()
	m.Enable()
	m.EnableTable("users")

	usersQuery := &Query{Table: "users", Action: ActionFind}
	ordersQuery := &Query{Table: "orders", Action: ActionFind}

	if !m.ShouldCache(usersQuery) {
		t.Error("should cache users table")
	}

	if m.ShouldCache(ordersQuery) {
		t.Error("should not cache orders table")
	}
}

// TestCacheManagerGenerateKey tests key generation.
// TestCacheManagerGenerateKey 测试键生成。
func TestCacheManagerGenerateKey(t *testing.T) {
	m := NewCacheManager()

	query1 := &Query{Table: "users", Action: ActionFind, Limit: 10}
	query2 := &Query{Table: "users", Action: ActionFind, Limit: 10}
	query3 := &Query{Table: "users", Action: ActionFind, Limit: 20}

	key1 := m.GenerateKey(query1)
	key2 := m.GenerateKey(query2)
	key3 := m.GenerateKey(query3)

	// Same queries should have same keys
	// 相同的查询应该有相同的键
	if key1 != key2 {
		t.Error("same queries should have same cache key")
	}

	// Different queries should have different keys
	// 不同的查询应该有不同的键
	if key1 == key3 {
		t.Error("different queries should have different cache keys")
	}
}

// TestCacheManagerSetGet tests set and get operations.
// TestCacheManagerSetGet 测试设置和获取操作。
func TestCacheManagerSetGet(t *testing.T) {
	m := NewCacheManager()
	m.Enable()

	query := &Query{Table: "users", Action: ActionFind}
	result := &Result{Success: true, Count: 5}

	m.Set(query, result)

	got, found := m.Get(query)
	if !found {
		t.Error("should find cached result")
	}
	if got.Count != 5 {
		t.Errorf("expected count 5, got %d", got.Count)
	}
}

// TestCacheManagerInvalidate tests cache invalidation.
// TestCacheManagerInvalidate 测试缓存失效。
func TestCacheManagerInvalidate(t *testing.T) {
	m := NewCacheManager()
	m.Enable()

	query := &Query{Table: "users", Action: ActionFind}
	m.Set(query, &Result{Success: true})

	m.Invalidate("users")

	_, found := m.Get(query)
	if found {
		t.Error("cache should be invalidated")
	}
}
