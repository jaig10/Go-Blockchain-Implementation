package main

import (
	"log"
	"time"

	"goblockcahinproject/models"   // Update with the correct module path
	"goblockcahinproject/services" // Update with the correct module path
)

func main() {
	db, err := services.InitializeLevelDB()
	if err != nil {
		log.Fatalf("Failed to initialize LevelDB: %v", err)
	}
	defer db.Close()

	blockchain := &services.Blockchain{
		LatestBlock: nil,
		Blocks:      make([]*models.Block, 0),
		db:          db,
	}

	go services.ListenForBlocks()

	// Simulate adding transactions
	for i := 1; i <= 10; i++ {
		transaction := map[string]interface{}{
			"txn": map[string]interface{}{
				"val": i,
				"ver": float64(i),
			},
		}
		blockchain.AddTransaction(transaction)
		time.Sleep(1 * time.Second) // Simulate delay
	}
}
