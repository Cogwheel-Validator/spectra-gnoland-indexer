package dataprocessor

import (
	"fmt"

	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	sqlDataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
)

// EventFormat represents the format type of the returned data
type EventFormat int

const (
	NativeFormat EventFormat = iota
	CompressedFormat
)

// EventResult holds the result of EventSolver with type discrimination
type EventResult struct {
	Format         EventFormat
	NativeEvents   []sqlDataTypes.Event // populated when Format == NativeFormat
	CompressedData []byte               // populated when Format == CompressedFormat
}

// IsNative returns true if the result contains native format data
func (er *EventResult) IsNative() bool {
	return er.Format == NativeFormat
}

// IsCompressed returns true if the result contains compressed format data
func (er *EventResult) IsCompressed() bool {
	return er.Format == CompressedFormat
}

// GetNativeEvents returns the native events if available, nil otherwise
func (er *EventResult) GetNativeEvents() []sqlDataTypes.Event {
	if er.Format == NativeFormat {
		return er.NativeEvents
	}
	return nil
}

// GetCompressedData returns the compressed data if available, nil otherwise
func (er *EventResult) GetCompressedData() []byte {
	if er.Format == CompressedFormat {
		return er.CompressedData
	}
	return nil
}

// TODO: when the dict training is done this will need to hold
// 2 types of storing the data
// 1. native postgres format
// 2. protobuf format
//
// Untill the training is done and safe to use focus on the native postgres format.

// EventSolver is a function that solves the event of a transaction
// it will solve the event of a transaction and return the event
//
// It can return data in two formats:
// 1. Native postgres format ([]sqlDataTypes.Event)
// 2. Compressed protobuf format ([]byte)
//
// Args:
//   - txResponse: a transaction response
//   - useCompressed: if true, returns compressed format; otherwise native format
//
// Returns:
//   - *EventResult: contains either native events or compressed data
//   - error: an error if the event solving fails
func EventSolver(txResponse *rpcClient.TxResponse, useCompressed bool) (*EventResult, error) {
	if useCompressed {
		// TODO: Implement compressed format when protobuf training is complete
		// This would involve:
		// 1. Converting to protobuf format
		// 2. Compressing with zstandard
		// 3. Returning the compressed bytes
		return nil, fmt.Errorf("compressed format not yet implemented")
	}

	// Native format implementation
	events := make([]sqlDataTypes.Event, 0, len(txResponse.Result.TxResult.ResponseBase.Events))
	for _, event := range txResponse.Result.TxResult.ResponseBase.Events {
		attributes := make([]sqlDataTypes.Attribute, 0, len(event.Attrs))
		for _, attribute := range event.Attrs {
			attributes = append(attributes, sqlDataTypes.Attribute{
				Key:   attribute.Key,
				Value: attribute.Value,
			})
		}
		events = append(events, sqlDataTypes.Event{
			AtType:     event.AtType,
			Type:       event.Type,
			Attributes: attributes,
			PkgPath:    event.PkgPath,
		})
	}

	return &EventResult{
		Format:       NativeFormat,
		NativeEvents: events,
	}, nil
}
