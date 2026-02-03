package goorm

import (
	"time"
)

// Model is the base struct that all models should embed.
// It provides standard fields: ID, CreatedAt, UpdatedAt, DeletedAt.
//
// Model 是所有模型都应该嵌入的基础结构体。
// 它提供标准字段：ID、CreatedAt、UpdatedAt、DeletedAt。
//
// Example / 示例:
//
//	type User struct {
//	    goorm.Model
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
type Model struct {
	// ID is the primary key, auto-incremented.
	// ID 是主键，自动递增。
	ID uint64 `json:"id" goorm:"primaryKey;autoIncrement"`

	// CreatedAt is automatically set when the record is created.
	// CreatedAt 在记录创建时自动设置。
	CreatedAt time.Time `json:"created_at" goorm:"autoCreateTime"`

	// UpdatedAt is automatically updated when the record is modified.
	// UpdatedAt 在记录修改时自动更新。
	UpdatedAt time.Time `json:"updated_at" goorm:"autoUpdateTime"`

	// DeletedAt is set when the record is soft-deleted.
	// DeletedAt 在记录软删除时设置。
	DeletedAt *time.Time `json:"deleted_at,omitempty" goorm:"index;softDelete"`
}

// TableNamer is an interface for models that want to customize their table name.
// If a model implements this interface, the returned name will be used as the table name.
//
// TableNamer 是一个接口，用于自定义表名。
// 如果模型实现了此接口，返回的名称将被用作表名。
//
// Example / 示例:
//
//	type User struct {
//	    goorm.Model
//	    Name string
//	}
//
//	func (User) TableName() string {
//	    return "sys_users"
//	}
type TableNamer interface {
	TableName() string
}

// ModelDescriber is an interface for models that provide a description.
// The description is used for AI understanding and schema documentation.
//
// ModelDescriber 是一个接口，用于提供模型描述。
// 描述用于 AI 理解和 Schema 文档。
type ModelDescriber interface {
	ModelDescription() string
}
