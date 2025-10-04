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
// it will process the blocks using async workers and store them in a channel
// it will then extract the blocks from the channel and insert them into the database
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
	blockChan := make(chan *sqlDataTypes.Blocks, len(blocks))
	wg := sync.WaitGroup{}
	wg.Add(len(blocks))

	for _, block := range blocks {
		go func(block *rpcClient.BlockResponse) {
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

			blockChan <- &sqlDataTypes.Blocks{
				Hash:      hash,
				Height:    height,
				Timestamp: block.Result.Block.Header.Time,
				ChainID:   block.Result.Block.Header.ChainID,
				Txs:       txs,
				ChainName: d.chainName,
			}
		}(block)
	}

	go func() {
		wg.Wait()
		close(blockChan)
	}()

	blocksData := make([]sqlDataTypes.Blocks, 0, len(blocks))
	for block := range blockChan {
		blocksData = append(blocksData, *block)
	}

	err := d.dbPool.InsertBlocks(blocksData)
	if err != nil {
		log.Printf("Failed to insert blocks: %v", err)
	}
	log.Printf("Blocks processed from %d to %d", fromHeight, toHeight)
}

// ProcessTransactions is a swarm method to process the transactions from a map of transactions and timestamps
// it will process the transactions using async workers and store them in a channel
// it will then extract the transactions from the channel and insert them into the database
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

	transactionChan := make(chan *sqlDataTypes.TransactionGeneral, len(transactions))
	wg := sync.WaitGroup{}
	wg.Add(len(transactions))

	for _, transaction := range transactions {
		go func(
			transaction TrasnactionsData,
			compressEvents bool) {
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

			// convert the gas wanetd and used from string to uint64
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

			// here the text event will return nil depending on the compressEvents
			transactionChan <- &sqlDataTypes.TransactionGeneral{
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

		}(transaction, compressEvents)
	}

	go func() {
		wg.Wait()
		close(transactionChan)
	}()

	transactionsData := make([]sqlDataTypes.TransactionGeneral, 0, len(transactions))
	for transaction := range transactionChan {
		transactionsData = append(transactionsData, *transaction)
	}

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

	// Phase 2: Concurrent message processing using channels
	type processedResult struct {
		dbGroups *decoder.DbMessageGroups
		err      error
	}

	resultChan := make(chan processedResult, len(transactions))
	wg2 := sync.WaitGroup{}
	wg2.Add(len(transactions))

	// Launch goroutines to process messages concurrently
	txIndex := 0
	for _, transaction := range transactions {
		if txIndex >= len(allDecodedMsgs) {
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
				resultChan <- processedResult{nil, nil}
				return
			}

			// Convert directly to database-ready messages with address IDs
			txHash, err := base64.StdEncoding.DecodeString(transaction.Response.GetHash())
			if err != nil {
				log.Printf("Failed to decode tx hash %s: %v", transaction.Response.GetHash(), err)
				resultChan <- processedResult{nil, err}
				return
			}

			dbMessageGroups, err := decodedMsg.ConvertToDbMessages(d.addressCache, txHash, d.chainName, transaction.Timestamp, decodedMsg.GetSigners())
			if err != nil {
				log.Printf("Failed to convert messages for tx %s: %v", transaction.Response.GetHash(), err)
				resultChan <- processedResult{nil, err}
				return
			}
			resultChan <- processedResult{dbMessageGroups, nil}

		}(transaction, allDecodedMsgs[txIndex])
		txIndex++
	}

	// Close result channel when all processing goroutines finish
	go func() {
		wg2.Wait()
		close(resultChan)
	}()

	// Aggregate results from all goroutines
	aggregatedDbGroups := &decoder.DbMessageGroups{
		MsgSend:   make([]sqlDataTypes.MsgSend, 0),
		MsgCall:   make([]sqlDataTypes.MsgCall, 0),
		MsgAddPkg: make([]sqlDataTypes.MsgAddPackage, 0),
		MsgRun:    make([]sqlDataTypes.MsgRun, 0),
	}

	for result := range resultChan {
		if result.err != nil {
			continue // Skip failed transactions
		}
		if result.dbGroups == nil {
			continue // Skip nil results
		}

		// Thread-safe aggregation (single-threaded collection)
		aggregatedDbGroups.MsgSend = append(aggregatedDbGroups.MsgSend, result.dbGroups.MsgSend...)
		aggregatedDbGroups.MsgCall = append(aggregatedDbGroups.MsgCall, result.dbGroups.MsgCall...)
		aggregatedDbGroups.MsgAddPkg = append(aggregatedDbGroups.MsgAddPkg, result.dbGroups.MsgAddPkg...)
		aggregatedDbGroups.MsgRun = append(aggregatedDbGroups.MsgRun, result.dbGroups.MsgRun...)
	}

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
// decoder.DbMessageGroups
// it should be used to create the data for the address_tx table
func createAddressTx(msg *decoder.DbMessageGroups) []sqlDataTypes.AddressTx {
	txAmount := len(msg.MsgSend) + len(msg.MsgCall) + len(msg.MsgAddPkg) + len(msg.MsgRun)
	if txAmount == 0 {
		return []sqlDataTypes.AddressTx{}
	}

	chanAddr := make(chan sqlDataTypes.AddressTx)
	wg := sync.WaitGroup{}
	wg.Add(txAmount)

	// Start a goroutine to collect all addresses
	addresses := make([]sqlDataTypes.AddressTx, 0)
	done := make(chan struct{})
	go func() {
		for address := range chanAddr {
			addresses = append(addresses, address)
		}
		done <- struct{}{}
	}()

	// Process MsgSend messages
	for _, msgItem := range msg.MsgSend {
		go func(msgItem sqlDataTypes.MsgSend) {
			defer wg.Done()
			txAddresses := msgItem.GetAllAddresses()
			timestamp := msgItem.Timestamp
			msgTypes := []string{msgItem.TableName()}

			for _, addr := range txAddresses.GetAddressList() {
				address := sqlDataTypes.AddressTx{
					Address:   addr,
					TxHash:    txAddresses.TxHash,
					ChainName: msgItem.ChainName,
					Timestamp: timestamp,
					MsgTypes:  msgTypes,
				}
				chanAddr <- address
			}
		}(msgItem)
	}

	// Process MsgCall messages
	for _, msgItem := range msg.MsgCall {
		go func(msgItem sqlDataTypes.MsgCall) {
			defer wg.Done()
			txAddresses := msgItem.GetAllAddresses()
			timestamp := msgItem.Timestamp
			msgTypes := []string{msgItem.TableName()}

			for _, addr := range txAddresses.GetAddressList() {
				address := sqlDataTypes.AddressTx{
					Address:   addr,
					TxHash:    txAddresses.TxHash,
					ChainName: msgItem.ChainName,
					Timestamp: timestamp,
					MsgTypes:  msgTypes,
				}
				chanAddr <- address
			}
		}(msgItem)
	}

	// Process MsgAddPkg messages
	for _, msgItem := range msg.MsgAddPkg {
		go func(msgItem sqlDataTypes.MsgAddPackage) {
			defer wg.Done()
			txAddresses := msgItem.GetAllAddresses()
			timestamp := msgItem.Timestamp
			msgTypes := []string{msgItem.TableName()}

			for _, addr := range txAddresses.GetAddressList() {
				address := sqlDataTypes.AddressTx{
					Address:   addr,
					TxHash:    txAddresses.TxHash,
					ChainName: msgItem.ChainName,
					Timestamp: timestamp,
					MsgTypes:  msgTypes,
				}
				chanAddr <- address
			}
		}(msgItem)
	}

	// Process MsgRun messages
	for _, msgItem := range msg.MsgRun {
		go func(msgItem sqlDataTypes.MsgRun) {
			defer wg.Done()
			txAddresses := msgItem.GetAllAddresses()
			timestamp := msgItem.Timestamp
			msgTypes := []string{msgItem.TableName()}

			for _, addr := range txAddresses.GetAddressList() {
				address := sqlDataTypes.AddressTx{
					Address:   addr,
					TxHash:    txAddresses.TxHash,
					ChainName: msgItem.ChainName,
					Timestamp: timestamp,
					MsgTypes:  msgTypes,
				}
				chanAddr <- address
			}
		}(msgItem)
	}

	// Wait for all workers to finish, then close channel
	go func() {
		wg.Wait()
		close(chanAddr)
	}()

	// Wait for the collector goroutine to finish
	<-done

	return addresses
}
