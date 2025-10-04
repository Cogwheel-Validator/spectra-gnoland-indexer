package database

import "time"

// BlockData represents the actual block data returned in the response body
type BlockData struct {
	Hash      string    `json:"hash" doc:"Block hash (base64 encoded)"`
	Height    uint64    `json:"height" doc:"Block height"`
	Timestamp time.Time `json:"timestamp" doc:"Block timestamp"`
	ChainID   string    `json:"chain_id" doc:"Chain identifier"`
	Txs       []string  `json:"txs" doc:"Transactions (base64 encoded)"`
	TxCount   int       `json:"tx_count" doc:"Number of transactions in the block"`
}

type Event struct {
	AtType     string      `json:"at_type" doc:"Event type"`
	Type       string      `json:"type" doc:"Event type"`
	Attributes []Attribute `json:"attributes" doc:"Event attributes"`
	PkgPath    string      `json:"pkg_path" doc:"Package path"`
}

type Attribute struct {
	Key   string `json:"key" doc:"Attribute key"`
	Value string `json:"value" doc:"Attribute value"`
}

type Amount struct {
	Amount string `json:"amount" doc:"Amount"`
	Denom  string `json:"denom" doc:"Denom"`
}

type BankSend struct {
	TxHash      string    `json:"tx_hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp   time.Time `json:"timestamp" doc:"Transaction timestamp"`
	FromAddress string    `json:"from_address" doc:"From address (addresses)"`
	ToAddress   string    `json:"to_address" doc:"To address (addresses)"`
	Amount      []Amount  `json:"amount" doc:"Amount"`
	Signers     []string  `json:"signers" doc:"Signers (addresses)"`
}

type MsgCall struct {
	TxHash     string    `json:"tx_hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp  time.Time `json:"timestamp" doc:"Transaction timestamp"`
	Caller     string    `json:"caller" doc:"Caller address (addresses)"`
	Send       []Amount  `json:"send" doc:"Send amount"`
	PkgPath    string    `json:"pkg_path" doc:"Package path"`
	FuncName   string    `json:"func_name" doc:"Function name"`
	Args       string    `json:"args" doc:"Arguments"`
	MaxDeposit []Amount  `json:"max_deposit" doc:"Max deposit"`
	Signers    []string  `json:"signers" doc:"Signers (addresses)"`
}

type MsgAddPackage struct {
	TxHash       string    `json:"tx_hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp    time.Time `json:"timestamp" doc:"Transaction timestamp"`
	Creator      string    `json:"creator" doc:"Creator address (addresses)"`
	PkgPath      string    `json:"pkg_path" doc:"Package path"`
	PkgName      string    `json:"pkg_name" doc:"Package name"`
	PkgFileNames []string  `json:"pkg_file_names" doc:"Package file names"`
	Send         []Amount  `json:"send" doc:"Send amount"`
	MaxDeposit   []Amount  `json:"max_deposit" doc:"Max deposit"`
	Signers      []string  `json:"signers" doc:"Signers (addresses)"`
}

type MsgRun struct {
	TxHash       string    `json:"tx_hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp    time.Time `json:"timestamp" doc:"Transaction timestamp"`
	Caller       string    `json:"caller" doc:"Caller address (addresses)"`
	PkgPath      string    `json:"pkg_path" doc:"Package path"`
	PkgName      string    `json:"pkg_name" doc:"Package name"`
	PkgFileNames []string  `json:"pkg_file_names" doc:"Package file names"`
	Send         []Amount  `json:"send" doc:"Send amount"`
	MaxDeposit   []Amount  `json:"max_deposit" doc:"Max deposit"`
	Signers      []string  `json:"signers" doc:"Signers (addresses)"`
}

type Transaction struct {
	TxHash      string    `json:"tx_hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp   time.Time `json:"timestamp" doc:"Transaction timestamp"`
	BlockHeight uint64    `json:"block_height" doc:"Block height"`
	TxEvents    []Event   `json:"tx_events" doc:"Transaction events"`
	GasUsed     uint64    `json:"gas_used" doc:"Gas used"`
	GasWanted   uint64    `json:"gas_wanted" doc:"Gas wanted"`
	Fee         Amount    `json:"fee" doc:"Fee"`
}
