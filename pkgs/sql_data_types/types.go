package sql_data_types

import (
	"reflect"
	"time"

	dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/src/db_init"
)

// GnoAddress represents a regular Gno address with database mapping information
// Stores:
// - Address (string)
// - ID (int32)
// - Chain ID (string)
// PRIMARY KEY (id), UNIQUE (address, chain_id)
type GnoAddress struct {
	// any of the values can't be a null value and there shouldn't be any duplicates
	Address string `db:"address" dbtype:"TEXT" nullable:"false" primary:"true" unique:"true"`
	ID      int32  `db:"id" dbtype:"INTEGER GENERATED ALWAYS AS IDENTITY" nullable:"false" primary:"false" unique:"true"`
	// use type enum chain_name from postgres
	ChainName string `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
}

// TableName returns the name of the table for the GnoAddress struct
func (g GnoAddress) TableName() string {
	return "gno_addresses"
}

// GetTableInfo returns the table info for the GnoAddress struct
func (g GnoAddress) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(g, g.TableName())
}

// GnoValidatorAddress represents a Gno validator address with database mapping information
// Same structure as GnoAddress but creates a separate table for validators
type GnoValidatorAddress struct {
	Address   string `db:"address" dbtype:"TEXT" nullable:"false" primary:"true" unique:"true"`
	ID        int32  `db:"id" dbtype:"INTEGER GENERATED ALWAYS AS IDENTITY" nullable:"false" primary:"false" unique:"true"`
	ChainName string `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
}

// TableName returns the name of the table for the GnoValidatorAddress struct
func (gv GnoValidatorAddress) TableName() string {
	return "gno_validators"
}

// GetTableInfo returns the table info for the GnoValidatorAddress struct
func (gv GnoValidatorAddress) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(gv, gv.TableName())
}

// Blocks represents a blockchain block with database mapping information
//
// Stores:
//   - Hash (bytea)
//   - Height (uint64)
//   - Timestamp (time.Time)
//   - Chain ID (string)
//   - Proposer address (int32)
//   - Txs ([]string)
//
// PRIMARY KEY (hash, height, timestamp, chain_id)
type Blocks struct {
	Hash      []byte    `db:"hash" dbtype:"bytea" nullable:"false" primary:"false"`
	Height    uint64    `db:"height" dbtype:"bigint" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainID   string    `db:"chain_id" dbtype:"TEXT" nullable:"false" primary:"false"`
	// proposer address is the validator address hence why this should be an integer
	ProposerAddress int32  `db:"proposer_address" dbtype:"integer" nullable:"false" primary:"false"`
	Txs             []byte `db:"txs" dbtype:"bytea" primary:"false" nullable:"true"` // can be a null value
	ChainName       string `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
}

func (b Blocks) TableName() string {
	return "blocks"
}

// GetTableInfo returns the table info for the Blocks struct
func (b Blocks) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(b, b.TableName())
}

// ValidatorBlockSigning represents a validator block signing with database mapping information
// Stores:
//   - Block height (uint64)
//   - Timestamp (time.Time)
//   - Signed validators (int32 all of the validators that signed the block)
//   - Chain ID (string)
//   - Missed validators (int32 all of the validators that missed the block)
//
// PRIMARY KEY (block_height, timestamp, chain_id)
type ValidatorBlockSigning struct {
	BlockHeight uint64    `db:"block_height" dbtype:"bigint" nullable:"false" primary:"true"`
	Timestamp   time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	SignedVals  []int32   `db:"address" dbtype:"integer" nullable:"false" primary:"false"`
	ChainName   string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"` // use type enum chain_name from postgres
	MissedVals  []int32   `db:"missed_vals" dbtype:"integer" nullable:"false" primary:"false"`
}

// TableName returns the name of the table for the ValidatorBlockSigning struct
func (vbs ValidatorBlockSigning) TableName() string {
	return "validator_block_signing"
}

// GetTableInfo returns the table info for the ValidatorBlockSigning struct
func (vbs ValidatorBlockSigning) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(vbs, vbs.TableName())
}

// AddressTx represents a transaction with database mapping information
//
// Stores:
// - Address (int32)
// - TxHash (bytea)
// - Chain ID (string)
// - Timestamp (time.Time)
// - MsgTypes ([]string)
// PRIMARY KEY (timestamp) because of timescaledb although it is not marked as primary it will be considered as such
type AddressTx struct {
	Address   int32     `db:"address" dbtype:"INTEGER" nullable:"false" primary:"false"`
	TxHash    []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"false"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"false"`
	MsgTypes  []string  `db:"msg_types" dbtype:"TEXT[]" nullable:"false" primary:"false"`
}

// TableName returns the name of the table for the AddressTx struct
func (at AddressTx) TableName() string {
	return "address_tx"
}

// GetTableInfo returns the table info for the AddressTx struct
func (at AddressTx) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(at, at.TableName())
}

// Fee is a postgres type that is used to store the fee of a transaction
//
// It is a custom type that is used to store the fee of a transaction
// Stores:
// - Amount (uint64)
// - Denom (string)
// PRIMARY KEY (amount, denom)
type Fee struct {
	Amount uint64 `db:"amount" dbtype:"NUMERIC"`
	Denom  string `db:"denom" dbtype:"TEXT"`
}

// TypeName returns the name of the type for the Fee struct
func (f Fee) TypeName() string {
	return "fee"
}

// GetSpecialTypeInfo returns the special type info for the Fee struct
func (f Fee) GetSpecialTypeInfo() (*dbinit.SpecialType, error) {
	return dbinit.GetSpecialTypeInfo(f, f.TypeName())
}

// TransactionGeneral represents a transaction general data with database mapping information
type TransactionGeneral struct {
	TxHash             []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"true"`
	ChainName          string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	Timestamp          time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	MsgTypes           []string  `db:"msg_types" dbtype:"TEXT[]" nullable:"false" primary:"false"`
	TxEvents           []byte    `db:"tx_events" dbtype:"bytea" nullable:"true" primary:"false"`            // in some cases can be null
	TxEventsCompressed []byte    `db:"tx_events_compressed" dbtype:"bytea" nullable:"true" primary:"false"` // for now it can be a null could be changed later
	GasUsed            uint64    `db:"gas_used" dbtype:"bigint" nullable:"false" primary:"false"`
	GasWanted          uint64    `db:"gas_wanted" dbtype:"bigint" nullable:"false" primary:"false"`
	Fee                Fee       `db:"fee" dbtype:"fee" nullable:"false" primary:"false"`
}

