package goorm

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

// HookType represents the type of hook.
// HookType 表示钩子的类型。
type HookType string

const (
	// HookBeforeCreate is called before creating a record.
	// HookBeforeCreate 在创建记录之前调用。
	HookBeforeCreate HookType = "before_create"

	// HookAfterCreate is called after creating a record.
	// HookAfterCreate 在创建记录之后调用。
	HookAfterCreate HookType = "after_create"

	// HookBeforeUpdate is called before updating records.
	// HookBeforeUpdate 在更新记录之前调用。
	HookBeforeUpdate HookType = "before_update"

	// HookAfterUpdate is called after updating records.
	// HookAfterUpdate 在更新记录之后调用。
	HookAfterUpdate HookType = "after_update"

	// HookBeforeDelete is called before deleting records.
	// HookBeforeDelete 在删除记录之前调用。
	HookBeforeDelete HookType = "before_delete"

	// HookAfterDelete is called after deleting records.
	// HookAfterDelete 在删除记录之后调用。
	HookAfterDelete HookType = "after_delete"

	// HookBeforeFind is called before querying records.
	// HookBeforeFind 在查询记录之前调用。
	HookBeforeFind HookType = "before_find"

	// HookAfterFind is called after querying records.
	// HookAfterFind 在查询记录之后调用。
	HookAfterFind HookType = "after_find"
)

// HookContext contains context for hook execution.
// HookContext 包含钩子执行的上下文。
type HookContext struct {
	// Context is the Go context.
	// Context 是 Go 上下文。
	Context context.Context

	// DB is the database instance.
	// DB 是数据库实例。
	DB *DB

	// Table is the table name.
	// Table 是表名。
	Table string

	// Action is the operation being performed.
	// Action 是正在执行的操作。
	Action Action

	// Query is the original query.
	// Query 是原始查询。
	Query *Query

	// Data is the data being created/updated.
	// Data 是正在创建/更新的数据。
	Data map[string]any

	// Result is the operation result (for after hooks).
	// Result 是操作结果（用于 after 钩子）。
	Result *Result

	// Skip indicates whether to skip the operation.
	// Skip 表示是否跳过操作。
	Skip bool

	// Error can be set to abort the operation.
	// Error 可以设置为中止操作。
	Error error
}

// HookFunc is the function signature for hooks.
// HookFunc 是钩子的函数签名。
type HookFunc func(ctx *HookContext) error

// HookManager manages hooks for database operations.
// HookManager 管理数据库操作的钩子。
type HookManager struct {
	// hooks maps table -> hook type -> handlers
	// hooks 映射 表 -> 钩子类型 -> 处理器
	hooks map[string]map[HookType][]HookFunc

	// globalHooks are applied to all tables
	// globalHooks 应用于所有表
	globalHooks map[HookType][]HookFunc
}

// NewHookManager creates a new hook manager.
// NewHookManager 创建新的钩子管理器。
func NewHookManager() *HookManager {
	return &HookManager{
		hooks:       make(map[string]map[HookType][]HookFunc),
		globalHooks: make(map[HookType][]HookFunc),
	}
}

// Register registers a hook for a specific table.
// Register 为特定表注册钩子。
func (m *HookManager) Register(table string, hookType HookType, fn HookFunc) {
	if m.hooks[table] == nil {
		m.hooks[table] = make(map[HookType][]HookFunc)
	}
	m.hooks[table][hookType] = append(m.hooks[table][hookType], fn)
}

// RegisterGlobal registers a global hook for all tables.
// RegisterGlobal 为所有表注册全局钩子。
func (m *HookManager) RegisterGlobal(hookType HookType, fn HookFunc) {
	m.globalHooks[hookType] = append(m.globalHooks[hookType], fn)
}

// Execute executes all hooks of the given type for the table.
// Execute 执行给定类型的所有钩子。
func (m *HookManager) Execute(ctx *HookContext, hookType HookType) error {
	// Execute global hooks first
	// 首先执行全局钩子
	for _, fn := range m.globalHooks[hookType] {
		if err := fn(ctx); err != nil {
			return err
		}
		if ctx.Skip || ctx.Error != nil {
			return ctx.Error
		}
	}

	// Execute table-specific hooks
	// 执行表特定的钩子
	if tableHooks, ok := m.hooks[ctx.Table]; ok {
		for _, fn := range tableHooks[hookType] {
			if err := fn(ctx); err != nil {
				return err
			}
			if ctx.Skip || ctx.Error != nil {
				return ctx.Error
			}
		}
	}

	return nil
}

// --- Model Interface Hooks ---
// --- 模型接口钩子 ---

// BeforeCreator is implemented by models that need before-create hooks.
// BeforeCreator 由需要创建前钩子的模型实现。
type BeforeCreator interface {
	BeforeCreate(ctx *HookContext) error
}

// AfterCreator is implemented by models that need after-create hooks.
// AfterCreator 由需要创建后钩子的模型实现。
type AfterCreator interface {
	AfterCreate(ctx *HookContext) error
}

// BeforeUpdater is implemented by models that need before-update hooks.
// BeforeUpdater 由需要更新前钩子的模型实现。
type BeforeUpdater interface {
	BeforeUpdate(ctx *HookContext) error
}

// AfterUpdater is implemented by models that need after-update hooks.
// AfterUpdater 由需要更新后钩子的模型实现。
type AfterUpdater interface {
	AfterUpdate(ctx *HookContext) error
}

// BeforeDeleter is implemented by models that need before-delete hooks.
// BeforeDeleter 由需要删除前钩子的模型实现。
type BeforeDeleter interface {
	BeforeDelete(ctx *HookContext) error
}

