package p2p

import (
	"context"
	"log"
	"time"

	"github.com/Halcyonic-01/AapdaSetu/backend/internal/peers"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// discoveryNotifee handles mDNS peer discovery events.
type discoveryNotifee struct {
	ctx       context.Context
	h         host.Host
	peerStore *peers.Store
}

// HandlePeerFound is triggered when mDNS detects a new peer on the local subnet.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	// Skip self-discovery
	if pi.ID == n.h.ID() {
		return
	}

	log.Printf("[mDNS Discovery] Discovered nearby peer: %s (%d addresses)", pi.ID, len(pi.Addrs))

	// Register peer's addresses in the libp2p host's internal peerstore
	n.h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.ConnectedAddrTTL)

	// Add to our application-level peer store
	if n.peerStore != nil {
		n.peerStore.AddOrUpdate(pi.ID, pi.Addrs, false)
	}

	// Attempt connection in background
	go func() {
		connCtx, cancel := context.WithTimeout(n.ctx, 10*time.Second)
		defer cancel()

		if err := n.h.Connect(connCtx, pi); err != nil {
			log.Printf("[mDNS Discovery] Connection attempt to %s failed: %v", pi.ID, err)
			return
		}

		log.Printf("[mDNS Discovery] Successfully established connection with peer: %s", pi.ID)
		if n.peerStore != nil {
			n.peerStore.SetConnected(pi.ID, true)
		}
	}()
}

// SetupDiscovery initializes the mDNS peer discovery service on the local network.
func SetupDiscovery(ctx context.Context, h host.Host, rendezvous string, peerStore *peers.Store) (mdns.Service, error) {
	notifee := &discoveryNotifee{
		ctx:       ctx,
		h:         h,
		peerStore: peerStore,
	}

	service := mdns.NewMdnsService(h, rendezvous, notifee)
	return service, nil
}
