package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"goblockcahinproject/models" // Update with the correct module path

	"github.com/syndtr/goleveldb/leveldb"
)

// Blockchain Struct
type Blockchain struct {
	LatestBlock *models.Block // Keep track of the latest block
	Blocks      []*models.Block
	db          *leveldb.DB // LevelDB instance for transactions
}

// Global block channel (outside Blockchain struct)
var BlockChannel = make(chan *models.Block)

// Initialize LevelDB with 1000 entries
func InitializeLevelDB() (*leveldb.DB, error) {
	db, err := leveldb.OpenFile("leveldb", nil)
	if err != nil {
		log.Fatalf("Failed to open LevelDB: %v", err)
	}

	// Populate with 1000 entries
	for i := 1; i <= 1000; i++ {
		key := fmt.Sprintf("SIM%d", i)
		value := map[string]interface{}{
			"val": i,
			"ver": 1.0,
		}
		valueBytes, _ := json.Marshal(value)
		db.Put([]byte(key), valueBytes, nil)
	}
	return db, nil
}

// Create a new block, using the previous block's hash
func (bc *Blockchain) CreateNewBlock() *models.Block {
	var prevHash string
	if bc.LatestBlock != nil {
		prevHash = bc.LatestBlock.BlockHash
	} else {
		prevHash = "0x000" // Genesis block prev hash
	}

	block := &models.Block{
		BlockNumber:   len(bc.Blocks) + 1,
		Timestamp:     time.Now(),
		Status:        models.Pending,
		BlockSize:     3, // Configurable Block Size
		PrevBlockHash: prevHash,
	}
	return block
}

// Add transaction to the blockchain and handle block size
func (bc *Blockchain) AddTransaction(txn map[string]interface{}) {
	// If no block exists or current block is full, create a new block
	if bc.LatestBlock == nil || len(bc.LatestBlock.Transactions) >= bc.LatestBlock.BlockSize {
		newBlock := bc.CreateNewBlock()
		bc.LatestBlock = newBlock
		bc.Blocks = append(bc.Blocks, newBlock)
	}

	// Add the transaction to the current block
	bc.LatestBlock.PushValidTransaction(txn, bc.db)

	// Check again if block is full after adding the transaction
	if len(bc.LatestBlock.Transactions) >= bc.LatestBlock.BlockSize {
		bc.LatestBlock.CommitBlock(bc.db) // Commit the block if full
		// Reset to indicate that a new block is needed
	}
}

// Push valid transaction to the block
func (b *models.Block) PushValidTransaction(txn map[string]interface{}, db *leveldb.DB) {
	for key, val := range txn {
		txMap := val.(map[string]interface{})

		tx := models.Transaction{
			ID:      models.GenerateTxnID(txn),
			TxData:  txMap,
			Version: txMap["ver"].(float64),
		}

		// Fetch from LevelDB and validate version
		dbValue, err := db.Get([]byte(key), nil)
		if err != nil {
			log.Printf("Error fetching key %s: %v", key, err)
			tx.Valid = false
		} else {
			var storedData map[string]interface{}
			json.Unmarshal(dbValue, &storedData)
			storedVersion := storedData["ver"].(float64)
			if storedVersion == txMap["ver"].(float64) {
				tx.Valid = true
			} else {
				tx.Valid = false
			}
		}

		b.Transactions = append(b.Transactions, tx)
	}
}

// Commit the block and update valid transactions in LevelDB
func (b *models.Block) CommitBlock(db *leveldb.DB) {
	startTime := time.Now() // Start timer
	b.UpdateBlockStatus(models.Committed)
	b.BlockHash = models.CalculateBlockHash(b)

	// Commit only valid transactions back to LevelDB
	for _, txn := range b.Transactions {
		if txn.Valid {
			valueBytes, _ := json.Marshal(txn.TxData)
			db.Put([]byte(txn.ID), valueBytes, nil)
		}
	}

	BlockChannel <- b // Send the block to the global channel for writing to file

	// Calculate processing time
	elapsed := time.Since(startTime)
	fmt.Printf("Block %d processing time: %s\n", b.BlockNumber, elapsed)
}

func (b *models.Block) UpdateBlockStatus(status models.BlockStatus) {
	b.Status = status
}

// File operations and block handling functions

func WriteBlockToFile(block *models.Block) {
	file, err := os.OpenFile("ledger.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	blockData, _ := json.Marshal(block)
	_, err = file.WriteString(string(blockData) + "\n")
	if err != nil {
		fmt.Println("Error writing block to file:", err)
	}
}

func ListenForBlocks() {
	for block := range BlockChannel {
		WriteBlockToFile(block)
		fmt.Printf("Block %d committed and written to file.\n", block.BlockNumber)
	}

}

func FetchBlockDetails(blockNumber int) *models.Block {
	file, err := os.Open("ledger.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	var block *models.Block
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var b models.Block
		err := decoder.Decode(&b)
		if err != nil {
			fmt.Println("Error decoding block:", err)
			return nil
		}
		if b.BlockNumber == blockNumber {
			block = &b
			break
		}
	}
	return block
}

func FetchAllBlocks() []*models.Block {
	file, err := os.Open("ledger.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	var blocks []*models.Block
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var b models.Block
		err := decoder.Decode(&b)
		if err != nil {
			fmt.Println("Error decoding block:", err)
			return nil
		}
		blocks = append(blocks, &b)
	}
	return blocks
}
