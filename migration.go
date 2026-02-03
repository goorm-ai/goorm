package goorm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MigrationAction represents the type of schema change.
// MigrationAction 表示架构变更的类型。
type MigrationAction string

const (
	MigrationActionCreateTable  MigrationAction = "CREATE_TABLE"
	MigrationActionDropTable    MigrationAction = "DROP_TABLE"
	MigrationActionAddColumn    MigrationAction = "ADD_COLUMN"
	MigrationActionDropColumn   MigrationAction = "DROP_COLUMN"
	MigrationActionModifyColumn MigrationAction = "MODIFY_COLUMN"
	MigrationActionAddIndex     MigrationAction = "ADD_INDEX"
	MigrationActionDropIndex    MigrationAction = "DROP_INDEX"
)

// MigrationChange represents a single schema change.
// MigrationChange 表示单个架构变更。
type MigrationChange struct {
	Action      MigrationAction
	Table       string
	Column      string
	OldType     string
	NewType     string
	SQL         string
	BackupTable string
	Destructive bool
}

// MigrationPlan represents a set of changes to apply.
// MigrationPlan 表示要应用的变更集。
type MigrationPlan struct {
	Changes   []MigrationChange
	Backups   []BackupInfo
	CreatedAt time.Time
}

// BackupInfo stores backup metadata.
// BackupInfo 存储备份元数据。
type BackupInfo struct {
	Name      string
	Table     string
	Column    string
	CreatedAt time.Time
	RowCount  int64
	SizeBytes int64
}

// Migrator handles database schema migrations.
// Migrator 处理数据库架构迁移。
type Migrator struct {
	db      *DB
	dialect Dialect
}

// NewMigrator creates a new migrator.
// NewMigrator 创建新的迁移器。
func NewMigrator(db *DB) *Migrator {
	return &Migrator{
		db:      db,
		dialect: db.dialect,
	}
}

// AutoSync synchronizes the database schema with registered models.
// In aggressive mode, it also removes columns/tables not in models.
//
// AutoSync 将数据库架构与已注册的模型同步。
// 在激进模式下，它还会删除模型中不存在的列/表。
func (m *Migrator) AutoSync(ctx context.Context) error {
	plan, err := m.Plan(ctx)
	if err != nil {
		return err
	}

	return m.Execute(ctx, plan)
}

// Plan generates a migration plan without executing.
// Plan 生成迁移计划但不执行。
func (m *Migrator) Plan(ctx context.Context) (*MigrationPlan, error) {
	plan := &MigrationPlan{
		Changes:   make([]MigrationChange, 0),
		Backups:   make([]BackupInfo, 0),
		CreatedAt: time.Now(),
	}

	// Get current database schema
	// 获取当前数据库架构
	dbTables, err := m.getDBTables(ctx)
	if err != nil {
		return nil, err
	}

	// Get registered models
	// 获取已注册的模型
	modelTables := m.db.registry.ListTables()

	// Build lookup maps
	// 构建查找映射
	dbTableMap := make(map[string]map[string]ColumnInfo)
	for _, table := range dbTables {
		dbTableMap[table.Name] = table.Columns
	}

	modelTableMap := make(map[string]*ModelMeta)
	for _, table := range modelTables {
		meta, _ := m.db.registry.Get(table.Name)
		modelTableMap[table.Name] = meta
	}

	// Find tables to create
	// 查找要创建的表
	for _, model := range modelTables {
		if _, exists := dbTableMap[model.Name]; !exists {
			meta, _ := m.db.registry.Get(model.Name)
			sql := m.generateCreateTableSQL(meta)
			plan.Changes = append(plan.Changes, MigrationChange{
				Action:      MigrationActionCreateTable,
				Table:       model.Name,
				SQL:         sql,
				Destructive: false,
			})
		}
	}

	// Find columns to add/modify
	// 查找要添加/修改的列
	for tableName, meta := range modelTableMap {
		dbCols, exists := dbTableMap[tableName]
		if !exists {
			continue // Table will be created
		}

		for _, field := range meta.Fields {
			dbCol, colExists := dbCols[field.ColumnName]
			if !colExists {
				// Add column
				// 添加列
				sql := m.generateAddColumnSQL(tableName, field)
				plan.Changes = append(plan.Changes, MigrationChange{
					Action:      MigrationActionAddColumn,
					Table:       tableName,
					Column:      field.ColumnName,
					NewType:     field.SQLType,
					SQL:         sql,
					Destructive: false,
				})
			} else {
				// Check if modification needed
				// 检查是否需要修改
				newType := m.dialect.GoTypeToSQL(field.GoType, field.Tags)
				if !m.typesCompatible(dbCol.Type, newType) {
					sql := m.generateModifyColumnSQL(tableName, field)
					plan.Changes = append(plan.Changes, MigrationChange{
						Action:      MigrationActionModifyColumn,
						Table:       tableName,
						Column:      field.ColumnName,
						OldType:     dbCol.Type,
						NewType:     newType,
						SQL:         sql,
						Destructive: true,
					})
				}
			}
		}
	}

	// Aggressive mode: find columns/tables to drop
	// 激进模式：查找要删除的列/表
	if m.db.config.Migration.Aggressive {
		// Find tables to drop
		for tableName := range dbTableMap {
			if _, exists := modelTableMap[tableName]; !exists {
				// Check if it's a system table
				if m.isSystemTable(tableName) {
					continue
				}
				plan.Changes = append(plan.Changes, MigrationChange{
					Action:      MigrationActionDropTable,
					Table:       tableName,
					SQL:         fmt.Sprintf("DROP TABLE %s", m.dialect.Quote(tableName)),
					Destructive: true,
				})
			}
		}

		// Find columns to drop
		for tableName, meta := range modelTableMap {
			dbCols, exists := dbTableMap[tableName]
			if !exists {
				continue
			}

			modelCols := make(map[string]bool)
			for _, field := range meta.Fields {
				modelCols[field.ColumnName] = true
			}

			for colName := range dbCols {
				if !modelCols[colName] {
					backupName := m.generateBackupName(tableName, colName)
					plan.Changes = append(plan.Changes, MigrationChange{
						Action:      MigrationActionDropColumn,
						Table:       tableName,
						Column:      colName,
						SQL:         m.generateDropColumnSQL(tableName, colName),
						BackupTable: backupName,
						Destructive: true,
					})
				}
			}
		}
	}

	return plan, nil
}

