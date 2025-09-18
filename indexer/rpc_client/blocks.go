package rpcclient

import (
	"fmt"
	"strconv"
	"time"
)

// BlockID is part of the struct for the block result
type BlockID struct {
	Hash  string `json:"hash"`
	Parts Parts  `json:"parts"`
}

// Parts is part of the struct for the block ID
type Parts struct {
	Total string `json:"total"`
	Hash  string `json:"hash"`
}

// BlockHeader is part of the struct for the block result
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

// BlockData is part of the struct for the block result
// Usually stored as a slice of strings or nil
type BlockData struct {
	Txs *[]string `json:"txs"` // it can be a slice of strings or nil
}

// Precommit is part of the struct for the block result
// Usually stored as a slice of precommits
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

// LastCommit is part of the struct for the block result
type LastCommit struct {
	BlockID    BlockID      `json:"block_id"`
	Precommits []*Precommit `json:"precommits"` // some of the slices can be nil
}

// BlockMeta is part of the struct for the block result
type BlockMeta struct {
	BlockID BlockID     `json:"block_id"`
	Header  BlockHeader `json:"header"`
}

// BlockInfo is part of the struct for the block result
type BlockInfo struct {
	Header     BlockHeader `json:"header"`
	Data       BlockData   `json:"data"`
	LastCommit LastCommit  `json:"last_commit"`
}

// BlockResult is part of the struct for the block response
type BlockResult struct {
	BlockMeta BlockMeta `json:"block_meta"`
	Block     BlockInfo `json:"block"`
}

// BlockResponse is the response from the rpc client for the block method
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
