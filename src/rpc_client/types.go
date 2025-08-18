package rpcclient

import (
	"net/rpc"
	"time"
)

type RpcGnoland struct {
	rpcURL  string
	port    *int
	client  *rpc.Client
	timeout time.Duration
}
type ValidatorsResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
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
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
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
				LastCommitHash     string      `json:"last_commit_hash"`
				DataHash           string      `json:"data_hash"`
				ValidatorsHash     string      `json:"validators_hash"`
				NextValidatorsHash string      `json:"next_validators_hash"`
				ConsensusHash      string      `json:"consensus_hash"`
				AppHash            string      `json:"app_hash"`
				LastResultsHash    interface{} `json:"last_results_hash"`
				ProposerAddress    string      `json:"proposer_address"`
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
				LastCommitHash     string      `json:"last_commit_hash"`
				DataHash           string      `json:"data_hash"`
				ValidatorsHash     string      `json:"validators_hash"`
				NextValidatorsHash string      `json:"next_validators_hash"`
				ConsensusHash      string      `json:"consensus_hash"`
				AppHash            string      `json:"app_hash"`
				LastResultsHash    interface{} `json:"last_results_hash"`
				ProposerAddress    string      `json:"proposer_address"`
			} `json:"header"`
			Data struct {
				Txs []string `json:"txs"`
			} `json:"data"`
			LastCommit struct {
				BlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total string `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"block_id"`
				Precommits []struct {
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
				} `json:"precommits"`
			} `json:"last_commit"`
		} `json:"block"`
	} `json:"result"`
}

type HealthResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
	} `json:"result"`
}
