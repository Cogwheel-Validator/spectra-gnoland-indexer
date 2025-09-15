package rpcclient

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rate_limit"
)

type RpcGnoland struct {
	rpcURL string
	client *http.Client
}

type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ValidatorsResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Error   *JsonRpcError   `json:"error,omitempty"`
	Result  ValidatorResult `json:"result"`
}

type ValidatorResult struct {
	BlockHeight string            `json:"block_height"`
	Validators  []ValidatorsSlice `json:"validators"`
}

type ValidatorsSlice struct {
	Address          string    `json:"address"`
	PubKey           ValPubKey `json:"pub_key"`
	VotingPower      string    `json:"voting_power"`
	ProposerPriority string    `json:"proposer_priority"`
}

type ValPubKey struct {
	Type  string `json:"@type"`
	Value string `json:"value"`
}

// Named structs for better testability and reusability
type BlockID struct {
	Hash  string `json:"hash"`
	Parts Parts  `json:"parts"`
}

type Parts struct {
	Total string `json:"total"`
	Hash  string `json:"hash"`
}

type BlockHeader struct {
	Version            string    `json:"version"`
	ChainID            string    `json:"chain_id"`
	Height             string    `json:"height"`
	Time               time.Time `json:"time"`
	NumTxs             string    `json:"num_txs"`
	TotalTxs           string    `json:"total_txs"`
	AppVersion         string    `json:"app_version"`
	LastBlockID        BlockID   `json:"last_block_id"`
	LastCommitHash     string    `json:"last_commit_hash"`
	DataHash           string    `json:"data_hash"`
	ValidatorsHash     string    `json:"validators_hash"`
	NextValidatorsHash string    `json:"next_validators_hash"`
	ConsensusHash      string    `json:"consensus_hash"`
	AppHash            string    `json:"app_hash"`
	LastResultsHash    any       `json:"last_results_hash"`
	ProposerAddress    string    `json:"proposer_address"`
}

type BlockData struct {
	Txs *[]string `json:"txs"` // it can be a slice of strings or nil
}

type Precommit struct {
	Type             int       `json:"type"`
	Height           string    `json:"height"`
	Round            string    `json:"round"`
	BlockID          BlockID   `json:"block_id"`
	Timestamp        time.Time `json:"timestamp"`
	ValidatorAddress string    `json:"validator_address"`
	ValidatorIndex   string    `json:"validator_index"`
	Signature        string    `json:"signature"`
}

type LastCommit struct {
	BlockID    BlockID      `json:"block_id"`
	Precommits []*Precommit `json:"precommits"` // some of the slices can be nil
}

type BlockMeta struct {
	BlockID BlockID     `json:"block_id"`
	Header  BlockHeader `json:"header"`
}

type BlockInfo struct {
	Header     BlockHeader `json:"header"`
	Data       BlockData   `json:"data"`
	LastCommit LastCommit  `json:"last_commit"`
}

type BlockResult struct {
	BlockMeta BlockMeta `json:"block_meta"`
	Block     BlockInfo `json:"block"`
}

type BlockResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Error   *JsonRpcError `json:"error,omitempty"`
	Result  BlockResult   `json:"result"`
}

// Helper methods for BlockResponse to safely access nested data
func (br *BlockResponse) GetHeight() (uint64, error) {
	if br == nil {
		return 0, fmt.Errorf("BlockResponse is nil")
	}
	return strconv.ParseUint(br.Result.Block.Header.Height, 10, 64)
}

func (br *BlockResponse) GetTimestamp() time.Time {
	if br == nil {
		return time.Time{}
	}
	return br.Result.Block.Header.Time
}

func (br *BlockResponse) GetChainID() string {
	if br == nil {
		return ""
	}
	return br.Result.Block.Header.ChainID
}

func (br *BlockResponse) GetProposerAddress() string {
	if br == nil {
		return ""
	}
	return br.Result.Block.Header.ProposerAddress
}

func (br *BlockResponse) GetBlockHash() string {
	if br == nil {
		return ""
	}
	return br.Result.BlockMeta.BlockID.Hash
}

func (br *BlockResponse) GetTxHashes() []string {
	if br == nil || br.Result.Block.Data.Txs == nil {
		return nil
	}
	return *br.Result.Block.Data.Txs
}

func (br *BlockResponse) GetPrecommits() []*Precommit {
	if br == nil {
		return nil
	}
	return br.Result.Block.LastCommit.Precommits
}

func (br *BlockResponse) IsValid() bool {
	return br != nil && br.Error == nil
}

type HealthResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Error   *JsonRpcError `json:"error,omitempty"`
	Result  interface{}   `json:"result"`
}

