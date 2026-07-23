package rpc

import (
	"fmt"
	"net/http"

	"github.com/tharanesh/go-kv/internal/storage"
)

// Server wraps our storage engine so we can expose it over HTTP
type Server struct {
	store *storage.KVStore
}

// handleSet grabs the key and value from the URL and saves it to the database
func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	val := r.URL.Query().Get("val")

	if key == "" || val == "" {
		http.Error(w, `{"error": "Missing key or val parameter"}`, http.StatusBadRequest)
		return
	}

	s.store.Set(key, val)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "success", "key": "%s", "value": "%s"}`+"\n", key, val)
}

// handleGet fetches data from the database based on the URL parameter
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if key == "" {
		http.Error(w, `{"error": "Missing key parameter"}`, http.StatusBadRequest)
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

// Start boots up the HTTP server on a specific port
func Start(port string, store *storage.KVStore) error {
	srv := &Server{store: store}

	// Route incoming URLs to our specific functions
	http.HandleFunc("/set", srv.handleSet)
	http.HandleFunc("/get", srv.handleGet)

	log.Printf("Node API listening on http://localhost:%s", port)
	
	// Start listening for network traffic
	return http.ListenAndServe(":"+port, nil)
}