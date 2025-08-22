package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func DecodeStdTxFromBase64(s string) (*std.Tx, error) {
	bz, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var tx std.Tx
	if err := amino.Unmarshal(bz, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

func GetDataFromStdTx(tx *std.Tx) (string, error) {
	fees := fmt.Sprint(tx.Fee.GasFee.Amount) + " " + string(tx.Fee.GasFee.Denom)
	for _, msg := range tx.GetMsgs() {
		switch m := msg.(type) {
		case bank.MsgSend:
			fromAddress := m.FromAddress.String()
			toAddress := m.ToAddress.String()
			amount := m.Amount.String()
			return fmt.Sprintf("bank.MsgSend: %s -> %s, %s, %s", fromAddress, toAddress, amount, fees), nil
		case vm.MsgCall:
			caller := m.Caller.String()
			pkgPath := m.PkgPath
			funcName := m.Func
			args := strings.Join(m.Args, ",")
			return fmt.Sprintf("vm.MsgCall: %s, %s, %s, %s, %s", caller, pkgPath, funcName, args, fees), nil
		case vm.MsgAddPackage:
			creator := m.Creator.String()
			pkgPath := m.Package.Path
			pkgName := m.Package.Name
			return fmt.Sprintf("vm.MsgAddPackage: %s, %s, %s, %s", creator, pkgPath, pkgName, fees), nil
		case vm.MsgRun:
			return fmt.Sprintf("vm.MsgRun: %s, %s, %s, %s", m.Caller, m.Package.Path, m.Package.Name, fees), nil
		default:
			return "", fmt.Errorf("unknown or unsupported message type: %T", m)
		}
	}
	return "", fmt.Errorf("no message found in transaction")
}

func main() {
	data := "CnQKCi92bS5tX2NhbGwSZgooZzFxODdwcHVrang4bXA3bjcwZmFsOGVwdDZhandyMGR0dGFmbGQ2OCIaZ25vLmxhbmQvci9wb3VscHkxMzM3L2hvbWUqDEFkZF9jb21tZW50czIQWU8gTEVTIEFNSVMgXG5cbhITCICt4gQSDDEwMDAwMDB1Z25vdBp+CjoKEy90bS5QdWJLZXlTZWNwMjU2azESIwohAyog3sIWA1G/Y43nCkI44R47Tx9uzsxu6C9KhDL2g4pAEkAc/Nv1EPDl2vCkaBC8E2e8exUhJMWf7VSltTCDmUb3Qi0g9DOF5qvd9cpP+7hS7mJ7uw7pdvaLiEqiPwEvxI9q"
	tx, err := DecodeStdTxFromBase64(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(GetDataFromStdTx(tx))
}
