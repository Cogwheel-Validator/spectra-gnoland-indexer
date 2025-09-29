package query_test

import (
	"testing"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/query"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

// MockRpcClient - focuses on tracking what was called
type MockRpcClient struct {
	GetBlockCalled             bool
	GetLatestBlockHeightCalled bool
	GetTxCalled                bool
	GetBlockCallCount          int
	GetTxCallCount             int
}

// Mock method for GetBlock
func (m *MockRpcClient) GetBlock(height uint64) (*rpcClient.BlockResponse, *rpcClient.RpcHeightError) {
	m.GetBlockCalled = true
	m.GetBlockCallCount++
	return &rpcClient.BlockResponse{}, nil
}

// Mock method for GetLatestBlockHeight
func (m *MockRpcClient) GetLatestBlockHeight() (uint64, *rpcClient.RpcHeightError) {
	m.GetLatestBlockHeightCalled = true
	return 1, nil
}

// Mock method for GetTx
func (m *MockRpcClient) GetTx(txHash string) (*rpcClient.TxResponse, *rpcClient.RpcStringError) {
	m.GetTxCalled = true
	m.GetTxCallCount++
	return &rpcClient.TxResponse{}, nil
}

// TestQueryOperator - tests the query operator
func TestQueryOperator(t *testing.T) {
	mockRpcClient := &MockRpcClient{}
	queryOperator := query.NewQueryOperator(mockRpcClient, nil, nil, nil, nil) // should be overwritten by the constructor

	// Test GetFromToBlocks - should call GetBlock multiple times (1 to 10 = 10 calls)
	queryOperator.GetFromToBlocks(1, 10)
	if !mockRpcClient.GetBlockCalled {
		t.Errorf("GetFromToBlocks should call GetBlock")
	}
	if mockRpcClient.GetBlockCallCount != 10 {
		t.Errorf("GetFromToBlocks(1, 10) should call GetBlock 10 times, got %d", mockRpcClient.GetBlockCallCount)
	}

	// Reset mock for next test
	mockRpcClient.GetLatestBlockHeightCalled = false

	// Test GetLatestBlockHeight
	_, err := queryOperator.GetLatestBlockHeight()
	if err != nil {
		t.Errorf("GetLatestBlockHeight should not return an error")
	}
	if !mockRpcClient.GetLatestBlockHeightCalled {
		t.Errorf("GetLatestBlockHeight should be called")
	}

	// Test GetTransactions - should call GetTx for each transaction hash
	txHashes := []string{"txHash1", "txHash2", "txHash3"}
	queryOperator.GetTransactions(txHashes)
	if !mockRpcClient.GetTxCalled {
		t.Errorf("GetTransactions should call GetTx")
	}
	if mockRpcClient.GetTxCallCount != len(txHashes) {
		t.Errorf("GetTransactions should call GetTx %d times, got %d", len(txHashes), mockRpcClient.GetTxCallCount)
	}
}
