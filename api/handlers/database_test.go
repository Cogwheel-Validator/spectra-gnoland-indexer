package handlers_test

import (
	"fmt"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
)

type MockDatabase struct {
	blocks       map[uint64]*database.BlockData
	transactions map[string]*database.Transaction
	addressTxs   map[string]*[]database.AddressTx
	blockSigners map[uint64]*database.BlockSigners
	latestBlock  *database.BlockData

	bankSend      map[string]*database.BankSend
	msgCall       map[string]*database.MsgCall
	msgAddPackage map[string]*database.MsgAddPackage
	msgRun        map[string]*database.MsgRun
	msgTypes      map[string]string

	shouldError bool
	errorMsg    string
}

func (m *MockDatabase) GetBlock(height uint64, chainName string) (*database.BlockData, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	block, ok := m.blocks[height]
	if !ok {
		return nil, fmt.Errorf("block not found")
	}
	return block, nil
}

func (m *MockDatabase) GetFromToBlocks(fromHeight uint64, toHeight uint64, chainName string) ([]*database.BlockData, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	var result []*database.BlockData
	for i := fromHeight; i <= toHeight; i++ {
		block, ok := m.blocks[i]
		if !ok {
			return nil, fmt.Errorf("block not found")
		}
		result = append(result, block)
	}
	return result, nil
}

func (m *MockDatabase) GetAllBlockSigners(chainName string, blockHeight uint64) (*database.BlockSigners, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	blockSigners, ok := m.blockSigners[blockHeight]
	if !ok {
		return nil, fmt.Errorf("block signers not found")
	}
	return blockSigners, nil
}

func (m *MockDatabase) GetTransaction(txHash string, chainName string) (*database.Transaction, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	transaction, ok := m.transactions[txHash]
	if !ok {
		return nil, fmt.Errorf("transaction not found")
	}
	return transaction, nil
}

func (m *MockDatabase) GetAddressTxs(address string, chainName string, fromTimestamp time.Time, toTimestamp time.Time) (*[]database.AddressTx, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	addressTxs, ok := m.addressTxs[address]
	if !ok {
		return nil, fmt.Errorf("address transactions not found")
	}
	return addressTxs, nil
}

func (m *MockDatabase) GetLatestBlockHeight(chainName string) (*database.BlockData, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	if m.latestBlock == nil {
		return nil, fmt.Errorf("latest block not found")
	}
	return m.latestBlock, nil
}

func (m *MockDatabase) GetLastXBlocks(chainName string, x uint64) ([]*database.BlockData, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	blocks := make([]*database.BlockData, 0, len(m.blocks))
	for _, block := range m.blocks {
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (m *MockDatabase) GetLastXTransactions(chainName string, x uint64) ([]*database.Transaction, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	transactions := make([]*database.Transaction, 0, len(m.transactions))
	for _, transaction := range m.transactions {
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (m *MockDatabase) GetMsgType(txHash string, chainName string) (string, error) {
	if m.shouldError {
		return "", fmt.Errorf("%s", m.errorMsg)
	}
	msgType, ok := m.msgTypes[txHash]
	if !ok {
		return "", fmt.Errorf("message type not found")
	}
	return msgType, nil
}

func (m *MockDatabase) GetBankSend(txHash string, chainName string) (*database.BankSend, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	bankSend, ok := m.bankSend[txHash]
	if !ok {
		return nil, fmt.Errorf("bank send not found")
	}
	return bankSend, nil
}

func (m *MockDatabase) GetMsgCall(txHash string, chainName string) (*database.MsgCall, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	msgCall, ok := m.msgCall[txHash]
	if !ok {
		return nil, fmt.Errorf("message call not found")
	}
	return msgCall, nil
}

func (m *MockDatabase) GetMsgAddPackage(txHash string, chainName string) (*database.MsgAddPackage, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	msgAddPackage, ok := m.msgAddPackage[txHash]
	if !ok {
		return nil, fmt.Errorf("message add package not found")
	}
	return msgAddPackage, nil
}

func (m *MockDatabase) GetMsgRun(txHash string, chainName string) (*database.MsgRun, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	msgRun, ok := m.msgRun[txHash]
	if !ok {
		return nil, fmt.Errorf("message run not found")
	}
	return msgRun, nil
}
