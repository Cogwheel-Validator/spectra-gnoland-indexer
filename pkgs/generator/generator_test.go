package generator

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAuthenticGeneration demonstrates the new authentic key and address generation
func TestAuthenticGeneration(t *testing.T) {
	// Create a new data generator
	generator := NewDataGenerator(500)

	// Test address generation
	// test up to a 1k addresses
	for range 500 {
		addr := generator.GenerateAddress()
		assert.True(t, ValidateAddress(addr))
	}

	// test up to a 1k public keys
	for range 500 {
		pubkey := generator.GeneratePubKey()
		assert.True(t, ValidatePubKey(pubkey))
	}

	// Test key pair access
	for range 500 {
		kp := generator.GetRandomKeyPair()
		assert.NotNil(t, kp)
		assert.NotEmpty(t, kp.Address)
		assert.NotEmpty(t, kp.PublicKey)
	}

	// Test synthetic transaction generation 500 times
	for range 500 {
		events, tx := generator.GenerateTransaction()
		assert.NotNil(t, tx)
		assert.NotNil(t, events.Events)
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
