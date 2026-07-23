package main

import (
	"fmt"
	"log"

	"github.com/tharanesh/go-kv/internal/rpc"
	"github.com/tharanesh/go-kv/internal/storage"
)

func main() {
	log.Println("Starting GO-KV Node...")

	// 1. Boot up the Storage Engine
	log.Println("Initializing thread-safe storage engine...")
	db := storage.NewKVStore()

	// 2. Start the RPC/HTTP Server on port 8080
	port := "8080"
	log.Printf("Starting network listener on port %s...", port)
	
	err := rpc.Start(port, db)
	if err != nil {
		log.Fatalf("Server crashed: %s", err)
	}
}