package dbinit

import (
	"context"
	"fmt"
	"strings"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
)

var l = logger.Get()

// Hypertable management for TimescaleDB
//
// Recommended TimescaleDB versions:
// - Community Edition 2.23.0+ (TimescaleDB)
// - Cloud edition (Tiger Data)
//

// GenerateCreateHypertableSQL generates a PostgreSQL CREATE TABLE statement with modern TimescaleDB hypertable syntax
//
// Parameters:
// - tableInfo: the table info for the table to create
// - params: the parameters for the hypertable
//
// Returns:
// - string: the SQL for the hypertable
//
// The function will generate a SQL statement for a hypertable based on the struct tags
// and the column info for the hypertable
// The SQL statement will be in the form of CREATE TABLE IF NOT EXISTS <tableName>
// (<column1> <column1Type>, <column2> <column2Type>, ...)
// WITH (tsdb.hypertable, tsdb.partition_column='<partitionColumn>', tsdb.chunk_interval='<chunkInterval>', tsdb.orderby='<orderBy>', tsdb.segmentby='<segmentBy>')
func GenerateCreateHypertableSQL(
	tableInfo *TableInfo,
	params HypertableParams,
) string {
	var columns []string
	var primaryKeys []string
	var uniqueKeys []string

	// Generate column definitions
	for _, col := range tableInfo.Columns {
		columnDef := fmt.Sprintf("%s %s", col.Name, col.DBType)

		if col.Nullable != nil && !*col.Nullable {
			columnDef += " NOT NULL"
		} else if col.Nullable != nil && *col.Nullable {
			columnDef += " NULL"
		}
	}
}

// AddCompressionPolicy is a method that adds the columnstore policy for the given tables.
//
// This function will only start this process however the whole process will run through the 3 steps
// This is third step in the process
// Parameters:
// - tableNames: a slice of table names to add the columnstore policy to
//
// Returns:
// - nil: if the program has a problem it will call log.Fatalf which will exit the program
//
// TThis function specifies the columnstore policy
func (init *DBInitializer) AddColumnstorePolicy(tableNames []string) {
	for _, tableName := range tableNames {
		sql := fmt.Sprintf(
			`
			CALL add_columnstore_policy('%s', INTERVAL '1 week');
			`, tableName)
		_, err := init.pool.Exec(context.Background(), sql)
		if err != nil {
			l.Error().
				Caller().
				Stack().
				Msgf(
					"failed to add columnstore policy for table %s: %v", tableName, err,
				)
		}
	}
}
