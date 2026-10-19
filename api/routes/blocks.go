package routes

import (
	"net/http"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
)

func RegisterBlocksRoutes(
	api huma.API,
	h *handlers.BlocksHandler,
	m *handlers.InMemoryHandler,
	s *handlers.BlockStreamHandler,
) {
	// Registered before /blocks/{height} so "stream" is never parsed as a height.
	sse.Register(api, huma.Operation{
		OperationID: "stream-blocks",
		Method:      http.MethodGet,
		Path:        "/blocks/stream",
		Summary:     "Stream New Blocks",
		Description: `Server-Sent Events stream of newly indexed blocks. Each "block" event carries the block data and uses the block height as its id.
		Send Last-Event-ID (EventSource does this on reconnect) to replay up to 50 missed blocks first.`,
	}, map[string]any{
		"block": database.BlockData{},
	}, s.StreamBlocks)
	huma.Get(api, "/blocks", h.GetLastXBlocks,
		func(op *huma.Operation) {
			op.Summary = "Get Last X Blocks"
			op.Description = "Retrieve the last X blocks data."
		})
	huma.Get(api, "/blocks/latest", h.GetLatestBlock,
		func(op *huma.Operation) {
			op.Summary = "Get Latest Block"
			op.Description = "Retrieve the latest block data."
		})
	huma.Get(api, "/blocks/{height}", h.GetBlock,
		func(op *huma.Operation) {
			op.Summary = "Get Block by Height"
			op.Description = "Retrieve block data by its height."
		})
	huma.Get(api, "/blocks/{from_height}/{to_height}", h.GetFromToBlocks,
		func(op *huma.Operation) {
			op.Summary = "Get Blocks in Height Range"
			op.Description = `Retrieve blocks data by height range.
			From height must be less than to height and the difference must be less than 100.`
		})
	huma.Get(api, "/blocks/{block_height}/signers", h.GetAllBlockSigners,
		func(op *huma.Operation) {
			op.Summary = "Get Block Signers"
			op.Description = "Retrieve all validators that signed a block by its height."
		})
	huma.Get(api, "/blocks/stats/count/recent", h.GetBlockCount24h,
		func(op *huma.Operation) {
			op.Summary = "Get Block Count (Last 24h)"
			op.Description = "Retrieve the total number of blocks produced in the last 24 hours."
		})
	huma.Get(api, "/blocks/stats/count/daily", h.GetBlockCountByDate,
		func(op *huma.Operation) {
			op.Summary = "Get Block Count by Day"
			op.Description = "Retrieve the block count per day within the given date range. Max range is 30 days."
		})
	huma.Get(api, "/blocks/stats/avg_time", m.GetAvgBlockProdTime,
		func(op *huma.Operation) {
			op.Summary = "Get Average Block Production Time"
			op.Description = "Retrieve the average time it takes to produce a block. Returns a float value in seconds."
		})
}
