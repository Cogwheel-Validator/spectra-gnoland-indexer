package dataprocessor

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/decoder"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	sqlDataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
)

// Constructor function for the DataProcessor struct
//
// Args:
//   - db: the database connection interface
//   - addressCache: the address cache interface
//   - validatorCache: the validator cache interface
//   - chainName: the name of the chain string
//
// Returns:
//   - *DataProcessor: the data processor
//
// The method will not throw an error if the data processor is not found, it will just return nil
func NewDataProcessor(
	db Database,
	addressCache AddressCache,
	validatorCache AddressCache,
	chainName string) *DataProcessor {
	return &DataProcessor{
		dbPool:         db,
		addressCache:   addressCache,
		validatorCache: validatorCache,
		chainName:      chainName,
	}
}

// ProcessValidatorAddresses is a method to process the validator addresses from a slice of blocks
// it will process the validator addresses from the blocks and store them in a sync.Map
// it will then extract the addresses from the sync.Map and insert them into the address cache
//
// Args:
//   - blocks: a slice of blocks
//   - fromHeight: the start height
//   - toHeight: the end height
//
// Returns:
//   - nil
//
// The method will not throw an error if the validator addresses are not found, it will just return nil
func (d *DataProcessor) ProcessValidatorAddresses(
	blocks []*rpcClient.BlockResponse,
	fromHeight uint64,
	toHeight uint64,
) {
	// sync.Map for thread-safe concurrent access and inser it as address/bool
	// the program should be able to avoid duplicates since it is a map
	var addressesMap sync.Map
	wg := sync.WaitGroup{}
	wg.Add(len(blocks))

	// Process blocks concurrently to extract addresses
	for _, block := range blocks {
		go func(block *rpcClient.BlockResponse) {
			defer wg.Done()

			// Process precommits
			precommits := block.Result.Block.LastCommit.Precommits
			for _, precommit := range precommits {
				if precommit != nil {
					addressesMap.Store(precommit.ValidatorAddress, true)
				}
			}

			// Process proposer
			proposer := block.Result.Block.Header.ProposerAddress
			addressesMap.Store(proposer, true)
		}(block)
	}

	wg.Wait()

	// Extract unique addresses from sync.Map
	addresses := make([]string, 0)
	addressesMap.Range(func(key, value interface{}) bool {
		addresses = append(addresses, key.(string))
		return true
	})

	// retry 3 times just for the sake of it
	d.validatorCache.AddressSolver(addresses, d.chainName, true, 3, nil)
	log.Printf("Validator addresses processed from %d to %d", fromHeight, toHeight)
}

// ProcessBlocks is a "swarm" method to process the blocks from a slice of blocks
// it will process the blocks using async workers and store them directly into a result slice
// it will then insert the blocks into the database
//
// Args:
//   - blocks: a slice of blocks
//   - fromHeight: the start height
//   - toHeight: the end height
//
// Returns:
//   - nil
//
// The method will not throw an error if the blocks are not found, it will just return nil
func (d *DataProcessor) ProcessBlocks(blocks []*rpcClient.BlockResponse, fromHeight uint64, toHeight uint64) {
	// Preallocate slice to avoid growing allocations
	blocksData := make([]sqlDataTypes.Blocks, len(blocks))
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	wg.Add(len(blocks))

	for idx, block := range blocks {
		go func(idx int, block *rpcClient.BlockResponse) {
			defer wg.Done()
			// decode base64 hash
			hash, err := base64.StdEncoding.DecodeString(block.Result.BlockMeta.BlockID.Hash)
			if err != nil {
				log.Printf("Failed to decode block hash %s: %v", block.Result.BlockMeta.BlockID.Hash, err)
				return
			}

			// convert from string to uint64
			height, err := strconv.ParseUint(block.Result.Block.Header.Height, 10, 64)
			if err != nil {
				log.Printf("Failed to parse block height %s: %v", block.Result.Block.Header.Height, err)
				return
			}

			// there should be an slice of strings but it can be nil
			// if slice exists we need to convert each slice from base64 to sha256
			// since it is shorter and better for the database
			var txs [][]byte
			if block.Result.Block.Data.Txs != nil {
				txs = make([][]byte, 0, len(*block.Result.Block.Data.Txs))
				for _, tx := range *block.Result.Block.Data.Txs {
					txHash, err := base64.StdEncoding.DecodeString(tx)
					// turn to sha256 and then turn it to raw bytes to
					// match the transaction hash
					txHashSha256 := sha256.Sum256(txHash)
					if err != nil {
						log.Printf("Failed to decode tx hash %s: %v", tx, err)
						continue
					}
					txs = append(txs, txHashSha256[:])
				}
			}

			// Use mutex only when writing to shared slice
			mu.Lock()
			blocksData[idx] = sqlDataTypes.Blocks{
				Hash:      hash,
				Height:    height,
				Timestamp: block.Result.Block.Header.Time,
				ChainID:   block.Result.Block.Header.ChainID,
				Txs:       txs,
				ChainName: d.chainName,
			}
			mu.Unlock()
		}(idx, block)
	}

	wg.Wait()

	err := d.dbPool.InsertBlocks(blocksData)
	if err != nil {
		log.Printf("Failed to insert blocks: %v", err)
	}
	log.Printf("Blocks processed from %d to %d", fromHeight, toHeight)
}

