package peers

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// PeerInfo holds connection and identification metadata for a mesh peer.
type PeerInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name,omitempty"`
	Addresses []string  `json:"addresses"`
	Connected bool      `json:"connected"`
	LastSeen  time.Time `json:"last_seen"`
}

// Store maintains a thread-safe registry of discovered and connected mesh peers.
type Store struct {
	mu    sync.RWMutex
	peers map[peer.ID]*PeerInfo
}

// NewStore initializes a new Peer Store.
func NewStore() *Store {
	return &Store{
		peers: make(map[peer.ID]*PeerInfo),
	}
}

// AddOrUpdate inserts or refreshes a peer's address list and connection status.
func (s *Store) AddOrUpdate(id peer.ID, addrs []multiaddr.Multiaddr, connected bool) *PeerInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	addrStrings := make([]string, 0, len(addrs))
	for _, a := range addrs {
		addrStrings = append(addrStrings, a.String())
	}

	p, exists := s.peers[id]
	if !exists {
		p = &PeerInfo{
			ID:        id.String(),
			Addresses: addrStrings,
			Connected: connected,
			LastSeen:  time.Now().UTC(),
		}
		s.peers[id] = p
		return p
	}

	// Update existing record
	if len(addrStrings) > 0 {
		p.Addresses = addrStrings
	}
	p.Connected = connected
	p.LastSeen = time.Now().UTC()
	return p
}

// SetConnected updates the connection state of a specific peer.
func (s *Store) SetConnected(id peer.ID, connected bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p, exists := s.peers[id]; exists {
		p.Connected = connected
		p.LastSeen = time.Now().UTC()
	} else if connected {
		s.peers[id] = &PeerInfo{
			ID:        id.String(),
			Addresses: []string{},
			Connected: true,
			LastSeen:  time.Now().UTC(),
		}
	}
}

// SetName assigns a human-friendly display name to a peer.
func (s *Store) SetName(id peer.ID, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p, exists := s.peers[id]; exists {
		p.Name = name
	}
}

// Get retrieves metadata for a specific peer ID.
func (s *Store) Get(id peer.ID) (*PeerInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, exists := s.peers[id]
	if !exists {
		return nil, false
	}
	// Return copy to prevent race conditions
	copyInfo := *p
	return &copyInfo, true
}

// List returns a slice of all known peers.
func (s *Store) List() []*PeerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*PeerInfo, 0, len(s.peers))
	for _, p := range s.peers {
		copyInfo := *p
		list = append(list, &copyInfo)
	}
	return list
}

// ConnectedPeers returns all peers currently marked connected.
func (s *Store) ConnectedPeers() []*PeerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*PeerInfo, 0)
	for _, p := range s.peers {
		if p.Connected {
			copyInfo := *p
			list = append(list, &copyInfo)
		}
	}
	return list
}

// Count returns the number of currently connected peers.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, p := range s.peers {
		if p.Connected {
			count++
		}
	}
	return count
}
