package main

import (
	"fmt"
	"log"
	"slices"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/src/database"
	dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/src/db_init"
	"github.com/spf13/cobra"
)

func main() {
	var dbHost string
	var dbPort int = 0
	var dbUser string
	var dbName string
	var dbPassword string
	var sslMode string

	var allowedSslModes = []string{"disable", "require", "verify-ca", "verify-full", "allow", "prefer"}

	// use cobra flags to insert dbhost dpport dbuser
	_ = &cobra.Command{
		Use:   "create-db",
		Short: "Create a new database",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbHost, _ = cmd.Flags().GetString("--db-host")
			dbPort, _ = cmd.Flags().GetInt("--db-port")
			dbUser, _ = cmd.Flags().GetString("--db-user")
			sslMode, _ = cmd.Flags().GetString("--ssl-mode")
			dbName, _ = cmd.Flags().GetString("--db-name")
			return nil
		},
	}

	// verify inputed data

	if !slices.Contains(allowedSslModes, sslMode) {
		log.Fatalf("invalid ssl mode: %s", sslMode)
	} else if sslMode == "" {
		sslMode = "disable" // if not specified, default to disable
	}

	if dbHost == "" {
		dbHost = "localhost" // if not specified, default to localhost
	}

	if dbPort == 0 {
		dbPort = 5432 // if not specified, default to 5432
	} else if dbPort > 65535 {
		log.Fatalf("invalid port: %d", dbPort)
	}

	// input the database password
	fmt.Scanf("Enter the database password: %s", &dbPassword)

	// create a new database connection
	db := database.NewTimescaleDb(database.DatabasePoolConfig{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Dbname:   dbName,
		Password: dbPassword,
		Sslmode:  sslMode,
	})

	// create a new database named "gnoland"
	// but check if the current database is "gnoland"
	currentDb, err := db.CheckCurrentDatabaseName()
	if err != nil {
		log.Fatalf("failed to check current database name: %v", err)
	}

	// if the current database is not "gnoland", create a new database named "gnoland"
	// and insert all of the tables and data from the "gnoland" database
	if currentDb != "gnoland" {
		// create a new database named "gnoland"
		err = database.CreateDatabase(db, "gnoland")
		if err != nil {
			log.Fatalf("failed to create database: %v", err)
		}
		// switch to the new database
		err = database.SwitchDatabase(db)
		if err != nil {
			log.Fatalf("failed to switch database: %v", err)
		}
		// insert all of the tables and data from the "gnoland" database

		// First create special types (custom postgres types that tables depend on)
		specialTypes := []sql_data_types.DBSpecialType{
			sql_data_types.Fee{},
		}

		// Initialize database initializer
		dbInit := dbinit.NewDBInitializer(db.GetPool())

		// Create special types first (they need to exist before tables that use them)
		for _, specialType := range specialTypes {
			err = dbInit.CreateSpecialTypeFromStruct(specialType, specialType.TypeName())
			if err != nil {
				log.Fatalf("failed to create special type %s: %v", specialType.TypeName(), err)
			}
		}

		// Create regular tables (non-time-series tables)
		regularTables := []sql_data_types.DBTable{
			sql_data_types.GnoAddress{},
			sql_data_types.GnoValidatorAddress{},
		}

		for _, dataType := range regularTables {
			err = dbInit.CreateTableFromStruct(dataType, dataType.TableName())
			if err != nil {
				log.Fatalf("failed to create table %s: %v", dataType.TableName(), err)
			}
		}

		// Create hypertables (time-series tables with timestamp columns)
		hypertables := []struct {
			table           sql_data_types.DBTable
			partitionColumn string
			chunkInterval   string
		}{
			{sql_data_types.Blocks{}, "timestamp", "1 week"},
			{sql_data_types.ValidatorBlockSigning{}, "timestamp", "1 week"},
			{sql_data_types.AddressTx{}, "timestamp", "1 week"},
			{sql_data_types.TransactionGeneral{}, "timestamp", "1 week"},
			{sql_data_types.MsgSend{}, "timestamp", "1 week"},
			{sql_data_types.MsgCall{}, "timestamp", "1 week"},
			{sql_data_types.MsgAddPackage{}, "timestamp", "1 week"},
			{sql_data_types.MsgRun{}, "timestamp", "1 week"},
		}

		for _, ht := range hypertables {
			err = dbInit.CreateHypertableFromStruct(ht.table, ht.table.TableName(), ht.partitionColumn, ht.chunkInterval)
			if err != nil {
				log.Fatalf("failed to create hypertable %s: %v", ht.table.TableName(), err)
			}
		}
	}

}
