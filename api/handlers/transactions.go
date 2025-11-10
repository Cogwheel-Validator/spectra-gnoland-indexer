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
	db        DatabaseHandler
	chainName string
}

func NewTransactionsHandler(db DatabaseHandler, chainName string) *TransactionsHandler {
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

	var message humatypes.TransactionMessage

	switch msgType {
	case "bank_msg_send":
		data, err := h.db.GetBankSend(txHashBase64, h.chainName)
		if err != nil {
			return nil, huma.Error404NotFound(fmt.Sprintf("Transaction with hash %s not found", input.TxHash), err)
		}
		message = humatypes.TransactionMessage{
			MessageType: msgType,
			TxHash:      data.TxHash,
			Timestamp:   data.Timestamp,
			Signers:     data.Signers,
			FromAddress: data.FromAddress,
			ToAddress:   data.ToAddress,
			Amount:      data.Amount,
		}
	case "vm_msg_call":
		data, err := h.db.GetMsgCall(txHashBase64, h.chainName)
		if err != nil {
			return nil, huma.Error404NotFound(fmt.Sprintf("Transaction with hash %s not found", input.TxHash), err)
		}
		message = humatypes.TransactionMessage{
			MessageType: msgType,
			TxHash:      data.TxHash,
			Timestamp:   data.Timestamp,
			Signers:     data.Signers,
			Caller:      data.Caller,
			Send:        data.Send,
			PkgPath:     data.PkgPath,
			FuncName:    data.FuncName,
			Args:        data.Args,
			MaxDeposit:  data.MaxDeposit,
		}
	case "vm_msg_add_package":
		data, err := h.db.GetMsgAddPackage(txHashBase64, h.chainName)
		if err != nil {
			return nil, huma.Error404NotFound(fmt.Sprintf("Transaction with hash %s not found", input.TxHash), err)
		}
		message = humatypes.TransactionMessage{
			MessageType:  msgType,
			TxHash:       data.TxHash,
			Timestamp:    data.Timestamp,
			Signers:      data.Signers,
			Creator:      data.Creator,
			PkgPath:      data.PkgPath,
			PkgName:      data.PkgName,
			PkgFileNames: data.PkgFileNames,
			Send:         data.Send,
			MaxDeposit:   data.MaxDeposit,
		}
	case "vm_msg_run":
		data, err := h.db.GetMsgRun(txHashBase64, h.chainName)
		if err != nil {
			return nil, huma.Error404NotFound(fmt.Sprintf("Transaction with hash %s not found", input.TxHash), err)
		}
		message = humatypes.TransactionMessage{
			MessageType:  msgType,
			TxHash:       data.TxHash,
			Timestamp:    data.Timestamp,
			Signers:      data.Signers,
			Caller:       data.Caller,
			PkgPath:      data.PkgPath,
			PkgName:      data.PkgName,
			PkgFileNames: data.PkgFileNames,
			Send:         data.Send,
			MaxDeposit:   data.MaxDeposit,
		}
	default:
		return nil, huma.Error400BadRequest("Transaction message type not found", nil)
	}

	return &humatypes.TransactionMessageGetOutput{
		Body: message,
	}, nil
}

func (h *TransactionsHandler) GetLastXTransactions(ctx context.Context, input *humatypes.TransactionGeneralListGetInput) (*humatypes.TransactionGeneralListGetOutput, error) {
	transactions, err := h.db.GetLastXTransactions(h.chainName, input.Amount)
	if err != nil {
		return nil, huma.Error404NotFound("Last x transactions not found", err)
	}
	response := &humatypes.TransactionGeneralListGetOutput{
		Body: make([]database.Transaction, 0, len(transactions)),
	}
	for _, transaction := range transactions {
		response.Body = append(response.Body, *transaction)
	}
	return response, nil
}
