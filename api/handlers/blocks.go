package handlers

import (
	"context"
	"fmt"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
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
	block, err := h.db.GetBlock(ctx, input.Height, h.chainName)
	if err != nil {
		return nil, huma.Error404NotFound(fmt.Sprintf("Block at height %d not found", input.Height), err)
	}

	response := &humatypes.BlockGetOutput{
		Body: block,
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
	if input.ToHeight-input.FromHeight > 100 {
		return nil, huma.Error400BadRequest("From height and to height difference must be less than 100", nil)
	}

	blocks, err := h.db.GetFromToBlocks(ctx, input.FromHeight, input.ToHeight, h.chainName)
	if err != nil {
		return nil, huma.Error404NotFound(fmt.Sprintf("Blocks from height %d to height %d not found", input.FromHeight, input.ToHeight), err)
	}
	response := &humatypes.FromToBlocksGetOutput{
		Body: blocks,
	}
	return response, nil
}

func (h *BlocksHandler) GetAllBlockSigners(
	ctx context.Context,
	input *humatypes.AllBlockSignersGetInput,
) (*humatypes.AllBlockSignersGetOutput, error) {
	// Fetch from database
	blockSigners, err := h.db.GetAllBlockSigners(ctx, h.chainName, input.BlockHeight)
	if err != nil {
		return nil, huma.Error404NotFound("Block signers not found", err)
	}
	response := &humatypes.AllBlockSignersGetOutput{
		Body: blockSigners,
	}
	return response, nil
}

// Get latest block height
func (h *BlocksHandler) GetLatestBlock(ctx context.Context, _ *humatypes.LatestBlockHeightGetInput) (*humatypes.LatestBlockHeightGetOutput, error) {
	block, err := h.db.GetLatestBlock(ctx, h.chainName)
	if err != nil {
		return nil, huma.Error404NotFound("Latest block height not found", err)
	}
	response := &humatypes.LatestBlockHeightGetOutput{
		Body: block,
	}
	return response, nil
}

// Get last x blocks
func (h *BlocksHandler) GetLastXBlocks(ctx context.Context, input *humatypes.LastXBlocksGetInput) (*humatypes.LastXBlocksGetOutput, error) {
	// Fetch from database
	blocks, err := h.db.GetLastXBlocks(ctx, h.chainName, input.Amount)
	if err != nil {
		return nil, huma.Error404NotFound("Last x blocks not found", err)
	}
	response := &humatypes.LastXBlocksGetOutput{
		Body: blocks,
	}
	return response, nil
}
