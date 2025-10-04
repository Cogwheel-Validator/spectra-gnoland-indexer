package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/danielgtaylor/huma/v2"
)

type TransactionsHandler struct {
	db        *database.TimescaleDb
	chainName string
}

func NewTransactionsHandler(db *database.TimescaleDb, chainName string) *TransactionsHandler {
	return &TransactionsHandler{db: db, chainName: chainName}
}

func (h *TransactionsHandler) GetTransactionBasic(
	ctx context.Context,
	input *humatypes.TransactionGetInput,
) (*humatypes.TransactionBasicGetOutput, error) {
	input.TxHash = strings.Trim(input.TxHash, " ")
	txHash, err := base64.URLEncoding.DecodeString(input.TxHash)
	txHashBase64 := base64.StdEncoding.EncodeToString(txHash)
	if err != nil {
		return nil, huma.Error400BadRequest("Transaction hash is not valid base64url encoded", err)
	}
	transaction, err := h.db.GetTransaction(txHashBase64, h.chainName)
	if err != nil {
		return nil, huma.Error404NotFound(fmt.Sprintf("Transaction with hash %s not found", input.TxHash), err)
	}
	return &humatypes.TransactionBasicGetOutput{
		Body: *transaction,
	}, nil
}

func (h *TransactionsHandler) GetTransactionMessage(
	ctx context.Context,
	input *humatypes.TransactionGetInput,
) (*humatypes.TransactionMessageGetOutput, error) {
	input.TxHash = strings.Trim(input.TxHash, " ")
	txHash, err := base64.URLEncoding.DecodeString(input.TxHash)
	txHashBase64 := base64.StdEncoding.EncodeToString(txHash)
	if err != nil {
		return nil, huma.Error400BadRequest("Transaction hash is not valid base64url encoded", err)
	}
	msgType, err := h.db.GetMsgType(txHashBase64, h.chainName)
	if err != nil {
		return nil, huma.Error404NotFound(fmt.Sprintf("Transaction with hash %s not found", input.TxHash), err)
	}
	switch msgType {
	case "bank_msg_send":
		data, err := h.db.GetBankSend(txHashBase64, h.chainName)
		if err != nil {
			return nil, huma.Error404NotFound(fmt.Sprintf("Transaction with hash %s not found", input.TxHash), err)
		}
		return &humatypes.TransactionMessageGetOutput{
			Body: *data,
		}, nil
	case "vm_msg_call":
		data, err := h.db.GetMsgCall(txHashBase64, h.chainName)
		if err != nil {
			return nil, huma.Error404NotFound(fmt.Sprintf("Transaction with hash %s not found", input.TxHash), err)
		}
		return &humatypes.TransactionMessageGetOutput{
			Body: *data,
		}, nil
	case "vm_msg_add_package":
		data, err := h.db.GetMsgAddPackage(txHashBase64, h.chainName)
		if err != nil {
			return nil, huma.Error404NotFound(fmt.Sprintf("Transaction with hash %s not found", input.TxHash), err)
		}
		return &humatypes.TransactionMessageGetOutput{
			Body: *data,
		}, nil
	case "vm_msg_run":
		data, err := h.db.GetMsgRun(txHashBase64, h.chainName)
		if err != nil {
			return nil, huma.Error404NotFound(fmt.Sprintf("Transaction with hash %s not found", input.TxHash), err)
		}
		return &humatypes.TransactionMessageGetOutput{
			Body: *data,
		}, nil
	}
	return nil, huma.Error400BadRequest("Transaction message type not found", nil)
}