// ProcessTransactions is a swarm method to process the transactions from a map of transactions and timestamps
// it will process the transactions using async workers and collect them in a preallocated slice
// it will then insert the transactions into the database
//
// Args:
//   - transactions: a map of transactions and timestamps
//   - compressEvents: if true, compress the events
//
// Returns:
//   - nil
//
// The method will not throw an error if the transactions are not found, it will just return nil
func (d *DataProcessor) ProcessTransactions(
	transactions []TrasnactionsData,
	compressEvents bool,
	fromHeight uint64,
	toHeight uint64) {

	// Preallocate slice to avoid growing allocations
	transactionsData := make([]sqlDataTypes.TransactionGeneral, len(transactions))
	var mu sync.Mutex
	var validCount int
	wg := sync.WaitGroup{}
	wg.Add(len(transactions))

	for idx, transaction := range transactions {
		go func(idx int, transaction TrasnactionsData) {
			defer wg.Done()
			txResult := transaction.Response.Result.TxResult

			decodedMsg := decoder.NewDecodedMsg(transaction.Response.Result.Tx)

			fee := decodedMsg.GetFee()
			msgTypes := decodedMsg.GetMsgTypes()

			// convert the tx hash from base64 to sha256
			txHash, err := base64.StdEncoding.DecodeString(transaction.Response.GetHash())
			if err != nil {
				return
			}

			// convert the gas wanted and used from string to uint64
			gasWanted, err := strconv.ParseUint(txResult.GasWanted, 10, 64)
			if err != nil {
				return
			}
			gasUsed, err := strconv.ParseUint(txResult.GasUsed, 10, 64)
			if err != nil {
				return
			}

			// solve the events
			events, err := EventSolver(transaction.Response, compressEvents)
			if err != nil {
				return
			}

			// Use mutex only when writing to shared slice
			mu.Lock()
			transactionsData[idx] = sqlDataTypes.TransactionGeneral{
				TxHash:             txHash,
				ChainName:          d.chainName,
				Timestamp:          transaction.Timestamp,
				BlockHeight:        transaction.BlockHeight,
				MsgTypes:           msgTypes,
				TxEvents:           events.GetNativeEvents(),
				TxEventsCompressed: events.GetCompressedData(),
				CompressionOn:      compressEvents,
				GasUsed:            gasUsed,
				GasWanted:          gasWanted,
				Fee:                fee,
			}
			validCount++
			mu.Unlock()

		}(idx, transaction)
	}

	wg.Wait()

	// It is more of a safety feature than nececity
	transactionsData = transactionsData[:validCount]

	err := d.dbPool.InsertTransactionsGeneral(transactionsData)
	if err != nil {
		log.Printf("Failed to insert transactions: %v", err)
	}
	log.Printf("Transactions processed from %d to %d", fromHeight, toHeight)
}

