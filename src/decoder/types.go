package decoder

import datatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/src/data_types"

type BasicTxData struct {
	TxHash  []byte
	Signers []string
	Memo    string
	Fee     datatypes.Fee
}
