package decoder

import (
	"time"

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

// Database-ready message types using address IDs instead of strings
// These types are optimized for storage using int32 address references

type DbMsgSend struct {
	TxHash      []byte
	ChainName   string
	FromAddress int32 // Address ID from cache
	ToAddress   int32 // Address ID from cache
	Amount      []Coin
	Timestamp   time.Time
}

type DbMsgCall struct {
	TxHash     []byte
	ChainName  string
	Caller     int32 // Address ID from cache
	Send       []Coin
	PkgPath    string
	FuncName   string
	Args       string
	MaxDeposit []Coin
	Timestamp  time.Time
}

type DbMsgAddPackage struct {
	TxHash     []byte
	ChainName  string
	Creator    int32 // Address ID from cache
	PkgPath    string
	PkgName    string
	Send       []Coin
	MaxDeposit []Coin
	Timestamp  time.Time
}

type DbMsgRun struct {
	TxHash     []byte
	ChainName  string
	Caller     int32 // Address ID from cache
	PkgPath    string
	PkgName    string
	Send       []Coin
	MaxDeposit []Coin
	Timestamp  time.Time
}
