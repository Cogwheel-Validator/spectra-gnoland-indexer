package rpcclient_test

import (
	"fmt"
	"testing"

	rpcclient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/src/rpc_client"
)

func TestGetValidators(t *testing.T) {
	fmt.Println("Starting tests")
	fmt.Println(
		"Make sure to have the rpc client running and change any hardcoded values to something that is valid\n",
		"--------------------------------",
	)
	rpcClient, err := rpcclient.NewRpcClient("https://gnoland-testnet-rpc.cogwheel.zone", nil)
	if err != nil {
		t.Fatalf("failed to create rpc client: %v", err)
	}

	height := uint64(539847)

	validators, rpcErr := rpcClient.GetValidators(height)
	if rpcErr != nil {
		t.Fatalf("failed to get validators: %v", rpcErr)
	}

	if validators == nil {
		t.Fatal("validators is nil")
	}

}

func TestGetBlock(t *testing.T) {
	rpcClient, err := rpcclient.NewRpcClient("https://gnoland-testnet-rpc.cogwheel.zone", nil)
	if err != nil {
		t.Fatalf("failed to create rpc client: %v", err)
	}

	height := uint64(539847)

	block, rpcErr := rpcClient.GetBlock(height)
	if rpcErr != nil {
		t.Fatalf("failed to get block: %v", rpcErr)
	}

	if block == nil {
		t.Fatal("block is nil")
	}

}

func TestGetTx(t *testing.T) {
	rpcClient, err := rpcclient.NewRpcClient("https://gnoland-testnet-rpc.cogwheel.zone", nil)
	if err != nil {
		t.Fatalf("failed to create rpc client: %v", err)
	}

	txHash := "SUuY8TthP9MlOs28bWDRlsuTan19TLca5fdTQJf1v4w="

	tx, rpcErr := rpcClient.GetTx(txHash)
	if rpcErr != nil {
		t.Fatalf("failed to get tx: %v", rpcErr)
	}

	if tx == nil {
		t.Fatal("tx is nil")
	}

	txHash = "invalid_tx_hash"

	tx, rpcErr = rpcClient.GetTx(txHash)
	if rpcErr == nil {
		t.Fatalf("expected error but got none")
	}

	if tx != nil {
		t.Fatalf("expected tx to be nil but got %v", tx)
	}
}

func TestHealthCheck(t *testing.T) {
	rpcClient, err := rpcclient.NewRpcClient("https://gnoland-testnet-rpc.cogwheel.zone", nil)
	if err != nil {
		t.Fatalf("failed to create rpc client: %v", err)
	}

	rpcErr := rpcClient.Health()
	if rpcErr != nil {
		t.Fatalf("failed to get health: %v", rpcErr)
	}
}

func TestFailedSetupClient(t *testing.T) {
	_, err := rpcclient.NewRpcClient("gnoland-testnet-rpc.cogwheel.zone", nil)
	// should fail because the url is not valid
	if err == nil {
		t.Fatalf("expected error but got none")
	}
}