// TableName returns the name of the table for the TransactionGeneral struct
func (tg TransactionGeneral) TableName() string {
	return "transaction_general"
}

// GetTableInfo returns the table info for the TransactionGeneral struct
func (tg TransactionGeneral) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(tg, tg.TableName())
}

// MsgSend represents a bank send message
type MsgSend struct {
	TxHash      []byte `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"false"`
	ChainName   string `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
	FromAddress string `db:"from_address" dbtype:"TEXT" nullable:"false" primary:"false"`
	// need to test this out later leave it as a possible null value
	ToAddress string    `db:"to_address" dbtype:"TEXT" nullable:"true" primary:"false"`
	Amount    string    `db:"amount" dbtype:"TEXT" nullable:"false" primary:"false"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"false"`
}

// TableName returns the name of the table for the MsgSend struct
func (ms MsgSend) TableName() string {
	return "bank_msg_send"
}

// GetTableInfo returns the table info for the MsgSend struct
func (ms MsgSend) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(ms, ms.TableName())
}

// A method to get the columns of the struct
// Useful in GnoMessage interface
func (ms MsgSend) TableColumns() []string {
	columns := make([]string, 0)
	fields := reflect.TypeOf(ms)
	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
}

// MsgCall represents a VM function call message
type MsgCall struct {
	TxHash    []byte `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"false"`
	ChainName string `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
	Caller    string `db:"caller" dbtype:"TEXT" nullable:"false" primary:"false"`
	PkgPath   string `db:"pkg_path" dbtype:"TEXT" nullable:"false" primary:"false"`
	FuncName  string `db:"func_name" dbtype:"TEXT" nullable:"false" primary:"false"`
	// could be a null value but maybe return an empty array, leave false for now
	Args      []string  `db:"args" dbtype:"TEXT[]" nullable:"false" primary:"false"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"false"`
}

func (mc MsgCall) TableName() string {
	return "vm_msg_call"
}

// GetTableInfo returns the table info for the MsgCall struct
func (mc MsgCall) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(mc, mc.TableName())
}

// A method to get the columns of the struct
// Useful in GnoMessage interface
func (mc MsgCall) TableColumns() []string {
	columns := make([]string, 0)
	fields := reflect.TypeOf(mc)
	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
}

// MsgAddPackage represents a VM package addition message
type MsgAddPackage struct {
	TxHash    []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"false"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
	Creator   string    `db:"creator" dbtype:"TEXT" nullable:"false" primary:"false"`
	PkgPath   string    `db:"pkg_path" dbtype:"TEXT" nullable:"false" primary:"false"`
	PkgName   string    `db:"pkg_name" dbtype:"TEXT" nullable:"false" primary:"false"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"false"`
}

func (ma MsgAddPackage) TableName() string {
	return "vm_msg_add_package"
}

// GetTableInfo returns the table info for the MsgAddPackage struct
func (ma MsgAddPackage) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(ma, ma.TableName())
}

// A method to get the columns of the struct
// Useful in GnoMessage interface
func (ma MsgAddPackage) TableColumns() []string {
	columns := make([]string, 0)
	fields := reflect.TypeOf(ma)
	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
}

// MsgRun represents a VM package run message
type MsgRun struct {
	TxHash    []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"false"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
	Caller    string    `db:"caller" dbtype:"TEXT" nullable:"false" primary:"false"`
	PkgPath   string    `db:"pkg_path" dbtype:"TEXT" nullable:"false" primary:"false"`
	PkgName   string    `db:"pkg_name" dbtype:"TEXT" nullable:"false" primary:"false"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"false"`
}

// A method to get the columns of the struct
// Useful in GnoMessage interface
func (mr MsgRun) TableColumns() []string {
	columns := make([]string, 0)
	// get the fields of the struct
	fields := reflect.TypeOf(mr)
	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
}

func (mr MsgRun) TableName() string {
	return "vm_msg_run"
}

// GetTableInfo returns the table info for the MsgRun struct
func (mr MsgRun) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(mr, mr.TableName())
}

type DataTypes interface {
	TableName(valTable bool) string
	TypeName() string
	GetTableInfo() (*dbinit.TableInfo, error)
}

// DBTable is an interface for structs that represent database tables
type DBTable interface {
	GetTableInfo() (*dbinit.TableInfo, error)
	TableName() string
}

// DBSpecialType is an interface for structs that represent custom database types
type DBSpecialType interface {
	GetSpecialTypeInfo() (*dbinit.SpecialType, error)
	TypeName() string
}

// An interface for Gno messages
//
// Methods:
// - TableColumns() []string: a method to get the columns of the struct
type GnoMessage interface {
	TableColumns() []string
}
