package goorm

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Registry holds registered models and their metadata.
// It provides methods for model registration, lookup, and schema generation.
//
// Registry 保存已注册的模型及其元数据。
// 它提供模型注册、查找和 Schema 生成的方法。
type Registry struct {
	// models maps table name to model metadata
	// models 将表名映射到模型元数据
	models map[string]*ModelMeta

	// mu protects concurrent access
	// mu 保护并发访问
	mu sync.RWMutex
}

// ModelMeta contains metadata about a registered model.
// ModelMeta 包含已注册模型的元数据。
type ModelMeta struct {
	// Type is the reflect.Type of the model
	// Type 是模型的 reflect.Type
	Type reflect.Type

	// TableName is the database table name
	// TableName 是数据库表名
	TableName string

	// ModelName is the Go struct name
	// ModelName 是 Go 结构体名
	ModelName string

	// Description is the model description
	// Description 是模型描述
	Description string

	// Fields contains field metadata
	// Fields 包含字段元数据
	Fields []*FieldMeta

	// PrimaryKey is the primary key field
	// PrimaryKey 是主键字段
	PrimaryKey *FieldMeta

	// Indexes contains index definitions
	// Indexes 包含索引定义
	Indexes []IndexSchema

	// Relations contains relation definitions
	// Relations 包含关联定义
	Relations []RelationSchema
}

// FieldMeta contains metadata about a model field.
// FieldMeta 包含模型字段的元数据。
type FieldMeta struct {
	// Name is the Go field name
	// Name 是 Go 字段名
	Name string

	// ColumnName is the database column name
	// ColumnName 是数据库列名
	ColumnName string

	// Type is the reflect.Type of the field
	// Type 是字段的 reflect.Type
	Type reflect.Type

	// SQLType is the SQL type for this field
	// SQLType 是此字段的 SQL 类型
	SQLType string

	// GoType is the Go type string representation
	// GoType 是 Go 类型的字符串表示
	GoType string

	// Nullable indicates if the field can be NULL
	// Nullable 表示字段是否可为 NULL
	Nullable bool

	// PrimaryKey indicates if this is the primary key
	// PrimaryKey 表示是否为主键
	PrimaryKey bool

	// AutoIncrement indicates if the field auto-increments
	// AutoIncrement 表示字段是否自动递增
	AutoIncrement bool

	// Unique indicates if the field has a unique constraint
	// Unique 表示字段是否有唯一约束
	Unique bool

	// Index indicates if the field should be indexed
	// Index 表示字段是否应该被索引
	Index bool

	// Default is the default value
	// Default 是默认值
	Default string

	// Description is the field description
	// Description 是字段描述
	Description string

	// Sensitive indicates if this is a sensitive field
	// Sensitive 表示是否为敏感字段
	Sensitive bool

	// Mask is the masking strategy for sensitive fields
	// Mask 是敏感字段的脱敏策略
	Mask string

	// JSONName is the JSON field name
	// JSONName 是 JSON 字段名
	JSONName string

	// Tags contains all parsed struct tags
	// Tags 包含所有解析的结构体标签
	Tags map[string]string
}

// NewRegistry creates a new model registry.
// NewRegistry 创建新的模型注册表。
func NewRegistry() *Registry {
	return &Registry{
		models: make(map[string]*ModelMeta),
	}
}

// Register registers a model with the registry.
// Register 向注册表注册一个模型。
func (r *Registry) Register(model any, naming NamingConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a struct, got %s", t.Kind())
	}

	// Determine table name
	// 确定表名
	tableName := ""
	if tn, ok := model.(TableNamer); ok {
		tableName = tn.TableName()
	}
	if tableName == "" {
		// Check for table tag on embedded Model
		// 检查嵌入的 Model 上的 table 标签
		if field, ok := t.FieldByName("Model"); ok {
			if tag := field.Tag.Get("table"); tag != "" {
				tableName = tag
			}
		}
	}
	if tableName == "" && naming.TableNamer != nil {
		tableName = naming.TableNamer(t.Name())
	}
	if tableName == "" {
		tableName = SnakeCasePlural(t.Name())
	}

	// Add table prefix
	// 添加表前缀
	if naming.TablePrefix != "" {
		tableName = naming.TablePrefix + tableName
	}

	// Get description
	// 获取描述
	description := ""
	if md, ok := model.(ModelDescriber); ok {
		description = md.ModelDescription()
	}
	if description == "" {
		if field, ok := t.FieldByName("Model"); ok {
			description = field.Tag.Get("desc")
		}
	}

	meta := &ModelMeta{
		Type:        t,
		TableName:   tableName,
		ModelName:   t.Name(),
		Description: description,
		Fields:      make([]*FieldMeta, 0),
	}

	// Parse fields
	// 解析字段
	if err := r.parseFields(t, meta, naming); err != nil {
		return err
	}

	r.models[tableName] = meta
	return nil
}

