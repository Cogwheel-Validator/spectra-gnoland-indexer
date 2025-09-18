package main

import (
	"encoding/base64"
	"encoding/hex"
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
	signatures := tx.GetSignatures()
	for _, signature := range signatures {
		pubKey := signature.PubKey.String()
		signatureHex := hex.EncodeToString(signature.Signature)

		fmt.Printf("signature: %v %v \n", pubKey, signatureHex)
	}
	signers := tx.GetSigners()

	for _, msg := range tx.GetMsgs() {
		switch m := msg.(type) {
		case bank.MsgSend:
			fromAddress := m.FromAddress.String()
			toAddress := m.ToAddress.String()
			amount := m.Amount.String()
			return fmt.Sprintf("bank.MsgSend: %s -> %s, amount: %s, fees: %s, signatures: %s, signers: %s \n", fromAddress, toAddress, amount, fees, signatures[0], signers), nil
		case vm.MsgCall:
			caller := m.Caller.String()
			pkgPath := m.PkgPath
			funcName := m.Func
			args := strings.Join(m.Args, ",")
			return fmt.Sprintf("vm.MsgCall, caller: %s, pkgPath: %s, funcName: %s, args: %s, fees: %s, signatures: %s, signers: %s \n", caller, pkgPath, funcName, args, fees, signatures[0], signers), nil
		case vm.MsgAddPackage:
			creator := m.Creator.String()
			pkgPath := m.Package.Path
			pkgName := m.Package.Name
			pkgFiles := m.Package.FileNames()
			pkgType := m.Package.Type
			return fmt.Sprintf(
				"vm.MsgAddPackage, creator: %s, pkgPath: %s, pkgName: %s, pkgFiles: %s, pkgType: %s, fees: %s, signatures: %s, signers: %s \n", creator, pkgPath, pkgName, pkgFiles, pkgType, fees, signatures[0], signers), nil
		case vm.MsgRun:
			caller := m.Caller.String()
			pkgPath := m.Package.Path
			pkgName := m.Package.Name
			pkgFiles := m.Package.FileNames()
			pkgType := m.Package.Type
			return fmt.Sprintf("vm.MsgRun, caller: %s, pkgPath: %s, pkgName: %s, pkgFiles: %s, pkgType: %s, fees: %s, signatures: %s, signers: %s \n", caller, pkgPath, pkgName, pkgFiles, pkgType, fees, signatures[0], signers), nil
		default:
			return "", fmt.Errorf("unknown or unsupported message type: %T", m)
		}
	}
	return "", fmt.Errorf("no message found in transaction")
}

func main() {
	data := "CrQFCgwvdm0ubV9hZGRwa2cSowUKKGcxamc4bXR1dHU5a2hoZndjNG54bXVoY3BmdGYwcGFqZGhmdnNxZjUS9gQKF21hdHRoZXdfc3RvcmFnZV9kZXBvc2l0Ektnbm8ubGFuZC9yL2cxamc4bXR1dHU5a2hoZndjNG54bXVoY3BmdGYwcGFqZGhmdnNxZjUvbWF0dGhld19zdG9yYWdlX2RlcG9zaXQacQoLZ25vbW9kLnRvbWwSYm1vZHVsZSA9ICJnbm8ubGFuZC9yL2cxamc4bXR1dHU5a2hoZndjNG54bXVoY3BmdGYwcGFqZGhmdnNxZjUvbWF0dGhld19zdG9yYWdlX2RlcG9zaXQiCmdubyA9ICIwLjkiGvYCCgtwYWNrYWdlLmdubxLmAnBhY2thZ2UgbWF0dGhld19zdG9yYWdlX2RlcG9zaXQKCnZhciBkYXRhID0gbWFrZShtYXBbc3RyaW5nXXN0cmluZykKCmZ1bmMgRGVwb3NpdChjdXIgcmVhbG0sIGtleSBzdHJpbmcpIHsKCWRhdGFba2V5XSA9ICIiCglmb3IgaSA6PSAwOyBpIDwgMTAwMDA7IGkrKyB7CgkJZGF0YVtrZXldID0gZGF0YVtrZXldICsgInRlc3R0ZXN0dGVzdHRlc3R0ZXN0dGVzdHRlc3R0ZXN0dGVzdHRlc3R0ZXN0dGVzdHRlc3R0ZXN0dGVzdHRlc3R0ZXN0dGVzdHRlc3R0ZXN0dGVzdHRlc3R0ZXN0dGVzdHRlc3R0ZXN0dGVzdHRlc3QiCgl9Cn0KCmZ1bmMgUmVsZWFzZShjdXIgcmVhbG0sIGtleSBzdHJpbmcpIHsKCWRlbGV0ZShkYXRhLCBrZXkpCn0iIgoTL2duby5NZW1QYWNrYWdlVHlwZRILCglNUFVzZXJBbGwSEwiAwtcvEgwxMDAwMDAwdWdub3Qafgo6ChMvdG0uUHViS2V5U2VjcDI1NmsxEiMKIQPhYTbbFx4y30iZNZQfBW4i+Jhj43OdCrfNSexCg5ydshJAeKL4UNZFxab5c3gSFfFIh6supP4w2K2kQlfOPe6z1wA49aQz81Y4ZbwwiFulqXusxR62nuA6pjSX9UpGN1oU/w=="
	tx, err := DecodeStdTxFromBase64(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(GetDataFromStdTx(tx))

	data2 := "Cn4KCi92bS5tX2NhbGwScAooZzFxMmhza3NuaDVxNTV4a20zZDh0dWwyOGdnOXV1d2wyZW04dmVyNiIXZ25vLmxhbmQvci9kZW1vL3Byb2ZpbGUqDlNldFN0cmluZ0ZpZWxkMgtEaXNwbGF5TmFtZTIObm9kZXJ1bm5lcmluZG8SEwiAh6cOEgwxMDAwMDAwdWdub3Qafgo6ChMvdG0uUHViS2V5U2VjcDI1NmsxEiMKIQNHbE/VfTVWkgdRu7mmAbi2MQuJrkFiIM9xHCA0e9/V+RJAVSpcFFNgOO8K/XrXfdUOW9SpDTwhsnYsWpWEWcOo0wVJRFtDVzTb2qkkm3vao2uteVkXIKrLh2vle2yGY5v8bA=="
	tx2, err := DecodeStdTxFromBase64(data2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(GetDataFromStdTx(tx2))
}