// ProcessMessages processes all messages from transactions using concurrent "swarm method"
// This method uses a two-phase concurrent approach:
// 1. Collect and resolve all addresses to IDs using concurrent workers and sync.Map
// 2. Convert messages to database-ready format with address IDs using concurrent processing
//
// Args:
//   - transactions: a map of transactions and timestamps
//   - fromHeight: the start height
//   - toHeight: the end height
//
// Returns:
//   - error: if processing fails
func (d *DataProcessor) ProcessMessages(
	transactions []TrasnactionsData,
	fromHeight uint64,
	toHeight uint64) error {

	// Phase 1: Concurrent address collection using sync.Map
	var addressesMap sync.Map
	addressCollectionChan := make(chan []*decoder.DecodedMsg, len(transactions))
	wg1 := sync.WaitGroup{}
	wg1.Add(len(transactions))

	// Launch goroutines to collect addresses concurrently
	for _, transaction := range transactions {
		go func(transaction TrasnactionsData) {
			defer wg1.Done()
			decodedMsg := decoder.NewDecodedMsg(transaction.Response.Result.Tx)
			if decodedMsg == nil {
				addressCollectionChan <- []*decoder.DecodedMsg{nil}
				return
			}

			// Collect addresses from this transaction and store in sync.Map
			addresses := decodedMsg.CollectAllAddresses()
			for _, address := range addresses {
				addressesMap.Store(address, true) // Thread-safe deduplication
			}

			addressCollectionChan <- []*decoder.DecodedMsg{decodedMsg}
		}(transaction)
	}

	// Close channel when all address collection goroutines finish
	go func() {
		wg1.Wait()
		close(addressCollectionChan)
	}()

	// Collect decoded messages
	allDecodedMsgs := make([]*decoder.DecodedMsg, 0, len(transactions))
	for decodedMsgs := range addressCollectionChan {
		allDecodedMsgs = append(allDecodedMsgs, decodedMsgs[0])
	}

	// Extract addresses from sync.Map and resolve to IDs
	allAddresses := make([]string, 0)
	addressesMap.Range(func(key, value interface{}) bool {
		allAddresses = append(allAddresses, key.(string))
		return true
	})

	if len(allAddresses) > 0 {
		d.addressCache.AddressSolver(allAddresses, d.chainName, false, 3, nil)
		log.Printf("Resolved %d unique addresses for messages from %d to %d", len(allAddresses), fromHeight, toHeight)
	}

	// Phase 2: Concurrent message processing without channels
	// Use a mutex for aggregation instead of channels
	aggregatedDbGroups := &decoder.DbMessageGroups{
		MsgSend:   make([]sqlDataTypes.MsgSend, 0),
		MsgCall:   make([]sqlDataTypes.MsgCall, 0),
		MsgAddPkg: make([]sqlDataTypes.MsgAddPackage, 0),
		MsgRun:    make([]sqlDataTypes.MsgRun, 0),
	}
	var aggregationMutex sync.Mutex

	wg2 := sync.WaitGroup{}
	wg2.Add(len(transactions))

	// Launch goroutines to process messages concurrently
	for idx, transaction := range transactions {
		// Guard against index out of bounds
		if idx >= len(allDecodedMsgs) {
			wg2.Done()
			continue
		}

		go func(transaction TrasnactionsData, decodedMsg *decoder.DecodedMsg) {
			defer wg2.Done()

			if decodedMsg == nil {
				// There might be an error here
				// but any kind of retry mechanism will probably not help
				// log it for.
				log.Printf("The transaction couldn't be decoded, tx hash: %s", transaction.Response.GetHash())
				return
			}

			// Convert directly to database-ready messages with address IDs
			txHash, err := base64.StdEncoding.DecodeString(transaction.Response.GetHash())
			if err != nil {
				log.Printf("Failed to decode tx hash %s: %v", transaction.Response.GetHash(), err)
				return
			}

			dbMessageGroups, err := decodedMsg.ConvertToDbMessages(d.addressCache, txHash, d.chainName, transaction.Timestamp, decodedMsg.GetSigners())
			if err != nil {
				log.Printf("Failed to convert messages for tx %s: %v", transaction.Response.GetHash(), err)
				return
			}

			if dbMessageGroups != nil {
				aggregationMutex.Lock()
				aggregatedDbGroups.MsgSend = append(aggregatedDbGroups.MsgSend, dbMessageGroups.MsgSend...)
				aggregatedDbGroups.MsgCall = append(aggregatedDbGroups.MsgCall, dbMessageGroups.MsgCall...)
				aggregatedDbGroups.MsgAddPkg = append(aggregatedDbGroups.MsgAddPkg, dbMessageGroups.MsgAddPkg...)
				aggregatedDbGroups.MsgRun = append(aggregatedDbGroups.MsgRun, dbMessageGroups.MsgRun...)
				aggregationMutex.Unlock()
			}

		}(transaction, allDecodedMsgs[idx])
	}

	wg2.Wait()

	// Create a slice of sqlDataType.AddressTx
	// we need to get all of the addresses from the aggregatedDbGroups
	addresses := createAddressTx(aggregatedDbGroups)
	err := d.dbPool.InsertAddressTx(addresses)
	if err != nil {
		return fmt.Errorf("failed to insert address tx: %w", err)
	}

	// Batch insert optimized messages
	if err := d.insertDbMessageGroups(aggregatedDbGroups); err != nil {
		return fmt.Errorf("failed to insert optimized messages: %w", err)
	}

	log.Printf("Messages processed concurrently from %d to %d: MsgSend=%d, MsgCall=%d, MsgAddPkg=%d, MsgRun=%d",
		fromHeight, toHeight,
		len(aggregatedDbGroups.MsgSend),
		len(aggregatedDbGroups.MsgCall),
		len(aggregatedDbGroups.MsgAddPkg),
		len(aggregatedDbGroups.MsgRun))

	return nil
}

