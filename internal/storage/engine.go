package storage

import "sync"

// KVStore represents our in-memory database.
// We MUST use a sync.RWMutex because in a web server, multiple clients
// might try to read and write data at the exact same millisecond.
type KVStore struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewKVStore creates a ready-to-use database engine.
func NewKVStore() *KVStore {
	return &KVStore{
		data: make(map[string]string),
	}
}

// Set safely adds or updates a key-value pair.
func (s *KVStore) Set(key, value string) {
	s.mu.Lock()         // Lock the door so no one else can write at the same time
	defer s.mu.Unlock() // Ensure we unlock the door when the function finishes

	s.data[key] = value
}

// Get safely retrieves a value by its key.
func (s *KVStore) Get(key string) (string, bool) {
	s.mu.RLock()         // Read-Lock allows multiple readers, but blocks writers
	defer s.mu.RUnlock()

	val, exists := s.data[key]
	return val, exists
}