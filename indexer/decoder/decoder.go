package decoder

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	dataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jackc/pgx/v5/pgtype"
)

// Decoder is a struct that contains the encoded transaction and the decoded transaction
// It is used to decode the transaction and get the data from it
type Decoder struct {
	encodedTx string
}

// NewDecoder creates a new Decoder struct
//
// Args:
//   - encodedTx: base64 encoded stdTx
//
// Returns:
//   - *Decoder: new Decoder struct
func NewDecoder(encodedTx string) *Decoder {
	return &Decoder{
		encodedTx: encodedTx,
	}
}

// DecodeStdTxFromBase64 decodes a base64 encoded stdTx
//
// The function decodes the data and unmarshals it into a std.Tx struct
// The struct contains the transaction data and the messages
//
// Args:
//   - s: base64 encoded stdTx
//
// Returns:
//   - *std.Tx: decoded stdTx
//   - error: if the base64 decoding or unmarshalling fails
func (d *Decoder) DecodeStdTxFromBase64() (*std.Tx, error) {
	bz, err := base64.StdEncoding.DecodeString(d.encodedTx)
	if err != nil {
		return nil, err
	}
	var tx std.Tx
	if err := amino.Unmarshal(bz, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

// GetMessageFromStdTx is a method that decodes the transaction and returns the appropriate basic tx data and messages
//
// The use case of this function is to decode the raw tx data and gather information about the transaction
// It relies on data imported direclty from the gnoland repo, so any changes to the gnoland repo will need
// to be reflected here, mostly...
// Some will need to be updated and added manually in the future but for now this is good enough
//
// Args:
//   - none
//
// Returns:
//   - BasicTxData: basic tx data
//   - []map[string]any: messages data in a map
//   - error: if the decoding or unmarshalling fails
func (d *Decoder) GetMessageFromStdTx() (BasicTxData, []map[string]any, error) {
	tx, err := d.DecodeStdTxFromBase64()
	if err != nil {
		return BasicTxData{}, nil, err
	}

	// Get transaction hash
	bz, err := base64.StdEncoding.DecodeString(d.encodedTx)
	if err != nil {
		return BasicTxData{}, nil, err
	}

	// Use sha256 and then we will use the hash as the primary key for the transaction
	txHash := sha256.Sum256(bz)

	signers := tx.GetSigners()
	signersString := make([]string, len(signers))
	for i, signer := range signers {
		signersString[i] = signer.String()
	}
	bigInt := big.NewInt(tx.Fee.GasFee.Amount)
	feeAmount := pgtype.Numeric{Int: bigInt, Valid: true}
	fee := dataTypes.Amount{
		Amount: feeAmount,
		Denom:  tx.Fee.GasFee.Denom,
	}

	basicTxData := BasicTxData{
		TxHash:  txHash[:],
		Signers: signersString,
		Memo:    tx.GetMemo(),
		Fee:     fee,
	}

	var messages []map[string]any

	// Process each message in the transaction
	for _, msg := range tx.GetMsgs() {
		switch m := msg.(type) {
		case bank.MsgSend:
			// amount should have something like 1000000 ugnot we just need to split it and convert it to uint64
			amount, err := extractCoins(m.Amount)
			if err != nil {
				amount = []Coin{}
			}
			messages = append(messages, map[string]any{
				"msg_type":     "bank_msg_send",
				"from_address": m.FromAddress.String(),
				"to_address":   m.ToAddress.String(),
				"amount":       amount,
			})
		case vm.MsgCall:
			caller := m.Caller.String()
			send, err := extractCoins(m.Send)
			if err != nil {
				send = []Coin{}
			}
			pkgPath := m.PkgPath
			// max deposit could be empty and there is a chance it will return an error
			// so we need to handle that
			maxDeposit, err := extractCoins(m.MaxDeposit)
			if err != nil {
				maxDeposit = []Coin{}
			}
			funcName := m.Func
			// combine the args into a string
			args := strings.Join(m.Args, ",")
			messages = append(messages, map[string]any{
				"msg_type":    "vm_msg_call",
				"caller":      caller,
				"pkg_path":    pkgPath,
				"func_name":   funcName,
				"args":        args,
				"send":        send,
				"max_deposit": maxDeposit,
			})
		case vm.MsgAddPackage:
			pkgPath := m.Package.Path
			pkgName := m.Package.Name
			pkgFileNames := m.Package.FileNames()
			creator := m.Creator.String()
			send, err := extractCoins(m.Send)
			if err != nil {
				send = []Coin{}
			}
			maxDeposit, err := extractCoins(m.MaxDeposit)
			if err != nil {
				maxDeposit = []Coin{}
			}
			messages = append(messages, map[string]any{
				"msg_type":       "vm_msg_add_package",
				"pkg_path":       pkgPath,
				"pkg_name":       pkgName,
				"pkg_file_names": pkgFileNames,
				"creator":        creator,
				"send":           send,
				"max_deposit":    maxDeposit,
			})

		case vm.MsgRun:
			caller := m.Caller.String()
			pkgPath := m.Package.Path
			pkgName := m.Package.Name
			pkgFileNames := m.Package.FileNames()
			send, err := extractCoins(m.Send)
			if err != nil {
				send = []Coin{}
			}
			// max deposit could be empty and there is a chance it will return an error
			// so we need to handle that
			maxDeposit, err := extractCoins(m.MaxDeposit)
			if err != nil {
				maxDeposit = []Coin{}
			}
			messages = append(messages, map[string]any{
				"msg_type":       "vm_msg_run",
				"caller":         caller,
				"pkg_path":       pkgPath,
				"pkg_name":       pkgName,
				"pkg_file_names": pkgFileNames,
				"send":           send,
				"max_deposit":    maxDeposit,
			})
		// case for AnyNewMessage add here:
		default:
			return BasicTxData{}, nil, fmt.Errorf("unknown or unsupported message type: %T", m)
		}
	}
	return basicTxData, messages, nil
}

// Local function to split the amount and denom
func extractCoins(amount std.Coins) ([]Coin, error) {
	// make a string and split it by space
	coins := make([]Coin, len(amount))
	for _, coin := range amount {
		coins = append(coins, Coin{
			Amount: coin.Amount,
			Denom:  coin.Denom,
		})
	}
	return coins, nil
}
