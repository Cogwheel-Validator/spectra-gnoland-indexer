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

var allowedSslModes = []string{"disable", "require", "verify-ca", "verify-full", "allow", "prefer"}

// dbParams holds common database connection parameters
type dbParams struct {
	host     string
	port     int
	user     string
	name     string
	password string
	sslMode  string
}

// parseCommonFlags extracts and validates common database flags
func parseCommonFlags(cmd *cobra.Command, defaultDbName string) (*dbParams, error) {
	params := &dbParams{}

	params.host, _ = cmd.Flags().GetString("db-host")
	params.port, _ = cmd.Flags().GetInt("db-port")
	params.user, _ = cmd.Flags().GetString("db-user")
	params.sslMode, _ = cmd.Flags().GetString("ssl-mode")
	params.name, _ = cmd.Flags().GetString("db-name")

	// Apply defaults
	if params.sslMode == "" {
		params.sslMode = "disable"
	}
	if params.host == "" {
		params.host = "localhost"
	}
	if params.port == 0 {
		params.port = 5432
	}
	if params.user == "" {
		params.user = "postgres"
	}
	if params.name == "" {
		params.name = defaultDbName
	}

	// Validate
	if !slices.Contains(allowedSslModes, params.sslMode) {
		return nil, fmt.Errorf("invalid ssl mode: %s", params.sslMode)
	}
	if params.port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", params.port)
	}

	return params, nil
}

// promptPassword prompts user for password input
func promptPassword() (string, error) {
	fmt.Print("Enter the database password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %v", err)
	}
	fmt.Println()
	return string(bytePassword), nil
}

// createDatabaseConfig creates a DatabasePoolConfig from dbParams
func (p *dbParams) createDatabaseConfig() database.DatabasePoolConfig {
	return database.DatabasePoolConfig{
		Host:                      p.host,
		Port:                      p.port,
		User:                      p.user,
		Dbname:                    p.name,
		Password:                  p.password,
		Sslmode:                   p.sslMode,
		PoolMaxConns:              10,
		PoolMinConns:              1,
		PoolMaxConnLifetime:       10 * time.Minute,
		PoolMaxConnIdleTime:       5 * time.Minute,
		PoolHealthCheckPeriod:     1 * time.Minute,
		PoolMaxConnLifetimeJitter: 1 * time.Minute,
	}
}

var rootCmd = &cobra.Command{
	Use:   "setup",
	Short: "Database setup tools for the gnoland indexer",
	Long:  `A collection of tools to set up and manage the database for the gnoland indexer.`,
}

var createDbCmd = &cobra.Command{
	Use:   "create-db",
	Short: "Create a new database named gnoland",
	Long: `Create a new database named gnoland for the indexer. It goes
		throught a lot of steps to create the database and insert the tables and data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("Initiating the cmd to set up the database for the indexer...")

		// Parse and validate common database flags
		params, err := parseCommonFlags(cmd, "postgres")
		if err != nil {
			log.Fatalf("failed to parse flags: %v", err)
		}

		// Prompt for password
		params.password, err = promptPassword()
		if err != nil {
			log.Fatalf("failed to read password: %v", err)
		}

		// Create database config
		dbConfig := params.createDatabaseConfig()

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

var createUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "Create a new user for the database",
	Long:  `Create a new user for the database. It will ask for the password and create the user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse and validate common database flags
		params, err := parseCommonFlags(cmd, "gnoland")
		if err != nil {
			log.Fatalf("failed to parse flags: %v", err)
		}

		// Get privilege flag
		privilege, _ := cmd.Flags().GetString("privilege")
		if privilege == "" {
			log.Fatalf("privilege is required")
		} else if privilege != "reader" && privilege != "writer" {
			log.Fatalf("invalid privilege: %s", privilege)
		}

		// get the user name from the flags
		userName, _ := cmd.Flags().GetString("user")
		if userName == "" {
			log.Fatalf("user name is required")
		}

		// Prompt for password
		params.password, err = promptPassword()
		if err != nil {
			log.Fatalf("failed to read password: %v", err)
		}

		// Create database config and connection
		dbConfig := params.createDatabaseConfig()
		db := database.NewTimescaleDbSetup(dbConfig)
		dbInit := dbinit.NewDBInitializer(db.GetPool())

		// Create a new user
		err = dbInit.CreateUser(userName)
		if err != nil {
			log.Fatalf("failed to create user: %v", err)
		}

		// Appoint privileges to the user
		err = dbInit.AppointPrivileges(params.user, privilege, []string{})
		if err != nil {
			log.Fatalf("failed to appoint privileges to user: %v", err)
		}

		log.Printf("Successfully created user %s with privilege %s", params.user, privilege)
		return nil
	},
}

func init() {
	// Common database flags to both commands
	for _, cmd := range []*cobra.Command{createDbCmd, createUserCmd} {
		cmd.Flags().String("db-host", "", "The database host, default is localhost")
		cmd.Flags().Int("db-port", 0, "The database port, default is 5432")
		cmd.Flags().String("db-user", "", "The database user, default is postgres")
		cmd.Flags().String("db-name", "", "The database name, default is postgres")
		cmd.Flags().String("ssl-mode", "", "The SSL mode for the database connection, default is disable")
	}

	// Add create-user command specific flags
	createUserCmd.Flags().String("privilege", "", "The privilege level for the user (reader or writer)")
	createUserCmd.Flags().String("user", "", "The user name for the user to create")

	rootCmd.AddCommand(createDbCmd)
	rootCmd.AddCommand(createUserCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute root command: %v", err)
	}
}
