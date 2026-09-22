package peers

import (
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func generateTestPeer(t *testing.T) (peer.ID, []multiaddr.Multiaddr) {
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	id, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		t.Fatalf("failed to derive peer ID: %v", err)
	}
	ma, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9001")
	if err != nil {
		t.Fatalf("failed to create multiaddr: %v", err)
	}
	return id, []multiaddr.Multiaddr{ma}
}

func TestPeerStoreAddAndList(t *testing.T) {
	store := NewStore()

	id1, addrs1 := generateTestPeer(t)
	id2, addrs2 := generateTestPeer(t)

	store.AddOrUpdate(id1, addrs1, true)
	store.AddOrUpdate(id2, addrs2, false)

	if store.Count() != 1 {
		t.Errorf("expected 1 connected peer, got %d", store.Count())
	}

	p1, found := store.Get(id1)
	if !found {
		t.Fatalf("expected to find peer %s", id1)
	}
	if !p1.Connected {
		t.Errorf("expected peer 1 to be connected")
	}

	all := store.List()
	if len(all) != 2 {
		t.Errorf("expected 2 total peers, got %d", len(all))
	}

	store.SetConnected(id2, true)
	if store.Count() != 2 {
		t.Errorf("expected 2 connected peers after update, got %d", store.Count())
	}

	store.SetConnected(id1, false)
	if store.Count() != 1 {
		t.Errorf("expected 1 connected peer after disconnect, got %d", store.Count())
	}
}
