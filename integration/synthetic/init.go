package synthetic

import (
	"log"

	addressCache "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/address_cache"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	dataProcessor "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/data_processor"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/orchestrator"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

// SyntheticIntegrationTestConfig holds configuration for synthetic integration tests
type SyntheticIntegrationTestConfig struct {
	DatabaseConfig database.DatabasePoolConfig
	ChainID        string
	MaxHeight      uint64
	FromHeight     uint64
	ToHeight       uint64
}

// RunSyntheticIntegrationTest runs a full integration test using synthetic data
// This tests the entire orchestrator pipeline with generated data and a real database
func RunSyntheticIntegrationTest(testConfig *SyntheticIntegrationTestConfig) error {
	log.Printf("Starting synthetic integration test from height %d to %d", testConfig.FromHeight, testConfig.ToHeight)

	// Initialize database
	db := database.NewTimescaleDb(testConfig.DatabaseConfig)
	log.Printf("Connected to database successfully")

	// Initialize address caches (required by data processor)
	validatorCache := addressCache.NewAddressCache(testConfig.ChainID, db, true)
	addrCache := addressCache.NewAddressCache(testConfig.ChainID, db, false)
	log.Printf("Initialized address caches")

	// Initialize data processor
	dataProc := dataProcessor.NewDataProcessor(db, addrCache, validatorCache, testConfig.ChainID)
	log.Printf("Initialized data processor")

	// Create synthetic query operator (this replaces the real RPC queries!)
	syntheticQueryOp := NewSyntheticQueryOperator(testConfig.ChainID, testConfig.MaxHeight)
	log.Printf("Created synthetic query operator with max height %d", testConfig.MaxHeight)

	// Create a mock database height interface for the orchestrator
	mockDbHeight := &MockDatabaseHeight{lastHeight: testConfig.FromHeight - 1}

	// Create a mock gnoland rpc clien
	mockGnoRpc := &MockGnolandRpcClient{latestHeight: testConfig.MaxHeight}

	// Create basic config for orchestrator
	orchConfig := &config.Config{
		MaxBlockChunkSize: 50, // Cap it to 50 for now
	}

	// Create orchestrator with synthetic components
	orch := orchestrator.NewOrchestrator(
		"historic", // Run in historic mode
		orchConfig,
		testConfig.ChainID,
		mockDbHeight,
		mockGnoRpc,
		dataProc,
		syntheticQueryOp, // Synthetic query operator
	)

	log.Printf("Created orchestrator with synthetic components")

	// Run the historic process - this will use synthetic data but process it through
	// the real data processor and store it in the real database
	orch.HistoricProcess(testConfig.FromHeight, testConfig.ToHeight)

	log.Printf("Synthetic integration test completed successfully!")
	return nil
}

// Mock implementations for the interfaces that orchestrator needs

// MockDatabaseHeight implements the DatabaseHeight interface
type MockDatabaseHeight struct {
	lastHeight uint64
}

func (m *MockDatabaseHeight) GetLastBlockHeight(chainName string) (uint64, error) {
	return m.lastHeight, nil
}

// MockGnolandRpcClient implements the GnolandRpcClient interface
type MockGnolandRpcClient struct {
	latestHeight uint64
}

func (m *MockGnolandRpcClient) GetLatestBlockHeight() (uint64, *rpcClient.RpcHeightError) {
	return m.latestHeight, nil
}
