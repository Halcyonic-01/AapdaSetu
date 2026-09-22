package p2p

import (
	"context"
	"fmt"
	"log"

	"github.com/Halcyonic-01/AapdaSetu/backend/internal/peers"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
)

// CreateHost instantiates a libp2p Host listening on the specified TCP port.
func CreateHost(ctx context.Context, listenPort int, peerStore *peers.Store) (host.Host, error) {
	listenAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort))
	if err != nil {
		return nil, fmt.Errorf("invalid listen multiaddr: %w", err)
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrs(listenAddr),
		libp2p.NATPortMap(), // Attempt UPnP / NAT-PMP port mapping if available
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// Register connection event listeners to track peer lifecycle
	if peerStore != nil {
		h.Network().Notify(&network.NotifyBundle{
			ConnectedF: func(n network.Network, c network.Conn) {
				remotePeer := c.RemotePeer()
				remoteAddr := c.RemoteMultiaddr()
				log.Printf("[P2P Mesh] Connected to peer: %s (%s)", remotePeer, remoteAddr)
				peerStore.AddOrUpdate(remotePeer, []multiaddr.Multiaddr{remoteAddr}, true)
			},
			DisconnectedF: func(n network.Network, c network.Conn) {
				remotePeer := c.RemotePeer()
				// Only mark disconnected if there are no remaining active connections to this peer
				if n.Connectedness(remotePeer) != network.Connected {
					log.Printf("[P2P Mesh] Disconnected from peer: %s", remotePeer)
					peerStore.SetConnected(remotePeer, false)
				}
			},
		})
	}

	return h, nil
}
