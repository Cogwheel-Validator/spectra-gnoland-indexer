package rpcclient

import "fmt"

// RpcHeightError represents an RPC error that includes the height context for retry purposes
type RpcHeightError struct {
	Height    uint64
	HasHeight bool // indicates if Height field is valid
	Err       error
}

// Error implements the error interface
func (e *RpcHeightError) Error() string {
	if e.HasHeight {
		return fmt.Sprintf("rpc error at height %d: %v", e.Height, e.Err)
	}
	return fmt.Sprintf("rpc error: %v", e.Err)
}

// RpcStringError represents an RPC error that includes string context (like tx hash) for retry purposes
type RpcStringError struct {
	Value    string
	HasValue bool // indicates if Value field is valid
	Err      error
}

// Error implements the error interface
func (e *RpcStringError) Error() string {
	if e.HasValue {
		return fmt.Sprintf("rpc error for %s: %v", e.Value, e.Err)
	}
	return fmt.Sprintf("rpc error: %v", e.Err)
}
