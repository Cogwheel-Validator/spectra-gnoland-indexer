package handlers_test

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionsHandler_GetLastXTransactions_Success(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	db := MockDatabase{
		transactions: map[string]*database.Transaction{
			"tx_hash_1": {
				TxHash:      "tx_hash_1",
				Timestamp:   fixedTime,
				BlockHeight: 42,
				TxEvents:    []database.Event{},
				GasUsed:     100,
				GasWanted:   100,
				Fee:         database.Amount{Amount: "100", Denom: "ugnot"},
				MsgTypes:    []string{"msg_send"},
			},
		},
	}

	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetLastXTransactions(context.Background(), &humatypes.TransactionGeneralListGetInput{Amount: 1})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 1, len(response.Body))
	assert.Equal(t, "tx_hash_1", response.Body[0].TxHash)
}

func TestTransactionsHandler_GetLastXTransactions_Fail(t *testing.T) {
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "error getting last x transactions",
	}
	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetLastXTransactions(context.Background(), &humatypes.TransactionGeneralListGetInput{Amount: 1})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "Last x transactions not found")
}

func TestTransactionsHandler_GetTransactionBasic_Success(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	txHash := base64.RawURLEncoding.EncodeToString([]byte("tx_hash_1"))

	db := MockDatabase{
		transactions: map[string]*database.Transaction{
			txHash: {
				TxHash:      txHash,
				Timestamp:   fixedTime,
				BlockHeight: 42,
				TxEvents:    []database.Event{},
				GasUsed:     100,
				GasWanted:   100,
				Fee:         database.Amount{Amount: "100", Denom: "ugnot"},
				MsgTypes:    []string{"msg_send"},
			},
		},
	}

	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetTransactionBasic(context.Background(), &humatypes.TransactionGetInput{TxHash: txHash})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, txHash, response.Body.TxHash)
}

func TestTransactionsHandler_GetTransactionBasic_Fail(t *testing.T) {
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "error getting transaction basic",
	}
	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetTransactionBasic(context.Background(), &humatypes.TransactionGetInput{TxHash: "invalid"})

	assert.Error(t, err)
	assert.Nil(t, response)
}
