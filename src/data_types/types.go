package datatypes

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
	ChainID string `db:"chain_id" dbtype:"chain_name" nullable:"false" primary:"false"`
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
	ChainID   string    `db:"chain_id" dbtype:"varchar" nullable:"false" primary:"false"`
	// proposer address is the validator address hence why this should be an integer
	ProposerAddress int32 `db:"proposer_address" dbtype:"integer" nullable:"false" primary:"false"`
	// Technically speaking we could set it to have a null in place however i think even null takes up space
	// so keep it like in the cosmos indexer
	Txs []string `db:"txs" dbtype:"[]TEXT" primary:"false"`
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
	ChainID     string    `db:"chain_id" dbtype:"chain_name" nullable:"false" primary:"true"` // use type enum chain_name from postgres
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
	ChainID   string    `db:"chain_id" dbtype:"chain_name" nullable:"false" primary:"false"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamp" nullable:"false" primary:"false"`
	MsgTypes  []string  `db:"msg_types" dbtype:"[]TEXT" nullable:"false" primary:"false"`
}

// TableName returns the name of the table for the AddressTx struct
func (at AddressTx) TableName() string {
	return "address_tx"
}
