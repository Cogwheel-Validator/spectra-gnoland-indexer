package generator

import (
	"fmt"
	"log"
	"testing"
)

// TestAuthenticGeneration demonstrates the new authentic key and address generation
func TestAuthenticGeneration(t *testing.T) {
	fmt.Println("Testing Authentic Address & PubKey Generation")

	// Create a new data generator
	generator := NewDataGenerator(500)

	fmt.Printf("Generated %d key pairs in the pool\n\n", len(generator.keyPairPool))

	// Test address generation
	fmt.Println("Generated Addresses (from pool):")
	// test up to a 1k addresses
	for i := range 500 {
		addr := generator.GenerateAddress()
		fmt.Printf("%d. %s (valid: %t)\n", i+1, addr, ValidateAddress(addr))
	}

	fmt.Println("\nGenerated Public Keys (from pool):")
	// test up to a 1k public keys
	for i := range 500 {
		pubkey := generator.GeneratePubKey()
		fmt.Printf("%d. %s\n", i+1, pubkey[:50]+"...")
		fmt.Printf("   Valid: %t\n", ValidatePubKey(pubkey))
	}

	// Test key pair access
	fmt.Println("\nRandom Key Pairs with Full Details:")
	for i := range 500 {
		kp := generator.GetRandomKeyPair()
		fmt.Printf("KeyPair %d:\n", i+1)
		fmt.Printf("  Address:    %s\n", kp.AddressBech32)
		fmt.Printf("  PubKey:     %s\n", kp.PubKeyBech32[:50]+"...")
		fmt.Printf("  PrivKey:    %s\n", kp.PrivKeyHex[:32]+"...")
		fmt.Println()
	}

	// Test synthetic transaction generation 500 times
	fmt.Println("Sample Transaction with Authentic Addresses:")
	for range 500 {
		events, tx := generator.GenerateTransaction()
		fmt.Printf("Transaction: %+v\n", tx)
		for i := range events.Events {
			event := &events.Events[i]
			fmt.Printf("\tEvent %d: %s\n", i+1, event.Type)
			for _, attr := range event.Attributes {
				// Check if this attribute contains an address
				if attr.Key == "from" || attr.Key == "to" || attr.Key == "Creator" || attr.Key == "Author" {
					if ValidateAddress(attr.GetStringValue()) {
						fmt.Printf("\t%s: %s (authentic)\n", attr.Key, attr.Value)
					} else {
						fmt.Printf("\t%s: %s (not authentic)\n", attr.Key, attr.Value)
					}
				} else {
					fmt.Printf("\t%s: %s\n", attr.Key, attr.Value)
				}
			}
			fmt.Println()
		}
	}
}

// RunAllTests runs all the test functions
func RunAllTests(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Test failed with panic: %v", r)
		}
	}()

	TestAuthenticGeneration(t)
}
