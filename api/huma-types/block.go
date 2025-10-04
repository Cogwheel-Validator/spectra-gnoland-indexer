package humatypes

import (
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
)

// BlockGetInput represents the input for getting a block by height
type BlockGetInput struct {
	Height uint64 `path:"height" minimum:"1" example:"12345" doc:"Block height to retrieve" required:"true"`
}

type FromToBlocksGetInput struct {
	FromHeight uint64 `path:"from_height" minimum:"1" example:"12345" doc:"From block height" required:"true"`
	ToHeight   uint64 `path:"to_height" minimum:"1" example:"12345" doc:"To block height" required:"true"`
}

// BlockGetOutput represents the response structure for a single block
type BlockGetOutput struct {
	Body database.BlockData
}

type FromToBlocksGetOutput struct {
	Body []database.BlockData
}
