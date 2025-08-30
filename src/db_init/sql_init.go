package dbinit

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
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

// Special Type is a postgres type that is to be used for a table
//
// It is used to create a type
type SpecialType struct {
	TypeName string
	Columns  []ColumnInfo
}

// GetTableInfo extracts database table information from a struct using reflection
// This function reads the struct tags and converts them to database metadata
//
// Args:
// - structType: the struct type to get the table info from
// - tableName: the name of the table to get the table info from
//
// Returns:
// - TableInfo: the table info for the table
// - error: if the function fails
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
			// stop the program here if the field is not set
			//
			// this function should always receive the sql data types
			// which should have the db tag set
			//
			// if the function receives a struct that is not a sql data type
			// kill the program here explicitly
			// TODO the cmd that will handle this will need to check if the db table already exists
			panic(fmt.Sprintf(
				`field %s is not set, to proceed with the program you need to set the db tag,
				this panic is related to %s table`,
				field.Name, tableName))
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
//
// Args:
// - tableInfo: the table info for the table to create
//
// Returns:
// - string: the SQL for the table
//
// The function will generate a SQL statement for a table based on the struct tags
// and the column info for the table
// The SQL statement will be in the form of CREATE TABLE IF NOT EXISTS <tableName>
// (<column1> <column1Type>, <column2> <column2Type>, ...)
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
		}
		// coment this out for now
		// better to have a explicit null value than to have a implicit null value
		//  else {
		// 	// skip because it's not set
		// 	continue
		// }

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

// GenerateCreateHypertableSQL generates a PostgreSQL CREATE TABLE statement with modern TimescaleDB hypertable syntax
//
// Args:
// - tableInfo: the table info for the table to create
// - chunkInterval: the interval to chunk the table by
// - partitionColumn: the column to partition the table by
//
// Returns:
// - string: the SQL for the hypertable
//
// The function will generate a SQL statement for a hypertable based on the struct tags
// and the column info for the hypertable
// The SQL statement will be in the form of CREATE TABLE IF NOT EXISTS <tableName>
// (<column1> <column1Type>, <column2> <column2Type>, ...)
// WITH (tsdb.hypertable, tsdb.partition_column='<partitionColumn>', tsdb.chunk_interval='<chunkInterval>')
func GenerateCreateHypertableSQL(tableInfo *TableInfo, chunkInterval string, partitionColumn string) string {
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

	// Add modern TimescaleDB hypertable configuration
	sql += fmt.Sprintf("\n) WITH (\n    tsdb.hypertable,\n    tsdb.partition_column='%s',\n    tsdb.chunk_interval='%s'\n);",
		partitionColumn, chunkInterval)

	return sql
}

// GenerateSpecialTypeSQL generates a PostgreSQL CREATE TYPE statement from struct metadata
//
// Args:
// - specialType: the special type to generate the SQL for
//
// Returns:
// - string: the SQL for the special type
//
// The function will generate a SQL statement for a special type based on the struct tags
// and the column info for the special type
// The SQL statement will be in the form of CREATE TYPE <typeName> AS
// (<column1> <column1Type>, <column2> <column2Type>, ...)
func GenerateSpecialTypeSQL(specialType *SpecialType) string {
	var columns []string

	for _, col := range specialType.Columns {
		columns = append(columns, fmt.Sprintf("%s %s", col.Name, col.DBType))
	}

	sql := fmt.Sprintf("CREATE TYPE %s AS (\n    %s\n);",
		specialType.TypeName,
		strings.Join(columns, ",\n    "))

	return sql
}

// CreateTableSQL creates a table in the database based on struct metadata
//
// Args:
// - t: the table info for the table to create
//
// Returns:
// - string: the SQL for the table
//
// The function will generate a SQL statement for a table based on the struct tags
// and the column info for the table
// The SQL statement will be in the form of CREATE TABLE IF NOT EXISTS <tableName>
// (<column1> <column1Type>, <column2> <column2Type>, ...)
func (t *TableInfo) CreateTableSQL() string {
	return GenerateCreateTableSQL(t)
}

// DBInitializer handles database initialization tasks
type DBInitializer struct {
	pool *pgxpool.Pool
}

// NewDBInitializer creates a new database initializer
func NewDBInitializer(pool *pgxpool.Pool) *DBInitializer {
	return &DBInitializer{pool: pool}
}

// TimescaleDBVersion represents a TimescaleDB version
type TimescaleDBVersion struct {
	Major int
	Minor int
	Patch int
}

// IsModernVersion returns true if this version supports the modern WITH syntax (2.19.3+)
func (v TimescaleDBVersion) IsModernVersion() bool {
	// Compare versions properly: 2.19.3+
	if v.Major > 2 {
		return true
	}
	if v.Major == 2 {
		if v.Minor > 19 {
			return true
		}
		if v.Minor == 19 && v.Patch >= 3 {
			return true
		}
	}
	return false
}

// GetTimescaleDBVersion detects the TimescaleDB version
func (db *DBInitializer) GetTimescaleDBVersion() (*TimescaleDBVersion, error) {
	var versionStr string
	err := db.pool.QueryRow(context.Background(), "SELECT extversion FROM pg_extension WHERE extname = 'timescaledb'").Scan(&versionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get TimescaleDB version: %w", err)
	}

	// Parse version string like "2.14.2" or "2.19.3"
	parts := strings.Split(versionStr, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch := 0
	if len(parts) >= 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			// Some versions might have additional suffixes, just use 0 for patch
			patch = 0
		}
	}

	return &TimescaleDBVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// CreateTableFromStruct creates a database table based on struct tags
