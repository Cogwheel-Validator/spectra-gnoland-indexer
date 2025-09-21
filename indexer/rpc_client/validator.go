package rpcclient

// ValidatorsResponse is the response from the rpc client for the validators method
// Probably won't be used as much since the signing data is part of the block data
// but it's still useful to have it
type ValidatorsResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Error   *JsonRpcError   `json:"error,omitempty"`
	Result  ValidatorResult `json:"result"`
}

type ValidatorResult struct {
	BlockHeight string            `json:"block_height"`
	Validators  []ValidatorsSlice `json:"validators"`
}

type ValidatorsSlice struct {
	Address          string    `json:"address"`
	PubKey           ValPubKey `json:"pub_key"`
	VotingPower      string    `json:"voting_power"`
	ProposerPriority string    `json:"proposer_priority"`
}

type ValPubKey struct {
	Type  string `json:"@type"`
	Value string `json:"value"`
}

func (vr *ValidatorsResponse) GetValidators() []ValidatorsSlice {
	return vr.Result.Validators
}

func (vr *ValidatorsResponse) GetBlockHeight() string {
	return vr.Result.BlockHeight
}

func (vr *ValidatorsResponse) IsValid() bool {
	return vr != nil && vr.Error == nil
}
