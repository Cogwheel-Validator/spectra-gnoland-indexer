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
	raw_data := "CnQKDS9iYW5rLk1zZ1NlbmQSYwooZzE0ODU4M3Q1eDY2enM2cDkwZWhhZDZsNHFlZmV5YWY1NHM2OXdxbBIoZzFrM2NjcDA5NzdhajMyeXJqeHBtdjA5d2hxZ2w3bjgzcjZjc2pyYxoNMTAwMDAwMDB1Z25vdBIRCIDaxAkSCjEwMDAwdWdub3Qafgo6ChMvdG0uUHViS2V5U2VjcDI1NmsxEiMKIQMcQg5ueuQfcNtMh5w7PjWaTKO72epQaK9kax/g3+C9NRJAYMsAIGMPOzc964cBKZ49xg+KBO/5T3RJLhjDzO9x5a5Kr4sX+78p3QLvOV6a30htoJc11XYGV5tymCLgewf8lg=="

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
