package handlers

import (
	"context"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/danielgtaylor/huma/v2"
)

type AddressHandler struct {
	db        DatabaseHandler
	chainName string
}

func NewAddressHandler(db DatabaseHandler, chainName string) *AddressHandler {
	return &AddressHandler{db: db, chainName: chainName}
}

func (h *AddressHandler) GetAddressTxs(
	ctx context.Context,
	input *humatypes.AddressGetInput,
) (*humatypes.AddressGetOutput, error) {
	address, err := h.db.GetAddressTxs(
		ctx,
		input.Address,
		h.chainName,
		input.FromTimestamp,
		input.ToTimestamp,
	)
	if err != nil {
		return nil, huma.Error404NotFound("Address not found", err)
	}
	return &humatypes.AddressGetOutput{
		Body: *address,
	}, nil
}
