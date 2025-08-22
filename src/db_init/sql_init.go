package dbinit

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TableInfo contains metadata about a database table structure
type TableInfo struct {
	TableName string
	Columns   []ColumnInfo
}

// ColumnInfo contains metadata about a database column
// Name and DBType are required
// Nullable, Primary, and Unique are optional
type ColumnInfo struct {
	Name     string
	DBType   string
	Nullable *bool
	Primary  *bool
	Unique   *bool
}

type SpecialType struct {
	TypeName string
	Columns  []ColumnInfo
}

// GetTableInfo extracts database table information from a struct using reflection
// This function reads the struct tags and converts them to database metadata
func GetTableInfo(structType interface{}, tableName string) (*TableInfo, error) {
	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v", t.Kind())
	}

	tableInfo := &TableInfo{
		TableName: tableName,
		Columns:   make([]ColumnInfo, 0),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Read struct tags
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue // Skip fields without db tag or explicitly ignored
		}

		dbType := field.Tag.Get("dbtype")
		if dbType == "" {
			return nil, fmt.Errorf("missing dbtype tag for field %s", field.Name)
		}

		nullable := field.Tag.Get("nullable") != "false" // default to nullable unless explicitly false
		primary := field.Tag.Get("primary") == "true"    // default to not primary unless explicitly true
		unique := field.Tag.Get("unique") == "false"     // default to not unique unless explicitly true

		columnInfo := ColumnInfo{
			Name:     dbTag,
			DBType:   dbType,
			Nullable: &nullable,
			Primary:  &primary,
			Unique:   &unique,
		}

		tableInfo.Columns = append(tableInfo.Columns, columnInfo)
	}

	return tableInfo, nil
}

// GenerateCreateTableSQL generates a PostgreSQL CREATE TABLE statement from struct metadata
func GenerateCreateTableSQL(tableInfo *TableInfo) string {
	var columns []string
	var primaryKeys []string
	var uniqueKeys []string

	// Generate column definitions
	for _, col := range tableInfo.Columns {
		columnDef := fmt.Sprintf("%s %s", col.Name, col.DBType)

		if col.Nullable != nil && !*col.Nullable {
			columnDef += " NOT NULL"
		} else if col.Nullable != nil && *col.Nullable == false {
			columnDef += " NULL"
		} else {
			// skip because it's not set
			continue
		}

		if col.Primary != nil && *col.Primary {
			primaryKeys = append(primaryKeys, col.Name)
		}

		if col.Unique != nil && *col.Unique {
			uniqueKeys = append(uniqueKeys, col.Name)
		}

		columns = append(columns, columnDef)
	}

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n    %s",
		tableInfo.TableName,
		strings.Join(columns, ",\n    "))

	if len(primaryKeys) > 0 {
		sql += fmt.Sprintf(",\n    PRIMARY KEY (%s)", strings.Join(primaryKeys, ", "))
	}

	if len(uniqueKeys) > 0 {
		sql += fmt.Sprintf(",\n    UNIQUE (%s)", strings.Join(uniqueKeys, ", "))
	}

	sql += "\n);"

	return sql
}

func GenerateSpecialTypeSQL(specialType *SpecialType) string {
	var columns []string

	for _, col := range specialType.Columns {
		columns = append(columns, fmt.Sprintf("%s %s", col.Name, col.DBType))
	}

	sql := fmt.Sprintf("CREATE TYPE IF NOT EXISTS %s AS (\n    %s",
		specialType.TypeName,
		strings.Join(columns, ",\n    "))

	return sql
}

// CreateTableSQL creates a table in the database based on struct metadata
func (t *TableInfo) CreateTableSQL() string {
	return GenerateCreateTableSQL(t)
	// TODO: add timescaledb hypertable creation and partition,compression logic setup and run the type creation sql first
}

// DBInitializer handles database initialization tasks
type DBInitializer struct {
	pool *pgxpool.Pool
}

// NewDBInitializer creates a new database initializer
func NewDBInitializer(pool *pgxpool.Pool) *DBInitializer {
	return &DBInitializer{pool: pool}
}

// CreateTableFromStruct creates a database table based on struct tags
func (db *DBInitializer) CreateTableFromStruct(structType interface{}, tableName string) error {
	tableInfo, err := GetTableInfo(structType, tableName)
	if err != nil {
		return fmt.Errorf("failed to get table info: %w", err)
	}

	sql := tableInfo.CreateTableSQL()

	_, err = db.pool.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	return nil
}
