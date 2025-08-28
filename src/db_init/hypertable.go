package dbinit

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// Hypertable is a table that is used to store time-series data in the TimescaleDB
//
// The program will use the create_hypertable function and will align with altering the tables
// with the SELECT or ALTER functions that work with version prior to v2.19.3 and later because someone might
// be using an older version of TimescaleDB
//
// However the recomendation should be to use the latest version of TimescaleDB.
// As for which edition of TimescaleDB is recomended, the answer is that it depends.
// The Tiger Data is the TimescaleDB Cloud edition.
// The Community Edition is the self hosted version.
// And the Apache TimescaleDB is the open source version.
//
// The recomendation is to use the atleast the Community TimescaleDB Edition. The Tiger Data does have some additional features
// but the Community Edition is very stable and it was tested with the cosmos indexer so it should defenetly work.
//
// TODO: for now even the cosmos indexer uses the same logic but in the future we should
// adopt the new logic with the latest version of TimescaleDB in case this gets deprecated

// ConvertToHypertables is a method that converts the given table names to hypertables
//
// This function will only start this process however the whole process will run through the 3 steps
// This is first step in the process
// Args:
// - tableNames: a slice of table names to convert to hypertables
//
// Returns:
// - nil: if the program has a problem it will call log.Fatalf which will exit the program
//
// The function will only set the hypertable chunk to 1 week, this is pretty much the default
// however this interval should be defined mostly by developer and system specs.
// The indexed data on a weekly basis must not be more than 25% of system RAM memory.
// However depending on the data and the system specs this might need to be adjusted to be shorter or longer.
// This is all in alpha stage and might be adjusted later. For now it will be hard coded to 1 week since this
// was an optimal setting for the cosmos indexer.
// For addition info search the Tiger Data for more info.
func (init *DBInitializer) ConvertToHypertables(tableNames []string) {
	for _, tableName := range tableNames {
		sql := fmt.Sprintf("SELECT create_hypertable('%s', 'timestamp', chunk_time_interval => INTERVAL '1 weeks')", tableName)
		_, err := init.pool.Exec(context.Background(), sql)
		if err != nil {
			log.Fatalf("failed to convert table %s to hypertable: %v", tableName, err)
		}
	}
}

// AlterCompressionSegments is a method that alters the compression segments for the given tables
//
// This function will only start this process however the whole process will run through the 3 steps
// This is second step in the process
// Args:
// - tables: a map of table names to their columns
//
// Returns:
// - nil: if the program has a problem it will call log.Fatalf which will exit the program
//
// The function will only set the compression segments to the given columns.
// The columns will be hard encoded for now depending on the table.
func (init *DBInitializer) AlterCompressionSegments(tables map[string][]string) {
	for tableName, columns := range tables {
		columnsString := strings.Join(columns, ", ")
		sql := fmt.Sprintf(
			`
			ALTER TABLE %s SET (
				timescaledb.compress = TRUE,
				timescaledb.compress_segmentby = %s
				timescaledb.compress_orderby = 'timestamp DESC'
			)
			`, tableName, columnsString)
		_, err := init.pool.Exec(context.Background(), sql)
		if err != nil {
			log.Fatalf("failed to alter compression segments for table %s: %v", tableName, err)
		}
	}
}

// AddCompressionPolicy is a method that adds the compression policy for the given tables
//
// This function will only start this process however the whole process will run through the 3 steps
// This is third step in the process
// Args:
// - tableNames: a slice of table names to add the compression policy to
//
// Returns:
// - nil: if the program has a problem it will call log.Fatalf which will exit the program
//
// TThis function specifies the compression policy
func (init *DBInitializer) AddCompressionPolicy(tableNames []string) {
	for _, tableName := range tableNames {
		sql := fmt.Sprintf(
			`
			SELECT add_compression_policy('%s', INTERVAL '1 week');
			`, tableName)
		_, err := init.pool.Exec(context.Background(), sql)
		if err != nil {
			log.Fatalf("failed to add compression policy for table %s: %v", tableName, err)
		}
	}
}
