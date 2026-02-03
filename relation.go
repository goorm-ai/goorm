package goorm

import (
	"fmt"
	"reflect"
)

// RelationType represents the type of relationship between models.
// RelationType 表示模型之间关系的类型。
type RelationType string

const (
	// RelationHasOne represents a one-to-one relationship.
	// RelationHasOne 表示一对一关系。
	RelationHasOne RelationType = "has_one"

	// RelationHasMany represents a one-to-many relationship.
	// RelationHasMany 表示一对多关系。
	RelationHasMany RelationType = "has_many"

	// RelationBelongsTo represents an inverse one-to-one or one-to-many.
	// RelationBelongsTo 表示反向一对一或一对多关系。
	RelationBelongsTo RelationType = "belongs_to"

	// RelationManyToMany represents a many-to-many relationship.
	// RelationManyToMany 表示多对多关系。
	RelationManyToMany RelationType = "many_to_many"
)

// Relation represents a relationship between models.
// Relation 表示模型之间的关系。
type Relation struct {
	// Type is the relationship type.
	// Type 是关系类型。
	Type RelationType

	// Model is the related model.
	// Model 是关联的模型。
	Model string

	// ForeignKey is the foreign key column.
	// ForeignKey 是外键列。
	ForeignKey string

	// ReferenceKey is the referenced column (usually primary key).
	// ReferenceKey 是引用的列（通常是主键）。
	ReferenceKey string

	// Through is the junction table for many-to-many.
	// Through 是多对多关系的关联表。
	Through string

	// JoinTableForeignKey is the foreign key in junction table for this model.
	// JoinTableForeignKey 是关联表中此模型的外键。
	JoinTableForeignKey string

	// JoinTableReferenceKey is the foreign key in junction table for related model.
	// JoinTableReferenceKey 是关联表中关联模型的外键。
	JoinTableReferenceKey string
}

// RelationLoader handles eager loading of relations.
// RelationLoader 处理关联的预加载。
type RelationLoader struct {
	db       *DB
	registry *Registry
}

// NewRelationLoader creates a new relation loader.
// NewRelationLoader 创建新的关联加载器。
func NewRelationLoader(db *DB) *RelationLoader {
	return &RelationLoader{
		db:       db,
		registry: db.registry,
	}
}

// LoadRelations loads related data into the result set.
// LoadRelations 将关联数据加载到结果集中。
func (l *RelationLoader) LoadRelations(data []map[string]any, table string, with []any) error {
	if len(data) == 0 || len(with) == 0 {
		return nil
	}

	meta, ok := l.registry.Get(table)
	if !ok {
		return fmt.Errorf("table %q not found", table)
	}

	for _, w := range with {
		if err := l.loadRelation(data, meta, w); err != nil {
			return err
		}
	}

	return nil
}

// loadRelation loads a single relation.
// loadRelation 加载单个关联。
func (l *RelationLoader) loadRelation(data []map[string]any, meta *ModelMeta, with any) error {
	var relationName string
	var nestedWith []any

	// Parse the "with" specification
	// 解析 "with" 规格
	switch v := with.(type) {
	case string:
		relationName = v
	case map[string]any:
		for key, val := range v {
			relationName = key
			if nested, ok := val.([]any); ok {
				nestedWith = nested
			}
			break
		}
	default:
		return fmt.Errorf("invalid with specification: %v", with)
	}

	// Find the relation in meta
	// 在 meta 中查找关系
	var relation *RelationSchema
	for _, r := range meta.Relations {
		if r.Name == relationName {
			relation = &r
			break
		}
	}

	if relation == nil {
		return fmt.Errorf("relation %q not found on table %q", relationName, meta.TableName)
	}

	// Load based on relation type
	// 根据关系类型加载
	switch RelationType(relation.Type) {
	case RelationHasOne:
		return l.loadHasOne(data, relation, nestedWith)
	case RelationHasMany:
		return l.loadHasMany(data, relation, nestedWith)
	case RelationBelongsTo:
		return l.loadBelongsTo(data, relation, nestedWith)
	case RelationManyToMany:
		return l.loadManyToMany(data, relation, nestedWith)
	default:
		return fmt.Errorf("unsupported relation type: %s", relation.Type)
	}
}

