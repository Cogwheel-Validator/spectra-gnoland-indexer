package handlers

import (
	"context"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
)

// DatabaseHandler interface for the database
type DatabaseHandler interface {
	// Addresses
	GetAddressTxs(ctx context.Context, address string, chainName string, fromTimestamp time.Time, toTimestamp time.Time) (*[]database.AddressTx, error)

	// Blocks
	GetBlock(ctx context.Context, height uint64, chainName string) (*database.BlockData, error)
	GetFromToBlocks(ctx context.Context, fromHeight uint64, toHeight uint64, chainName string) ([]*database.BlockData, error)
	GetAllBlockSigners(ctx context.Context, chainName string, blockHeight uint64) (*database.BlockSigners, error)
	GetLatestBlock(ctx context.Context, chainName string) (*database.BlockData, error)
	GetLastXBlocks(ctx context.Context, chainName string, x uint64) ([]*database.BlockData, error)

	// Transactions
	GetTransaction(ctx context.Context, txHash string, chainName string) (*database.Transaction, error)
	GetLastXTransactions(ctx context.Context, chainName string, x uint64) ([]*database.Transaction, error)
	GetMsgTypes(ctx context.Context, txHash string, chainName string) ([]string, error)
	GetBankSend(ctx context.Context, txHash string, chainName string) ([]*database.BankSend, error)
	GetMsgCall(ctx context.Context, txHash string, chainName string) ([]*database.MsgCall, error)
	GetMsgAddPackage(ctx context.Context, txHash string, chainName string) ([]*database.MsgAddPackage, error)
	GetMsgRun(ctx context.Context, txHash string, chainName string) ([]*database.MsgRun, error)
}
