package decoder

import (
	"fmt"
	"math/big"
	"time"

	dataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
	"github.com/jackc/pgx/v5/pgtype"
)

// NewDecodedMsg creates a new DecodedMsg struct
//
// Args:
//   - encodedTx: the encoded transaction
//
// Returns:
//   - *DecodedMsg: the decoded message
//   - error: an error if the decoding fails
//
// The method will not throw an error if the decoded message is not found, it will just return nil
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

// GetBasicData returns the basic data of the decoded message
//
// Returns:
//   - BasicTxData: the basic data of the decoded message
//
// The method will not throw an error if the basic data is not found, it will just return nil
func (dm *DecodedMsg) GetBasicData() BasicTxData {
	return dm.BasicData
}

// GetMessages returns the messages of the decoded message
//
// Returns:
//   - []map[string]any: the messages of the decoded message
//
// The method will not throw an error if the messages are not found, it will just return nil
func (dm *DecodedMsg) GetMessages() []map[string]any {
	return dm.Messages
}

// GetMsgTypes returns the message types of the decoded message
//
// Returns:
//   - []string: the message types of the decoded message
//
// The method will not throw an error if the message types are not found, it will just return nil
func (dm *DecodedMsg) GetMsgTypes() []string {
	msgTypes := make([]string, len(dm.Messages))
	for _, message := range dm.Messages {
		msgTypes = append(msgTypes, message["msg_type"].(string))
	}
	return msgTypes
}

// GetSigners returns the signers of the decoded message
//
// Returns:
//   - []string: the signers of the decoded message
//
// The method will not throw an error if the signers are not found, it will just return nil
func (dm *DecodedMsg) GetSigners() []string {
	return dm.BasicData.Signers
}

// GetMemo returns the memo of the decoded message
//
// Returns:
//   - string: the memo of the decoded message
//
// The method will not throw an error if the memo is not found, it will just return nil
func (dm *DecodedMsg) GetMemo() string {
	return dm.BasicData.Memo
}

// GetFee returns the fee of the decoded message
//
// Returns:
//   - dataTypes.Amount: the fee of the decoded message
//
// The method will not throw an error if the fee is not found, it will just return nil
func (dm *DecodedMsg) GetFee() dataTypes.Amount {
	return dm.BasicData.Fee
}

// CollectAllAddresses extracts all unique addresses from the decoded message
// This includes signers and all addresses from individual messages
func (dm *DecodedMsg) CollectAllAddresses() []string {
	addressSet := make(map[string]bool)

	// Add signers from transaction
	for _, signer := range dm.BasicData.Signers {
		addressSet[signer] = true
	}

	// Add addresses from each message
	for _, msgMap := range dm.Messages {
		msgType, ok := msgMap["msg_type"].(string)
		if !ok {
			continue
		}

		switch msgType {
		case "bank_msg_send":
			if fromAddr, ok := msgMap["from_address"].(string); ok {
				addressSet[fromAddr] = true
			}
			if toAddr, ok := msgMap["to_address"].(string); ok {
				addressSet[toAddr] = true
			}

		case "vm_msg_call":
			if caller, ok := msgMap["caller"].(string); ok {
				addressSet[caller] = true
			}

		case "vm_msg_add_package":
			if creator, ok := msgMap["creator"].(string); ok {
				addressSet[creator] = true
			}

		case "vm_msg_run":
			if caller, ok := msgMap["caller"].(string); ok {
				addressSet[caller] = true
			}
		}
	}

	// Convert set to slice
	addresses := make([]string, 0, len(addressSet))
	for addr := range addressSet {
		addresses = append(addresses, addr)
	}

	return addresses
}

// MessageGroups holds different message types grouped for batch insertion
type MessageGroups struct {
	MsgSend   []MsgSend
	MsgCall   []MsgCall
	MsgAddPkg []MsgAddPackage
	MsgRun    []MsgRun
}

// DbMessageGroups holds database-ready message types with address IDs
type DbMessageGroups struct {
	MsgSend   []dataTypes.MsgSend
	MsgCall   []dataTypes.MsgCall
	MsgAddPkg []dataTypes.MsgAddPackage
	MsgRun    []dataTypes.MsgRun
}

