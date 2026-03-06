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
			err := h.getBankSendResponse(ctx, msgType, txHashBase64, h.chainName, &response)
			if err != nil {
				return nil, err
			}
		case "vm_msg_call":
			err := h.getMsgCallResponse(ctx, msgType, txHashBase64, h.chainName, &response)
			if err != nil {
				return nil, err
			}
		case "vm_msg_add_package":
			err := h.getMsgAddPackageResponse(ctx, msgType, txHashBase64, h.chainName, &response)
			if err != nil {
				return nil, err
			}
		case "vm_msg_run":
			err := h.getMsgRunResponse(ctx, msgType, txHashBase64, h.chainName, &response)
			if err != nil {
				return nil, err
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

// Helper method that collects msg call data from the database and adds it to the response
func (h *TransactionsHandler) getMsgCallResponse(
	ctx context.Context,
	msgType string,
	txHash string,
	chainName string,
	response *map[int16]humatypes.TransactionMessage,
) error {
	data, err := h.db.GetMsgCall(ctx, txHash, chainName)
	if err != nil {
		return huma.Error400BadRequest(fmt.Sprintf("Failed to fetch %s data for transaction %s", "vm_msg_call", txHash), err)
	}
	for _, d := range data {
		index := d.MessageCounter
		(*response)[index] = humatypes.TransactionMessage{
			MessageType: msgType,
			TxHash:      d.TxHash,
			Timestamp:   d.Timestamp,
			Signers:     d.Signers,
			Caller:      d.Caller,
			Send:        d.Send,
			PkgPath:     d.PkgPath,
			FuncName:    d.FuncName,
			Args:        d.Args,
			MaxDeposit:  d.MaxDeposit,
		}
	}
	return nil
}

// Helper method that collects add package data from the database and adds it to the response
func (h *TransactionsHandler) getMsgAddPackageResponse(
	ctx context.Context,
	msgType string,
	txHash string,
	chainName string,
	response *map[int16]humatypes.TransactionMessage,
) error {
	data, err := h.db.GetMsgAddPackage(ctx, txHash, chainName)
	if err != nil {
		return huma.Error400BadRequest(fmt.Sprintf("Failed to fetch %s data for transaction %s", "vm_msg_add_package", txHash), err)
	}
	for _, d := range data {
		index := d.MessageCounter
		(*response)[index] = humatypes.TransactionMessage{
			MessageType:  msgType,
			TxHash:       d.TxHash,
			Timestamp:    d.Timestamp,
			Signers:      d.Signers,
			Creator:      d.Creator,
			PkgPath:      d.PkgPath,
			PkgName:      d.PkgName,
			PkgFileNames: d.PkgFileNames,
			Send:         d.Send,
			MaxDeposit:   d.MaxDeposit,
		}
	}
	return nil
}

// Helper method that collects msg run data from the database and adds it to the response
func (h *TransactionsHandler) getMsgRunResponse(
	ctx context.Context,
	msgType string,
	txHash string,
	chainName string,
	response *map[int16]humatypes.TransactionMessage,
) error {
	data, err := h.db.GetMsgRun(ctx, txHash, chainName)
	if err != nil {
		return huma.Error400BadRequest(fmt.Sprintf("Failed to fetch %s data for transaction %s", "vm_msg_run", txHash), err)
	}
	for _, d := range data {
		index := d.MessageCounter
		(*response)[index] = humatypes.TransactionMessage{
			MessageType:  msgType,
			TxHash:       d.TxHash,
			Timestamp:    d.Timestamp,
			Signers:      d.Signers,
			Caller:       d.Caller,
			PkgPath:      d.PkgPath,
			PkgName:      d.PkgName,
			PkgFileNames: d.PkgFileNames,
			Send:         d.Send,
			MaxDeposit:   d.MaxDeposit,
		}
	}
	return nil
}

// Helper method that collects bank send data from the database and adds it to the response
func (h *TransactionsHandler) getBankSendResponse(
	ctx context.Context,
	msgType string,
	txHash string,
	chainName string,
	response *map[int16]humatypes.TransactionMessage,
) error {
	data, err := h.db.GetBankSend(ctx, txHash, chainName)
	if err != nil {
		return huma.Error400BadRequest(fmt.Sprintf("Failed to fetch %s data for transaction %s", "bank_msg_send", txHash), err)
	}
	for _, d := range data {
		index := d.MessageCounter
		(*response)[index] = humatypes.TransactionMessage{
			MessageType: msgType,
			TxHash:      d.TxHash,
			Timestamp:   d.Timestamp,
			Signers:     d.Signers,
			FromAddress: d.FromAddress,
			ToAddress:   d.ToAddress,
			Amount:      d.Amount,
		}
	}
	return nil
}
