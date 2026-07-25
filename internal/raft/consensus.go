package raft

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type State int

const (
	Follower State = iota
	Candidate
	Leader
)

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

type Node struct {
	mu sync.Mutex 

	ID    string   
	Peers []string 

	State       State 
	CurrentTerm int   
	VotedFor    string 

	LastHeartbeat time.Time 
}

func NewNode(id string, peers []string) *Node {
	return &Node{
		ID:            id,
		Peers:         peers,
		State:         Follower,
		CurrentTerm:   0,
		LastHeartbeat: time.Now(),
	}
}

func (rn *Node) RunElectionTimer() {
	timeoutDuration := time.Duration(150+rand.Intn(150)) * time.Millisecond

	for {
		time.Sleep(10 * time.Millisecond) 

		rn.mu.Lock()
		if rn.State != Leader && time.Since(rn.LastHeartbeat) >= timeoutDuration {
			log.Printf("Node %s: Election timeout reached! No leader detected.", rn.ID)
			rn.startElection()
			timeoutDuration = time.Duration(150+rand.Intn(150)) * time.Millisecond
		}
		rn.mu.Unlock()
	}
}

func (rn *Node) startElection() {
	rn.State = Candidate
	rn.CurrentTerm++      
	rn.VotedFor = rn.ID   
	rn.LastHeartbeat = time.Now() 

	log.Printf("Node %s: Starting election for Term %d. Becoming %s.", rn.ID, rn.CurrentTerm, rn.State)

	if len(rn.Peers) == 0 {
		rn.becomeLeader()
		return
	}

	votes := 1 
	requiredVotes := (len(rn.Peers) / 2) + 1

	for _, peerPort := range rn.Peers {
		go func(port string) {
			url := fmt.Sprintf("http://localhost:%s/raft/vote?term=%d&candidate=%s", port, rn.CurrentTerm, rn.ID)
			resp, err := http.Get(url)
			if err != nil {
				return 
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			if string(body) == "yes" {
				rn.mu.Lock()
				if rn.State == Candidate {
					votes++
					if votes >= requiredVotes {
						rn.becomeLeader()
					}
				}
				rn.mu.Unlock()
			}
		}(peerPort)
	}
}

func (rn *Node) becomeLeader() {
	rn.State = Leader
	log.Printf("Node %s: 👑 I AM THE NEW LEADER for Term %d! 👑", rn.ID, rn.CurrentTerm)
	go rn.broadcastHeartbeats()
}

func (rn *Node) broadcastHeartbeats() {
	for {
		rn.mu.Lock()
		if rn.State != Leader {
			rn.mu.Unlock()
			return 
		}
		rn.mu.Unlock()

		for _, peerPort := range rn.Peers {
			go func(port string) {
				url := fmt.Sprintf("http://localhost:%s/raft/heartbeat?term=%d&leader=%s", port, rn.CurrentTerm, rn.ID)
				http.Get(url) 
			}(peerPort)
		}
		time.Sleep(50 * time.Millisecond) 
	}
}

// Replicate sends the newly saved data to all followers
func (rn *Node) Replicate(key, val string) {
	for _, peerPort := range rn.Peers {
		go func(port string) {
			url := fmt.Sprintf("http://localhost:%s/raft/sync?key=%s&val=%s", port, key, val)
			http.Get(url) // Fire and forget the clone command
		}(peerPort)
	}
}

func (rn *Node) RequestVote(termStr, candidateID string) bool {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	term, _ := strconv.Atoi(termStr)
	
	if term > rn.CurrentTerm {
		rn.CurrentTerm = term
		rn.State = Follower
		rn.VotedFor = candidateID
		rn.LastHeartbeat = time.Now()
		log.Printf("Node %s: 🗳️ Voted for Candidate %s in Term %d", rn.ID, candidateID, term)
		return true
	}
	return false
}

func (rn *Node) AppendEntries(termStr, leaderID string) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	term, _ := strconv.Atoi(termStr)
	
	if term >= rn.CurrentTerm {
		if rn.State == Leader && rn.ID != leaderID {
			log.Printf("Node %s: 😲 Discovered higher-term Leader %s. Stepping down!", rn.ID, leaderID)
		}
		rn.CurrentTerm = term
		rn.State = Follower
		rn.LastHeartbeat = time.Now()
	}
}
