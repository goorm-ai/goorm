package goorm

import (
	"testing"
)

// TestMigratorGenerateCreateTableSQL tests CREATE TABLE generation.
// TestMigratorGenerateCreateTableSQL 测试 CREATE TABLE 生成。
func TestMigratorGenerateCreateTableSQL(t *testing.T) {
	meta := &ModelMeta{
		TableName: "users",
		ModelName: "User",
		Fields: []*FieldMeta{
			{
				Name:          "ID",
				ColumnName:    "id",
				GoType:        "uint64",
				PrimaryKey:    true,
				AutoIncrement: true,
			},
			{
				Name:       "Name",
				ColumnName: "name",
				GoType:     "string",
				Tags:       map[string]string{"size": "100"},
			},
			{
				Name:       "Email",
				ColumnName: "email",
				GoType:     "string",
				Unique:     true,
			},
			{
				Name:       "Age",
				ColumnName: "age",
				GoType:     "int",
				Nullable:   true,
			},
		},
	}

	// Test PostgreSQL
	pgMigrator := &Migrator{dialect: &PostgresDialect{}}
	sql := pgMigrator.generateCreateTableSQL(meta)

	if !containsAll(sql, []string{"CREATE TABLE", `"users"`, `"id"`, `"name"`, "VARCHAR", "UNIQUE"}) {
		t.Errorf("PostgreSQL CREATE TABLE missing expected parts: %s", sql)
	}

	// Test MySQL
	mysqlMigrator := &Migrator{dialect: &MySQLDialect{}}
	sql = mysqlMigrator.generateCreateTableSQL(meta)

	if !containsAll(sql, []string{"CREATE TABLE", "`users`", "`id`", "`name`", "VARCHAR"}) {
		t.Errorf("MySQL CREATE TABLE missing expected parts: %s", sql)
	}
}

// TestMigratorGenerateAddColumnSQL tests ADD COLUMN generation.
// TestMigratorGenerateAddColumnSQL 测试 ADD COLUMN 生成。
func TestMigratorGenerateAddColumnSQL(t *testing.T) {
	field := &FieldMeta{
		Name:       "Phone",
		ColumnName: "phone",
		GoType:     "string",
		Tags:       map[string]string{"size": "20"},
	}

	pgMigrator := &Migrator{dialect: &PostgresDialect{}}
	sql := pgMigrator.generateAddColumnSQL("users", field)

	if !containsAll(sql, []string{"ALTER TABLE", `"users"`, "ADD COLUMN", `"phone"`, "VARCHAR"}) {
		t.Errorf("ADD COLUMN missing expected parts: %s", sql)
	}
}

// TestMigratorTypesCompatible tests type compatibility checking.
// TestMigratorTypesCompatible 测试类型兼容性检查。
func TestMigratorTypesCompatible(t *testing.T) {
	m := &Migrator{}

	tests := []struct {
		dbType    string
		modelType string
		expected  bool
	}{
		{"integer", "INTEGER", true},
		{"int4", "INTEGER", true},
		{"bigint", "BIGINT", true},
		{"character varying", "VARCHAR(255)", true},
		{"text", "VARCHAR(255)", true},
		{"boolean", "BOOLEAN", true},
		{"timestamp with time zone", "TIMESTAMP", true},
		{"bytea", "DATETIME", false}, // completely different types
	}

	for _, tt := range tests {
		t.Run(tt.dbType+"_"+tt.modelType, func(t *testing.T) {
			result := m.typesCompatible(tt.dbType, tt.modelType)
			if result != tt.expected {
				t.Errorf("typesCompatible(%q, %q) = %v, want %v",
					tt.dbType, tt.modelType, result, tt.expected)
			}
		})
	}
}

// TestMigratorIsSystemTable tests system table detection.
// TestMigratorIsSystemTable 测试系统表检测。
func TestMigratorIsSystemTable(t *testing.T) {
	m := &Migrator{}

	systemTables := []string{
		"_backup_users_20260103",
		"pg_stat_activity",
		"sql_features",
		"_goorm_migrations",
	}

	for _, table := range systemTables {
		if !m.isSystemTable(table) {
			t.Errorf("isSystemTable(%q) = false, want true", table)
		}
	}

	normalTables := []string{"users", "orders", "products"}
	for _, table := range normalTables {
		if m.isSystemTable(table) {
			t.Errorf("isSystemTable(%q) = true, want false", table)
		}
	}
}

// TestMigratorGenerateBackupName tests backup name generation.
// TestMigratorGenerateBackupName 测试备份名生成。
func TestMigratorGenerateBackupName(t *testing.T) {
	m := &Migrator{}

	// With column
	name := m.generateBackupName("users", "phone")
	if !containsHelper(name, "_backup_users_phone_") {
		t.Errorf("backup name should contain '_backup_users_phone_': %s", name)
	}

	// Without column
	name = m.generateBackupName("users", "")
	if !containsHelper(name, "_backup_users_") {
		t.Errorf("backup name should contain '_backup_users_': %s", name)
	}
}