//
// Args:
// - structType: the struct type to create the table from
// - tableName: the name of the table to create
//
// Returns:
// - nil: if the function is successful
// - error: if the function fails
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

// GetSpecialTypeInfo extracts database type information from a struct using reflection
func GetSpecialTypeInfo(structType interface{}, typeName string) (*SpecialType, error) {
	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v", t.Kind())
	}

	specialType := &SpecialType{
		TypeName: typeName,
		Columns:  make([]ColumnInfo, 0),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			return nil, fmt.Errorf("field %s is not set, to proceed with the program you need to set the db tag, this panic is related to %s type", field.Name, typeName)
		}

		dbType := field.Tag.Get("dbtype")
		if dbType == "" {
			return nil, fmt.Errorf("missing dbtype tag for field %s", field.Name)
		}

		columnInfo := ColumnInfo{
			Name:   dbTag,
			DBType: dbType,
		}

		specialType.Columns = append(specialType.Columns, columnInfo)
	}

	return specialType, nil
}

// CreateSpecialTypeFromStruct creates a database type based on struct tags
//
// Args:
// - structType: the struct type to create the special type from
// - typeName: the name of the special type to create
//
// Returns:
// - nil: if the function is successful
// - error: if the function fails
func (db *DBInitializer) CreateSpecialTypeFromStruct(structType interface{}, typeName string) error {
	specialType, err := GetSpecialTypeInfo(structType, typeName)
	if err != nil {
		return fmt.Errorf("failed to get special type info: %w", err)
	}

	sql := GenerateSpecialTypeSQL(specialType)

	_, err = db.pool.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("failed to create special type %s: %w", typeName, err)
	}

	return nil
}

// CreateHypertableFromStruct creates a hypertable using the appropriate method based on TimescaleDB version
//
// Args:
// - structType: the struct type to create the hypertable from
// - tableName: the name of the table to create the hypertable from
// - partitionColumn: the column to partition the table by
// - chunkInterval: the interval to chunk the table by
//
// Returns:
// - nil: if the function is successful
// - error: if the function fails
func (db *DBInitializer) CreateHypertableFromStruct(structType interface{}, tableName, partitionColumn, chunkInterval string) error {
	// First get the table info
	tableInfo, err := GetTableInfo(structType, tableName)
	if err != nil {
		return fmt.Errorf("failed to get table info: %w", err)
	}

	// Try to detect TimescaleDB version
	version, err := db.GetTimescaleDBVersion()
	if err != nil {
		// If we can't detect version, fall back to legacy method
		log.Printf("Warning: Could not detect TimescaleDB version (%v), using legacy method", err)
		return db.createHypertableLegacy(tableInfo, partitionColumn, chunkInterval)
	}

	// Use modern syntax for 2.19.3+, legacy for older versions
	if version.IsModernVersion() {
		log.Printf("Using modern hypertable syntax for TimescaleDB %d.%d.%d", version.Major, version.Minor, version.Patch)
		return db.createHypertableModern(tableInfo, partitionColumn, chunkInterval)
	} else {
		log.Printf("Using legacy hypertable syntax for TimescaleDB %d.%d.%d", version.Major, version.Minor, version.Patch)
		return db.createHypertableLegacy(tableInfo, partitionColumn, chunkInterval)
	}
}

// createHypertableModern creates a hypertable using the modern WITH syntax
func (db *DBInitializer) createHypertableModern(tableInfo *TableInfo, partitionColumn, chunkInterval string) error {
	sql := GenerateCreateHypertableSQL(tableInfo, chunkInterval, partitionColumn)

	_, err := db.pool.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("failed to create hypertable %s using modern syntax: %w", tableInfo.TableName, err)
	}

	log.Printf("Successfully created hypertable: %s", tableInfo.TableName)
	return nil
}

// createHypertableLegacy creates a hypertable using the legacy 3-step process
//
// Args:
// - tableInfo: the table info for the table to create
// - partitionColumn: the column to partition the table by
// - chunkInterval: the interval to chunk the table by
//
// Returns:
// - nil: if the function is successful
// - error: if the function fails
func (db *DBInitializer) createHypertableLegacy(tableInfo *TableInfo, partitionColumn, chunkInterval string) error {
	// Step 1: Create regular table
	sql := tableInfo.CreateTableSQL()
	_, err := db.pool.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableInfo.TableName, err)
	}

	// Step 2: Convert to hypertable
	hypertableSQL := fmt.Sprintf("SELECT create_hypertable('%s', '%s', chunk_time_interval => INTERVAL '%s')",
		tableInfo.TableName, partitionColumn, chunkInterval)
	_, err = db.pool.Exec(context.Background(), hypertableSQL)
	if err != nil {
		return fmt.Errorf("failed to convert table %s to hypertable: %w", tableInfo.TableName, err)
	}

	log.Printf("Successfully created hypertable (legacy): %s", tableInfo.TableName)
	return nil
}

// CreateTypeEnum creates a type enum in the database
// This function should be used to create type enums, for now only one enum is created at a time
func (db *DBInitializer) CreateChainTypeEnum(enumValues []string) error {
	sql := fmt.Sprintln("CREATE TYPE chain_name AS ENUM ()")
	_, err := db.pool.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("failed to create type enum chain_name: %w", err)
	}
	for _, enumValue := range enumValues {
		sql = fmt.Sprintf("ALTER TYPE chain_name ADD VALUE '%s'", enumValue)
		_, err = db.pool.Exec(context.Background(), sql)
		if err != nil {
			return fmt.Errorf("failed to add value %s to type enum chain_name: %w", enumValue, err)
		}
	}
	return nil
}
