package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/danielgtaylor/huma/v2"
)

type TransactionsHandler struct {
	db        DatabaseHandler
	chainName string
}

func NewTransactionsHandler(db DatabaseHandler, chainName string) *TransactionsHandler {
	return &TransactionsHandler{db: db, chainName: chainName}
}

// GetTransactionBasic retrieves basic transaction details by tx hash
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
	transaction, err := h.db.GetTransaction(ctx, txHashBase64, h.chainName)
	if err != nil {
		return nil, huma.Error404NotFound(fmt.Sprintf("Transaction with hash %s not found", input.TxHash), err)
	}
	return &humatypes.TransactionBasicGetOutput{
		Body: *transaction,
	}, nil
}

// GetTransactionMessage retrieves all messages within a transaction by tx hash
func (h *TransactionsHandler) GetTransactionMessage(
	ctx context.Context,
	input *humatypes.TransactionGetInput,
) (*humatypes.TransactionMessageGetOutput, error) {
	input.TxHash = strings.Trim(input.TxHash, " ")
	txHash, err := base64.URLEncoding.DecodeString(input.TxHash)
	txHashBase64 := base64.StdEncoding.EncodeToString(txHash)
	response := make(map[int16]humatypes.TransactionMessage)
	if err != nil {
		return nil, huma.Error400BadRequest("Transaction hash is not valid base64url encoded", err)
	}
	msgTypes, err := h.db.GetMsgTypes(ctx, txHashBase64, h.chainName)
	if err != nil {
		return nil, huma.Error404NotFound(fmt.Sprintf("Transaction with hash %s not found", input.TxHash), err)
	}

	for _, msgType := range msgTypes {
		switch msgType {
		case "bank_msg_send":
			data, err := h.db.GetBankSend(ctx, txHashBase64, h.chainName)
			if err != nil {
				return nil, huma.Error404NotFound(fmt.Sprintf("Failed to fetch %s data for transaction %s", msgType, input.TxHash), err)
			}
			for _, data := range data {
				index := data.MessageCounter
				response[index] = humatypes.TransactionMessage{
					MessageType: msgType,
					TxHash:      data.TxHash,
					Timestamp:   data.Timestamp,
					Signers:     data.Signers,
					FromAddress: data.FromAddress,
					ToAddress:   data.ToAddress,
					Amount:      data.Amount,
				}
			}
		case "vm_msg_call":
			data, err := h.db.GetMsgCall(ctx, txHashBase64, h.chainName)
			if err != nil {
				return nil, huma.Error404NotFound(fmt.Sprintf("Failed to fetch %s data for transaction %s", msgType, input.TxHash), err)
			}
			for _, data := range data {
				index := data.MessageCounter
				response[index] = humatypes.TransactionMessage{
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
			}
		case "vm_msg_add_package":
			data, err := h.db.GetMsgAddPackage(ctx, txHashBase64, h.chainName)
			if err != nil {
				return nil, huma.Error404NotFound(fmt.Sprintf("Failed to fetch %s data for transaction %s", msgType, input.TxHash), err)
			}
			for _, data := range data {
				index := data.MessageCounter
				response[index] = humatypes.TransactionMessage{
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
			}
		case "vm_msg_run":
			data, err := h.db.GetMsgRun(ctx, txHashBase64, h.chainName)
			if err != nil {
				return nil, huma.Error404NotFound(fmt.Sprintf("Failed to fetch %s data for transaction %s", msgType, input.TxHash), err)
			}
			for _, data := range data {
				index := data.MessageCounter
				response[index] = humatypes.TransactionMessage{
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
			}
		default:
			return nil, huma.Error400BadRequest("Transaction message type not found", nil)
		}
	}
	return &humatypes.TransactionMessageGetOutput{
		Body: response,
	}, nil
}

// Get tx by limit and limit and cursor
func (h *TransactionsHandler) GetTransactionsByCursor(ctx context.Context, input *humatypes.TransactionGeneralListByCursorGetInput) (*humatypes.TransactionGeneralListByCursorGetOutput, error) {
	transactions, err := h.db.GetTransactionsByCursor(ctx, h.chainName, input.Cursor, input.Limit)
	if err != nil {
		return nil, huma.Error404NotFound("Transactions by cursor not found", err)
	}
	return &humatypes.TransactionGeneralListByCursorGetOutput{
		Body: transactions,
	}, nil
}
