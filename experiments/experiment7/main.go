package main

import (
	"encoding/hex"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

func generateKeyPair() (crypto.PrivKey, crypto.PubKey, crypto.Address) {
	// Generate a new private key using secp256k1
	privKey := secp256k1.GenPrivKey()

	// Derive the public key from the private key
	pubKey := privKey.PubKey()

	// Derive the address from the public key
	address := pubKey.Address()

	return privKey, pubKey, address
}

func testKeyGeneration() {
	fmt.Println("=== Testing Key Generation ===")

	// Generate 5 key pairs for testing
	for i := 0; i < 5; i++ {
		privKey, pubKey, address := generateKeyPair()

		fmt.Printf("Key Pair %d:\n", i+1)
		fmt.Printf("  Private Key: %s\n", hex.EncodeToString(privKey.Bytes()))
		fmt.Printf("  Public Key:  %s\n", pubKey.String())
		fmt.Printf("  Address:     %s\n", address.String())
		fmt.Printf("  Address Bech32: %s\n", crypto.AddressToBech32(address))
		fmt.Printf("  PubKey Bech32:  %s\n", crypto.PubKeyToBech32(pubKey))
		fmt.Println()
	}
}

func main() {
	testKeyGeneration()
}
