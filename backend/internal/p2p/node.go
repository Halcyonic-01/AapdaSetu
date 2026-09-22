package p2p

import (
	"context"
	"fmt"
	"log"

	"github.com/Halcyonic-01/AapdaSetu/backend/config"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/peers"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
)

// Node wraps the libp2p host, mDNS discovery, and PubSub systems.
type Node struct {
	cfg       *config.Config
	host      host.Host
	mdns      mdns.Service
	pubsub    *PubSubManager
	peerStore *peers.Store
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewNode instantiates an AapdaSetu P2P Mesh Node.
func NewNode(parentCtx context.Context, cfg *config.Config, peerStore *peers.Store) (*Node, error) {
	ctx, cancel := context.WithCancel(parentCtx)

	// 1. Initialize libp2p Host
	h, err := CreateHost(ctx, cfg.P2PPort, peerStore)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create host on port %d: %w", cfg.P2PPort, err)
	}

	// 2. Initialize GossipSub router
	ps, err := NewPubSubManager(ctx, h)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to init pubsub: %w", err)
	}

	// 3. Initialize mDNS Discovery
	discovery, err := SetupDiscovery(ctx, h, cfg.Rendezvous, peerStore)
	if err != nil {
		ps.Close()
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to init mDNS discovery: %w", err)
	}

	node := &Node{
		cfg:       cfg,
		host:      h,
		mdns:      discovery,
		pubsub:    ps,
		peerStore: peerStore,
		ctx:       ctx,
		cancel:    cancel,
	}

	log.Printf("[P2P Mesh] Node initialized with Peer ID: %s", h.ID())
	for _, addr := range h.Addrs() {
		log.Printf("[P2P Mesh] Listening on: %s/p2p/%s", addr, h.ID())
	}

	return node, nil
}

// Host returns the underlying libp2p host.
func (n *Node) Host() host.Host {
	return n.host
}

// ID returns the peer ID of this node.
func (n *Node) ID() peer.ID {
	return n.host.ID()
}

// Addrs returns all multiaddresses this node is listening on.
func (n *Node) Addrs() []string {
	addrs := make([]string, 0, len(n.host.Addrs()))
	for _, a := range n.host.Addrs() {
		addrs = append(addrs, fmt.Sprintf("%s/p2p/%s", a, n.host.ID()))
	}
	return addrs
}

// PubSub returns the PubSub manager.
func (n *Node) PubSub() *PubSubManager {
	return n.pubsub
}

// PeerStore returns the peer store registry.
func (n *Node) PeerStore() *peers.Store {
	return n.peerStore
}

// ConnectDirect dials a peer directly by multiaddr string (e.g. for testing or ad-hoc links).
func (n *Node) ConnectDirect(ctx context.Context, peerAddr string) error {
	maddr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		return fmt.Errorf("invalid multiaddr: %w", err)
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return fmt.Errorf("failed to extract peer info: %w", err)
	}

	if err := n.host.Connect(ctx, *info); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	if n.peerStore != nil {
		n.peerStore.AddOrUpdate(info.ID, info.Addrs, true)
	}
	return nil
}

// Close gracefully terminates the node and all network services.
func (n *Node) Close() error {
	log.Println("[P2P Mesh] Stopping P2P node services...")
	n.cancel()

	if n.mdns != nil {
		if err := n.mdns.Close(); err != nil {
			log.Printf("[P2P Mesh] Error closing mDNS: %v", err)
		}
	}

	if n.pubsub != nil {
		n.pubsub.Close()
	}

	if n.host != nil {
		if err := n.host.Close(); err != nil {
			return fmt.Errorf("error closing host: %w", err)
		}
	}

	log.Println("[P2P Mesh] Node stopped successfully.")
	return nil
}
