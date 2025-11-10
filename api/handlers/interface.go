package handlers

import (
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
)

// DatabaseHandler interface for the database
type DatabaseHandler interface {
	// Addresses
	GetAddressTxs(address string, chainName string, fromTimestamp time.Time, toTimestamp time.Time) (*[]database.AddressTx, error)

	// Blocks
	GetBlock(height uint64, chainName string) (*database.BlockData, error)
	GetFromToBlocks(fromHeight uint64, toHeight uint64, chainName string) ([]*database.BlockData, error)
	GetAllBlockSigners(chainName string, blockHeight uint64) (*database.BlockSigners, error)
	GetLatestBlock(chainName string) (*database.BlockData, error)
	GetLastXBlocks(chainName string, x uint64) ([]*database.BlockData, error)

	// Transactions
	GetTransaction(txHash string, chainName string) (*database.Transaction, error)
	GetLastXTransactions(chainName string, x uint64) ([]*database.Transaction, error)
	GetMsgType(txHash string, chainName string) (string, error)
	GetBankSend(txHash string, chainName string) (*database.BankSend, error)
	GetMsgCall(txHash string, chainName string) (*database.MsgCall, error)
	GetMsgAddPackage(txHash string, chainName string) (*database.MsgAddPackage, error)
	GetMsgRun(txHash string, chainName string) (*database.MsgRun, error)
}