// Named structs for TxResponse
type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Event struct {
	AtType  string           `json:"@type"`
	Type    string           `json:"type"`
	Attrs   []EventAttribute `json:"attrs"`
	PkgPath string           `json:"pkg_path"`
}

type ResponseBase struct {
	Error  interface{} `json:"Error"`
	Data   string      `json:"Data"`
	Events []Event     `json:"Events"`
	Log    string      `json:"Log"`
	Info   string      `json:"Info"`
}

type TxResult struct {
	ResponseBase ResponseBase `json:"ResponseBase"`
	GasWanted    string       `json:"GasWanted"`
	GasUsed      string       `json:"GasUsed"`
}

type TxResultData struct {
	Hash     string   `json:"hash"`
	Height   string   `json:"height"`
	Index    int      `json:"index"`
	TxResult TxResult `json:"tx_result"`
	Tx       string   `json:"tx"`
}

type TxResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Error   *JsonRpcError `json:"error,omitempty"`
	Result  TxResultData  `json:"result"`
}

// Helper methods for TxResponse to safely access nested data
func (tr *TxResponse) GetHash() string {
	if tr == nil {
		return ""
	}
	return tr.Result.Hash
}

func (tr *TxResponse) GetHeight() (uint64, error) {
	if tr == nil {
		return 0, fmt.Errorf("TxResponse is nil")
	}
	return strconv.ParseUint(tr.Result.Height, 10, 64)
}

func (tr *TxResponse) GetEvents() []Event {
	if tr == nil {
		return nil
	}
	return tr.Result.TxResult.ResponseBase.Events
}

func (tr *TxResponse) GetGasWanted() (uint64, error) {
	if tr == nil {
		return 0, fmt.Errorf("TxResponse is nil")
	}
	return strconv.ParseUint(tr.Result.TxResult.GasWanted, 10, 64)
}

func (tr *TxResponse) GetGasUsed() (uint64, error) {
	if tr == nil {
		return 0, fmt.Errorf("TxResponse is nil")
	}
	return strconv.ParseUint(tr.Result.TxResult.GasUsed, 10, 64)
}

func (tr *TxResponse) GetTx() string {
	if tr == nil {
		return ""
	}
	return tr.Result.Tx
}

func (tr *TxResponse) GetIndex() int {
	if tr == nil {
		return 0
	}
	return tr.Result.Index
}

func (tr *TxResponse) IsValid() bool {
	return tr != nil && tr.Error == nil
}

func (tr *TxResponse) HasError() bool {
	return tr != nil && tr.Result.TxResult.ResponseBase.Error != nil
}

// Test Builder Functions for easy synthetic data creation

// NewTestBlockResponse creates a BlockResponse with sensible defaults for testing
func NewTestBlockResponse(height uint64, chainID string) *BlockResponse {
	heightStr := strconv.FormatUint(height, 10)
	timestamp := time.Now()

	return &BlockResponse{
		Jsonrpc: "2.0",
		ID:      1,
		Result: BlockResult{
			BlockMeta: BlockMeta{
				BlockID: BlockID{
					Hash: fmt.Sprintf("test-block-hash-%d", height),
					Parts: Parts{
						Total: "1",
						Hash:  fmt.Sprintf("test-parts-hash-%d", height),
					},
				},
				Header: BlockHeader{
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
			Block: BlockInfo{
				Header: BlockHeader{
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
				LastCommit: LastCommit{
					BlockID: BlockID{
						Hash: fmt.Sprintf("test-prev-hash-%d", height-1),
						Parts: Parts{
							Total: "1",
							Hash:  fmt.Sprintf("test-prev-parts-%d", height-1),
						},
					},
					Precommits: nil, // No precommits by default
				},
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

type Client interface {
	Health() error
	GetValidators(height uint64) (*ValidatorsResponse, *RpcHeightError)
	GetBlock(height uint64) (*BlockResponse, *RpcHeightError)
	GetLatestBlockHeight() (uint64, *RpcHeightError)
	GetTx(txHash string) (*TxResponse, *RpcStringError)
	GetAbciQuery(path string, data string, height *uint64, prove *bool) (any, error)
}

type RateLimiter interface {
	Allow() bool
	Wait()
	Close()
	GetStatus() rate_limit.ChannelRateLimiterStatus
}

// RateLimitedRpcClient wraps the original RPC client with rate limiting
//
// The struct contains the client and the rate limiter
// The client is the original RPC client
// The rate limiter is the rate limiter for the client
type RateLimitedRpcClient struct {
	client      Client
	rateLimiter RateLimiter
}
