package rpcclient

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"strings"
	"time"
)

func NewRpcClient(rpcURL string, rpcPort *int, timeout *time.Duration) (*RpcGnoland, error) {
	if rpcURL == "" {
		return nil, errors.New("rpcURL is required")
	}

	if strings.HasPrefix(rpcURL, "https") && rpcPort == nil {
		rpcPort = new(int)
		*rpcPort = 443
	}

	if rpcPort == nil {
		// assume it uses the default rpc port 26657 if not specified
		rpcPort = new(int)
		*rpcPort = 26657
	}

	// default timeout is 10 seconds
	// maybe increase it later?
	if timeout == nil {
		timeout = new(time.Duration)
		*timeout = 10 * time.Second
	}

	hostPort := net.JoinHostPort(rpcURL, fmt.Sprintf("%d", *rpcPort))
	client, err := net.Dial("tcp", hostPort)
	if err != nil {
		return nil, err
	}

	return &RpcGnoland{
		rpcURL:  rpcURL,
		port:    rpcPort,
		client:  rpc.NewClient(client),
		timeout: *timeout,
	}, nil
}

// Only add methods that will be used by the indexer.
// Add future methods here.
const (
	Validators = "validators"
	Block      = "block"
	AbciQuery  = "abci_query"
	// might be useful for health check
	Health = "health"
	Tx     = "tx"
)

// Sync call to get the health of the rpc client.
//
// Returns:
// - error: if the call fails
func (r *RpcGnoland) Health() error {
	var response HealthResponse
	// There is no need to use Go here because we are not waiting for a response and we need
	// to be sure that the call is successful. Hence why blocking is fine in this case.
	err := r.client.Call(Health, nil, &response)
	if err != nil {
		return fmt.Errorf("failed to get health: %v", err)
	}

	return nil
}

// Async call to get validators from the rpc client.
//
// Args:
// - height: the height of the block to get the validators for
//
// Returns:
// - *ValidatorsResponse: the response from the rpc client
// - error: if the call fails
func (r *RpcGnoland) GetValidators(height uint64) (*ValidatorsResponse, error) {
	timeout := time.NewTimer(r.timeout)
	response := &ValidatorsResponse{}
	call := r.client.Go(Validators, height, response, nil)
	select {
	case <-timeout.C:
		return nil, fmt.Errorf("timeout waiting for validators")
	case err := <-call.Done:
		if err != nil {
			return nil, fmt.Errorf("failed to get validators: %v", err)
		}
	}
	return response, nil
}

// Async call to get a block from the rpc client.
//
// Args:
// - height: the height of the block to get
//
// Returns:
// - *BlockResponse: the response from the rpc client
// - error: if the call fails
func (r *RpcGnoland) GetBlock(height uint64) (*BlockResponse, error) {
	timeout := time.NewTimer(r.timeout)
	response := &BlockResponse{}
	call := r.client.Go(Block, height, response, nil)
	select {
	case <-timeout.C:
		return nil, fmt.Errorf("timeout waiting for block")
	case err := <-call.Done:
		if err != nil {
			return nil, fmt.Errorf("failed to get block: %v", err)
		}
	}

	return response, nil
}
