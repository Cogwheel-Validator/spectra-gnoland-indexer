package sql_data_types

import (
	"reflect"
	"time"

	dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/db_init"
)

// TxAddresses groups all addresses involved in a single transaction
// It stores in a set like data structure to avoid duplicates
// all addresses for the same transaction hash together
type TxAddresses struct {
	TxHash    []byte
	Addresses map[int32]struct{}
}

// NewTxAddresses creates a new TxAddresses with the given transaction hash
func NewTxAddresses(txHash []byte) *TxAddresses {
	return &TxAddresses{
		TxHash:    txHash,
		Addresses: make(map[int32]struct{}),
	}
}

// AddAddress adds an address to the set
// this will probably overwrite the address if it already exists but there should be no duplicates
func (ta *TxAddresses) AddAddress(addressID int32) {
	ta.Addresses[addressID] = struct{}{}
}

// GetAddressList returns a slice of all address IDs.
// Returns a slice of all address IDs.
func (ta *TxAddresses) GetAddressList() []int32 {
	addresses := make([]int32, 0, len(ta.Addresses))
	for addr := range ta.Addresses {
		addresses = append(addresses, addr)
	}
	return addresses
}

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
	ChainName string `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false" unique:"true"`
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
// Stores:
// - Address (string)
// - ID (int32)
// - Chain Name (string)
// PRIMARY KEY (id), UNIQUE (address, chain_name)
type GnoValidatorAddress struct {
	Address   string `db:"address" dbtype:"TEXT" nullable:"false" primary:"true" unique:"true"`
	ID        int32  `db:"id" dbtype:"INTEGER GENERATED ALWAYS AS IDENTITY" nullable:"false" primary:"false" unique:"true"`
	ChainName string `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false" unique:"true"`
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
//   - Chain Name (string)
//
// PRIMARY KEY (height, timestamp, chain_name)
type Blocks struct {
	Hash      []byte    `db:"hash" dbtype:"bytea" nullable:"false" primary:"false"`
	Height    uint64    `db:"height" dbtype:"bigint" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainID   string    `db:"chain_id" dbtype:"TEXT" nullable:"false" primary:"false"`
	Txs       [][]byte  `db:"txs" dbtype:"bytea[]" primary:"false" nullable:"true"` // can be a null value
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
}

func (b Blocks) TableName() string {
	return "blocks"
}

// GetTableInfo returns the table info for the Blocks struct
func (b Blocks) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(b, b.TableName())
}

func (b Blocks) TableColumns() []string {
	columns := make([]string, 0)
	fields := reflect.TypeOf(b)
	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
}

// ValidatorBlockSigning represents a validator block signing with database mapping information
// Stores:
//   - Block height (uint64)
//   - Timestamp (time.Time)
//   - Proposer (int32)
//   - Signed validators (int32 all of the validators that signed the block)
//   - Chain ID (string)
//   - Missed validators (int32 all of the validators that missed the block)
//
// PRIMARY KEY (block_height, timestamp, chain_id)
type ValidatorBlockSigning struct {
	BlockHeight uint64    `db:"block_height" dbtype:"bigint" nullable:"false" primary:"true"`
	Timestamp   time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	Proposer    int32     `db:"proposer" dbtype:"integer" nullable:"false" primary:"false"`
	SignedVals  []int32   `db:"signed_vals" dbtype:"integer[]" nullable:"false" primary:"false"`
	ChainName   string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"` // use type enum chain_name from postgres
	// MissedVals  []int32   `db:"missed_vals" dbtype:"integer" nullable:"false" primary:"false"`
	// can't confirm who is in the active set without making a smart contract query to the DAO smart contract
	// which could cost a lot of gas if the program checks it constantly, also in "historical" mode when the program
	// is running it can't check previous active set without having some over engeneered system to track the votes from the DAO
	// so only store the signed validators because that is the only thing we can gather and confirm they did sign
}

// TableName returns the name of the table for the ValidatorBlockSigning struct
func (vbs ValidatorBlockSigning) TableName() string {
	return "validator_block_signing"
}

// GetTableInfo returns the table info for the ValidatorBlockSigning struct
func (vbs ValidatorBlockSigning) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(vbs, vbs.TableName())
}

