package generator

import (
	"fmt"
	"strconv"
	"time"

	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

// Data generated to be used for making a synthetic block response
type GenBlock struct {
	Height          uint64
	ChainID         string
	Timestamp       time.Time
	NumTxs          int
	ProposerAddress string
}

// NewTestBlockResponse creates a BlockResponse with sensible defaults for testing
func NewTestBlockResponse(
	height uint64,
	chainID string) *rpcClient.BlockResponse {
	heightStr := strconv.FormatUint(height, 10)
	timestamp := time.Now()

	return &rpcClient.BlockResponse{
		Jsonrpc: "2.0",
		ID:      1,
		Result: rpcClient.BlockResult{
			BlockMeta: rpcClient.BlockMeta{
				BlockID: rpcClient.BlockID{
					Hash: fmt.Sprintf("test-block-hash-%d", height),
					Parts: rpcClient.Parts{
						Total: "1",
						Hash:  fmt.Sprintf("test-parts-hash-%d", height),
					},
				},
				Header: rpcClient.BlockHeader{
					Version:         "test-version",
					ChainID:         chainID,
					Height:          heightStr,
					Time:            timestamp,
					NumTxs:          "0",
					TotalTxs:        heightStr,
					AppVersion:      "test-app",
					ProposerAddress: "test-proposer-address",
				},
			},
			Block: rpcClient.BlockInfo{
				Header: rpcClient.BlockHeader{
					Version:         "test-version",
					ChainID:         chainID,
					Height:          heightStr,
					Time:            timestamp,
					NumTxs:          "0",
					TotalTxs:        heightStr,
					AppVersion:      "test-app",
					ProposerAddress: "test-proposer-address",
				},
				Data: BlockData{
					Txs: nil, // No transactions by default
				},
				LastCommit: nil,
			},
		},
	}
}

// WithTransactions adds transaction hashes to a test BlockResponse
func (br *BlockResponse) WithTransactions(txHashes []string) *BlockResponse {
	br.Result.Block.Data.Txs = &txHashes
	br.Result.Block.Header.NumTxs = strconv.Itoa(len(txHashes))
	br.Result.BlockMeta.Header.NumTxs = strconv.Itoa(len(txHashes))
	return br
}

// WithPrecommits adds precommits to a test BlockResponse
func (br *BlockResponse) WithPrecommits(precommits []*Precommit) *BlockResponse {
	br.Result.Block.LastCommit.Precommits = precommits
	return br
}

// NewTestTxResponse creates a TxResponse with sensible defaults for testing
func NewTestTxResponse(hash string, height uint64) *TxResponse {
	heightStr := strconv.FormatUint(height, 10)

	return &TxResponse{
		Jsonrpc: "2.0",
		ID:      1,
		Result: TxResultData{
			Hash:   hash,
			Height: heightStr,
			Index:  0,
			TxResult: TxResult{
				ResponseBase: ResponseBase{
					Error:  nil,
					Data:   "test-data",
					Events: []Event{}, // No events by default
					Log:    "test-log",
					Info:   "test-info",
				},
				GasWanted: "100000",
				GasUsed:   "50000",
			},
			Tx: "test-tx-data",
		},
	}
}

// WithEvents adds events to a test TxResponse
func (tr *TxResponse) WithEvents(events []Event) *TxResponse {
	tr.Result.TxResult.ResponseBase.Events = events
	return tr
}

// WithError adds an error to a test TxResponse
func (tr *TxResponse) WithError(err interface{}) *TxResponse {
	tr.Result.TxResult.ResponseBase.Error = err
	return tr
}

// NewTestEvent creates a test Event with sensible defaults
// should be part of the integration tests but keep it here for now as an idea
func NewTestEvent(eventType string, pkgPath string) Event {
	return Event{
		AtType:  "test-type",
		Type:    eventType,
		PkgPath: pkgPath,
		Attrs: []EventAttribute{
			{Key: "test-key", Value: "test-value"},
		},
	}
}

// NewTestPrecommit creates a test Precommit with sensible defaults
// should be part of the integration tests but keep it here for now as an idea
func NewTestPrecommit(validatorAddress string, height uint64) *Precommit {
	return &Precommit{
		Type:             1,
		Height:           strconv.FormatUint(height, 10),
		Round:            "0",
		BlockID:          BlockID{Hash: fmt.Sprintf("test-hash-%d", height), Parts: Parts{Total: "1", Hash: "test-parts"}},
		Timestamp:        time.Now(),
		ValidatorAddress: validatorAddress,
		ValidatorIndex:   "0",
		Signature:        "test-signature",
	}
}
