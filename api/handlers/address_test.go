package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressHandler_GetAddressTxs_Success(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	addressTxData := []database.AddressTx{
		{Hash: "tx_hash_1", Timestamp: fixedTime, MsgTypes: []string{"msg_send"}},
		{Hash: "tx_hash_2", Timestamp: fixedTime, MsgTypes: []string{"msg_send"}},
		{Hash: "tx_hash_3", Timestamp: fixedTime, MsgTypes: []string{"msg_send"}},
	}

	db := MockDatabase{
		addressTxs: map[string]*[]database.AddressTx{
			"gno_address_1": &addressTxData,
		},
	}
	handler := handlers.NewAddressHandler(&db, "gnoland")
	response, err := handler.GetAddressTxs(context.Background(), &humatypes.AddressGetInput{Address: "gno_address_1"})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 3, len(response.Body))
	assert.Equal(t, "tx_hash_1", response.Body[0].Hash)
}

func TestAddressHandler_GetAddressTxs_Fail(t *testing.T) {
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "error getting address transactions",
	}
	handler := handlers.NewAddressHandler(&db, "gnoland")
	response, err := handler.GetAddressTxs(context.Background(), &humatypes.AddressGetInput{Address: "gno_address_1"})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "Address not found")
}
