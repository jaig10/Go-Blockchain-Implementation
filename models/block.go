package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// Enum for BlockStatus
type BlockStatus int

const (
	Committed BlockStatus = iota
	Pending
)

func (bs BlockStatus) String() string {
	return [...]string{"Committed", "Pending"}[bs]
}

// Transaction Struct
type Transaction struct {
	ID      string                 `json:"id"`
	TxData  map[string]interface{} `json:"txData"`
	Valid   bool                   `json:"valid"`
	Version float64                `json:"version"`
}

// Block Struct
type Block struct {
	BlockNumber   int           `json:"blockNumber"`
	Transactions  []Transaction `json:"txns"`
	Timestamp     time.Time     `json:"timestamp"`
	Status        BlockStatus   `json:"blockStatus"`
	PrevBlockHash string        `json:"prevBlockHash"`
	BlockHash     string        `json:"blockHash"`
	BlockSize     int           // Block size limit
}

// Utility functions for hashing and txn ID generation
func GenerateTxnID(txn map[string]interface{}) string {
	txData, _ := json.Marshal(txn)
	hasher := sha256.New()
	hasher.Write(txData)
	return hex.EncodeToString(hasher.Sum(nil))
}

func CalculateBlockHash(block *Block) string {
	hasher := sha256.New()
	blockData, _ := json.Marshal(block)
	hasher.Write(blockData)
	return hex.EncodeToString(hasher.Sum(nil))
}
