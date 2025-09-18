package generator

import (
	"fmt"
	"log"
	"strings"
)

// TestAuthenticGeneration demonstrates the new authentic key and address generation
func TestAuthenticGeneration() {
	fmt.Println("=== Testing Authentic Address & PubKey Generation ===")

	// Create a new data generator
	generator := NewDataGenerator()

	fmt.Printf("Generated %d key pairs in the pool\n\n", len(generator.keyPairPool))

	// Test address generation
	fmt.Println("Generated Addresses (from pool):")
	for i := 0; i < 10; i++ {
		addr := generator.GenerateAddress()
		fmt.Printf("%d. %s (valid: %t)\n", i+1, addr, ValidateAddress(addr))
	}

	fmt.Println("\nGenerated Public Keys (from pool):")
	for i := 0; i < 5; i++ {
		pubkey := generator.GeneratePubKey()
		fmt.Printf("%d. %s\n", i+1, pubkey[:50]+"...")
		fmt.Printf("   Valid: %t\n", ValidatePubKey(pubkey))
	}

	// Test key pair access
	fmt.Println("\nRandom Key Pairs with Full Details:")
	for i := 0; i < 3; i++ {
		kp := generator.GetRandomKeyPair()
		fmt.Printf("KeyPair %d:\n", i+1)
		fmt.Printf("  Address:    %s\n", kp.AddressBech32)
		fmt.Printf("  PubKey:     %s\n", kp.PubKeyBech32[:50]+"...")
		fmt.Printf("  PrivKey:    %s\n", kp.PrivKeyHex[:32]+"...")
		fmt.Println()
	}

	// Test synthetic transaction generation with authentic addresses
	fmt.Println("Sample Transaction with Authentic Addresses:")
	tx := generator.GenerateTransaction()

	for i, event := range tx.Events {
		fmt.Printf("Event %d: %s\n", i+1, event.Type)
		for _, attr := range event.Attrs {
			// Check if this attribute contains an address
			if attr.Key == "from" || attr.Key == "to" || attr.Key == "Creator" || attr.Key == "Author" {
				if ValidateAddress(attr.Value) {
					fmt.Printf("  %s: %s ✓ (authentic)\n", attr.Key, attr.Value)
				} else {
					fmt.Printf("  %s: %s ✗ (not authentic)\n", attr.Key, attr.Value)
				}
			} else {
				fmt.Printf("  %s: %s\n", attr.Key, attr.Value)
			}
		}
		fmt.Println()
	}
}

// ExampleUsage shows how to use the generator in your integration tests
func ExampleUsage() {
	fmt.Println("=== Example Usage for Integration Tests ===")

	generator := NewDataGenerator()

	// Generate authentic data for your tests
	fromAddr := generator.GenerateAddress()
	toAddr := generator.GenerateAddress()
	pubKey := generator.GeneratePubKey()

	fmt.Printf("From Address: %s\n", fromAddr)
	fmt.Printf("To Address:   %s\n", toAddr)
	fmt.Printf("Public Key:   %s\n", pubKey)

	// Validate they are authentic
	fmt.Printf("From Address Valid: %t\n", ValidateAddress(fromAddr))
	fmt.Printf("To Address Valid:   %t\n", ValidateAddress(toAddr))
	fmt.Printf("Public Key Valid:   %t\n", ValidatePubKey(pubKey))

	// You can also get the full key pair for more advanced operations
	keyPair := generator.GetRandomKeyPair()
	fmt.Printf("\nFull KeyPair for advanced operations:\n")
	fmt.Printf("Address: %s\n", keyPair.AddressBech32)
	fmt.Printf("PubKey:  %s\n", keyPair.PubKeyBech32)

	// The private key is also available if needed for signing
	fmt.Printf("PrivKey available for signing operations: %t\n", keyPair.PrivateKey != nil)
}

// RunAllTests runs all the test functions
func RunAllTests() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Test failed with panic: %v", r)
		}
	}()

	TestAuthenticGeneration()
	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")
	ExampleUsage()
}
