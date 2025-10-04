package dataprocessor

import (
	"time"

	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	sqlDataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
)

// Define interface for what DataProcessor needs from database
type Database interface {
	InsertBlocks(blocks []sqlDataTypes.Blocks) error
	InsertValidatorBlockSignings(validatorBlockSignings []sqlDataTypes.ValidatorBlockSigning) error
	InsertTransactionsGeneral(transactionsGeneral []sqlDataTypes.TransactionGeneral) error
	InsertMsgSend(messages []sqlDataTypes.MsgSend) error
	InsertMsgCall(messages []sqlDataTypes.MsgCall) error
	InsertMsgAddPackage(messages []sqlDataTypes.MsgAddPackage) error
	InsertMsgRun(messages []sqlDataTypes.MsgRun) error
	InsertAddressTx(addresses []sqlDataTypes.AddressTx) error
}

// Define interface for what DataProcessor needs from AddressCache
type AddressCache interface {
	AddressSolver(address []string, chainName string, insertValidators bool, retryAttempts uint8, oneByOne *bool)
	GetAddress(address string) int32
}

type DataProcessor struct {
	dbPool         Database
	addressCache   AddressCache
	validatorCache AddressCache
	chainName      string
}

type TrasnactionsData struct {
	Response    *rpcClient.TxResponse
	Timestamp   time.Time
	BlockHeight uint64
}