// insertDbMessageGroups performs optimized batch insertions using address IDs
func (d *DataProcessor) insertDbMessageGroups(groups *decoder.DbMessageGroups) error {
	var insertErrors []error

	// Insert DbMsgSend messages with address IDs
	if len(groups.MsgSend) > 0 {
		if err := d.dbPool.InsertMsgSend(groups.MsgSend); err != nil {
			insertErrors = append(insertErrors, fmt.Errorf("failed to insert DbMsgSend: %w", err))
		}
	}

	// Insert DbMsgCall messages with address IDs
	if len(groups.MsgCall) > 0 {
		if err := d.dbPool.InsertMsgCall(groups.MsgCall); err != nil {
			insertErrors = append(insertErrors, fmt.Errorf("failed to insert DbMsgCall: %w", err))
		}
	}

	// Insert DbMsgAddPackage messages with address IDs
	if len(groups.MsgAddPkg) > 0 {
		if err := d.dbPool.InsertMsgAddPackage(groups.MsgAddPkg); err != nil {
			insertErrors = append(insertErrors, fmt.Errorf("failed to insert DbMsgAddPackage: %w", err))
		}
	}

	// Insert DbMsgRun messages with address IDs
	if len(groups.MsgRun) > 0 {
		if err := d.dbPool.InsertMsgRun(groups.MsgRun); err != nil {
			insertErrors = append(insertErrors, fmt.Errorf("failed to insert DbMsgRun: %w", err))
		}
	}

	// Combine all errors if any occurred
	if len(insertErrors) > 0 {
		var errorMessages []string
		for _, err := range insertErrors {
			errorMessages = append(errorMessages, err.Error())
		}
		return fmt.Errorf("multiple insertion errors: %s", strings.Join(errorMessages, "; "))
	}

	return nil
}

func (d *DataProcessor) ProcessValidatorSignings(
	blocks []*rpcClient.BlockResponse,
	fromHeight uint64,
	toHeight uint64) {

	validatorChan := make(chan *sqlDataTypes.ValidatorBlockSigning, len(blocks))
	wg := sync.WaitGroup{}
	wg.Add(len(blocks))

	// Process blocks concurrently
	for _, block := range blocks {
		go func(block *rpcClient.BlockResponse) {
			defer wg.Done()

			signedVals := struct {
				Proposer   int32
				SignedVals []int32
			}{
				Proposer:   d.validatorCache.GetAddress(block.Result.Block.Header.ProposerAddress),
				SignedVals: make([]int32, 0),
			}
			precommits := block.Result.Block.LastCommit.Precommits
			for _, precommit := range precommits {
				if precommit != nil {
					signedVals.SignedVals = append(
						signedVals.SignedVals, d.validatorCache.GetAddress(precommit.ValidatorAddress),
					)
				}
			}

			height, err := strconv.ParseUint(block.Result.Block.Header.Height, 10, 64)
			if err != nil {
				log.Printf("Failed to parse block height %s: %v", block.Result.Block.Header.Height, err)
				return
			}

			validatorChan <- &sqlDataTypes.ValidatorBlockSigning{
				BlockHeight: height,
				Timestamp:   block.Result.Block.Header.Time,
				Proposer:    signedVals.Proposer,
				SignedVals:  signedVals.SignedVals,
				ChainName:   d.chainName,
			}
		}(block)
	}

	// Close channel when all goroutines finish
	go func() {
		wg.Wait()
		close(validatorChan)
	}()

	validatorData := make([]sqlDataTypes.ValidatorBlockSigning, 0, len(blocks))
	for validator := range validatorChan {
		validatorData = append(validatorData, *validator)
	}

	err := d.dbPool.InsertValidatorBlockSignings(validatorData)
	if err != nil {
		log.Printf("Failed to insert validator block signings: %v", err)
	}
	log.Printf("Validator block signings processed from %d to %d", fromHeight, toHeight)
}

