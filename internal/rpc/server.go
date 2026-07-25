package rpc

import (
	"fmt"
	"net/http"

	"github.com/tharanesh/go-kv/internal/raft"
	"github.com/tharanesh/go-kv/internal/storage"
)

// Server wraps our storage engine and raft node so we can expose them over HTTP
type Server struct {
	store *storage.KVStore
	raft  *raft.Node
}

// handleSet grabs the key and value from the URL and saves it to the database
func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
	// RULE 1: Only the Leader can accept writes!
	if s.raft.State != raft.Leader {
		http.Error(w, `{"error": "Access Denied. I am a Follower. Please talk to the Leader!"}`+"\n", http.StatusForbidden)
		return
	}

	key := r.URL.Query().Get("key")
	val := r.URL.Query().Get("val")

	if key == "" || val == "" {
		http.Error(w, `{"error": "Missing key or val parameter"}`+"\n", http.StatusBadRequest)
		return
	}

	// 1. The Leader saves the data to its own vault
	s.store.Set(key, val)
	
	// 2. The Leader commands all Followers to clone this data
	s.raft.Replicate(key, val)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "success", "message": "Data replicated across cluster", "key": "%s", "value": "%s"}`+"\n", key, val)
}

// handleGet fetches data from the database based on the URL parameter
// NOTE: In a real cluster, you can usually read (GET) from any node!
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if key == "" {
		http.Error(w, `{"error": "Missing key parameter"}`+"\n", http.StatusBadRequest)
		return
	}

	val, exists := s.store.Get(key)
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error": "Key not found"}`+"\n")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "success", "key": "%s", "value": "%s"}`+"\n", key, val)
}

// handleSync is the internal ear for Followers to receive data clones from the Leader
func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	val := r.URL.Query().Get("val")

	s.store.Set(key, val) // The Follower obediently saves the data
	w.Write([]byte("ok"))
}

// handleVote lets this node receive vote requests from other candidates
func (s *Server) handleVote(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")
	candidate := r.URL.Query().Get("candidate")

	granted := s.raft.RequestVote(term, candidate)
	if granted {
		w.Write([]byte("yes"))
	} else {
		w.Write([]byte("no"))
	}
}

// handleHeartbeat receives "I am alive" messages from the current leader
func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")
	leader := r.URL.Query().Get("leader")

	s.raft.AppendEntries(term, leader)
	w.Write([]byte("ok"))
}

// Start boots up the HTTP server on a specific port
func Start(port string, store *storage.KVStore, raftNode *raft.Node) error {
	srv := &Server{store: store, raft: raftNode}

	// Client Endpoints
	http.HandleFunc("/set", srv.handleSet)
	http.HandleFunc("/get", srv.handleGet)
	
	// Raft Internal Network Endpoints
	http.HandleFunc("/raft/vote", srv.handleVote)
	http.HandleFunc("/raft/heartbeat", srv.handleHeartbeat)
	http.HandleFunc("/raft/sync", srv.handleSync) // NEW: Data replication endpoint
	
	return http.ListenAndServe(":"+port, nil)
}