// Execute applies a migration plan.
// Execute 应用迁移计划。
func (m *Migrator) Execute(ctx context.Context, plan *MigrationPlan) error {
	// Group changes: non-destructive first
	// 分组变更：先执行非破坏性的
	var nonDestructive, destructive []MigrationChange
	for _, change := range plan.Changes {
		if change.Destructive {
			destructive = append(destructive, change)
		} else {
			nonDestructive = append(nonDestructive, change)
		}
	}

	// Execute non-destructive changes first
	// 先执行非破坏性变更
	for _, change := range nonDestructive {
		if err := m.executeChange(ctx, change); err != nil {
			return fmt.Errorf("failed to execute %s on %s: %w", change.Action, change.Table, err)
		}
	}

	// Create backups for destructive changes if enabled
	// 如果启用，为破坏性变更创建备份
	if m.db.config.Migration.AutoBackup {
		for _, change := range destructive {
			if change.BackupTable != "" {
				if err := m.createBackup(ctx, change); err != nil {
					return fmt.Errorf("failed to backup %s.%s: %w", change.Table, change.Column, err)
				}
			}
		}
	}

	// Execute destructive changes
	// 执行破坏性变更
	for _, change := range destructive {
		if err := m.executeChange(ctx, change); err != nil {
			return fmt.Errorf("failed to execute %s on %s: %w", change.Action, change.Table, err)
		}
	}

	return nil
}

// executeChange executes a single migration change.
// executeChange 执行单个迁移变更。
func (m *Migrator) executeChange(ctx context.Context, change MigrationChange) error {
	_, err := m.db.sqlDB.ExecContext(ctx, change.SQL)
	return err
}

// createBackup creates a backup before destructive changes.
// createBackup 在破坏性变更前创建备份。
func (m *Migrator) createBackup(ctx context.Context, change MigrationChange) error {
	var sql string
	if change.Column != "" {
		// Backup single column
		// 备份单列
		sql = fmt.Sprintf(
			"CREATE TABLE %s AS SELECT %s, %s FROM %s",
			m.dialect.Quote(change.BackupTable),
			m.dialect.Quote("id"),
			m.dialect.Quote(change.Column),
			m.dialect.Quote(change.Table),
		)
	} else {
		// Backup entire table
		// 备份整个表
		sql = fmt.Sprintf(
			"CREATE TABLE %s AS SELECT * FROM %s",
			m.dialect.Quote(change.BackupTable),
			m.dialect.Quote(change.Table),
		)
	}

	_, err := m.db.sqlDB.ExecContext(ctx, sql)
	return err
}