// createAddressTx is a private func that creates a slice of sqlDataTypes.AddressTx from a
// decoder.DbMessageGroups using concurrent workers with mutex-based aggregation
// it should be used to create the data for the address_tx table
func createAddressTx(msg *decoder.DbMessageGroups) []sqlDataTypes.AddressTx {
	txAmount := len(msg.MsgSend) + len(msg.MsgCall) + len(msg.MsgAddPkg) + len(msg.MsgRun)
	if txAmount == 0 {
		return []sqlDataTypes.AddressTx{}
	}

	var addresses []sqlDataTypes.AddressTx
	var mu sync.Mutex
	wg := sync.WaitGroup{}

	// Process MsgSend messages
	wg.Add(len(msg.MsgSend))
	for _, msgItem := range msg.MsgSend {
		go func(msgItem sqlDataTypes.MsgSend) {
			defer wg.Done()
			txAddresses := msgItem.GetAllAddresses()
			timestamp := msgItem.Timestamp
			msgTypes := []string{msgItem.TableName()}

			// Collect addresses for this message
			msgAddressList := make([]sqlDataTypes.AddressTx, 0, len(txAddresses.GetAddressList()))
			for _, addr := range txAddresses.GetAddressList() {
				msgAddressList = append(msgAddressList, sqlDataTypes.AddressTx{
					Address:   addr,
					TxHash:    txAddresses.TxHash,
					ChainName: msgItem.ChainName,
					Timestamp: timestamp,
					MsgTypes:  msgTypes,
				})
			}

			// Thread-safe aggregation
			mu.Lock()
			addresses = append(addresses, msgAddressList...)
			mu.Unlock()
		}(msgItem)
	}

	// Process MsgCall messages
	wg.Add(len(msg.MsgCall))
	for _, msgItem := range msg.MsgCall {
		go func(msgItem sqlDataTypes.MsgCall) {
			defer wg.Done()
			txAddresses := msgItem.GetAllAddresses()
			timestamp := msgItem.Timestamp
			msgTypes := []string{msgItem.TableName()}

			msgAddressList := make([]sqlDataTypes.AddressTx, 0, len(txAddresses.GetAddressList()))
			for _, addr := range txAddresses.GetAddressList() {
				msgAddressList = append(msgAddressList, sqlDataTypes.AddressTx{
					Address:   addr,
					TxHash:    txAddresses.TxHash,
					ChainName: msgItem.ChainName,
					Timestamp: timestamp,
					MsgTypes:  msgTypes,
				})
			}

			mu.Lock()
			addresses = append(addresses, msgAddressList...)
			mu.Unlock()
		}(msgItem)
	}

	// Process MsgAddPkg messages
	wg.Add(len(msg.MsgAddPkg))
	for _, msgItem := range msg.MsgAddPkg {
		go func(msgItem sqlDataTypes.MsgAddPackage) {
			defer wg.Done()
			txAddresses := msgItem.GetAllAddresses()
			timestamp := msgItem.Timestamp
			msgTypes := []string{msgItem.TableName()}

			msgAddressList := make([]sqlDataTypes.AddressTx, 0, len(txAddresses.GetAddressList()))
			for _, addr := range txAddresses.GetAddressList() {
				msgAddressList = append(msgAddressList, sqlDataTypes.AddressTx{
					Address:   addr,
					TxHash:    txAddresses.TxHash,
					ChainName: msgItem.ChainName,
					Timestamp: timestamp,
					MsgTypes:  msgTypes,
				})
			}

			mu.Lock()
			addresses = append(addresses, msgAddressList...)
			mu.Unlock()
		}(msgItem)
	}

	// Process MsgRun messages
	wg.Add(len(msg.MsgRun))
	for _, msgItem := range msg.MsgRun {
		go func(msgItem sqlDataTypes.MsgRun) {
			defer wg.Done()
			txAddresses := msgItem.GetAllAddresses()
			timestamp := msgItem.Timestamp
			msgTypes := []string{msgItem.TableName()}

			msgAddressList := make([]sqlDataTypes.AddressTx, 0, len(txAddresses.GetAddressList()))
			for _, addr := range txAddresses.GetAddressList() {
				msgAddressList = append(msgAddressList, sqlDataTypes.AddressTx{
					Address:   addr,
					TxHash:    txAddresses.TxHash,
					ChainName: msgItem.ChainName,
					Timestamp: timestamp,
					MsgTypes:  msgTypes,
				})
			}

			mu.Lock()
			addresses = append(addresses, msgAddressList...)
			mu.Unlock()
		}(msgItem)
	}

	wg.Wait()
	return addresses
}
