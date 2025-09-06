package dataprocessor

import (
	"encoding/base64"
	"log"
	"strconv"
	"sync"
	"time"

	addressCache "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/address_cache"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/decoder"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	sqlDataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
)

func NewDataProcessor(
	db *database.TimescaleDb,
	addressCache *addressCache.AddressCache,
	validatorCache *addressCache.AddressCache,
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
				addressesMap.Store(precommit.ValidatorAddress, true)
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
	d.addressCache.AddressSolver(addresses, d.chainName, true, 3, nil)
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
				return
			}

			// convert from string to uint64
			height, err := strconv.ParseUint(block.Result.Block.Header.Height, 10, 64)
			if err != nil {
				return
			}

			// there should be an slice of strings but it can be nil
			// if slice exists we need to convert each slice from base64 to sha256
			// since it is shorter and better for the database
			var txs [][]byte
			if block.Result.Block.Data.Txs != nil {
				for _, tx := range *block.Result.Block.Data.Txs {
					txHash, err := base64.StdEncoding.DecodeString(tx)
					if err != nil {
						return
					}
					txs = append(txs, txHash)
				}
			}

			blockChan <- &sqlDataTypes.Blocks{
				Hash:            hash,
				Height:          height,
				Timestamp:       block.Result.Block.Header.Time,
				ChainID:         block.Result.Block.Header.ChainID,
				ProposerAddress: d.addressCache.GetAddress(block.Result.Block.Header.ProposerAddress),
				Txs:             txs,
				ChainName:       d.chainName,
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

	d.dbPool.InsertBlocks(blocksData)
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
// However this function is not finished yet, it only records the transaction general data
// but it needs to look into the decoded message and record the data from it
func (d *DataProcessor) ProcessTransactions(
	transactions map[*rpcClient.TxResponse]time.Time,
	compressEvents bool,
	fromHeight uint64,
	toHeight uint64) {

	transactionChan := make(chan *sqlDataTypes.TransactionGeneral, len(transactions))
	wg := sync.WaitGroup{}
	wg.Add(len(transactions))

	for transaction, timestamp := range transactions {
		go func(
			transaction *rpcClient.TxResponse,
			timestamp time.Time,
			compressEvents bool) {
			defer wg.Done()
			txResult := transaction.Result.TxResult

			decodedMsg := decoder.NewDecodedMsg(transaction.Result.Tx)

			fee := decodedMsg.GetFee()
			msgTypes := decodedMsg.GetMsgTypes()

			// convert the tx hash from base64 to sha256
			txHash, err := base64.StdEncoding.DecodeString(transaction.Result.Hash)
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
			events, err := EventSolver(transaction, compressEvents)
			if err != nil {
				return
			}

			// here the text event will return nil depending on the compressEvents
			transactionChan <- &sqlDataTypes.TransactionGeneral{
				TxHash:             txHash,
				ChainName:          d.chainName,
				Timestamp:          timestamp,
				MsgTypes:           msgTypes,
				TxEvents:           events.GetNativeEvents(),
				TxEventsCompressed: events.GetCompressedData(),
				GasUsed:            gasUsed,
				GasWanted:          gasWanted,
				Fee:                fee,
			}

		}(transaction, timestamp, compressEvents)
	}

	go func() {
		wg.Wait()
		close(transactionChan)
	}()

	transactionsData := make([]sqlDataTypes.TransactionGeneral, 0, len(transactions))
	for transaction := range transactionChan {
		transactionsData = append(transactionsData, *transaction)
	}

	d.dbPool.InsertTransactionsGeneral(transactionsData)
	log.Printf("Transactions processed from %d to %d", fromHeight, toHeight)
}
