package query

import (
	rpcclient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

type QueryOperator struct {
	rpcClient *rpcclient.RateLimitedRpcClient
}