func (vbs ValidatorBlockSigning) TableColumns() []string {
	columns := make([]string, 0)
	fields := reflect.TypeOf(vbs)
	numFields := fields.NumField()
	for i := range numFields {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
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
func (at AddressTx) TableColumns() []string {
	columns := make([]string, 0)
	fields := reflect.TypeOf(at)
	numFields := fields.NumField()
	for i := range numFields {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
}

// TransactionGeneral represents a transaction general data with database mapping information
//
// Stores:
// - TxHash (bytea)
// - ChainName (string)
// - Timestamp (time.Time)
// - MsgTypes (string[])
// - TxEvents (Event[])
// - TxEventsCompressed (bytea)
// - GasUsed (uint64)
// - GasWanted (uint64)
// - Fee (Fee)
//
// PRIMARY KEY (tx_hash, chain_name, timestamp)
//
// INFO about this type!
// This project is open source hence for the sake of wider addoption this table storer both compressed and native format
// But what kind of data will be stored should be managed by the config.
// It is not recommended to use both modes at the same time.
type TransactionGeneral struct {
	TxHash      []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"true"`
	ChainName   string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	Timestamp   time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	BlockHeight uint64    `db:"block_height" dbtype:"bigint" nullable:"false" primary:"false"`
	MsgTypes    []string  `db:"msg_types" dbtype:"TEXT[]" nullable:"false" primary:"false"`
	// tx events in the future there should be an option to have this compressed
	// for now only store the native format but keep the option to have it compressed
	TxEvents           []Event `db:"tx_events" dbtype:"event[]" nullable:"true" primary:"false"`
	TxEventsCompressed []byte  `db:"tx_events_compressed" dbtype:"bytea" nullable:"true" primary:"false"`
	CompressionOn      bool    `db:"compression_on" dbtype:"boolean" nullable:"false" primary:"false"`
	GasUsed            uint64  `db:"gas_used" dbtype:"bigint" nullable:"false" primary:"false"`
	GasWanted          uint64  `db:"gas_wanted" dbtype:"bigint" nullable:"false" primary:"false"`
	Fee                Amount  `db:"fee" dbtype:"amount" nullable:"false" primary:"false"`
}

// TableName returns the name of the table for the TransactionGeneral struct
func (tg TransactionGeneral) TableName() string {
	return "transaction_general"
}

// GetTableInfo returns the table info for the TransactionGeneral struct
func (tg TransactionGeneral) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(tg, tg.TableName())
}

func (tg TransactionGeneral) TableColumns() []string {
	columns := make([]string, 0)
	fields := reflect.TypeOf(tg)
	numFields := fields.NumField()
	for i := range numFields {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
}

// GetTxHash returns the tx hash of the transaction general
//
// Returns:
//   - []byte: the tx hash of the transaction general
func (tg *TransactionGeneral) GetTxHash() []byte {
	return tg.TxHash
}

func (tg *TransactionGeneral) GetMessageTypes() []string {
	return tg.MsgTypes
}

// MsgSend represents a bank send message
//
// Stores:
// - TxHash (bytea)
// - Timestamp (time.Time)
// - ChainName (string)
// - FromAddress (int32)
// - ToAddress (int32)
// - Amount (Amount[])
// - Signers (int32[])
// - MessageCounter (int16)
//
// PRIMARY KEY (tx_hash, chain_name, timestamp)
type MsgSend struct {
	TxHash    []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	// gno address, pull from the gno_addresses table
	FromAddress int32 `db:"from_address" dbtype:"INTEGER" nullable:"false" primary:"false"`
	// gno address, pull from the gno_addresses table
	ToAddress      int32    `db:"to_address" dbtype:"INTEGER" nullable:"true" primary:"false"`
	Amount         []Amount `db:"amount" dbtype:"amount[]" nullable:"false" primary:"false"`
	Signers        []int32  `db:"signers" dbtype:"INTEGER[]" nullable:"false" primary:"false"`
	MessageCounter int16    `db:"message_counter" dbtype:"smallint" nullable:"false" primary:"true"`
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

// GetAllAddresses returns all the addresses that are involved in the message
// it will return the from address, to address, and signers in a single TxAddresses struct
// This prevents duplicates by grouping all addresses for the same transaction
//
// Returns:
//   - *TxAddresses: grouped addresses for this transaction
func (ms *MsgSend) GetAllAddresses() *TxAddresses {
	txAddresses := NewTxAddresses(ms.TxHash)
	txAddresses.AddAddress(ms.FromAddress)
	if ms.ToAddress != 0 {
		txAddresses.AddAddress(ms.ToAddress)
	}
	for _, address := range ms.Signers {
		txAddresses.AddAddress(address)
	}
	return txAddresses
}

// MsgCall represents a VM function call message
//
// Stores:
//   - TxHash (bytea)
//   - Timestamp (time.Time)
//   - ChainName (string)
//   - Caller (int32)
//   - PkgPath (string)
//   - FuncName (string)
//   - Args (string)
//   - Send (Amount[])
//   - MaxDeposit (Amount[])
//   - Signers (int32[])
//   - MessageCounter (int16)
//
// PRIMARY KEY (tx_hash, chain_name, timestamp)
type MsgCall struct {
	TxHash    []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	// gno address, pull from the gno_addresses table
	Caller         int32    `db:"caller" dbtype:"INTEGER" nullable:"false" primary:"false"`
	PkgPath        string   `db:"pkg_path" dbtype:"TEXT" nullable:"true" primary:"false"`
	FuncName       string   `db:"func_name" dbtype:"TEXT" nullable:"true" primary:"false"`
	Args           string   `db:"args" dbtype:"TEXT" nullable:"true" primary:"false"`
	Send           []Amount `db:"send" dbtype:"amount[]" nullable:"true" primary:"false"`
	MaxDeposit     []Amount `db:"max_deposit" dbtype:"amount[]" nullable:"true" primary:"false"`
	Signers        []int32  `db:"signers" dbtype:"INTEGER[]" nullable:"false" primary:"false"`
	MessageCounter int16    `db:"message_counter" dbtype:"smallint" nullable:"false" primary:"true"`
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
	numFields := fields.NumField()
	for i := range numFields {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
}

// GetAllAddresses returns all the addresses that are involved in the message
// Groups the caller and signers for this transaction
//
// Returns:
//   - *TxAddresses: grouped addresses for this transaction
func (mc *MsgCall) GetAllAddresses() *TxAddresses {
	txAddresses := NewTxAddresses(mc.TxHash)
	txAddresses.AddAddress(mc.Caller)
	for _, addr := range mc.Signers {
		txAddresses.AddAddress(addr)
	}
	return txAddresses
}

// MsgAddPackage represents a VM package addition message
//
// Stores:
// - TxHash (bytea)
// - ChainName (string)
// - Creator (string)
// - PkgPath (string)
// - PkgName (string)
// - PkgFileNames (string[])
// - Send (Amount[])
// - MaxDeposit (Amount[])
// - Signers (int32[])
// - Timestamp (time.Time)
// - MessageCounter (int16)
//
// PRIMARY KEY (tx_hash, chain_name, timestamp)
type MsgAddPackage struct {
	TxHash    []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	// gno address, pull from the gno_addresses table
	Creator      int32    `db:"creator" dbtype:"INTEGER" nullable:"false" primary:"false"`
	PkgPath      string   `db:"pkg_path" dbtype:"TEXT" nullable:"true" primary:"false"`
	PkgName      string   `db:"pkg_name" dbtype:"TEXT" nullable:"true" primary:"false"`
	PkgFileNames []string `db:"pkg_file_names" dbtype:"TEXT[]" nullable:"true" primary:"false"`
	Send         []Amount `db:"send" dbtype:"amount[]" nullable:"true" primary:"false"`
	MaxDeposit   []Amount `db:"max_deposit" dbtype:"amount[]" nullable:"true" primary:"false"`
	// signers are the addresses that signed the transaction
	Signers        []int32 `db:"signers" dbtype:"INTEGER[]" nullable:"false" primary:"false"`
	MessageCounter int16   `db:"message_counter" dbtype:"smallint" nullable:"false" primary:"true"`
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
	numFields := fields.NumField()
	for i := range numFields {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
}

// GetAllAddresses returns all the addresses that are involved in the message
// Groups the creator and signers for this transaction
//
// Returns:
//   - *TxAddresses: grouped addresses for this transaction
func (ma *MsgAddPackage) GetAllAddresses() *TxAddresses {
	txAddresses := NewTxAddresses(ma.TxHash)
	txAddresses.AddAddress(ma.Creator)
	for _, addr := range ma.Signers {
		txAddresses.AddAddress(addr)
	}
	return txAddresses
}

// MsgRun represents a VM package run message
//
// Stores:
// - TxHash (bytea)
// - Timestamp (time.Time)
// - ChainName (string)
// - Caller (int32)
// - PkgPath (string)
// - PkgName (string)
// - PkgFileNames (string[])
// - Send (Amount[])
// - MaxDeposit (Amount[])
// - Signers (int32[])
// - MessageCounter (int16)
//
// PRIMARY KEY (tx_hash, chain_name, timestamp)
type MsgRun struct {
	TxHash    []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	// gno address, pull from the gno_addresses table
	Caller       int32    `db:"caller" dbtype:"INTEGER" nullable:"false" primary:"false"`
	PkgPath      string   `db:"pkg_path" dbtype:"TEXT" nullable:"true" primary:"false"`
	PkgName      string   `db:"pkg_name" dbtype:"TEXT" nullable:"true" primary:"false"`
	PkgFileNames []string `db:"pkg_file_names" dbtype:"TEXT[]" nullable:"true" primary:"false"`
	Send         []Amount `db:"send" dbtype:"amount[]" nullable:"true" primary:"false"`
	MaxDeposit   []Amount `db:"max_deposit" dbtype:"amount[]" nullable:"true" primary:"false"`
	// signers are the addresses that signed the transaction
	Signers        []int32 `db:"signers" dbtype:"INTEGER[]" nullable:"false" primary:"false"`
	MessageCounter int16   `db:"message_counter" dbtype:"smallint" nullable:"false" primary:"true"`
}

// A method to get the columns of the struct
// Useful in GnoMessage interface
func (mr MsgRun) TableColumns() []string {
	columns := make([]string, 0)
	// get the fields of the struct
	fields := reflect.TypeOf(mr)
	numFields := fields.NumField()
	for i := range numFields {
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

// GetAllAddresses returns all the addresses that are involved in the message
// Groups the caller and signers for this transaction
//
// Returns:
//   - *TxAddresses: grouped addresses for this transaction
func (mr *MsgRun) GetAllAddresses() *TxAddresses {
	txAddresses := NewTxAddresses(mr.TxHash)
	txAddresses.AddAddress(mr.Caller)
	for _, addr := range mr.Signers {
		txAddresses.AddAddress(addr)
	}
	return txAddresses
}

// DBTable is an interface for structs that represent database tables
type DBTable interface {
	GetTableInfo() (*dbinit.TableInfo, error)
	TableName() string
}

// An interface for Gno messages
//
// Methods:
// - TableColumns() []string: a method to get the columns of the struct
type GnoMessage interface {
	TableColumns() []string
}
