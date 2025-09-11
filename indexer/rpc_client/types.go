package rpcclient

import (
	"net/http"
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
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Error   *JsonRpcError `json:"error,omitempty"`
	Result  struct {
		BlockHeight string `json:"block_height"`
		Validators  []struct {
			Address string `json:"address"`
			PubKey  struct {
				Type  string `json:"@type"`
				Value string `json:"value"`
			} `json:"pub_key"`
			VotingPower      string `json:"voting_power"`
			ProposerPriority string `json:"proposer_priority"`
		} `json:"validators"`
	} `json:"result"`
}

type BlockResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Error   *JsonRpcError `json:"error,omitempty"`
	Result  struct {
		BlockMeta struct {
			BlockID struct {
				Hash  string `json:"hash"`
				Parts struct {
					Total string `json:"total"`
					Hash  string `json:"hash"`
				} `json:"parts"`
			} `json:"block_id"`
			Header struct {
				Version     string    `json:"version"`
				ChainID     string    `json:"chain_id"`
				Height      string    `json:"height"`
				Time        time.Time `json:"time"`
				NumTxs      string    `json:"num_txs"`
				TotalTxs    string    `json:"total_txs"`
				AppVersion  string    `json:"app_version"`
				LastBlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total string `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"last_block_id"`
				LastCommitHash     string `json:"last_commit_hash"`
				DataHash           string `json:"data_hash"`
				ValidatorsHash     string `json:"validators_hash"`
				NextValidatorsHash string `json:"next_validators_hash"`
				ConsensusHash      string `json:"consensus_hash"`
				AppHash            string `json:"app_hash"`
				LastResultsHash    any    `json:"last_results_hash"`
				ProposerAddress    string `json:"proposer_address"`
			} `json:"header"`
		} `json:"block_meta"`
		Block struct {
			Header struct {
				Version     string    `json:"version"`
				ChainID     string    `json:"chain_id"`
				Height      string    `json:"height"`
				Time        time.Time `json:"time"`
				NumTxs      string    `json:"num_txs"`
				TotalTxs    string    `json:"total_txs"`
				AppVersion  string    `json:"app_version"`
				LastBlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total string `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"last_block_id"`
				LastCommitHash     string `json:"last_commit_hash"`
				DataHash           string `json:"data_hash"`
				ValidatorsHash     string `json:"validators_hash"`
				NextValidatorsHash string `json:"next_validators_hash"`
				ConsensusHash      string `json:"consensus_hash"`
				AppHash            string `json:"app_hash"`
				LastResultsHash    any    `json:"last_results_hash"`
				ProposerAddress    string `json:"proposer_address"`
			} `json:"header"`
			Data struct {
				Txs *[]string `json:"txs"` // it can be a slice of strings or nil
			} `json:"data"`
			LastCommit struct {
				BlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total string `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"block_id"`
				Precommits []*struct { // some of the slices can be nil
					Type    int    `json:"type"`
					Height  string `json:"height"`
					Round   string `json:"round"`
					BlockID struct {
						Hash  string `json:"hash"`
						Parts struct {
							Total string `json:"total"`
							Hash  string `json:"hash"`
						} `json:"parts"`
					} `json:"block_id"`
					Timestamp        time.Time `json:"timestamp"`
					ValidatorAddress string    `json:"validator_address"`
					ValidatorIndex   string    `json:"validator_index"`
					Signature        string    `json:"signature"`
				} `json:"precommits"` // some of the slices can be nil
			} `json:"last_commit"`
		} `json:"block"`
	} `json:"result"`
}

type HealthResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Error   *JsonRpcError `json:"error,omitempty"`
	Result  interface{}   `json:"result"`
}

type TxResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Error   *JsonRpcError `json:"error,omitempty"`
	Result  struct {
		Hash     string `json:"hash"`
		Height   string `json:"height"`
		Index    int    `json:"index"`
		TxResult struct {
			ResponseBase struct {
				Error  interface{} `json:"Error"`
				Data   string      `json:"Data"`
				Events []struct {
					AtType string `json:"@type"`
					Type   string `json:"type"`
					Attrs  []struct {
						Key   string `json:"key"`
						Value string `json:"value"`
					} `json:"attrs"`
					PkgPath string `json:"pkg_path"`
				} `json:"Events"`
				Log  string `json:"Log"`
				Info string `json:"Info"`
			} `json:"ResponseBase"`
			GasWanted string `json:"GasWanted"`
			GasUsed   string `json:"GasUsed"`
		} `json:"tx_result"`
		Tx string `json:"tx"`
	} `json:"result"`
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