// generateCreateTableSQL generates CREATE TABLE SQL.
// generateCreateTableSQL 生成 CREATE TABLE SQL。
func (m *Migrator) generateCreateTableSQL(meta *ModelMeta) string {
	var sb strings.Builder

	sb.WriteString("CREATE TABLE ")
	sb.WriteString(m.dialect.Quote(meta.TableName))
	sb.WriteString(" (\n")

	columns := make([]string, 0, len(meta.Fields))
	for _, field := range meta.Fields {
		col := m.generateColumnDef(field)
		columns = append(columns, "  "+col)
	}

	sb.WriteString(strings.Join(columns, ",\n"))
	sb.WriteString("\n)")

	return sb.String()
}

// generateColumnDef generates a column definition.
// generateColumnDef 生成列定义。
func (m *Migrator) generateColumnDef(field *FieldMeta) string {
	var parts []string

	parts = append(parts, m.dialect.Quote(field.ColumnName))

	// Determine SQL type
	// 确定 SQL 类型
	sqlType := field.SQLType
	if sqlType == "" {
		sqlType = m.dialect.GoTypeToSQL(field.GoType, field.Tags)
	}

	// Handle auto-increment primary key
	// 处理自增主键
	if field.PrimaryKey && field.AutoIncrement {
		// SQLite requires: INTEGER PRIMARY KEY AUTOINCREMENT
		// PostgreSQL uses: SERIAL PRIMARY KEY
		// MySQL uses: INT AUTO_INCREMENT PRIMARY KEY
		// SQLite 需要：INTEGER PRIMARY KEY AUTOINCREMENT
		// PostgreSQL 使用：SERIAL PRIMARY KEY
		// MySQL 使用：INT AUTO_INCREMENT PRIMARY KEY
		switch m.dialect.Name() {
		case "sqlite", "sqlite3":
			parts = append(parts, "INTEGER PRIMARY KEY AUTOINCREMENT")
		default:
			parts = append(parts, m.dialect.AutoIncrementClause())
			parts = append(parts, "PRIMARY KEY")
		}
	} else {
		parts = append(parts, sqlType)

		if field.PrimaryKey {
			parts = append(parts, "PRIMARY KEY")
		}

		if !field.Nullable && !field.PrimaryKey {
			parts = append(parts, "NOT NULL")
		}

		if field.Unique {
			parts = append(parts, "UNIQUE")
		}

		if field.Default != "" {
			parts = append(parts, "DEFAULT "+field.Default)
		}
	}

	return strings.Join(parts, " ")
}

// generateAddColumnSQL generates ADD COLUMN SQL.
// generateAddColumnSQL 生成 ADD COLUMN SQL。
func (m *Migrator) generateAddColumnSQL(table string, field *FieldMeta) string {
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s",
		m.dialect.Quote(table),
		m.generateColumnDef(field),
	)
}

// generateModifyColumnSQL generates MODIFY COLUMN SQL.
// generateModifyColumnSQL 生成 MODIFY COLUMN SQL。
func (m *Migrator) generateModifyColumnSQL(table string, field *FieldMeta) string {
	sqlType := field.SQLType
	if sqlType == "" {
		sqlType = m.dialect.GoTypeToSQL(field.GoType, field.Tags)
	}

	// Different syntax for different databases
	// 不同数据库的语法不同
	switch m.dialect.Name() {
	case "postgres":
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s",
			m.dialect.Quote(table),
			m.dialect.Quote(field.ColumnName),
			sqlType,
		)
	case "mysql":
		return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s",
			m.dialect.Quote(table),
			m.generateColumnDef(field),
		)
	default:
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s",
			m.dialect.Quote(table),
			m.dialect.Quote(field.ColumnName),
			sqlType,
		)
	}
}

// generateDropColumnSQL generates DROP COLUMN SQL.
// generateDropColumnSQL 生成 DROP COLUMN SQL。
func (m *Migrator) generateDropColumnSQL(table, column string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s",
		m.dialect.Quote(table),
		m.dialect.Quote(column),
	)
}

// generateBackupName generates a backup table name.
// generateBackupName 生成备份表名。
func (m *Migrator) generateBackupName(table, column string) string {
	timestamp := time.Now().Format("20060102_150405")
	if column != "" {
		return fmt.Sprintf("_backup_%s_%s_%s", table, column, timestamp)
	}
	return fmt.Sprintf("_backup_%s_%s", table, timestamp)
}

// getDBTables gets the current database schema.
// getDBTables 获取当前数据库架构。
func (m *Migrator) getDBTables(ctx context.Context) ([]DBTable, error) {
	switch m.dialect.Name() {
	case "postgres":
		return m.getPostgresTables(ctx)
	case "mysql":
		return m.getMySQLTables(ctx)
	case "sqlite", "sqlite3":
		return m.getSQLiteTables(ctx)
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", m.dialect.Name())
	}
}

