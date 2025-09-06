package decoder

import dataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"

type DecodedMsg struct {
	BasicData BasicTxData
	Messages  []map[string]any
}

func NewDecodedMsg(encodedTx string) *DecodedMsg {
	decoder := NewDecoder(encodedTx)
	basicData, messages, err := decoder.GetMessageFromStdTx()
	if err != nil {
		return nil
	}
	return &DecodedMsg{
		BasicData: basicData,
		Messages:  messages,
	}
}

func (dm *DecodedMsg) GetBasicData() BasicTxData {
	return dm.BasicData
}

func (dm *DecodedMsg) GetMessages() []map[string]any {
	return dm.Messages
}

func (dm *DecodedMsg) GetMsgTypes() []string {
	msgTypes := make([]string, len(dm.Messages))
	for _, message := range dm.Messages {
		msgTypes = append(msgTypes, message["msg_type"].(string))
	}
	return msgTypes
}

func (dm *DecodedMsg) GetSigners() []string {
	return dm.BasicData.Signers
}

func (dm *DecodedMsg) GetMemo() string {
	return dm.BasicData.Memo
}

func (dm *DecodedMsg) GetFee() dataTypes.Fee {
	return dm.BasicData.Fee
}
