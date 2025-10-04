package humatypes

import "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"

type TransactionGetInput struct {
	TxHash string `path:"tx_hash" doc:"Transaction hash (base64url encoded)" required:"true"`
}

type TransactionBasicGetOutput struct {
	Body database.Transaction
}

type TransactionMessageGetOutput struct {
	Body any // database.MsgRun | database.MsgCall | database.MsgAddPackage | database.BankSend
}