// getPostgresTables gets PostgreSQL database schema.
// getPostgresTables 获取 PostgreSQL 数据库架构。
func (m *Migrator) getPostgresTables(ctx context.Context) ([]DBTable, error) {
	query := `
		SELECT table_name, column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema = 'public'
		ORDER BY table_name, ordinal_position
	`
	return m.queryDBTables(ctx, query)
}

// getMySQLTables gets MySQL database schema.
// getMySQLTables 获取 MySQL 数据库架构。
func (m *Migrator) getMySQLTables(ctx context.Context) ([]DBTable, error) {
	query := `
		SELECT table_name, column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		ORDER BY table_name, ordinal_position
	`
	return m.queryDBTables(ctx, query)
}

// getSQLiteTables gets SQLite database schema.
// getSQLiteTables 获取 SQLite 数据库架构。
func (m *Migrator) getSQLiteTables(ctx context.Context) ([]DBTable, error) {
	// Get list of tables
	// 获取表列表
	rows, err := m.db.sqlDB.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return nil, err
	}

	var tableNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return nil, err
		}
		tableNames = append(tableNames, name)
	}
	rows.Close()

	tables := make([]DBTable, 0, len(tableNames))

	for _, tableName := range tableNames {
		// Get columns for each table using PRAGMA
		// 使用 PRAGMA 获取每个表的列
		colRows, err := m.db.sqlDB.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
		if err != nil {
			return nil, err
		}

		columns := make(map[string]ColumnInfo)
		for colRows.Next() {
			var cid int
			var name, colType string
			var notNull, pk int
			var dfltValue *string

			if err := colRows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
				colRows.Close()
				return nil, err
			}

			columns[name] = ColumnInfo{
				Name:     name,
				Type:     colType,
				Nullable: notNull == 0,
				Default:  dfltValue,
			}
		}
		colRows.Close()

		tables = append(tables, DBTable{
			Name:    tableName,
			Columns: columns,
		})
	}

	return tables, nil
}

// queryDBTables executes a schema query and returns table info.
// queryDBTables 执行架构查询并返回表信息。
func (m *Migrator) queryDBTables(ctx context.Context, query string) ([]DBTable, error) {
	rows, err := m.db.sqlDB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tablesMap := make(map[string]map[string]ColumnInfo)

	for rows.Next() {
		var tableName, colName, dataType, isNullable string
		var colDefault *string

		if err := rows.Scan(&tableName, &colName, &dataType, &isNullable, &colDefault); err != nil {
			return nil, err
		}

		if _, exists := tablesMap[tableName]; !exists {
			tablesMap[tableName] = make(map[string]ColumnInfo)
		}

		tablesMap[tableName][colName] = ColumnInfo{
			Name:     colName,
			Type:     dataType,
			Nullable: isNullable == "YES",
			Default:  colDefault,
		}
	}

	tables := make([]DBTable, 0, len(tablesMap))
	for name, columns := range tablesMap {
		tables = append(tables, DBTable{
			Name:    name,
			Columns: columns,
		})
	}

	return tables, nil
}

// DBTable represents a database table.
// DBTable 表示数据库表。
type DBTable struct {
	Name    string
	Columns map[string]ColumnInfo
}

// ColumnInfo represents a database column.
// ColumnInfo 表示数据库列。
type ColumnInfo struct {
	Name     string
	Type     string
	Nullable bool
	Default  *string
}

// typesCompatible checks if two SQL types are compatible.
// typesCompatible 检查两个 SQL 类型是否兼容。
func (m *Migrator) typesCompatible(dbType, modelType string) bool {
	// Normalize types for comparison
	// 规范化类型以进行比较
	dbType = strings.ToUpper(dbType)
	modelType = strings.ToUpper(modelType)

	// Handle common equivalences
	// 处理常见等价类型
	equivalences := map[string][]string{
		"INTEGER":   {"INT", "INT4", "SERIAL"},
		"BIGINT":    {"INT8", "BIGSERIAL"},
		"VARCHAR":   {"CHARACTER VARYING", "TEXT"},
		"BOOLEAN":   {"BOOL"},
		"TIMESTAMP": {"TIMESTAMP WITH TIME ZONE", "TIMESTAMPTZ"},
	}

	for key, values := range equivalences {
		if strings.Contains(dbType, key) || strings.Contains(modelType, key) {
			for _, v := range values {
				if strings.Contains(dbType, v) || strings.Contains(modelType, v) {
					return true
				}
			}
		}
	}

	return strings.HasPrefix(dbType, modelType) || strings.HasPrefix(modelType, dbType)
}

// isSystemTable checks if a table is a system table.
// isSystemTable 检查表是否为系统表。
func (m *Migrator) isSystemTable(name string) bool {
	systemPrefixes := []string{"_backup_", "pg_", "sql_", "_goorm_"}
	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
