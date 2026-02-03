package goorm

import (
	"errors"
	"testing"
)

// TestHookManager tests the hook manager.
// TestHookManager 测试钩子管理器。
func TestHookManager(t *testing.T) {
	m := NewHookManager()

	callCount := 0
	m.Register("users", HookBeforeCreate, func(ctx *HookContext) error {
		callCount++
		return nil
	})

	ctx := &HookContext{Table: "users"}
	if err := m.Execute(ctx, HookBeforeCreate); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected hook to be called once, got %d", callCount)
	}

	// Should not call hook for different table
	// 不应该为不同的表调用钩子
	ctx = &HookContext{Table: "orders"}
	if err := m.Execute(ctx, HookBeforeCreate); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("hook should not be called for different table")
	}
}

// TestGlobalHook tests global hooks.
// TestGlobalHook 测试全局钩子。
func TestGlobalHook(t *testing.T) {
	m := NewHookManager()

	callCount := 0
	m.RegisterGlobal(HookBeforeCreate, func(ctx *HookContext) error {
		callCount++
		return nil
	})

	// Should be called for any table
	// 应该为任何表调用
	ctx := &HookContext{Table: "users"}
	m.Execute(ctx, HookBeforeCreate)

	ctx = &HookContext{Table: "orders"}
	m.Execute(ctx, HookBeforeCreate)

	if callCount != 2 {
		t.Errorf("expected global hook to be called twice, got %d", callCount)
	}
}

// TestHookError tests hook error handling.
// TestHookError 测试钩子错误处理。
func TestHookError(t *testing.T) {
	m := NewHookManager()

	expectedErr := errors.New("hook error")
	m.Register("users", HookBeforeCreate, func(ctx *HookContext) error {
		return expectedErr
	})

	ctx := &HookContext{Table: "users"}
	err := m.Execute(ctx, HookBeforeCreate)

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

// TestHookSkip tests hook skip functionality.
// TestHookSkip 测试钩子跳过功能。
func TestHookSkip(t *testing.T) {
	m := NewHookManager()

	hook1Called := false
	hook2Called := false

	m.Register("users", HookBeforeCreate, func(ctx *HookContext) error {
		hook1Called = true
		ctx.Skip = true
		return nil
	})

	m.Register("users", HookBeforeCreate, func(ctx *HookContext) error {
		hook2Called = true
		return nil
	})

	ctx := &HookContext{Table: "users"}
	m.Execute(ctx, HookBeforeCreate)

	if !hook1Called {
		t.Error("first hook should be called")
	}
	if hook2Called {
		t.Error("second hook should be skipped")
	}
}

// TestTimestampHook tests the timestamp hook.
// TestTimestampHook 测试时间戳钩子。
func TestTimestampHook(t *testing.T) {
	config := NamingConfig{
		CreatedAtField: "created_at",
		UpdatedAtField: "updated_at",
	}

	hook := TimestampHook(config)

	// Test create action
	// 测试创建操作
	ctx := &HookContext{
		Action: ActionCreate,
		Data:   make(map[string]any),
	}

	if err := hook(ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if _, exists := ctx.Data["created_at"]; !exists {
		t.Error("created_at should be set")
	}
	if _, exists := ctx.Data["updated_at"]; !exists {
		t.Error("updated_at should be set")
	}

	// Test update action
	// 测试更新操作
	ctx = &HookContext{
		Action: ActionUpdate,
		Data:   make(map[string]any),
	}

	if err := hook(ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if _, exists := ctx.Data["updated_at"]; !exists {
		t.Error("updated_at should be set on update")
	}
}

// TestSoftDeleteHook tests the soft delete hook.
// TestSoftDeleteHook 测试软删除钩子。
func TestSoftDeleteHook(t *testing.T) {
	hook := SoftDeleteHook("deleted_at")

	query := &Query{
		Table:  "users",
		Action: ActionDelete,
		Where:  []Condition{{Field: "id", Op: OpEqual, Value: 1}},
	}

	ctx := &HookContext{
		Action: ActionDelete,
		Query:  query,
	}

	if err := hook(ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Action should be changed to update
	// Action 应该被改为更新
	if ctx.Query.Action != ActionUpdate {
		t.Errorf("expected action to be update, got %s", ctx.Query.Action)
	}

	// Data should have deleted_at
	// Data 应该有 deleted_at
	if _, exists := ctx.Query.Data["deleted_at"]; !exists {
		t.Error("deleted_at should be set in query data")
	}
}

// TestValidationHook tests the validation hook.
// TestValidationHook 测试验证钩子。
func TestValidationHook(t *testing.T) {
	validators := map[string]FieldValidator{
		"email": func(value any) error {
			email, ok := value.(string)
			if !ok || email == "" {
				return errors.New("email is required")
			}
			return nil
		},
	}

	hook := ValidationHook(validators)

	// Valid data
	// 有效数据
	ctx := &HookContext{
		Data: map[string]any{
			"email": "test@example.com",
		},
	}

	if err := hook(ctx); err != nil {
		t.Errorf("unexpected error for valid data: %v", err)
	}

	// Invalid data
	// 无效数据
	ctx = &HookContext{
		Data: map[string]any{
			"email": "",
		},
	}

	if err := hook(ctx); err == nil {
		t.Error("expected error for invalid data")
	}
}

// TestHookOrder tests that hooks are executed in order.
// TestHookOrder 测试钩子按顺序执行。
func TestHookOrder(t *testing.T) {
	m := NewHookManager()

	order := []int{}

	m.RegisterGlobal(HookBeforeCreate, func(ctx *HookContext) error {
		order = append(order, 1)
		return nil
	})

	m.Register("users", HookBeforeCreate, func(ctx *HookContext) error {
		order = append(order, 2)
		return nil
	})

	m.Register("users", HookBeforeCreate, func(ctx *HookContext) error {
		order = append(order, 3)
		return nil
	})

	ctx := &HookContext{Table: "users"}
	m.Execute(ctx, HookBeforeCreate)

	if len(order) != 3 {
		t.Errorf("expected 3 hooks to be called, got %d", len(order))
	}

	// Global hooks should run first
	// 全局钩子应该首先运行
	if order[0] != 1 {
		t.Error("global hook should run first")
	}

	// Table hooks should run in registration order
	// 表钩子应该按注册顺序运行
	if order[1] != 2 || order[2] != 3 {
		t.Error("table hooks should run in order")
	}
}
