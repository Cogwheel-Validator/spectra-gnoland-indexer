package sql_data_types

import "time"

// GnoAddress represents a Gno address with database mapping information
// There are 2 tables for this struct:
// - gno_addresses (for the regular addresses)
// - gno_validators (for the validator addresses)
//
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
func (g GnoAddress) TableName(valTable bool) string {
	if valTable {
		return "gno_validators"
	}
	return "gno_addresses"
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
	Timestamp time.Time `db:"timestamp" dbtype:"timestamp" nullable:"false" primary:"true"`
	ChainID   string    `db:"chain_id" dbtype:"TEXT" nullable:"false" primary:"false"`
	// proposer address is the validator address hence why this should be an integer
	ProposerAddress int32  `db:"proposer_address" dbtype:"integer" nullable:"false" primary:"false"`
	Txs             []byte `db:"txs" dbtype:"bytea" primary:"false" nullable:"true"` // can be a null value
	ChainName       string `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
}

func (b Blocks) TableName(valTable bool) string {
	if valTable {
		return "blocks_val"
	}
	return "blocks"
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
	Timestamp   time.Time `db:"timestamp" dbtype:"timestamp" nullable:"false" primary:"true"`
	SignedVals  []int32   `db:"address" dbtype:"integer" nullable:"false" primary:"false"`
	ChainName   string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"` // use type enum chain_name from postgres
	MissedVals  []int32   `db:"missed_vals" dbtype:"integer" nullable:"false" primary:"false"`
}

// TableName returns the name of the table for the ValidatorBlockSigning struct
func (vbs ValidatorBlockSigning) TableName() string {
	return "validator_block_signing"
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
	Timestamp time.Time `db:"timestamp" dbtype:"timestamp" nullable:"false" primary:"false"`
	MsgTypes  []string  `db:"msg_types" dbtype:"[]TEXT" nullable:"false" primary:"false"`
}

// TableName returns the name of the table for the AddressTx struct
func (at AddressTx) TableName() string {
	return "address_tx"
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

// TransactionGeneral represents a transaction general data with database mapping information
type TransactionGeneral struct {
	TxHash             []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"true"`
	ChainName          string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	Timestamp          time.Time `db:"timestamp" dbtype:"timestamp" nullable:"false" primary:"true"`
	MsgTypes           []string  `db:"msg_types" dbtype:"[]TEXT" nullable:"false" primary:"false"`
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

// MsgSend represents a bank send message
type MsgSend struct {
	TxHash      []byte `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"false"`
	ChainName   string `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
	FromAddress string `db:"from_address" dbtype:"TEXT" nullable:"false" primary:"false"`
	// need to test this out later leave it as a possible null value
	ToAddress string    `db:"to_address" dbtype:"TEXT" nullable:"true" primary:"false"`
	Amount    string    `db:"amount" dbtype:"TEXT" nullable:"false" primary:"false"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamp" nullable:"false" primary:"false"`
}

// TableName returns the name of the table for the MsgSend struct
func (ms MsgSend) TableName() string {
	return "bank_msg_send"
}

// MsgCall represents a VM function call message
type MsgCall struct {
	TxHash    []byte `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"false"`
	ChainName string `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
	Caller    string `db:"caller" dbtype:"TEXT" nullable:"false" primary:"false"`
	PkgPath   string `db:"pkg_path" dbtype:"TEXT" nullable:"false" primary:"false"`
	FuncName  string `db:"func_name" dbtype:"TEXT" nullable:"false" primary:"false"`
	// could be a null value but maybe return an empty array, leave false for now
	Args      []string  `db:"args" dbtype:"[]TEXT" nullable:"false" primary:"false"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamp" nullable:"false" primary:"false"`
}

func (mc MsgCall) TableName() string {
	return "vm_msg_call"
}

// MsgAddPackage represents a VM package addition message
type MsgAddPackage struct {
	TxHash    []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"false"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
	Creator   string    `db:"creator" dbtype:"TEXT" nullable:"false" primary:"false"`
	PkgPath   string    `db:"pkg_path" dbtype:"TEXT" nullable:"false" primary:"false"`
	PkgName   string    `db:"pkg_name" dbtype:"TEXT" nullable:"false" primary:"false"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamp" nullable:"false" primary:"false"`
}

func (ma MsgAddPackage) TableName() string {
	return "vm_msg_add_package"
}

// MsgRun represents a VM package run message
type MsgRun struct {
	TxHash    []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"false"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false"`
	Caller    string    `db:"caller" dbtype:"TEXT" nullable:"false" primary:"false"`
	PkgPath   string    `db:"pkg_path" dbtype:"TEXT" nullable:"false" primary:"false"`
	PkgName   string    `db:"pkg_name" dbtype:"TEXT" nullable:"false" primary:"false"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamp" nullable:"false" primary:"false"`
}

func (mr MsgRun) TableName() string {
	return "vm_msg_run"
}