// ConvertToDbMessages converts MessageGroups to DbMessageGroups using address cache
//
// Part of the 2 phase message processing,
// this method will convert to types that can be inserted into the database
// not the best solution, but the program needs a 2 step method regardless of the types
// the program needs to insert the addresses ids into the database hence this complex system
//
// Args:
//   - addressResolver: an address resolver to get the address IDs
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//   - timestamp: the timestamp of the transaction
//   - signers: the signers of the transaction
//
// Returns:
//   - *DbMessageGroups: a DbMessageGroups struct
//   - error: an error if the conversion fails
func (mg *MessageGroups) ConvertToDbMessages(
	addressResolver AddressResolver,
	txHash []byte,
	chainName string,
	timestamp time.Time,
	signers []string,
) *DbMessageGroups {
	dbGroups := &DbMessageGroups{
		MsgSend:   make([]dataTypes.MsgSend, len(mg.MsgSend)),
		MsgCall:   make([]dataTypes.MsgCall, len(mg.MsgCall)),
		MsgAddPkg: make([]dataTypes.MsgAddPackage, len(mg.MsgAddPkg)),
		MsgRun:    make([]dataTypes.MsgRun, len(mg.MsgRun)),
	}

	// Convert MsgSend
	for i, msg := range mg.MsgSend {
		// Convert amount from []Coin to dataTypes.Amount
		amount := make([]dataTypes.Amount, len(msg.Amount))
		for j, amt := range msg.Amount {
			bigInt := big.NewInt(int64(amt.Amount))
			amount[j] = dataTypes.Amount{
				Amount: pgtype.Numeric{Int: bigInt, Valid: true},
				Denom:  amt.Denom,
			}
		}
		// Convert signers to address IDs
		signerIds := make([]int32, len(signers))
		for k, signer := range signers {
			signerIds[k] = addressResolver.GetAddress(signer)
		}

		dbGroups.MsgSend[i] = dataTypes.MsgSend{
			TxHash:      txHash,
			ChainName:   chainName,
			FromAddress: addressResolver.GetAddress(msg.FromAddress),
			ToAddress:   addressResolver.GetAddress(msg.ToAddress),
			Amount:      amount,
			Signers:     signerIds,
			Timestamp:   timestamp,
		}
	}

	// Convert MsgCall
	for i, msg := range mg.MsgCall {
		// Convert send from []Coin to dataTypes.Amount
		send := make([]dataTypes.Amount, len(msg.Send))
		for j, amt := range msg.Send {
			bigInt := big.NewInt(int64(amt.Amount))
			send[j] = dataTypes.Amount{
				Amount: pgtype.Numeric{Int: bigInt, Valid: true},
				Denom:  amt.Denom,
			}
		}
		// Convert maxDeposit from []Coin to dataTypes.Amount
		maxDeposit := make([]dataTypes.Amount, len(msg.MaxDeposit))
		for j, amt := range msg.MaxDeposit {
			bigInt := big.NewInt(int64(amt.Amount))
			maxDeposit[j] = dataTypes.Amount{
				Amount: pgtype.Numeric{Int: bigInt, Valid: true},
				Denom:  amt.Denom,
			}
		}
		// Convert signers to address IDs
		signerIds := make([]int32, len(signers))
		for k, signer := range signers {
			signerIds[k] = addressResolver.GetAddress(signer)
		}

		dbGroups.MsgCall[i] = dataTypes.MsgCall{
			TxHash:     txHash,
			ChainName:  chainName,
			Caller:     addressResolver.GetAddress(msg.Caller),
			Send:       send,
			PkgPath:    msg.PkgPath,
			FuncName:   msg.FuncName,
			Args:       msg.Args,
			MaxDeposit: maxDeposit,
			Signers:    signerIds,
			Timestamp:  timestamp,
		}
	}

	// Convert MsgAddPackage
	for i, msg := range mg.MsgAddPkg {
		// Convert send from []Coin to dataTypes.Amount
		send := make([]dataTypes.Amount, len(msg.Send))
		for j, amt := range msg.Send {
			bigInt := big.NewInt(int64(amt.Amount))
			send[j] = dataTypes.Amount{
				Amount: pgtype.Numeric{Int: bigInt, Valid: true},
				Denom:  amt.Denom,
			}
		}
		// Convert maxDeposit from []Coin to dataTypes.Amount
		maxDeposit := make([]dataTypes.Amount, len(msg.MaxDeposit))
		for j, amt := range msg.MaxDeposit {
			bigInt := big.NewInt(int64(amt.Amount))
			maxDeposit[j] = dataTypes.Amount{
				Amount: pgtype.Numeric{Int: bigInt, Valid: true},
				Denom:  amt.Denom,
			}
		}
		// Convert signers to address IDs
		signerIds := make([]int32, len(signers))
		for k, signer := range signers {
			signerIds[k] = addressResolver.GetAddress(signer)
		}

		dbGroups.MsgAddPkg[i] = dataTypes.MsgAddPackage{
			TxHash:     txHash,
			ChainName:  chainName,
			Creator:    addressResolver.GetAddress(msg.Creator),
			PkgPath:    msg.PkgPath,
			PkgName:    msg.PkgName,
			Send:       send,
			MaxDeposit: maxDeposit,
			Signers:    signerIds,
			Timestamp:  timestamp,
		}
	}

	// Convert MsgRun
	for i, msg := range mg.MsgRun {
		// Convert send from []Coin to dataTypes.Amount
		send := make([]dataTypes.Amount, len(msg.Send))
		for j, amt := range msg.Send {
			bigInt := big.NewInt(int64(amt.Amount))
			send[j] = dataTypes.Amount{
				Amount: pgtype.Numeric{Int: bigInt, Valid: true},
				Denom:  amt.Denom,
			}
		}
		// Convert maxDeposit from []Coin to dataTypes.Amount
		maxDeposit := make([]dataTypes.Amount, len(msg.MaxDeposit))
		for j, amt := range msg.MaxDeposit {
			bigInt := big.NewInt(int64(amt.Amount))
			maxDeposit[j] = dataTypes.Amount{
				Amount: pgtype.Numeric{Int: bigInt, Valid: true},
				Denom:  amt.Denom,
			}
		}
		// Convert signers to address IDs
		signerIds := make([]int32, len(signers))
		for k, signer := range signers {
			signerIds[k] = addressResolver.GetAddress(signer)
		}

		dbGroups.MsgRun[i] = dataTypes.MsgRun{
			TxHash:     txHash,
			ChainName:  chainName,
			Caller:     addressResolver.GetAddress(msg.Caller),
			PkgPath:    msg.PkgPath,
			PkgName:    msg.PkgName,
			Send:       send,
			MaxDeposit: maxDeposit,
			Signers:    signerIds,
			Timestamp:  timestamp,
		}
	}

	return dbGroups
}

