package decoder

import (
	datatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
)

type BasicTxData struct {
	TxHash []byte
	// gno addresses
	Signers []string
	Memo    string
	Fee     datatypes.Amount
}

type Coin struct {
	Amount uint64
	Denom  string
}

type MsgSend struct {
	MsgType string
	// gno address
	FromAddress string
	// gno address
	ToAddress string
	Amount    []Coin
}

type MsgCall struct {
	MsgType string
	// gno address
	Caller     string
	Send       []Coin
	PkgPath    string
	FuncName   string
	Args       string
	MaxDeposit []Coin
}

type MsgAddPackage struct {
	MsgType string
	// gno address
	Creator      string
	PkgPath      string
	PkgName      string
	PkgFileNames []string
	Send         []Coin
	MaxDeposit   []Coin
}

type MsgRun struct {
	MsgType string
	// gno address
	Caller       string
	PkgPath      string
	PkgName      string
	PkgFileNames []string
	Send         []Coin
	MaxDeposit   []Coin
}

// AddressResolver interface to make the code testable and flexible
// the iterface is related to the type struct AddressCache
//
// The only method needed is GetAddress which returns the address id
type AddressResolver interface {
	GetAddress(address string) int32
}

// DecodedMsg struct to hold the basic data and messages of the decoded message
//
// The struct contains the basic data and messages of the decoded message
// The basic data contains the tx hash, signers, memo, and fee
// The messages contains the messages of the decoded message
// The messages are stored in a map with the message type as the key and the message as the value
type DecodedMsg struct {
	BasicData BasicTxData
	Messages  []map[string]any
}
