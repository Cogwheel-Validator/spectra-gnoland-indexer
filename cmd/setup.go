package main

import (
	"fmt"
	"log"
	"slices"
	"syscall"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/database"
	dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/db_init"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var rootCmd = &cobra.Command{
	Use:   "create-db",
	Short: "Create a new database named gnoland",
	Long: `Create a new database named gnoland for the indexer. It goes
		throught a lot of steps to create the database and insert the tables and data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("Initiating the cmd to set up the database for the indexer...")
		var dbHost string
		var dbPort int
		var dbUser string
		var dbName string
		var dbPassword string
		var sslMode string

		var allowedSslModes = []string{"disable", "require", "verify-ca", "verify-full", "allow", "prefer"}

		dbHost, _ = cmd.Flags().GetString("db-host")
		dbPort, _ = cmd.Flags().GetInt("db-port")
		dbUser, _ = cmd.Flags().GetString("db-user")
		sslMode, _ = cmd.Flags().GetString("ssl-mode")
		dbName, _ = cmd.Flags().GetString("db-name")

		if sslMode == "" {
			sslMode = "disable" // if not specified, default to disable
		}

		if !slices.Contains(allowedSslModes, sslMode) {
			log.Fatalf("invalid ssl mode: %s", sslMode)
		}

		if dbHost == "" {
			dbHost = "localhost" // if not specified, default to localhost
		}

		if dbPort == 0 {
			dbPort = 5432 // if not specified, default to 5432
		} else if dbPort > 65535 {
			log.Fatalf("invalid port: %d", dbPort)
		}

		if dbUser == "" {
			dbUser = "postgres" // if not specified, default to postgres
		}

		if dbName == "" {
			dbName = "postgres" // if not specified, default to postgres
		}

		fmt.Print("Enter the database password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf("failed to read password: %v", err)
		}
		dbPassword = string(bytePassword)
		fmt.Println()

		// create database config
		dbConfig := database.DatabasePoolConfig{
			Host:                      dbHost,
			Port:                      dbPort,
			User:                      dbUser,
			Dbname:                    dbName,
			Password:                  dbPassword,
			Sslmode:                   sslMode,
			PoolMaxConns:              10,
			PoolMinConns:              1,
			PoolMaxConnLifetime:       10 * time.Minute,
			PoolMaxConnIdleTime:       5 * time.Minute,
			PoolHealthCheckPeriod:     1 * time.Minute,
			PoolMaxConnLifetimeJitter: 1 * time.Minute,
		}

		// create a new database connection
		db := database.NewTimescaleDbSetup(dbConfig)

		// create a new database named "gnoland"
		// but check if the current database is "gnoland"
		currentDb, err := db.CheckCurrentDatabaseName()
		if err != nil {
			log.Fatalf("failed to check current database name: %v", err)
		}
		log.Printf("Logged in into the database %s", currentDb)

		// if the current database is not "gnoland", create a new database named "gnoland"
		// and insert all of the tables and data from the "gnoland" database
		if currentDb != "gnoland" {
			// create a new database named "gnoland"
			log.Printf("Creating a new database named %s", "gnoland")
			err = database.CreateDatabase(db, "gnoland")
			if err != nil {
				log.Fatalf("failed to create database: %v", err)
			}
			// switch to the new database
			log.Printf("Switching to the new database %s", "gnoland")
			// only for now later add dbName value, it is only for the testing now
			err = database.SwitchDatabase(db, dbConfig, "gnoland")
			if err != nil {
				log.Fatalf("failed to switch database: %v", err)
			}
			// insert all of the tables and data from the "gnoland" database
			// First create special types (custom postgres types that tables depend on)
			// and type enums
			specialTypes := []sql_data_types.DBSpecialType{
				sql_data_types.Amount{},
				sql_data_types.Attribute{}, // this needs to be inserted prior to event type
				sql_data_types.Event{},
			}
			typeEnums := []string{
				"gnoland",
			}

			// Initialize database initializer
			dbInit := dbinit.NewDBInitializer(db.GetPool())

			// Create special types first (they need to exist before tables that use them)
			log.Printf("Inserting all of the special types into the %s database", "gnoland")
			for _, specialType := range specialTypes {
				err = dbInit.CreateSpecialTypeFromStruct(specialType, specialType.TypeName())
				if err != nil {
					log.Fatalf("failed to create special type %s: %v", specialType.TypeName(), err)
				}
			}

			// Create type enums
			log.Printf("Inserting all of the type enums into the %s database", "gnoland")
			err = dbInit.CreateChainTypeEnum(typeEnums)
			if err != nil {
				log.Fatalf("failed to create type enum %s: %v", typeEnums, err)
			}

			// Create regular tables (non-time-series tables)
			log.Printf("Inserting all of the regular tables into the %s database", "gnoland")
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
			log.Printf("Inserting all of the hypertables into the %s database", "gnoland")
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
			log.Printf("Successfully created all of the hypertables into the %s database", "gnoland")
		} else {
			log.Printf("The current database is %s, and it already exists", currentDb)
			// TODO else if the current database is "gnoland" then we need to check if the tables exist
			// and if they don't exist then we need to create them
			// also any kind of future updates to the database should be done here
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().String("db-host", "", "The database host, default is localhost")
	rootCmd.PersistentFlags().Int("db-port", 0, "The database port, default is 5432")
	rootCmd.PersistentFlags().String("db-user", "", "The database user, default is postgres")
	rootCmd.PersistentFlags().String("db-name", "", "The database name, default is postgres")
	rootCmd.PersistentFlags().String("ssl-mode", "", "The SSL mode for the database connection, default is disable")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute root command: %v", err)
	}
}
