package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"testing"
)

// In Gnoland txs are marked with base64 not in hex like in other Cosmos SDK chains
// To get the txhash we need to query the block method and under the data field we can find the txs
// This data reveals basic info such as tx type, action happening, which package was used and what occured
// To get something that resembles txhash we need then to encode the data to sha256 and then to base64
// This can be later used to query the tx from the rpc client via the tx method
// Same data is pressent within the tx query method so this should be useful for the indexer
func TestEncodeDecode(t *testing.T) {
	raw_data := "CnQKCi92bS5tX2NhbGwSZgooZzFxODdwcHVrang4bXA3bjcwZmFsOGVwdDZhandyMGR0dGFmbGQ2OCIaZ25vLmxhbmQvci9wb3VscHkxMzM3L2hvbWUqDEFkZF9jb21tZW50czIQWU8gTEVTIEFNSVMgXG5cbhITCICt4gQSDDEwMDAwMDB1Z25vdBp+CjoKEy90bS5QdWJLZXlTZWNwMjU2azESIwohAyog3sIWA1G/Y43nCkI44R47Tx9uzsxu6C9KhDL2g4pAEkAc/Nv1EPDl2vCkaBC8E2e8exUhJMWf7VSltTCDmUb3Qi0g9DOF5qvd9cpP+7hS7mJ7uw7pdvaLiEqiPwEvxI9q"

	// decode the raw data from base64 to string
	base64Decoded, err := base64.StdEncoding.DecodeString(raw_data)
	if err != nil {
		log.Fatal(err)
	}

	// print the decoded string
	fmt.Println(string(base64Decoded))

	// encode to sha256
	hash := sha256.Sum256(base64Decoded)
	fmt.Println(hex.EncodeToString(hash[:]))

	// encode the encoded sha256 to base64
	base64Encoded := base64.StdEncoding.EncodeToString(hash[:])
	fmt.Println(base64Encoded)
}
