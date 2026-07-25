package main

import (
	"flag"
	"log"
	"strings"

	"github.com/tharanesh/go-kv/internal/raft"
	"github.com/tharanesh/go-kv/internal/rpc"
	"github.com/tharanesh/go-kv/internal/storage"
)

func main() {
	// Parse command line arguments to allow multiple nodes on one computer
	port := flag.String("port", "8080", "The port this node will run on")
	peersFlag := flag.String("peers", "", "Comma-separated list of peer ports (e.g., 8081,8082)")
	flag.Parse()

	var peers []string
	if *peersFlag != "" {
		peers = strings.Split(*peersFlag, ",")
	}

	log.Printf("Starting GO-KV Node on port %s...", *port)

	// 1. Boot up the Storage Engine
	db := storage.NewKVStore()

	// 2. Boot up the Raft Consensus Brain
	raftNode := raft.NewNode(*port, peers)
	go raftNode.RunElectionTimer()

	// 3. Start the RPC/HTTP Server
	log.Printf("Starting network listener on port %s...", *port)
	err := rpc.Start(*port, db, raftNode)
	if err != nil {
		log.Fatalf("Server crashed: %s", err)
	}
}
