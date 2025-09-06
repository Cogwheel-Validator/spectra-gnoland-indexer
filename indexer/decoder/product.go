package decoder

import (
	"fmt"
	"strings"
	"time"

	dataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
)

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

// MessageGroups holds different message types grouped for batch insertion
type MessageGroups struct {
	MsgSend   []dataTypes.MsgSend
	MsgCall   []dataTypes.MsgCall
	MsgAddPkg []dataTypes.MsgAddPackage
	MsgRun    []dataTypes.MsgRun
}

// ConvertToStructuredMessages converts the decoded message maps to structured message types
// Returns MessageGroups with messages organized by type for efficient batch insertion
func (dm *DecodedMsg) ConvertToStructuredMessages(chainName string, timestamp time.Time) (*MessageGroups, error) {
	groups := &MessageGroups{
		MsgSend:   make([]dataTypes.MsgSend, 0),
		MsgCall:   make([]dataTypes.MsgCall, 0),
		MsgAddPkg: make([]dataTypes.MsgAddPackage, 0),
		MsgRun:    make([]dataTypes.MsgRun, 0),
	}

	for _, msgMap := range dm.Messages {
		msgType, ok := msgMap["msg_type"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid msg_type")
		}

		switch msgType {
		case "bank_msg_send":
			msg, err := dm.convertToMsgSend(msgMap, chainName, timestamp)
			if err != nil {
				return nil, fmt.Errorf("failed to convert bank_msg_send: %w", err)
			}
			groups.MsgSend = append(groups.MsgSend, *msg)

		case "vm_msg_call":
			msg, err := dm.convertToMsgCall(msgMap, chainName, timestamp)
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_call: %w", err)
			}
			groups.MsgCall = append(groups.MsgCall, *msg)

		case "vm_msg_add_package":
			msg, err := dm.convertToMsgAddPackage(msgMap, chainName, timestamp)
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_add_package: %w", err)
			}
			groups.MsgAddPkg = append(groups.MsgAddPkg, *msg)

		case "vm_msg_run":
			msg, err := dm.convertToMsgRun(msgMap, chainName, timestamp)
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_run: %w", err)
			}
			groups.MsgRun = append(groups.MsgRun, *msg)

		default:
			return nil, fmt.Errorf("unknown message type: %s", msgType)
		}
	}

	return groups, nil
}

// A convert method to convert a map data type to a MsgSend struct
//
// Args:
//   - msgMap: a map of data types
//   - chainName: the name of the chain
//   - timestamp: the timestamp of the transaction
//
// Returns:
//   - *dataTypes.MsgSend: a MsgSend struct
//   - error: an error if the conversion fails
func (dm *DecodedMsg) convertToMsgSend(msgMap map[string]any, chainName string, timestamp time.Time) (*dataTypes.MsgSend, error) {
	fromAddress, ok := msgMap["from_address"].(string)
	if !ok {
		return nil, fmt.Errorf("missing from_address")
	}

	toAddress, ok := msgMap["to_address"].(string)
	if !ok {
		return nil, fmt.Errorf("missing to_address")
	}

	// Convert amount from []Coin to string representation
	amount, err := dm.convertAmountToString(msgMap["amount"])
	if err != nil {
		return nil, fmt.Errorf("failed to convert amount: %w", err)
	}

	return &dataTypes.MsgSend{
		TxHash:      dm.BasicData.TxHash,
		ChainName:   chainName,
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
		Timestamp:   timestamp,
	}, nil
}

// A convert method to convert a map data type to a MsgCall struct
//
// Args:
//   - msgMap: a map of data types
//   - chainName: the name of the chain
//   - timestamp: the timestamp of the transaction
//
// Returns:
//   - *dataTypes.MsgCall: a MsgCall struct
//   - error: an error if the conversion fails
func (dm *DecodedMsg) convertToMsgCall(msgMap map[string]any, chainName string, timestamp time.Time) (*dataTypes.MsgCall, error) {
	caller, ok := msgMap["caller"].(string)
	if !ok {
		return nil, fmt.Errorf("missing caller")
	}

	pkgPath, ok := msgMap["pkg_path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_path")
	}

	funcName, ok := msgMap["func_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing func_name")
	}

	// Convert args from string to []string
	argsStr, ok := msgMap["args"].(string)
	if !ok {
		return nil, fmt.Errorf("missing args")
	}

	var args []string
	if argsStr != "" {
		args = strings.Split(argsStr, ",")
	}

	return &dataTypes.MsgCall{
		TxHash:    dm.BasicData.TxHash,
		ChainName: chainName,
		Caller:    caller,
		PkgPath:   pkgPath,
		FuncName:  funcName,
		Args:      args,
		Timestamp: timestamp,
	}, nil
}

// A convert method to convert a map data type to a MsgAddPackage struct
//
// Args:
//   - msgMap: a map of data types
//   - chainName: the name of the chain
//   - timestamp: the timestamp of the transaction
//
// Returns:
//   - *dataTypes.MsgAddPackage: a MsgAddPackage struct
//   - error: an error if the conversion fails
func (dm *DecodedMsg) convertToMsgAddPackage(msgMap map[string]any, chainName string, timestamp time.Time) (*dataTypes.MsgAddPackage, error) {
	creator, ok := msgMap["creator"].(string)
	if !ok {
		return nil, fmt.Errorf("missing creator")
	}

	pkgPath, ok := msgMap["pkg_path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_path")
	}

	pkgName, ok := msgMap["pkg_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_name")
	}

	return &dataTypes.MsgAddPackage{
		TxHash:    dm.BasicData.TxHash,
		ChainName: chainName,
		Creator:   creator,
		PkgPath:   pkgPath,
		PkgName:   pkgName,
		Timestamp: timestamp,
	}, nil
}

// A convert method to convert a map data type to a MsgRun struct
//
// Args:
//   - msgMap: a map of data types
//   - chainName: the name of the chain
//   - timestamp: the timestamp of the transaction
//
// Returns:
//   - *dataTypes.MsgRun: a MsgRun struct
//   - error: an error if the conversion fails
func (dm *DecodedMsg) convertToMsgRun(msgMap map[string]any, chainName string, timestamp time.Time) (*dataTypes.MsgRun, error) {
	caller, ok := msgMap["caller"].(string)
	if !ok {
		return nil, fmt.Errorf("missing caller")
	}

	pkgPath, ok := msgMap["pkg_path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_path")
	}

	pkgName, ok := msgMap["pkg_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_name")
	}

	return &dataTypes.MsgRun{
		TxHash:    dm.BasicData.TxHash,
		ChainName: chainName,
		Caller:    caller,
		PkgPath:   pkgPath,
		PkgName:   pkgName,
		Timestamp: timestamp,
	}, nil
}

// Helper function to convert amount from various formats to string
// for now leave it like this, until further reaserch maybe the progema should use special type for any amount
func (dm *DecodedMsg) convertAmountToString(amountInterface any) (string, error) {
	switch amt := amountInterface.(type) {
	case []Coin:
		if len(amt) == 0 {
			return "0ugnot", nil
		}
		// Convert first coin to string format like "1000000ugnot"
		return fmt.Sprintf("%d%s", amt[0].Amount, amt[0].Denom), nil
	case string:
		return amt, nil
	default:
		return "", fmt.Errorf("unsupported amount format: %T", amountInterface)
	}
}
