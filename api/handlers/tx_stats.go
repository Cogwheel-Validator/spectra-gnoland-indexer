package handlers

import (
	"context"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
)

// GetTxStatsSummary returns transaction activity over the last `hours` hours.
func (h *TransactionsHandler) GetTxStatsSummary(
	ctx context.Context,
	input *humatypes.TxStatsSummaryGetInput,
) (*humatypes.TxStatsSummaryGetOutput, error) {
	summary, err := h.db.GetTxStatsSummary(ctx, h.chainName, input.Hours)
	if err != nil {
		return nil, mapDbError("GetTxStatsSummary", "transaction summary not found", err)
	}
	return &humatypes.TxStatsSummaryGetOutput{Body: summary}, nil
}
