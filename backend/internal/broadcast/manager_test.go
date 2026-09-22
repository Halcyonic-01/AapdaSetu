package broadcast

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Halcyonic-01/AapdaSetu/backend/config"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/p2p"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/peers"
)

func TestMessageStore(t *testing.T) {
	store := NewMessageStore(5)

	m1 := NewMessage("peer1", "Node1", "Message 1", false, "info")
	m2 := NewMessage("peer2", "Node2", "Message 2", true, "critical")

	store.Add(m1)
	store.Add(m2)
	// Duplicate ID shouldn't add twice
	store.Add(m1)

	if len(store.List()) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(store.List()))
	}

	recent := store.ListRecent(1)
	if len(recent) != 1 || recent[0].ID != m2.ID {
		t.Fatalf("expected recent message to be m2")
	}
}

func TestBroadcastMultiNodeChatAndAlert(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cfg1 := &config.Config{
		P2PPort:    9881,
		HTTPPort:   8881,
		Rendezvous: "aapdasetu-broadcast-test",
		ChatTopic:  "test-chat-topic",
		AlertTopic: "test-alert-topic",
		NodeName:   "Alpha",
	}
	cfg2 := &config.Config{
		P2PPort:    9882,
		HTTPPort:   8882,
		Rendezvous: "aapdasetu-broadcast-test",
		ChatTopic:  "test-chat-topic",
		AlertTopic: "test-alert-topic",
		NodeName:   "Bravo",
	}

	store1 := peers.NewStore()
	store2 := peers.NewStore()

	node1, err := p2p.NewNode(ctx, cfg1, store1)
	if err != nil {
		t.Fatalf("failed to create node1: %v", err)
	}
	defer node1.Close()

	node2, err := p2p.NewNode(ctx, cfg2, store2)
	if err != nil {
		t.Fatalf("failed to create node2: %v", err)
	}
	defer node2.Close()

	// Connect nodes
	addr1 := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", cfg1.P2PPort, node1.ID())
	if err := node2.ConnectDirect(ctx, addr1); err != nil {
		t.Fatalf("failed to connect node2 to node1: %v", err)
	}

	msgStore1 := NewMessageStore(100)
	msgStore2 := NewMessageStore(100)

	mgr1, err := NewManager(ctx, node1, cfg1, msgStore1)
	if err != nil {
		t.Fatalf("failed to create mgr1: %v", err)
	}
	defer mgr1.Close()

	mgr2, err := NewManager(ctx, node2, cfg2, msgStore2)
	if err != nil {
		t.Fatalf("failed to create mgr2: %v", err)
	}
	defer mgr2.Close()

	// Wait for GossipSub heartbeat
	time.Sleep(1 * time.Second)

	// Node 1 sends a chat message
	chatPayload := "Need status report from Sector B"
	if _, err := mgr1.SendChat(ctx, chatPayload); err != nil {
		t.Fatalf("failed to send chat: %v", err)
	}

	// Verify Node 2 received it
	var receivedChat *ChatMessage
	for i := 0; i < 20; i++ {
		list := msgStore2.List()
		for _, m := range list {
			if m.Body == chatPayload {
				receivedChat = m
				break
			}
		}
		if receivedChat != nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if receivedChat == nil {
		t.Fatalf("node 2 failed to receive chat message from node 1")
	}

	// Node 2 sends an Emergency Broadcast
	alertPayload := "FLASH FLOOD WARNING - MOVE TO HIGH GROUND"
	if _, err := mgr2.SendBroadcast(ctx, alertPayload, "critical"); err != nil {
		t.Fatalf("failed to send emergency broadcast: %v", err)
	}

	// Verify Node 1 received it
	var receivedAlert *ChatMessage
	for i := 0; i < 20; i++ {
		list := msgStore1.List()
		for _, m := range list {
			if m.Emergency && m.Body == alertPayload {
				receivedAlert = m
				break
			}
		}
		if receivedAlert != nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if receivedAlert == nil {
		t.Fatalf("node 1 failed to receive emergency broadcast from node 2")
	}
}
