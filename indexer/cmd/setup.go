package cmd

import (
	"context"
	"time"

	dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/db_init"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
	"github.com/spf13/cobra"
)

const defaultDBUser = "postgres"

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

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Database setup tools",
	Long:  `A collection of tools to set up and manage the database for the gnoland indexer.`,
}

func init() {
	// Add subcommands
	setupCmd.AddCommand(createDbCmd)
	setupCmd.AddCommand(createUserCmd)
	setupCmd.AddCommand(createConfigCmd)

	// Common flags for both database setup commands
	for _, cmd := range []*cobra.Command{createDbCmd, createUserCmd} {
		cmd.Flags().StringP("db-host", "b", "", "The database host, default is localhost")
		cmd.Flags().IntP("db-port", "p", 0, "The database port, default is 5432")
		cmd.Flags().StringP("db-user", "u", "", "The database user, default is postgres")
		cmd.Flags().StringP("db-name", "d", "", "The database name, default is postgres")
		cmd.Flags().StringP("ssl-mode", "s", "", "The SSL mode for the database connection, default is disable")
	}

	// create-user specific flags
	createUserCmd.Flags().StringP("privilege", "r", "", "The privilege level for the user (reader or writer)")
	createUserCmd.Flags().String("user", "", "The user name for the user to create")

	// create-db specific flags
	createDbCmd.Flags().String("new-db-name", "", "The database name to create, default is gnoland")
	createDbCmd.Flags().String("chain-name", "", "The chain name for the database type enum, default is gnoland")

	// create-config specific flags
	createConfigCmd.Flags().StringP("config", "c", "config.yml", "The config file name, default is config.yml")
	createConfigCmd.Flags().BoolP("overwrite", "o", false, "Overwrite the existing config file, default is false")
}

