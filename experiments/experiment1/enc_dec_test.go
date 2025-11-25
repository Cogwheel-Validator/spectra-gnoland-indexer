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
	raw_data := "CpABCgovdm0ubV9jYWxsEoEBCihnMTcyOTBjd3ZtcmFwdnA4Njl4Zm5oaGF3YThzbTllZHB1ZnphdDdkIhZnbm8ubGFuZC9yL2dub3N3YXAvZ25zKghUcmFuc2ZlcjIoZzEyejc3bHk0aHdyeWtqMnphZmNoZmR2bGxmZm5wbDU1d2dzcHVjbDIJMTAwMDAwMDAwCpwBCgovdm0ubV9jYWxsEo0BCihnMTcyOTBjd3ZtcmFwdnA4Njl4Zm5oaGF3YThzbTllZHB1ZnphdDdkIiJnbm8ubGFuZC9yL2dub3N3YXAvdGVzdF90b2tlbi91c2RjKghUcmFuc2ZlcjIoZzEyejc3bHk0aHdyeWtqMnphZmNoZmR2bGxmZm5wbDU1d2dzcHVjbDIJMTAwMDAwMDAwCpsBCgovdm0ubV9jYWxsEowBCihnMTcyOTBjd3ZtcmFwdnA4Njl4Zm5oaGF3YThzbTllZHB1ZnphdDdkIiFnbm8ubGFuZC9yL2dub3N3YXAvdGVzdF90b2tlbi9iYXIqCFRyYW5zZmVyMihnMTJ6NzdseTRod3J5a2oyemFmY2hmZHZsbGZmbnBsNTV3Z3NwdWNsMgkxMDAwMDAwMDAKmwEKCi92bS5tX2NhbGwSjAEKKGcxNzI5MGN3dm1yYXB2cDg2OXhmbmhoYXdhOHNtOWVkcHVmemF0N2QiIWduby5sYW5kL3IvZ25vc3dhcC90ZXN0X3Rva2VuL2JheioIVHJhbnNmZXIyKGcxMno3N2x5NGh3cnlrajJ6YWZjaGZkdmxsZmZucGw1NXdnc3B1Y2wyCTEwMDAwMDAwMAqbAQoKL3ZtLm1fY2FsbBKMAQooZzE3MjkwY3d2bXJhcHZwODY5eGZuaGhhd2E4c205ZWRwdWZ6YXQ3ZCIhZ25vLmxhbmQvci9nbm9zd2FwL3Rlc3RfdG9rZW4vb2JsKghUcmFuc2ZlcjIoZzEyejc3bHk0aHdyeWtqMnphZmNoZmR2bGxmZm5wbDU1d2dzcHVjbDIJMTAwMDAwMDAwCpsBCgovdm0ubV9jYWxsEowBCihnMTcyOTBjd3ZtcmFwdnA4Njl4Zm5oaGF3YThzbTllZHB1ZnphdDdkIiFnbm8ubGFuZC9yL2dub3N3YXAvdGVzdF90b2tlbi9mb28qCFRyYW5zZmVyMihnMTJ6NzdseTRod3J5a2oyemFmY2hmZHZsbGZmbnBsNTV3Z3NwdWNsMgkxMDAwMDAwMDAKmwEKCi92bS5tX2NhbGwSjAEKKGcxNzI5MGN3dm1yYXB2cDg2OXhmbmhoYXdhOHNtOWVkcHVmemF0N2QiIWduby5sYW5kL3IvZ25vc3dhcC90ZXN0X3Rva2VuL3F1eCoIVHJhbnNmZXIyKGcxMno3N2x5NGh3cnlrajJ6YWZjaGZkdmxsZmZucGw1NXdnc3B1Y2wyCTEwMDAwMDAwMBITCICEr18SDDEwMDAwMDB1Z25vdBp+CjoKEy90bS5QdWJLZXlTZWNwMjU2azESIwohAtGyA6l3UIrUup5z7yXo90bXDcXUOmLiK34YPffgQ6pAEkCg8pAtBv6Fhw98bKYdqrEX2UrcjYvIUpbGdxMyc5Zpfl1cMNo8G6vzpnFaQESX+7eZIFTfO5BCFjVJLfC5Ur5C"

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

	// decode the last base64 to the sha256
	rawByte, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The final base64 decoded has %v bytes", len(rawByte))
	fmt.Printf("The first decoded base64 has %v bytes", len(base64Decoded))
}
