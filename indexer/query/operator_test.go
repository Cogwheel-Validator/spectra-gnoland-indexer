package query_test

import (
	"sync"
	"testing"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/query"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	"github.com/stretchr/testify/assert"
)

// MockRpcClient - focuses on tracking what was called
type MockRpcClient struct {
	mu                         sync.Mutex
	GetBlockCalled             bool
	GetLatestBlockHeightCalled bool
	GetTxCalled                bool
	GetBlockCallCount          int
	GetTxCallCount             int
	GetCommitCalled            bool
	GetCommitCallCount         int
}

// Mock method for GetBlock
func (m *MockRpcClient) GetBlock(height uint64) (*rpcClient.BlockResponse, *rpcClient.RpcHeightError) {
	m.mu.Lock()
	m.GetBlockCalled = true
	m.GetBlockCallCount++
	m.mu.Unlock()
	return &rpcClient.BlockResponse{}, nil
}

// Mock method for GetLatestBlockHeight
func (m *MockRpcClient) GetLatestBlockHeight() (uint64, *rpcClient.RpcHeightError) {
	m.mu.Lock()
	m.GetLatestBlockHeightCalled = true
	m.mu.Unlock()
	return 1, nil
}

// Mock method for GetTx
func (m *MockRpcClient) GetTx(txHash string) (*rpcClient.TxResponse, *rpcClient.RpcStringError) {
	m.mu.Lock()
	m.GetTxCalled = true
	m.GetTxCallCount++
	m.mu.Unlock()
	return &rpcClient.TxResponse{}, nil
}

// Mock method for GetCommit
func (m *MockRpcClient) GetCommit(height uint64) (*rpcClient.CommitResponse, *rpcClient.RpcCommitError) {
	m.mu.Lock()
	m.GetCommitCalled = true
	m.GetCommitCallCount++
	m.mu.Unlock()
	return &rpcClient.CommitResponse{}, nil
}

// TestQueryOperator - tests the query operator
func TestQueryOperator(t *testing.T) {
	mockRpcClient := &MockRpcClient{}
	queryOperator := query.NewQueryOperator(mockRpcClient, nil, nil, nil, nil) // should be overwritten by the constructor

	// Test GetFromToBlocks - should call GetBlock multiple times (1 to 10 = 10 calls)
	queryOperator.GetFromToBlocks(1, 10)
	assert.True(t, mockRpcClient.GetBlockCalled)
	assert.Equal(t, 10, mockRpcClient.GetBlockCallCount)

	// Reset mock for next test
	mockRpcClient.GetLatestBlockHeightCalled = false

	// Test GetLatestBlockHeight
	_, err := queryOperator.GetLatestBlockHeight()
	assert.NoError(t, err)
	assert.True(t, mockRpcClient.GetLatestBlockHeightCalled)

	// Test GetTransactions - should call GetTx for each transaction hash
	txHashes := []string{"txHash1", "txHash2", "txHash3"}
	queryOperator.GetTransactions(txHashes)
	assert.True(t, mockRpcClient.GetTxCalled)
	assert.Equal(t, len(txHashes), mockRpcClient.GetTxCallCount)

	// Test GetFromToCommits - should call GetCommit multiple times (1 to 10 = 10 calls)
	queryOperator.GetFromToCommits(1, 10)
	assert.True(t, mockRpcClient.GetCommitCalled)

	assert.Equal(t, 10, mockRpcClient.GetCommitCallCount)
}
