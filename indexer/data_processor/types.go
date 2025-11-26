package dataprocessor

import (
	"context"
	"time"

	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	sqlDataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
)

// Define interface for what DataProcessor needs from database
type Database interface {
	InsertBlocks(ctx context.Context, blocks []sqlDataTypes.Blocks) error
	InsertValidatorBlockSignings(ctx context.Context, validatorBlockSignings []sqlDataTypes.ValidatorBlockSigning) error
	InsertTransactionsGeneral(ctx context.Context, transactionsGeneral []sqlDataTypes.TransactionGeneral) error
	InsertMsgSend(ctx context.Context, messages []sqlDataTypes.MsgSend) error
	InsertMsgCall(ctx context.Context, messages []sqlDataTypes.MsgCall) error
	InsertMsgAddPackage(ctx context.Context, messages []sqlDataTypes.MsgAddPackage) error
	InsertMsgRun(ctx context.Context, messages []sqlDataTypes.MsgRun) error
	InsertAddressTx(ctx context.Context, addresses []sqlDataTypes.AddressTx) error
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
