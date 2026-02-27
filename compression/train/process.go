package train

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/events_proto"
	"github.com/valyala/gozstd"
	"google.golang.org/protobuf/proto"
)

// Collect collects the transactions from the database
//
// Usage:
//
// # Used to collect the transactions from the database
//
// Parameters:
//   - db: the database connection pool
//   - chainName: the name of the chain
//   - amount: the amount of transactions to collect
//
// Returns:
//   - [][]byte: the transactions events in serialized protobuf format
//   - error: if the transactions fail to collect
func CollectEvents(db *database.TimescaleDb, chainName string, amount uint64) ([][]byte, error) {
	// define the limits and offset
	if amount > 50000 {
		return nil, fmt.Errorf("amount cannot be greater than 50000")
	}
	if amount == 0 {
		return nil, fmt.Errorf("amount cannot be 0")
	}
	var limit uint64
	var offset uint64
	var goroutines int
	if amount <= 100 {
		limit = amount
	} else {
		limit = 100
	}

	// use goroutines to get all of the needed data faster
	if offset > 0 {
		goroutines = int(offset)
	} else {
		goroutines = 1
	}

	transactions := make([]*database.Transaction, 0)
	wg := sync.WaitGroup{}
	wg.Add(goroutines)
	mu := sync.Mutex{}
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			offset := uint64(i) * limit
			transactions, err := db.GetTransactionsByOffset(context.Background(), chainName, limit, offset)
			if err != nil {
				log.Printf("failed to get transactions by offset: %v", err)
				return
			}
			mu.Lock()
			transactions = append(transactions, transactions...)
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	events := make([][]byte, 0)
	for _, transaction := range transactions {
		txEvents := &transaction.TxEvents
		if len(*txEvents) > 0 {
			for _, event := range *txEvents {
				protoAttrs := make([]*events_proto.Attribute, 0)
				for _, attribute := range event.Attributes {
					protoAttrs = append(protoAttrs, events_proto.NewAttributeFromString(attribute.Key, attribute.Value))
				}
				protoEv := &events_proto.Event{
					AtType:     event.AtType,
					Type:       event.Type,
					Attributes: protoAttrs,
					PkgPath:    &event.PkgPath,
				}
				bs, err := proto.Marshal(protoEv)
				if err != nil {
					log.Printf("failed to marshal event: %v", err)
					continue
				}
				events = append(events, bs)
			}
		}
	}
	return events, nil
}

// BuildZstdDict builds the zstd dict from the events
func BuildZstdDict(events [][]byte) []byte {
	// limit zstd dict to 64KB
	return gozstd.BuildDict(events, 64*1024)
}
