package handlers

import (
	"context"
	"fmt"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/danielgtaylor/huma/v2"
)

// BlocksHandler handles block-related API requests
type BlocksHandler struct {
	db        DatabaseHandler
	chainName string
}

// NewBlocksHandler creates a new blocks handler
func NewBlocksHandler(db DatabaseHandler, chainName string) *BlocksHandler {
	return &BlocksHandler{db: db, chainName: chainName}
}

// GetBlock retrieves a block by height
func (h *BlocksHandler) GetBlock(ctx context.Context, input *humatypes.BlockGetInput) (*humatypes.BlockGetOutput, error) {
	// Fetch from database
	block, err := h.db.GetBlock(input.Height, h.chainName)
	if err != nil {
		return nil, huma.Error404NotFound(fmt.Sprintf("Block at height %d not found", input.Height), err)
	}

	response := &humatypes.BlockGetOutput{
		Body: database.BlockData{
			Hash:      block.Hash,
			Height:    block.Height,
			Timestamp: block.Timestamp,
			ChainID:   block.ChainID,
			Txs:       block.Txs,
			TxCount:   len(block.Txs),
		},
	}
	return response, nil
}

// Get from block height a to block height b
func (h *BlocksHandler) GetFromToBlocks(
	ctx context.Context,
	input *humatypes.FromToBlocksGetInput,
) (*humatypes.FromToBlocksGetOutput, error) {
	// Fetch from database
	// validate input
	if input.FromHeight > input.ToHeight {
		return nil, huma.Error400BadRequest("From height must be less than to height", nil)
	}
	blocks, err := h.db.GetFromToBlocks(input.FromHeight, input.ToHeight, h.chainName)
	if err != nil {
		return nil, huma.Error404NotFound(fmt.Sprintf("Blocks from height %d to height %d not found", input.FromHeight, input.ToHeight), err)
	}
	response := &humatypes.FromToBlocksGetOutput{
		Body: make([]database.BlockData, 0, len(blocks)),
	}
	for _, block := range blocks {
		response.Body = append(response.Body, database.BlockData{
			Hash:      block.Hash,
			Height:    block.Height,
			Timestamp: block.Timestamp,
			ChainID:   block.ChainID,
			Txs:       block.Txs,
			TxCount:   len(block.Txs),
		})
	}
	return response, nil
}

func (h *BlocksHandler) GetAllBlockSigners(
	ctx context.Context,
	input *humatypes.AllBlockSignersGetInput,
) (*humatypes.AllBlockSignersGetOutput, error) {
	// Fetch from database
	blockSigners, err := h.db.GetAllBlockSigners(h.chainName, input.BlockHeight)
	if err != nil {
		return nil, huma.Error404NotFound("Block signers not found", err)
	}
	response := &humatypes.AllBlockSignersGetOutput{
		Body: database.BlockSigners{
			BlockHeight: blockSigners.BlockHeight,
			Proposer:    blockSigners.Proposer,
			SignedVals:  blockSigners.SignedVals,
		},
	}
	return response, nil
}

// Get latest block height
func (h *BlocksHandler) GetLatestBlockHeight(ctx context.Context, _ *humatypes.LatestBlockHeightGetInput) (*humatypes.LatestBlockHeightGetOutput, error) {
	block, err := h.db.GetLatestBlockHeight(h.chainName)
	if err != nil {
		return nil, huma.Error404NotFound("Latest block height not found", err)
	}
	response := &humatypes.LatestBlockHeightGetOutput{
		Body: database.BlockData{
			Hash:      block.Hash,
			Height:    block.Height,
			Timestamp: block.Timestamp,
			ChainID:   block.ChainID,
			Txs:       block.Txs,
			TxCount:   len(block.Txs),
		},
	}
	return response, nil
}

// Get last x blocks
func (h *BlocksHandler) GetLastXBlocks(ctx context.Context, input *humatypes.LastXBlocksGetInput) (*humatypes.LastXBlocksGetOutput, error) {
	// Fetch from database
	blocks, err := h.db.GetLastXBlocks(h.chainName, input.Amount)
	if err != nil {
		return nil, huma.Error404NotFound("Last x blocks not found", err)
	}
	response := &humatypes.LastXBlocksGetOutput{
		Body: make([]database.BlockData, 0, len(blocks)),
	}
	for _, block := range blocks {
		response.Body = append(response.Body, database.BlockData{
			Hash:      block.Hash,
			Height:    block.Height,
			Timestamp: block.Timestamp,
			ChainID:   block.ChainID,
			Txs:       block.Txs,
			TxCount:   len(block.Txs),
		})
	}
	return response, nil
}