var createDbCmd = &cobra.Command{
	Use:   "create-db",
	Short: "Create a new database named gnoland",
	Long: `Create a new database named gnoland for the indexer. It goes\n
	through a lot of steps to create the database and insert the tables and data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return createDatabaseSetup(cmd)
	},
}

func createDatabaseSetup(cmd *cobra.Command) error {
	l := logger.Get()
	l.Info().Msg("initiating database setup for the indexer")

	params, err := parseCommonFlags(cmd, "postgres")
	if err != nil {
		l.Error().Err(err).Msg("failed to parse flags")
		return err
	}

	newDbName := getFlagStringWithDefault(cmd, "new-db-name", "gnoland")
	chainName := getFlagStringWithDefault(cmd, "chain-name", "gnoland")

	params.password, err = promptPassword()
	if err != nil {
		l.Error().Err(err).Msg("failed to read password")
		return err
	}

	dbConfig := params.createDatabaseConfig()
	db := database.NewTimescaleDbSetup(dbConfig)

	currentDb, err := checkCurrentDatabase(db)
	if err != nil {
		return err
	}
	l.Info().Str("db", currentDb).Msg("logged into database")

	if currentDb != newDbName {
		return initializeNewDatabase(db, dbConfig, newDbName, chainName)
	}

	l.Info().Str("db", currentDb).Msg("database already exists, skipping creation")
	return nil
}

func getFlagStringWithDefault(cmd *cobra.Command, flagName, defaultValue string) string {
	value, _ := cmd.Flags().GetString(flagName)
	if value == "" {
		return defaultValue
	}
	return value
}

func checkCurrentDatabase(db *database.TimescaleDb) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return db.CheckCurrentDatabaseName(ctx)
}

func initializeNewDatabase(db *database.TimescaleDb, dbConfig database.DatabasePoolConfig, newDbName, chainName string) error {
	l := logger.Get()

	l.Info().Str("db", newDbName).Msg("creating new database")
	if err := database.CreateDatabase(db, newDbName); err != nil {
		l.Error().Err(err).Msg("failed to create database")
		return err
	}

	l.Info().Str("db", newDbName).Msg("switching to new database")
	if err := database.SwitchDatabase(db, dbConfig, newDbName); err != nil {
		l.Error().Err(err).Msg("failed to switch database")
		return err
	}

	dbInit := dbinit.NewDBInitializer(db.GetPool())

	if err := createDatabaseTypes(dbInit, chainName); err != nil {
		return err
	}

	if err := createRegularTables(dbInit, chainName); err != nil {
		return err
	}

	if err := createHypertables(dbInit, chainName); err != nil {
		return err
	}

	l.Info().Str("chain", chainName).Msg("successfully created all hypertables")
	return nil
}

func createDatabaseTypes(dbInit *dbinit.DBInitializer, chainName string) error {
	l := logger.Get()

	specialTypes := []sql_data_types.DBSpecialType{
		sql_data_types.Amount{},
		sql_data_types.Attribute{},
		sql_data_types.Event{},
	}

	l.Info().Str("chain", chainName).Msg("inserting special types")
	for _, specialType := range specialTypes {
		if err := dbInit.CreateSpecialTypeFromStruct(specialType, specialType.TypeName()); err != nil {
			l.Error().Err(err).Str("type", specialType.TypeName()).Msg("failed to create special type")
			return err
		}
	}

	typeEnums := []string{chainName}
	l.Info().Str("chain", chainName).Msg("inserting type enums")
	if err := dbInit.CreateChainTypeEnum(typeEnums); err != nil {
		l.Error().Err(err).Strs("enums", typeEnums).Msg("failed to create type enum")
		return err
	}

	return nil
}

func createRegularTables(dbInit *dbinit.DBInitializer, chainName string) error {
	l := logger.Get()

	regularTables := []sql_data_types.DBTable{
		sql_data_types.GnoAddress{},
		sql_data_types.GnoValidatorAddress{},
		sql_data_types.ApiKey{},
	}

	l.Info().Str("chain", chainName).Msg("inserting regular tables")
	for _, dataType := range regularTables {
		if err := dbInit.CreateTableFromStruct(dataType, dataType.TableName()); err != nil {
			l.Error().Err(err).Str("table", dataType.TableName()).Msg("failed to create table")
			return err
		}
	}

	return nil
}

func createHypertables(dbInit *dbinit.DBInitializer, chainName string) error {
	l := logger.Get()

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

	l.Info().Str("chain", chainName).Msg("inserting hypertables")
	for _, ht := range hypertables {
		if err := dbInit.CreateHypertableFromStruct(ht.table, ht.table.TableName(), ht.partitionColumn, ht.chunkInterval); err != nil {
			l.Error().Err(err).Str("table", ht.table.TableName()).Msg("failed to create hypertable")
			return err
		}
	}

	return nil
}

var createUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "Create a new user for the database",
	Long:  `Create a new user for the database. It will ask for the password and create the user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.Get()

		// Parse and validate common database flags
		params, err := parseCommonFlags(cmd, "gnoland")
		if err != nil {
			l.Fatal().Err(err).Msg("failed to parse flags")
			return err
		}

		// Get privilege flag
		privilege, _ := cmd.Flags().GetString("privilege")
		if privilege == "" {
			l.Fatal().Msg("privilege is required")
			return cmd.Usage()
		} else if privilege != "reader" && privilege != "writer" && privilege != "keymgr" {
			l.Fatal().Str("privilege", privilege).Msg("invalid privilege, must be reader, writer, or keymgr")
			return cmd.Usage()
		}

		// get the user name from the flags
		userName, _ := cmd.Flags().GetString("user")
		if userName == "" {
			l.Fatal().Msg("user name is required")
			return cmd.Usage()
		}

		// Prompt for password
		params.password, err = promptPassword()
		if err != nil {
			l.Fatal().Err(err).Msg("failed to read password")
			return err
		}

		// Create database config and connection
		dbConfig := params.createDatabaseConfig()
		db := database.NewTimescaleDbSetup(dbConfig)
		dbInit := dbinit.NewDBInitializer(db.GetPool())

		// Create a new user
		err = dbInit.CreateUser(userName)
		if err != nil {
			l.Fatal().Err(err).Str("user", userName).Msg("failed to create user")
			return err
		}

		// Appoint privileges to the user
		err = dbInit.AppointPrivileges(userName, privilege, sql_data_types.AllTableNames())
		if err != nil {
			l.Fatal().Err(err).Str("user", userName).Str("privilege", privilege).Msg("failed to appoint privileges")
			return err
		}

		l.Info().Str("user", userName).Str("privilege", privilege).Msg("successfully created user")
		return nil
	},
}

var createConfigCmd = &cobra.Command{
	Use:   "create-config",
	Short: "Generate a config with default values.",
	Long: `Generate a config with default values. It will make a config file with default values. 
	You can add --overwrite to overwrite the existing config file. And you can use --config to specifly the path`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.Get()

		// get the config file name from the flags
		configFileName, _ := cmd.Flags().GetString("config")
		if configFileName == "" {
			configFileName = "config.yaml"
		}
		overwrite, _ := cmd.Flags().GetBool("overwrite")
		if err := createConfig(overwrite, configFileName); err != nil {
			return err
		}
		l.Info().Str("file", configFileName).Msg("successfully created config file")
		return nil
	},
}
