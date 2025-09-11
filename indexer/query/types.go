package query

import (
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

type QueryOperator struct {
	rpcClient RpcClient
}

type RpcClient interface {
	GetBlock(height uint64) (*rpcClient.BlockResponse, *rpcClient.RpcHeightError)
	GetLatestBlockHeight() (uint64, *rpcClient.RpcHeightError)
	GetTx(txHash string) (*rpcClient.TxResponse, *rpcClient.RpcStringError)
}