// loadHasOne loads a has-one relation.
// loadHasOne 加载一对一关系。
func (l *RelationLoader) loadHasOne(data []map[string]any, rel *RelationSchema, nestedWith []any) error {
	// Collect all IDs
	// 收集所有 ID
	ids := make([]any, 0, len(data))
	idMap := make(map[any]int)
	for i, row := range data {
		if id, ok := row[rel.ReferenceKey]; ok {
			ids = append(ids, id)
			idMap[id] = i
		}
	}

	if len(ids) == 0 {
		return nil
	}

	// Query related records
	// 查询关联记录
	query := &Query{
		Table:  rel.Model,
		Action: ActionFind,
		Where: []Condition{
			{Field: rel.ForeignKey, Op: OpIn, Value: ids},
		},
	}

	result := l.db.ExecuteQuery(l.db.ctx, query)
	if !result.Success {
		return fmt.Errorf("failed to load relation: %s", result.Error.Message)
	}

	// Load nested relations
	// 加载嵌套关系
	if len(nestedWith) > 0 {
		if err := l.LoadRelations(result.Data, rel.Model, nestedWith); err != nil {
			return err
		}
	}

	// Map results back to parent
	// 将结果映射回父级
	for _, relRow := range result.Data {
		if fkVal, ok := relRow[rel.ForeignKey]; ok {
			if idx, exists := idMap[fkVal]; exists {
				data[idx][rel.Name] = relRow
			}
		}
	}

	return nil
}

