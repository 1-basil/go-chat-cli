package main

import (
	"fmt"
	"sync"
	"time"
)

// Peer represents a discovered peer on the LAN.
type Peer struct {
	Username string
	IP       string
	Port     string
	LastSeen time.Time
}

// PeerTable is a thread-safe in-memory registry of discovered peers.
type PeerTable struct {
	mu    sync.RWMutex
	peers map[string]*Peer
}

// NewPeerTable creates an empty peer table.
func NewPeerTable() *PeerTable {
	return &PeerTable{
		peers: make(map[string]*Peer),
	}
}

// Set adds or updates a peer entry.
func (pt *PeerTable) Set(username, ip, port string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.peers[username] = &Peer{
		Username: username,
		IP:       ip,
		Port:     port,
		LastSeen: time.Now(),
	}
}

// Get retrieves a peer by username. Returns nil if not found.
func (pt *PeerTable) Get(username string) *Peer {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	p, ok := pt.peers[username]
	if !ok {
		return nil
	}
	return p
}

// All returns a snapshot of all peers.
func (pt *PeerTable) All() []*Peer {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	result := make([]*Peer, 0, len(pt.peers))
	for _, p := range pt.peers {
		result = append(result, p)
	}
	return result
}

// Cleanup removes peers not seen for the given duration.
func (pt *PeerTable) Cleanup(maxAge time.Duration) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	now := time.Now()
	for k, p := range pt.peers {
		if now.Sub(p.LastSeen) > maxAge {
			delete(pt.peers, k)
		}
	}
}

// PrintAll prints the peer table to stdout.
func (pt *PeerTable) PrintAll() {
	peers := pt.All()
	if len(peers) == 0 {
		fmt.Println("No online users found.")
		return
	}
	fmt.Println("Online users")
	fmt.Println("------------")
	for _, p := range peers {
		fmt.Printf("%-15s %s:%s\n", p.Username, p.IP, p.Port)
	}
}
