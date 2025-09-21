package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
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

func printAllData(tx *std.Tx) {
	fmt.Printf("Transaction contains %d messages:\n", len(tx.Msgs))
	fmt.Printf("Basic tx data: %v, %v, %v, %v\n", tx.Fee, tx.Memo, tx.GetSigners(), tx.Signatures)
	fmt.Println(tx.Msgs)
	for i, msg := range tx.Msgs {
		fmt.Printf("Message %d: %T\n", i+1, msg)
		switch m := msg.(type) {
		case vm.MsgCall:
			fmt.Printf("  - vm.MsgCall: caller=%s, pkg=%s, func=%s\n", m.Caller, m.PkgPath, m.Func)
		case vm.MsgAddPackage:
			fmt.Printf("  - vm.MsgAddPackage: creator=%s, pkg=%s\n", m.Creator, m.Package.Path)
		case vm.MsgRun:
			fmt.Printf("  - vm.MsgRun: caller=%s, pkg=%s\n", m.Caller, m.Package.Path)
		case bank.MsgSend:
			fmt.Printf("  - bank.MsgSend: from=%s, to=%s, amount=%s\n", m.FromAddress, m.ToAddress, m.Amount)
		}
	}
}

func encodeData() {
	// try to recreate and encode the data
	address, err := crypto.AddressFromString("g1q2hsksnh5q55xkm3d8tul28gg9uuwl2em8ver6")
	if err != nil {
		log.Fatalf("Error creating address: %v", err)
	}
	// lets use real pubkey for now
	pubkey, err := crypto.PubKeyFromBech32("gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqdrkcn740564dys82xamnfsphzmrzzuf4eqkygx0wywzqdrmml2lj6rxjtk")
	if err != nil {
		log.Fatalf("Error creating pubkey: %v", err)
	}
	msgs := []std.Msg{
		vm.MsgCall{
			Caller:  address,
			PkgPath: "gno.land/r/demo/profile",
			Func:    "SetStringField",
			MaxDeposit: []std.Coin{
				{
					Amount: 1000000,
					Denom:  "ugnot",
				},
			},
			Args: []string{"Ready for lift off"},
		},
	}
	tx := std.Tx{
		Msgs: msgs,
	}
	tx.Fee = std.Fee{
		GasFee: std.Coin{
			Amount: 1000000,
			Denom:  "ugnot",
		},
		GasWanted: 1000000,
	}
	tx.Memo = "Ready for lift off"
	tx.Signatures = []std.Signature{
		{PubKey: pubkey, Signature: []byte("signature")},
	}

	// now it should encode the data and then the bytes need to be encoded to base64
	bz, err := amino.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}
	base64Encoded := base64.StdEncoding.EncodeToString(bz)
	fmt.Println("base64Encoded: ", base64Encoded)
}

func main() {
	data := "Cn4KCi92bS5tX2NhbGwScAooZzFxMmhza3NuaDVxNTV4a20zZDh0dWwyOGdnOXV1d2wyZW04dmVyNiIXZ25vLmxhbmQvci9kZW1vL3Byb2ZpbGUqDlNldFN0cmluZ0ZpZWxkMgtEaXNwbGF5TmFtZTIObm9kZXJ1bm5lcmluZG8SEwiAh6cOEgwxMDAwMDAwdWdub3Qafgo6ChMvdG0uUHViS2V5U2VjcDI1NmsxEiMKIQNHbE/VfTVWkgdRu7mmAbi2MQuJrkFiIM9xHCA0e9/V+RJAVSpcFFNgOO8K/XrXfdUOW9SpDTwhsnYsWpWEWcOo0wVJRFtDVzTb2qkkm3vao2uteVkXIKrLh2vle2yGY5v8bA=="
	tx, err := DecodeStdTxFromBase64(data)
	if err != nil {
		log.Fatal(err)
	}
	printAllData(tx)

	encodeData()
	// print the first encoded data it will not match this but i think there should be some similaraties
	fmt.Println("first encoded data: ", data)
}