// loadHasMany loads a has-many relation.
// loadHasMany 加载一对多关系。
func (l *RelationLoader) loadHasMany(data []map[string]any, rel *RelationSchema, nestedWith []any) error {
	// Collect all IDs
	// 收集所有 ID
	ids := make([]any, 0, len(data))
	for _, row := range data {
		if id, ok := row[rel.ReferenceKey]; ok {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return nil
	}

	// Query related records
	// 查询关联记录
	query := &Query{
		Table:  rel.Model,
		Action: ActionFind,
		Where: []Condition{
			{Field: rel.ForeignKey, Op: OpIn, Value: ids},
		},
	}

	result := l.db.ExecuteQuery(l.db.ctx, query)
	if !result.Success {
		return fmt.Errorf("failed to load relation: %s", result.Error.Message)
	}

	// Load nested relations
	// 加载嵌套关系
	if len(nestedWith) > 0 {
		if err := l.LoadRelations(result.Data, rel.Model, nestedWith); err != nil {
			return err
		}
	}

	// Build ID to parent index map
	// 构建 ID 到父索引的映射
	idToIdx := make(map[any]int)
	for i, row := range data {
		if id, ok := row[rel.ReferenceKey]; ok {
			idToIdx[id] = i
			// Initialize empty array
			data[i][rel.Name] = []map[string]any{}
		}
	}

	// Map results back to parent
	// 将结果映射回父级
	for _, relRow := range result.Data {
		if fkVal, ok := relRow[rel.ForeignKey]; ok {
			if idx, exists := idToIdx[fkVal]; exists {
				arr := data[idx][rel.Name].([]map[string]any)
				data[idx][rel.Name] = append(arr, relRow)
			}
		}
	}

	return nil
}

// loadBelongsTo loads a belongs-to relation.
// loadBelongsTo 加载属于关系。
func (l *RelationLoader) loadBelongsTo(data []map[string]any, rel *RelationSchema, nestedWith []any) error {
	// Collect all foreign key values
	// 收集所有外键值
	fkValues := make([]any, 0, len(data))
	fkMap := make(map[any][]int)

	for i, row := range data {
		if fkVal, ok := row[rel.ForeignKey]; ok && fkVal != nil {
			if _, exists := fkMap[fkVal]; !exists {
				fkValues = append(fkValues, fkVal)
			}
			fkMap[fkVal] = append(fkMap[fkVal], i)
		}
	}

	if len(fkValues) == 0 {
		return nil
	}

	// Query related records
	// 查询关联记录
	query := &Query{
		Table:  rel.Model,
		Action: ActionFind,
		Where: []Condition{
			{Field: rel.ReferenceKey, Op: OpIn, Value: fkValues},
		},
	}

	result := l.db.ExecuteQuery(l.db.ctx, query)
	if !result.Success {
		return fmt.Errorf("failed to load relation: %s", result.Error.Message)
	}

	// Load nested relations
	// 加载嵌套关系
	if len(nestedWith) > 0 {
		if err := l.LoadRelations(result.Data, rel.Model, nestedWith); err != nil {
			return err
		}
	}

	// Build reference map
	// 构建引用映射
	refMap := make(map[any]map[string]any)
	for _, relRow := range result.Data {
		if refVal, ok := relRow[rel.ReferenceKey]; ok {
			refMap[refVal] = relRow
		}
	}

	// Map results back to parent
	// 将结果映射回父级
	for fkVal, indices := range fkMap {
		if relRow, exists := refMap[fkVal]; exists {
			for _, idx := range indices {
				data[idx][rel.Name] = relRow
			}
		}
	}

	return nil
}

// loadManyToMany loads a many-to-many relation.
// loadManyToMany 加载多对多关系。
func (l *RelationLoader) loadManyToMany(data []map[string]any, rel *RelationSchema, nestedWith []any) error {
	// Collect all IDs
	// 收集所有 ID
	ids := make([]any, 0, len(data))
	idToIdx := make(map[any]int)

	for i, row := range data {
		if id, ok := row[rel.ReferenceKey]; ok {
			ids = append(ids, id)
			idToIdx[id] = i
			// Initialize empty array
			data[i][rel.Name] = []map[string]any{}
		}
	}

	if len(ids) == 0 {
		return nil
	}

	// Query junction table to get related IDs
	// 查询关联表获取关联 ID
	junctionQuery := &Query{
		Table:  rel.JoinTable,
		Action: ActionFind,
		Where: []Condition{
			{Field: rel.JoinForeignKey, Op: OpIn, Value: ids},
		},
	}

	junctionResult := l.db.ExecuteQuery(l.db.ctx, junctionQuery)
	if !junctionResult.Success {
		return fmt.Errorf("failed to load junction: %s", junctionResult.Error.Message)
	}

	// Build mapping and collect related IDs
	// 构建映射并收集关联 ID
	junctionMap := make(map[any][]any) // parent ID -> []related IDs
	relatedIDs := make([]any, 0)

	for _, jRow := range junctionResult.Data {
		parentID := jRow[rel.JoinForeignKey]
		relatedID := jRow[rel.JoinReferenceKey]
		junctionMap[parentID] = append(junctionMap[parentID], relatedID)
		relatedIDs = append(relatedIDs, relatedID)
	}

	if len(relatedIDs) == 0 {
		return nil
	}

	// Query related records
	// 查询关联记录
	relatedQuery := &Query{
		Table:  rel.Model,
		Action: ActionFind,
		Where: []Condition{
			{Field: rel.ReferenceKey, Op: OpIn, Value: relatedIDs},
		},
	}

	relatedResult := l.db.ExecuteQuery(l.db.ctx, relatedQuery)
	if !relatedResult.Success {
		return fmt.Errorf("failed to load related: %s", relatedResult.Error.Message)
	}

	// Load nested relations
	// 加载嵌套关系
	if len(nestedWith) > 0 {
		if err := l.LoadRelations(relatedResult.Data, rel.Model, nestedWith); err != nil {
			return err
		}
	}

	// Build related record map
	// 构建关联记录映射
	relatedMap := make(map[any]map[string]any)
	for _, relRow := range relatedResult.Data {
		if refVal, ok := relRow[rel.ReferenceKey]; ok {
			relatedMap[refVal] = relRow
		}
	}

	// Map results back to parent
	// 将结果映射回父级
	for parentID, relatedIDList := range junctionMap {
		if idx, exists := idToIdx[parentID]; exists {
			arr := data[idx][rel.Name].([]map[string]any)
			for _, relID := range relatedIDList {
				if relRow, found := relatedMap[relID]; found {
					arr = append(arr, relRow)
				}
			}
			data[idx][rel.Name] = arr
		}
	}

	return nil
}

// parseRelations parses relation definitions from struct tags.
// parseRelations 从结构体标签解析关联定义。
func parseRelations(t reflect.Type) []RelationSchema {
	relations := make([]RelationSchema, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Check for relation tag
		// 检查关联标签
		relTag := field.Tag.Get("rel")
		if relTag == "" {
			continue
		}

		rel := RelationSchema{
			Name:         field.Name,
			Type:         relTag,
			Model:        field.Tag.Get("model"),
			ForeignKey:   field.Tag.Get("fk"),
			ReferenceKey: field.Tag.Get("ref"),
		}

		// Default reference key
		// 默认引用键
		if rel.ReferenceKey == "" {
			rel.ReferenceKey = "id"
		}

		// Default foreign key based on relation type
		// 根据关系类型设置默认外键
		if rel.ForeignKey == "" {
			switch RelationType(relTag) {
			case RelationBelongsTo:
				// Model name + _id
				rel.ForeignKey = SnakeCase(rel.Model) + "_id"
			case RelationHasOne, RelationHasMany:
				// Current model name + _id
				rel.ForeignKey = SnakeCase(t.Name()) + "_id"
			}
		}

		// Many-to-many specific
		// 多对多特有
		if RelationType(relTag) == RelationManyToMany {
			rel.JoinTable = field.Tag.Get("through")
			rel.JoinForeignKey = field.Tag.Get("join_fk")
			rel.JoinReferenceKey = field.Tag.Get("join_ref")

			// Default join table
			// 默认关联表
			if rel.JoinTable == "" {
				names := []string{SnakeCase(t.Name()), SnakeCase(rel.Model)}
				if names[0] > names[1] {
					names[0], names[1] = names[1], names[0]
				}
				rel.JoinTable = names[0] + "_" + names[1]
			}
		}

		relations = append(relations, rel)
	}

	return relations
}
