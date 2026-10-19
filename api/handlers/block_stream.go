package handlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/blockstream"
	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/danielgtaylor/huma/v2/sse"
)

const (
	// blockStreamPollInterval is how often the shared poller checks for newly
	// indexed blocks while at least one stream subscriber is connected.
	blockStreamPollInterval = time.Second
	// blockStreamPingInterval keeps idle connections (and proxies) alive.
	blockStreamPingInterval = 15 * time.Second
	// blockStreamRetryMs tells EventSource how long to wait before reconnecting.
	blockStreamRetryMs = 3000
)

// BlockStreamDbHandler is the database surface the live block stream needs.
type BlockStreamDbHandler interface {
	GetBlocksAfter(ctx context.Context, chainName string, after uint64, limit uint64) ([]*database.BlockData, error)
	GetLatestBlockHeight(ctx context.Context, chainName string) (uint64, error)
}

// BlockStreamHandler serves the live block stream (Server-Sent Events). All
// connections share one poller.
type BlockStreamHandler struct {
	db        BlockStreamDbHandler
	chainName string
	hub       *blockstream.Hub[*database.BlockData]
}

// NewBlockStreamHandler creates a BlockStreamHandler.
func NewBlockStreamHandler(db BlockStreamDbHandler, chainName string) *BlockStreamHandler {
	hub := blockstream.NewHub(
		chainName,
		blockStreamPollInterval,
		func(ctx context.Context, after uint64, limit int) ([]*database.BlockData, error) {
			// #nosec G115 -- limit is a small positive constant
			return db.GetBlocksAfter(ctx, chainName, after, uint64(limit))
		},
		func(ctx context.Context) (uint64, error) {
			return db.GetLatestBlockHeight(ctx, chainName)
		},
		func(b *database.BlockData) uint64 { return b.Height },
	)
	return &BlockStreamHandler{db: db, chainName: chainName, hub: hub}
}

// StreamBlocks pushes every newly indexed block as a "block" event whose id
// is the block height.
func (h *BlockStreamHandler) StreamBlocks(
	ctx context.Context,
	input *humatypes.BlockStreamInput,
	send sse.Sender,
) {
	// Subscribe before backfilling so nothing falls in the gap; duplicates
	// are skipped via lastSent.
	sub, unsubscribe := h.hub.Subscribe()
	defer unsubscribe()

	if err := send(sse.Message{Retry: blockStreamRetryMs, Comment: "connected"}); err != nil {
		return
	}

	lastSent, err := h.backfill(ctx, input.LastEventID, send)
	if err != nil {
		return
	}

	ping := time.NewTicker(blockStreamPingInterval)
	defer ping.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case b, ok := <-sub:
			// A closed channel means this subscriber was dropped for falling
			// behind; the client reconnects with Last-Event-ID and backfills.
			if !ok {
				return
			}
			if b.Height <= lastSent {
				continue
			}
			if sendErr := sendBlock(send, b); sendErr != nil {
				return
			}
			lastSent = b.Height
		case <-ping.C:
			if pingErr := send.Comment("ping"); pingErr != nil {
				return
			}
		}
	}
}

// backfill replays blocks after lastEventID (if any) and returns the last
// height sent. A non-nil error means the client went away.
func (h *BlockStreamHandler) backfill(
	ctx context.Context,
	lastEventID uint64,
	send sse.Sender,
) (uint64, error) {
	if lastEventID == 0 {
		return 0, nil
	}
	blocks, err := h.db.GetBlocksAfter(ctx, h.chainName, lastEventID, blockstream.DefaultBatchLimit)
	if err != nil {
		slog.Error("block stream backfill", "err", err)
		return 0, nil
	}
	var lastSent uint64
	for _, b := range blocks {
		if err := sendBlock(send, b); err != nil {
			return lastSent, err
		}
		lastSent = b.Height
	}
	return lastSent, nil
}

func sendBlock(send sse.Sender, b *database.BlockData) error {
	return send(sse.Message{ID: int(b.Height), Data: b}) // #nosec G115 -- block heights fit in int64
}