// ConvertToStructuredMessages converts the decoded message maps to structured message types
// Returns MessageGroups with messages organized by type for efficient batch insertion
func (dm *DecodedMsg) ConvertToStructuredMessages(chainName string, timestamp time.Time) (*MessageGroups, error) {
	groups := &MessageGroups{
		MsgSend:   make([]MsgSend, 0),
		MsgCall:   make([]MsgCall, 0),
		MsgAddPkg: make([]MsgAddPackage, 0),
		MsgRun:    make([]MsgRun, 0),
	}

	for _, msgMap := range dm.Messages {
		msgType, ok := msgMap["msg_type"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid msg_type")
		}

		switch msgType {
		case "bank_msg_send":
			msg, err := dm.convertToMsgSend(msgMap)
			if err != nil {
				return nil, fmt.Errorf("failed to convert bank_msg_send: %w", err)
			}
			groups.MsgSend = append(groups.MsgSend, *msg)

		case "vm_msg_call":
			msg, err := dm.convertToMsgCall(msgMap)
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_call: %w", err)
			}
			groups.MsgCall = append(groups.MsgCall, *msg)

		case "vm_msg_add_package":
			msg, err := dm.convertToMsgAddPackage(msgMap)
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_add_package: %w", err)
			}
			groups.MsgAddPkg = append(groups.MsgAddPkg, *msg)

		case "vm_msg_run":
			msg, err := dm.convertToMsgRun(msgMap)
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
func (dm *DecodedMsg) convertToMsgSend(msgMap map[string]any) (*MsgSend, error) {
	msgType, ok := msgMap["msg_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing msg_type")
	}

	fromAddress, ok := msgMap["from_address"].(string)
	if !ok {
		return nil, fmt.Errorf("missing from_address")
	}

	toAddress, ok := msgMap["to_address"].(string)
	if !ok {
		return nil, fmt.Errorf("missing to_address")
	}

	// Convert amount from []Coin to string representation
	amount, ok := msgMap["amount"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing amount")
	}

	return &MsgSend{
		MsgType:     msgType,
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
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
func (dm *DecodedMsg) convertToMsgCall(msgMap map[string]any) (*MsgCall, error) {
	msgType, ok := msgMap["msg_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing msg_type")
	}

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

	send, ok := msgMap["send"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing send")
	}

	maxDeposit, ok := msgMap["max_deposit"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing max_deposit")
	}

	return &MsgCall{
		MsgType:    msgType,
		Caller:     caller,
		PkgPath:    pkgPath,
		FuncName:   funcName,
		Args:       argsStr,
		Send:       send,
		MaxDeposit: maxDeposit,
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
func (dm *DecodedMsg) convertToMsgAddPackage(msgMap map[string]any) (*MsgAddPackage, error) {
	msgType, ok := msgMap["msg_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing msg_type")
	}

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

	send, ok := msgMap["send"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing send")
	}

	maxDeposit, ok := msgMap["max_deposit"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing max_deposit")
	}

	return &MsgAddPackage{
		MsgType:    msgType,
		Creator:    creator,
		PkgPath:    pkgPath,
		PkgName:    pkgName,
		Send:       send,
		MaxDeposit: maxDeposit,
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
func (dm *DecodedMsg) convertToMsgRun(msgMap map[string]any) (*MsgRun, error) {
	msgType, ok := msgMap["msg_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing msg_type")
	}

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

	send, ok := msgMap["send"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing send")
	}

	maxDeposit, ok := msgMap["max_deposit"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing max_deposit")
	}

	return &MsgRun{
		MsgType:    msgType,
		Caller:     caller,
		PkgPath:    pkgPath,
		PkgName:    pkgName,
		Send:       send,
		MaxDeposit: maxDeposit,
	}, nil
}