// parseFields parses struct fields into field metadata.
// parseFields 将结构体字段解析为字段元数据。
func (r *Registry) parseFields(t reflect.Type, meta *ModelMeta, naming NamingConfig) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Handle embedded structs
		// 处理嵌入的结构体
		if field.Anonymous {
			if field.Type.Kind() == reflect.Struct {
				if err := r.parseFields(field.Type, meta, naming); err != nil {
					return err
				}
			}
			continue
		}

		// Skip unexported fields
		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		// Skip fields marked with "-"
		// 跳过标记为 "-" 的字段
		if field.Tag.Get("goorm") == "-" || field.Tag.Get("json") == "-" {
			continue
		}

		fieldMeta := r.parseField(field, naming)
		meta.Fields = append(meta.Fields, fieldMeta)

		if fieldMeta.PrimaryKey {
			meta.PrimaryKey = fieldMeta
		}
	}

	return nil
}

// parseField parses a single struct field.
// parseField 解析单个结构体字段。
func (r *Registry) parseField(field reflect.StructField, naming NamingConfig) *FieldMeta {
	fm := &FieldMeta{
		Name:   field.Name,
		Type:   field.Type,
		GoType: field.Type.String(),
		Tags:   make(map[string]string),
	}

	// Determine column name
	// 确定列名
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" && jsonTag != "-" {
		parts := strings.Split(jsonTag, ",")
		fm.JSONName = parts[0]
		fm.ColumnName = parts[0]
	}
	if fm.ColumnName == "" {
		if naming.ColumnNamer != nil {
			fm.ColumnName = naming.ColumnNamer(field.Name)
		} else {
			fm.ColumnName = SnakeCase(field.Name)
		}
	}

	// Parse goorm tag
	// 解析 goorm 标签
	goormTag := field.Tag.Get("goorm")
	if goormTag != "" {
		for _, part := range strings.Split(goormTag, ";") {
			kv := strings.SplitN(part, ":", 2)
			key := strings.TrimSpace(kv[0])
			value := ""
			if len(kv) == 2 {
				value = strings.TrimSpace(kv[1])
			}

			fm.Tags[key] = value

			switch strings.ToLower(key) {
			case "primarykey":
				fm.PrimaryKey = true
			case "autoincrement":
				fm.AutoIncrement = true
			case "unique":
				fm.Unique = true
			case "index":
				fm.Index = true
			case "default":
				fm.Default = value
			case "column":
				fm.ColumnName = value
			case "type":
				fm.SQLType = value
			}
		}
	}

	// Parse other tags
	// 解析其他标签
	fm.Description = field.Tag.Get("desc")
	fm.Sensitive = field.Tag.Get("sensitive") == "true"
	fm.Mask = field.Tag.Get("mask")

	if field.Tag.Get("unique") == "true" {
		fm.Unique = true
	}

	// Determine nullability
	// 确定可空性
	fm.Nullable = field.Type.Kind() == reflect.Ptr ||
		field.Type.Kind() == reflect.Slice ||
		field.Type.Kind() == reflect.Map

	return fm
}

// Get returns the model metadata for a table name.
// Get 返回表名的模型元数据。
func (r *Registry) Get(tableName string) (*ModelMeta, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	meta, ok := r.models[tableName]
	return meta, ok
}

// ListTables returns information about all registered tables.
// ListTables 返回所有已注册表的信息。
func (r *Registry) ListTables() []TableInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tables := make([]TableInfo, 0, len(r.models))
	for _, meta := range r.models {
		columns := make([]string, len(meta.Fields))
		for i, f := range meta.Fields {
			columns[i] = f.ColumnName
		}

		pk := "id"
		if meta.PrimaryKey != nil {
			pk = meta.PrimaryKey.ColumnName
		}

		tables = append(tables, TableInfo{
			Name:        meta.TableName,
			Model:       meta.ModelName,
			Description: meta.Description,
			Columns:     columns,
			PrimaryKey:  pk,
		})
	}

	return tables
}

// GetSchema returns the full schema for a table.
// GetSchema 返回表的完整 Schema。
func (r *Registry) GetSchema(tableName string) (*TableSchema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	meta, ok := r.models[tableName]
	if !ok {
		return nil, fmt.Errorf("table %q not found", tableName)
	}

	columns := make([]ColumnSchema, len(meta.Fields))
	for i, f := range meta.Fields {
		columns[i] = ColumnSchema{
			Name:        f.ColumnName,
			Type:        f.SQLType,
			GoType:      f.GoType,
			Nullable:    f.Nullable,
			Primary:     f.PrimaryKey,
			Unique:      f.Unique,
			Default:     f.Default,
			Description: f.Description,
			Sensitive:   f.Sensitive,
			Mask:        f.Mask,
		}
	}

	return &TableSchema{
		Table:       meta.TableName,
		Model:       meta.ModelName,
		Description: meta.Description,
		Columns:     columns,
		Indexes:     meta.Indexes,
		Relations:   meta.Relations,
	}, nil
}
