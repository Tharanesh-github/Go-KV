package raft

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

// State defines the 3 possible roles a node can have in the Raft cluster
type State int

const (
	Follower State = iota
	Candidate
	Leader
)

// String helper to print the state nicely in our terminal
func (s State) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}

// Node represents the brain of our distributed consensus engine.
type Node struct {
	mu sync.Mutex // Mutex to prevent race conditions when multiple threads read/write state

	ID    string   // This node's unique address (e.g., "localhost:8080")
	Peers []string // The addresses of the other nodes in the cluster

	State       State // Follower, Candidate, or Leader
	CurrentTerm int   // The current election cycle version (starts at 0)
	VotedFor    string // Who this node voted for in the current term

	// Last time we heard from the leader. Used to trigger elections!
	LastHeartbeat time.Time 
}

// NewNode creates a fresh, brand new Follower node
func NewNode(id string, peers []string) *Node {
	return &Node{
		ID:            id,
		Peers:         peers,
		State:         Follower,
		CurrentTerm:   0,
		LastHeartbeat: time.Now(),
	}
}

// RunElectionTimer is a background thread (goroutine) that constantly checks
// if the leader has died. If the leader dies, it triggers an election.
func (rn *Node) RunElectionTimer() {
	// Raft requires a RANDOMIZED timeout (e.g., between 150ms and 300ms)
	// This prevents all followers from starting an election at the exact same millisecond!
	timeoutDuration := time.Duration(150+rand.Intn(150)) * time.Millisecond

	for {
		time.Sleep(10 * time.Millisecond) // Check every 10ms

		rn.mu.Lock()
		// If we haven't heard a heartbeat in a while, and we aren't already the leader...
		if rn.State != Leader && time.Since(rn.LastHeartbeat) >= timeoutDuration {
			log.Printf("Node %s: Election timeout reached! No leader detected.", rn.ID)
			rn.startElection()
			
			// Reset the randomized timeout for the next potential election
			timeoutDuration = time.Duration(150+rand.Intn(150)) * time.Millisecond
		}
		rn.mu.Unlock()
	}
}

// startElection promotes the node to Candidate and asks others for votes
func (rn *Node) startElection() {
	rn.State = Candidate
	rn.CurrentTerm++      // Advance to the next election cycle
	rn.VotedFor = rn.ID   // Vote for ourselves
	rn.LastHeartbeat = time.Now() // Reset timer

	log.Printf("Node %s: Starting election for Term %d. Becoming %s.", rn.ID, rn.CurrentTerm, rn.State)

	// NOTE: In the next phase, we will add the actual HTTP network calls here
	// to literally send "Vote For Me!" requests to the other nodes in the rn.Peers list.
	
	// Simulated automatic win for now if it has no peers (to prevent crashing)
	if len(rn.Peers) == 0 {
		log.Printf("Node %s: I have no peers. I win the election by default!", rn.ID)
		rn.becomeLeader()
	}
}

// becomeLeader promotes the node to Leader and starts broadcasting heartbeats
func (rn *Node) becomeLeader() {
	rn.State = Leader
	log.Printf("Node %s: 👑 I AM THE NEW LEADER for Term %d! 👑", rn.ID, rn.CurrentTerm)

	// Start a background thread to send heartbeats so followers don't revolt
	go rn.broadcastHeartbeats()
}

// broadcastHeartbeats sends continuous pings to followers
func (rn *Node) broadcastHeartbeats() {
	for {
		rn.mu.Lock()
		if rn.State != Leader {
			rn.mu.Unlock()
			return // If I am no longer the leader, stop sending heartbeats!
		}
		rn.mu.Unlock()

		// log.Printf("Node %s: Thump-thump... sending heartbeat to peers...", rn.ID)
		
		// NOTE: In the next phase, we will add the HTTP code here to actually
		// send the heartbeat payload to the follower nodes.

		time.Sleep(50 * time.Millisecond) // Heartbeats are sent every 50ms
	}
}