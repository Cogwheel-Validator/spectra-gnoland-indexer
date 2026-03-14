package handlers

import (
	"context"
	"time"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/danielgtaylor/huma/v2"
)

type AddressHandler struct {
	db        AddressDbHandler
	chainName string
}

func NewAddressHandler(db AddressDbHandler, chainName string) *AddressHandler {
	return &AddressHandler{db: db, chainName: chainName}
}

func (h *AddressHandler) GetDailyActiveAccount(
	ctx context.Context,
	input *humatypes.DailyActiveAccountGetInput,
) (*humatypes.DailyActiveAccountGetOutput, error) {
	// validate input
	if !input.StartDate.Before(input.EndDate) {
		return nil, huma.Error400BadRequest("start_date must be before end_date", nil)
	}
	if input.EndDate.Sub(input.StartDate) > 24*time.Hour*30 {
		return nil, huma.Error400BadRequest("end_date must be within 30 days of start_date", nil)
	}

	data, err := h.db.GetDailyActiveAccount(ctx, h.chainName, input.StartDate, input.EndDate)
	if err != nil {
		return nil, huma.Error404NotFound("Daily active account data not found", err)
	}
	return &humatypes.DailyActiveAccountGetOutput{Body: data}, nil
}

func (h *AddressHandler) GetAddressTxs(
	ctx context.Context,
	input *humatypes.AddressGetInput,
) (*humatypes.AddressGetOutput, error) {
	var fromTs, toTs *time.Time
	if !input.FromTimestamp.IsZero() {
		fromTs = &input.FromTimestamp
	}
	if !input.ToTimestamp.IsZero() {
		toTs = &input.ToTimestamp
	}
	var limit, page *uint64
	if input.Limit != 0 {
		limit = &input.Limit
	}
	if input.Page != 0 {
		page = &input.Page
	}
	var cursor *string
	if input.Cursor != "" {
		cursor = &input.Cursor
	}
	addressTxs, nextCursor, txCount, err := h.db.GetAddressTxs(
		ctx,
		input.Address,
		h.chainName,
		fromTs,
		toTs,
		limit,
		page,
		cursor,
	)
	if err != nil {
		return nil, huma.Error404NotFound("Address not found", err)
	}
	return &humatypes.AddressGetOutput{
		Body: humatypes.AddressTxsBody{
			AddressTxs: *addressTxs,
			TxCount:    txCount,
			NextCursor: nextCursor,
		},
	}, nil
}