// AfterDeleter is implemented by models that need after-delete hooks.
// AfterDeleter 由需要删除后钩子的模型实现。
type AfterDeleter interface {
	AfterDelete(ctx *HookContext) error
}

// BeforeFinder is implemented by models that need before-find hooks.
// BeforeFinder 由需要查询前钩子的模型实现。
type BeforeFinder interface {
	BeforeFind(ctx *HookContext) error
}

// AfterFinder is implemented by models that need after-find hooks.
// AfterFinder 由需要查询后钩子的模型实现。
type AfterFinder interface {
	AfterFind(ctx *HookContext) error
}

// RegisterModelHooks registers hooks from model interface methods.
// RegisterModelHooks 从模型接口方法注册钩子。
func (m *HookManager) RegisterModelHooks(model any, tableName string) {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		v = reflect.New(t)
	}

	modelInstance := v.Interface()

	if bc, ok := modelInstance.(BeforeCreator); ok {
		m.Register(tableName, HookBeforeCreate, func(ctx *HookContext) error {
			return bc.BeforeCreate(ctx)
		})
	}

	if ac, ok := modelInstance.(AfterCreator); ok {
		m.Register(tableName, HookAfterCreate, func(ctx *HookContext) error {
			return ac.AfterCreate(ctx)
		})
	}

	if bu, ok := modelInstance.(BeforeUpdater); ok {
		m.Register(tableName, HookBeforeUpdate, func(ctx *HookContext) error {
			return bu.BeforeUpdate(ctx)
		})
	}

	if au, ok := modelInstance.(AfterUpdater); ok {
		m.Register(tableName, HookAfterUpdate, func(ctx *HookContext) error {
			return au.AfterUpdate(ctx)
		})
	}

	if bd, ok := modelInstance.(BeforeDeleter); ok {
		m.Register(tableName, HookBeforeDelete, func(ctx *HookContext) error {
			return bd.BeforeDelete(ctx)
		})
	}

	if ad, ok := modelInstance.(AfterDeleter); ok {
		m.Register(tableName, HookAfterDelete, func(ctx *HookContext) error {
			return ad.AfterDelete(ctx)
		})
	}

	if bf, ok := modelInstance.(BeforeFinder); ok {
		m.Register(tableName, HookBeforeFind, func(ctx *HookContext) error {
			return bf.BeforeFind(ctx)
		})
	}

	if af, ok := modelInstance.(AfterFinder); ok {
		m.Register(tableName, HookAfterFind, func(ctx *HookContext) error {
			return af.AfterFind(ctx)
		})
	}
}

// --- Built-in Hooks ---
// --- 内置钩子 ---

// TimestampHook automatically sets created_at and updated_at.
// TimestampHook 自动设置 created_at 和 updated_at。
func TimestampHook(config NamingConfig) HookFunc {
	return func(ctx *HookContext) error {
		if ctx.Data == nil {
			ctx.Data = make(map[string]any)
		}

		now := time.Now()

		switch ctx.Action {
		case ActionCreate:
			// Set created_at if not present
			// 如果不存在则设置 created_at
			createdField := config.CreatedAtField
			if createdField == "" {
				createdField = "created_at"
			}
			if _, exists := ctx.Data[createdField]; !exists {
				ctx.Data[createdField] = now
			}

			// Set updated_at if not present
			// 如果不存在则设置 updated_at
			updatedField := config.UpdatedAtField
			if updatedField == "" {
				updatedField = "updated_at"
			}
			if _, exists := ctx.Data[updatedField]; !exists {
				ctx.Data[updatedField] = now
			}

		case ActionUpdate:
			// Always update updated_at
			// 总是更新 updated_at
			updatedField := config.UpdatedAtField
			if updatedField == "" {
				updatedField = "updated_at"
			}
			ctx.Data[updatedField] = now
		}

		return nil
	}
}

// SoftDeleteHook converts delete to update with deleted_at.
// SoftDeleteHook 将删除转换为带 deleted_at 的更新。
func SoftDeleteHook(deletedAtField string) HookFunc {
	if deletedAtField == "" {
		deletedAtField = "deleted_at"
	}

	return func(ctx *HookContext) error {
		if ctx.Action != ActionDelete {
			return nil
		}

		// Convert delete to update with deleted_at = NOW()
		// 将删除转换为 deleted_at = NOW() 的更新
		ctx.Query.Action = ActionUpdate
		ctx.Query.Data = map[string]any{
			deletedAtField: "NOW()",
		}
		ctx.Action = ActionUpdate

		return nil
	}
}

// AuditHook logs all database operations.
// AuditHook 记录所有数据库操作。
func AuditHook(logger Logger) HookFunc {
	return func(ctx *HookContext) error {
		if logger == nil {
			return nil
		}

		logger.Info("DB operation",
			"table", ctx.Table,
			"action", ctx.Action,
		)
		return nil
	}
}

// ValidationHook validates data before create/update.
// ValidationHook 在创建/更新之前验证数据。
func ValidationHook(validators map[string]FieldValidator) HookFunc {
	return func(ctx *HookContext) error {
		if ctx.Data == nil {
			return nil
		}

		for field, validator := range validators {
			if value, exists := ctx.Data[field]; exists {
				if err := validator(value); err != nil {
					return fmt.Errorf("validation failed for field %q: %w", field, err)
				}
			}
		}

		return nil
	}
}

// FieldValidator is a function that validates a field value.
// FieldValidator 是验证字段值的函数。
type FieldValidator func(value any) error
